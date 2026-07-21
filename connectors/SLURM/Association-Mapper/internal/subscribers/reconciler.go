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
	"log/slog"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
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

// reconcile compares what the cluster has against what the allocations say it
// should have, and applies the difference in both directions: it writes what
// is missing or drifted, and removes access that should no longer exist.
//
// Removal matters more than creation here. A lost creation event only delays
// access, but a lost deactivation leaves a member able to submit indefinitely
// with nothing reporting it.
func (a *AssociationSubscriber) reconcile(ctx context.Context) {
	sweepCtx, cancel := context.WithTimeout(ctx, reconcileSweepTimeout)
	defer cancel()

	clusters, err := a.coreService.ListComputeClusters(sweepCtx)
	if err != nil {
		slog.Error("Association reconciler: failed to list compute clusters", "error", err)
		return
	}

	for _, cluster := range clusters {
		a.reconcileCluster(sweepCtx, cluster)
	}
}

func (a *AssociationSubscriber) reconcileCluster(ctx context.Context, cluster models.ComputeCluster) {
	clusterUsers, err := a.coreService.ListComputeClusterUsersByCluster(ctx, cluster.ID)
	if err != nil {
		slog.Error("Association reconciler: failed to list cluster users", "cluster_id", cluster.ID, "error", err)
		return
	}

	// One read per cluster. On failure fall back to writing every
	// association, which costs extra calls but still converges.
	existing, err := a.existingAssociations(ctx, cluster.Name)
	if err != nil {
		slog.Warn("Association reconciler: could not read current associations, writing unconditionally",
			"cluster", cluster.Name, "error", err)
		existing = nil
	}

	desired := make(map[assocKey]struct{})
	written := 0
	for _, csu := range clusterUsers {
		if csu.ProvisionedAt == nil {
			continue
		}
		memberships, err := a.coreService.ListAllocationsForUser(ctx, csu.UserID)
		if err != nil {
			slog.Error("Association reconciler: failed to list memberships", "user_id", csu.UserID, "error", err)
			// Without this user's memberships the desired set is incomplete,
			// so pruning could revoke access that is actually valid.
			return
		}
		for _, membership := range memberships {
			if membership.MembershipStatus != models.ACTIVE {
				continue
			}
			records, err := a.desiredAssociationsForMembership(ctx, membership)
			if err != nil {
				slog.Error("Association reconciler: failed to resolve desired associations",
					"membership_id", membership.ID, "error", err)
				return
			}
			for _, record := range records {
				desired[keyOf(record)] = struct{}{}
			}
			// The grace only holds back the write. The association is still
			// desired, so a member provisioned moments ago is never pruned.
			if !a.readyForAssociation(csu) {
				continue
			}
			for _, record := range records {
				if got, ok := existing[keyOf(record)]; ok && sameLimits(record, got) {
					continue
				}
				if err := a.slurmClient.UpsertAssociation(record); err != nil {
					slog.Error("Association reconciler: failed to upsert association",
						"membership_id", membership.ID, "error", err)
					continue
				}
				written++
			}
		}
	}

	removed := a.pruneStaleAssociations(ctx, cluster, existing, desired)
	slog.Debug("Association reconciler pass complete",
		"cluster", cluster.Name, "written", written, "removed", removed)
}

// pruneStaleAssociations removes member associations the allocations no longer
// call for. Two things are never touched: associations on accounts Custos does
// not manage, and the account-level records that carry an allocation's own
// limits rather than a member's access.
func (a *AssociationSubscriber) pruneStaleAssociations(ctx context.Context, cluster models.ComputeCluster, existing map[assocKey]client.Association, desired map[assocKey]struct{}) int {
	if len(existing) == 0 {
		return 0
	}
	// An empty desired set almost always means a failed lookup rather than
	// nobody being entitled to anything, and acting on it would revoke the
	// whole cluster. Refuse.
	if len(desired) == 0 {
		slog.Warn("Association reconciler: nothing is desired on this cluster, skipping removal",
			"cluster", cluster.Name, "existing", len(existing))
		return 0
	}

	managed, err := a.managedAccounts(ctx, cluster.ID)
	if err != nil {
		slog.Error("Association reconciler: could not list allocations, skipping removal",
			"cluster", cluster.Name, "error", err)
		return 0
	}

	removed := 0
	for key, assoc := range existing {
		if assoc.User == "" {
			continue
		}
		if _, ok := managed[assoc.Account]; !ok {
			continue
		}
		if _, ok := desired[key]; ok {
			continue
		}
		filter := client.AssocFilter{
			Cluster:   cluster.Name,
			Account:   assoc.Account,
			User:      assoc.User,
			Partition: assoc.Partition,
		}
		if err := a.slurmClient.DeleteAssociation(filter); err != nil {
			slog.Error("Association reconciler: failed to remove stale association",
				"account", assoc.Account, "user", assoc.User, "partition", assoc.Partition, "error", err)
			continue
		}
		slog.Info("Removed stale association",
			"account", assoc.Account, "user", assoc.User, "partition", assoc.Partition)
		removed++
	}
	return removed
}

// managedAccounts is the set of Slurm account names Custos owns on a cluster.
func (a *AssociationSubscriber) managedAccounts(ctx context.Context, clusterID string) (map[string]struct{}, error) {
	allocations, err := a.coreService.ListComputeAllocationsByCluster(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	names := make(map[string]struct{}, len(allocations))
	for _, allocation := range allocations {
		names[allocation.Name] = struct{}{}
	}
	return names, nil
}

// existingAssociations indexes the cluster's current associations by key.
func (a *AssociationSubscriber) existingAssociations(ctx context.Context, clusterName string) (map[assocKey]client.Association, error) {
	assocs, err := a.slurmClient.ListAssociations(client.AssocFilter{Cluster: clusterName})
	if err != nil {
		return nil, err
	}
	byKey := make(map[assocKey]client.Association, len(assocs))
	for _, assoc := range assocs {
		byKey[keyOf(assoc)] = assoc
	}
	return byKey, nil
}

// readyForAssociation reports whether the account has been provisioned long
// enough for the cluster to resolve it.
func (a *AssociationSubscriber) readyForAssociation(csu models.ComputeClusterUser) bool {
	if csu.ProvisionedAt == nil {
		return false
	}
	grace := a.provisionGrace
	if grace <= 0 {
		grace = DefaultProvisionGrace
	}
	return time.Since(*csu.ProvisionedAt) >= grace
}
