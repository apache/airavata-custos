// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package subscribers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"sync"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/drift"
	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/connectors/health"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// DefaultReconcileInterval is how often the sweep runs when the connector
// config does not set association_reconcile_interval.
const DefaultReconcileInterval = 5 * time.Minute

// DefaultProvisionGrace holds a just-provisioned account back for one sweep.
// The stamp means the registry write finished, but the account still has to
// reach the directory and the cluster's name cache, which Custos cannot
// observe. How long that takes is site-specific, hence configurable.
const DefaultProvisionGrace = 30 * time.Second

// StartReconciler runs the sweep until ctx is cancelled, beginning with an
// immediate pass so a restart backfills what it missed.
//
// The sweep, not the membership event, is what guarantees associations exist.
// The bus is in-process with no persistence or replay, so a restart or a
// failed handler loses an event for good and the member cannot submit jobs,
// silently. Re-declaring every pass makes a lost event cost latency instead.
//
// TODO: when the bus is durable, publish a provisioned event and act on it for
// near-instant creation, keeping this sweep as the backstop against drift.
func (a *AssociationSubscriber) StartReconciler(ctx context.Context) {
	interval := a.reconcileInterval
	if interval <= 0 {
		interval = DefaultReconcileInterval
	}
	slog.Info("Starting SLURM association reconciler", "interval", interval)

	a.reconcile(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping SLURM association reconciler", "reason", ctx.Err())
			return
		case <-ticker.C:
			a.reconcile(ctx)
		}
	}
}

// reconcile syncs every cluster: Custos pushes what it owns and what it did not
// cause is left on the cluster as a conflict.
//
// Removal matters more than creation here. A lost creation event only delays
// access, but a lost deactivation leaves a member able to submit indefinitely
// with nothing reporting it.
func (a *AssociationSubscriber) reconcile(ctx context.Context) {
	ctx, span := tracing.Start(ctx, "slurm.reconcile")
	defer span.End()
	sweepCtx, cancel := context.WithTimeout(ctx, reconcileSweepTimeout)
	defer cancel()

	clusters, err := a.coreService.ListComputeClusters(sweepCtx)
	if err != nil {
		slog.ErrorContext(ctx, "Association reconciler: failed to list compute clusters", "error", err)
		return
	}

	for _, cluster := range clusters {
		conflicts, err := a.syncCluster(sweepCtx, cluster)
		if err != nil {
			slog.ErrorContext(ctx, "Association reconciler: sync failed", "cluster", cluster.Name, "error", err)
			continue
		}
		slog.DebugContext(ctx, "Association reconciler pass complete", "cluster", cluster.Name, "conflicts", len(conflicts))
	}
}

// syncCluster pushes what Custos owns and returns the conflicts it left alone.
// It writes nothing when the cluster cannot be read.
func (a *AssociationSubscriber) syncCluster(ctx context.Context, cluster models.ComputeCluster) ([]drift.Conflict, error) {
	// Events and the sweep never overlap on a cluster: two passes would push the same diff twice.
	lock, _ := a.clusterLocks.LoadOrStore(cluster.ID, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()

	snap, err := a.Snapshot(ctx, cluster)
	if err != nil {
		return nil, err
	}
	return snap.Conflicts, errors.Join(a.apply(ctx, cluster, snap)...)
}

// apply writes the push to the cluster and clears the marks it satisfied. A
// failed write is logged and left for the next sweep; the rest still goes out.
func (a *AssociationSubscriber) apply(ctx context.Context, cluster models.ComputeCluster, snap Snapshot) []error {
	var errs []error
	// A mark clears only once every write for its target succeeded; otherwise the next pass retries.
	failedAccounts, failedMembers := map[string]bool{}, map[string]bool{}
	fail := func(what string, err error, attrs ...any) {
		slog.ErrorContext(ctx, "Association reconciler: "+what, append(attrs, "cluster", cluster.Name, "error", err)...)
		errs = append(errs, fmt.Errorf("%s: %w", what, err))
	}
	failedQoS := false
	failWrite := func(what string, err error, account, user string) {
		failedQoS = true
		failedAccounts[account] = true
		failedMembers[account+"/"+user] = true
		fail(what, err, "account", account, "user", user)
	}
	// Definitions go first: an account's list can name a QoS created in this pass.
	for _, q := range snap.Push.QoS {
		if err := a.slurmClient.PostQoS(q); err != nil {
			fail("failed to post qos", err, "qos", q.Name)
			failedQoS = true
		}
	}

	// The default account is set before the association so no allocation ever becomes it,
	// and set again whenever SLURM reports another default. A default on a system account,
	// or one SLURM does not report, is left alone.
	defaultIsNone, liveAccounts := map[string]bool{}, map[string]bool{}
	for _, assoc := range snap.Have.Associations {
		liveAccounts[assoc.Account] = true
		switch {
		case assoc.User == "":
		case assoc.IsDefault == nil:
			defaultIsNone[assoc.User] = true
		case *assoc.IsDefault && (assoc.Account == noAllocation || snap.Want.SystemAccounts[assoc.Account]):
			defaultIsNone[assoc.User] = true
		}
	}
	for _, name := range slices.Sorted(maps.Keys(snap.Want.Accounts)) {
		for _, user := range slices.Sorted(maps.Keys(snap.Want.Accounts[name].Members)) {
			if m := snap.Want.Accounts[name].Members[user]; !m.Active || m.Held || defaultIsNone[user] {
				continue
			}
			defaultIsNone[user] = true
			if err := a.provisionNoAllocationDefault(cluster.Name, user); err != nil {
				failWrite("failed to set default account", err, name, user)
			}
		}
	}

	// Diff queues a missing account's own association before its members', so the account is created first.
	written, uncreated := 0, map[string]bool{}
	for _, assoc := range snap.Push.Associations {
		if snap.Want.Accounts[assoc.Account].Members[assoc.User].Held {
			continue
		}
		if assoc.User == "" && !liveAccounts[assoc.Account] {
			if err := a.createAccount(ctx, cluster, snap.Want.Accounts[assoc.Account].AllocationID); err != nil {
				uncreated[assoc.Account] = true
				fail("failed to create account", err, "account", assoc.Account)
			}
		}
		if uncreated[assoc.Account] {
			failedAccounts[assoc.Account], failedMembers[assoc.Account+"/"+assoc.User] = true, true
			continue
		}
		assoc = completeAssociation(cluster, assoc)
		if err := a.slurmClient.UpsertAssociation(assoc); err != nil {
			failWrite("failed to upsert association", err, assoc.Account, assoc.User)
			continue
		}
		written++
	}
	for _, filter := range snap.Push.DeleteAssociations {
		filter.Cluster = cluster.Name
		if err := a.slurmClient.DeleteAssociation(filter); err != nil {
			failWrite("failed to remove association", err, filter.Account, filter.User)
			continue
		}
		slog.InfoContext(ctx, "Removed association", "account", filter.Account, "user", filter.User)
	}

	accountOf := map[string]string{}
	memberOf := map[string]string{}
	for account, desired := range snap.Want.Accounts {
		accountOf[desired.AllocationID] = account
		for user, member := range desired.Members {
			memberOf[member.MembershipID] = account + "/" + user
		}
	}
	clear := func(kind service.SlurmPendingKind, marks []drift.Mark, failed func(id string) bool) {
		for _, mark := range marks {
			if failed(mark.ID) {
				continue
			}
			if err := a.coreService.SetSlurmPending(ctx, kind, mark.ID, &mark.Since); err != nil {
				fail("failed to clear pending mark", err, "kind", kind, "id", mark.ID)
			}
		}
	}
	clear(service.SlurmPendingAllocation, snap.Push.ClearAllocations, func(id string) bool { return failedAccounts[accountOf[id]] })
	clear(service.SlurmPendingMembership, snap.Push.ClearMemberships, func(id string) bool { return failedMembers[memberOf[id]] })
	// A definition may be what a failed account or member push depends on; keep its mark until every write lands.
	clear(service.SlurmPendingQoS, snap.Push.ClearQoS, func(string) bool { return failedQoS })
	slog.DebugContext(ctx, "Association reconciler applied push", "cluster", cluster.Name,
		"written", written, "removed", len(snap.Push.DeleteAssociations), "qos", len(snap.Push.QoS))
	return errs
}

// completeAssociation fills what Diff cannot know for an association it creates.
func completeAssociation(cluster models.ComputeCluster, assoc client.Association) client.Association {
	if assoc.Cluster == "" {
		assoc.Cluster = cluster.Name
	}
	if assoc.User == "" && assoc.ParentAccount == "" {
		assoc.ParentAccount = rootAccount
	}
	return assoc
}

// readyForAssociation reports whether the account has been provisioned long
// enough for the cluster to resolve it.
func (a *AssociationSubscriber) readyForAssociation(provisionedAt *time.Time) bool {
	if provisionedAt == nil {
		return false
	}
	grace := a.provisionGrace
	if grace <= 0 {
		grace = DefaultProvisionGrace
	}
	return time.Since(*provisionedAt) >= grace
}

// Snapshot is the cluster's desired and live state and the conflicts between them.
type Snapshot struct {
	Want      drift.Desired
	Have      drift.Live
	Push      drift.Push
	Conflicts []drift.Conflict
}

// Snapshot reads Custos and the cluster afresh and diffs them without writing either.
func (a *AssociationSubscriber) Snapshot(ctx context.Context, cluster models.ComputeCluster) (Snapshot, error) {
	want, err := a.desiredState(ctx, cluster)
	if err != nil {
		return Snapshot{}, fmt.Errorf("desired state: %w", err)
	}
	have, err := a.liveState(cluster)
	health.Record(a.name, cluster.ID, err)
	if err != nil {
		slog.ErrorContext(ctx, "Association reconciler: failed to read live state", "cluster", cluster.Name, "error", err)
		return Snapshot{}, fmt.Errorf("live state: %w", service.ErrUnavailable)
	}
	want.SystemAccounts = map[string]bool{rootAccount: true, noAllocation: true}
	for _, assoc := range have.Associations {
		if _, managed := want.Accounts[assoc.Account]; !managed {
			want.SystemAccounts[assoc.Account] = true
		}
	}
	push, conflicts := drift.Diff(want, have)
	return Snapshot{Want: want, Have: have, Push: push, Conflicts: conflicts}, nil
}

// rootAccount is the parent of every allocation account.
const rootAccount = "root"

// desiredState is what Custos wants on the cluster.
func (a *AssociationSubscriber) desiredState(ctx context.Context, cluster models.ComputeCluster) (drift.Desired, error) {
	want := drift.Desired{
		QoS:             map[string]models.ComputeClusterQoS{},
		OtherClusterQoS: map[string]bool{},
		Accounts:        map[string]drift.DesiredAccount{},
	}
	clusters, err := a.coreService.ListComputeClusters(ctx)
	if err != nil {
		return want, fmt.Errorf("list compute clusters: %w", err)
	}
	for _, served := range clusters {
		definitions, err := a.coreService.ListComputeClusterQoS(ctx, served.ID, "")
		if err != nil {
			return want, fmt.Errorf("list compute cluster qos: %w", err)
		}
		for _, q := range definitions {
			if served.ID != cluster.ID {
				want.OtherClusterQoS[q.Name] = true
				continue
			}
			want.QoS[q.Name] = q
			if q.IsDefault {
				want.DefaultQoS = q.Name
			}
		}
	}

	allocations, err := a.coreService.ListComputeAllocationsByCluster(ctx, cluster.ID)
	if err != nil {
		return want, fmt.Errorf("list compute allocations: %w", err)
	}
	for _, allocation := range allocations {
		account, err := a.desiredAccount(ctx, allocation)
		if err != nil {
			return want, fmt.Errorf("allocation %s: %w", allocation.Name, err)
		}
		want.Accounts[SlurmAccount(allocation.Name)] = account
	}
	return want, nil
}

func (a *AssociationSubscriber) desiredAccount(ctx context.Context, allocation models.ComputeAllocation) (drift.DesiredAccount, error) {
	qosNames, err := GrantedQoSNames(ctx, a.coreService, allocation.ComputeClusterID, allocation.ID)
	if err != nil {
		return drift.DesiredAccount{}, fmt.Errorf("list allocation qos: %w", err)
	}
	mappings, err := a.coreService.ListResourceMappingsForAllocation(ctx, allocation.ID)
	if err != nil {
		return drift.DesiredAccount{}, fmt.Errorf("list resource mappings: %w", err)
	}
	rows := make([]resourceLimit, len(mappings))
	for i, m := range mappings {
		rows[i] = resourceLimit{m.ComputeAllocationResourceID, m.ResourceAmount, m.ResourceTime}
	}
	limits, err := a.limitsOf(ctx, rows)
	if err != nil {
		return drift.DesiredAccount{}, err
	}
	members, err := a.coreService.ListMembersForAllocation(ctx, allocation.ID)
	if err != nil {
		return drift.DesiredAccount{}, fmt.Errorf("list members: %w", err)
	}

	active := allocation.Status == models.ACTIVE
	account := drift.DesiredAccount{
		AllocationID: allocation.ID,
		Active:       active,
		PendingSince: allocation.SlurmPendingSince,
		QoS:          qosNames,
		Limits:       limits,
		Members:      map[string]drift.DesiredMember{},
	}
	for _, member := range members {
		// A member without a compute cluster user has nothing to associate.
		if member.ProvisionedAt == nil || member.LocalUsername == "" {
			continue
		}
		desired := drift.DesiredMember{
			MembershipID: member.ID,
			Active:       active && member.MembershipStatus == models.ACTIVE,
			PendingSince: member.SlurmPendingSince,
			Held:         !a.readyForAssociation(member.ProvisionedAt),
		}
		if desired.Active {
			overrides, err := a.coreService.ListOverridesForMembership(ctx, member.ID)
			if err != nil {
				return account, fmt.Errorf("list overrides for membership: %w", err)
			}
			rows := make([]resourceLimit, len(overrides))
			for i, o := range overrides {
				rows[i] = resourceLimit{o.ComputeAllocationResourceID, o.OverrideResourceAmount, o.OverrideResourceTime}
			}
			if desired.Limits, err = a.limitsOf(ctx, rows); err != nil {
				return account, err
			}
		}
		account.Members[member.LocalUsername] = desired
	}
	return account, nil
}

// liveState reads the cluster's QoS, associations and partitions from slurmrestd.
func (a *AssociationSubscriber) liveState(cluster models.ComputeCluster) (drift.Live, error) {
	qos, err := a.slurmClient.ListQoS()
	if err != nil {
		return drift.Live{}, fmt.Errorf("list qos: %w", err)
	}
	associations, err := a.slurmClient.ListAssociations(client.AssocFilter{Cluster: cluster.Name})
	if err != nil {
		return drift.Live{}, fmt.Errorf("list associations: %w", err)
	}
	partitions, err := a.slurmClient.ListPartitions(cluster.Name)
	if err != nil {
		return drift.Live{}, fmt.Errorf("list partitions: %w", err)
	}
	return drift.Live{QoS: qos, Associations: associations, Partitions: partitions}, nil
}

// GrantedQoSNames returns the names of the QoS the allocation grants, ordered by name.
func GrantedQoSNames(ctx context.Context, core service.CoreService, clusterID, allocationID string) ([]string, error) {
	granted, err := core.ListComputeClusterQoS(ctx, clusterID, allocationID)
	names := make([]string, len(granted))
	for i, q := range granted {
		names[i] = q.Name
	}
	return names, err
}
