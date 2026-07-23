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
	"strings"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
)

// tresFor builds a SLURM TRES from a stored resource type and count. A GRES
// resource is stored joined as "gres/gpu" but SLURM wants it split into type
// and name; a plain resource like "cpu" has no name.
func tresFor(resourceType string, count int64) client.TRES {
	if typ, name, ok := strings.Cut(resourceType, "/"); ok {
		return client.TRES{Type: typ, Name: name, Count: count}
	}
	return client.TRES{Type: resourceType, Count: count}
}

// errNotProvisioned means the account does not exist on the cluster yet, so
// writing the association now would make slurmctld cache a failed uid lookup
// and reject the user's jobs. The reconciler retries these.
var errNotProvisioned = errors.New("cluster account not provisioned yet")

// assocKey identifies an association the way Slurm does. Cluster is not part
// of it because callers scope their lookups to one cluster already.
type assocKey struct {
	account   string
	user      string
	partition string
}

func keyOf(a client.Association) assocKey {
	return assocKey{account: a.Account, user: a.User, partition: a.Partition}
}

// sameLimits reports whether two associations already carry the same limits.
// Only the limits are compared: the key fields match by construction.
func sameLimits(want, got client.Association) bool {
	return sameTRES(want.Limits.GrpTRES, got.Limits.GrpTRES) &&
		sameTRES(want.Limits.GrpTRESMins, got.Limits.GrpTRESMins)
}

func sameTRES(want, got []client.TRES) bool {
	if len(want) != len(got) {
		return false
	}
	counts := make(map[string]int64, len(want))
	for _, t := range want {
		counts[t.Type] = t.Count
	}
	for _, t := range got {
		if c, ok := counts[t.Type]; !ok || c != t.Count {
			return false
		}
	}
	return true
}

// upsertAssociationsForMembership writes the association for every resource on
// the membership's allocation, without checking what the cluster already has.
// The event paths use this: an event fires because something changed.
func (a *AssociationSubscriber) upsertAssociationsForMembership(ctx context.Context, membership models.ComputeAllocationMembership) error {
	return a.syncAssociationsForMembership(ctx, membership, nil)
}

// syncAssociationsForMembership writes the associations this membership should
// have, skipping any that already match existing, the state already on the
// cluster. Pass nil to skip that comparison and write every record.
//
// Every caller goes through here on purpose. Slurm keys an association by
// (cluster, account, user, partition) and its upsert is last-write-wins, so a
// caller that built the record without limits would wipe limits another caller
// had set. Folding the per-member overrides in here keeps all writers
// producing the same record for the same key.
func (a *AssociationSubscriber) syncAssociationsForMembership(ctx context.Context, membership models.ComputeAllocationMembership, existing map[assocKey]client.Association) error {
	desired, err := a.desiredAssociationsForMembership(ctx, membership)
	if err != nil {
		return err
	}
	for _, association := range desired {
		if got, ok := existing[keyOf(association)]; ok && sameLimits(association, got) {
			continue
		}
		if err := a.slurmClient.UpsertAssociation(association); err != nil {
			return fmt.Errorf("upsert association for partition %s: %w", association.Partition, err)
		}
		slog.Info("Upserted association", "association", association)
	}
	return nil
}

// desiredAssociationsForMembership builds the associations this membership
// should have on the cluster, without writing anything. An inactive allocation
// grants nothing, so it yields an empty set rather than an error.
func (a *AssociationSubscriber) desiredAssociationsForMembership(ctx context.Context, membership models.ComputeAllocationMembership) ([]client.Association, error) {
	allocation, err := a.coreService.GetComputeAllocation(ctx, membership.ComputeAllocationID)
	if err != nil {
		return nil, fmt.Errorf("get compute allocation: %w", err)
	}
	if !activeAllocation(*allocation) {
		return nil, nil
	}
	cluster, err := a.coreService.GetComputeCluster(ctx, allocation.ComputeClusterID)
	if err != nil {
		return nil, fmt.Errorf("get compute cluster: %w", err)
	}
	csu, err := a.coreService.GetComputeClusterUserByPair(ctx, cluster.ID, membership.UserID)
	if err != nil {
		return nil, fmt.Errorf("get compute cluster user: %w", err)
	}
	if csu.ProvisionedAt == nil {
		return nil, errNotProvisioned
	}

	resources, err := a.coreService.ListResourcesForAllocation(ctx, allocation.ID)
	if err != nil {
		return nil, fmt.Errorf("list resources for allocation: %w", err)
	}
	if len(resources) == 0 {
		// Nothing to map onto a partition. Skipping beats guessing a
		// partition name the cluster may not have.
		slog.Warn("Allocation has no resources, no association written",
			"allocation_id", allocation.ID, "user_id", membership.UserID)
		return nil, nil
	}

	overrides, err := a.coreService.ListOverridesForMembership(ctx, membership.ID)
	if err != nil {
		return nil, fmt.Errorf("list overrides for membership: %w", err)
	}
	overrideByResource := make(map[string]models.ComputeAllocationMembershipResourceOverride, len(overrides))
	for _, o := range overrides {
		overrideByResource[o.ComputeAllocationResourceID] = o
	}

	out := make([]client.Association, 0, len(resources))
	for _, resource := range resources {
		out = append(out, client.Association{
			Account:   allocation.Name,
			Cluster:   cluster.Name,
			User:      csu.LocalUsername,
			Partition: resource.Name,
			QoS:       []string{"normal"},
			Limits:    limitsFor(resource, overrideByResource[resource.ID]),
		})
	}
	return out, nil
}

// limitsFor turns a per-member override into association limits. A zero
// override means the member draws on the allocation-wide limits, so the
// association carries none of its own.
func limitsFor(resource models.ComputeAllocationResource, override models.ComputeAllocationMembershipResourceOverride) client.AssocLimits {
	var limits client.AssocLimits
	if override.OverrideResourceAmount > 0 {
		limits.GrpTRES = []client.TRES{tresFor(resource.ResourceType, override.OverrideResourceAmount)}
	}
	if override.OverrideResourceTime > 0 {
		limits.GrpTRESMins = []client.TRES{tresFor(resource.ResourceType, override.OverrideResourceTime)}
	}
	return limits
}
