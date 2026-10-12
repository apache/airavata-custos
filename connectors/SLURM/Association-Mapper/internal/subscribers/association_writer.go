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
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

// tresFor builds a SLURM TRES from a stored resource type and count. A GRES
// resource is stored joined as "gres/gpu" but SLURM wants it split into type
// and name; a plain resource like "cpu" has no name.
func tresFor(resourceType string, count int64) client.TRES {
	if typ, name, ok := strings.Cut(resourceType, "/"); ok {
		return client.TRES{Type: typ, Name: name, Count: float64(count)}
	}
	return client.TRES{Type: resourceType, Count: float64(count)}
}

// errNotProvisioned means the account does not exist on the cluster yet, so
// writing the association now would make slurmctld cache a failed uid lookup
// and reject the user's jobs. The reconciler retries these.
var errNotProvisioned = errors.New("cluster account not provisioned yet")

// errEmptyResourceType means the resource row has an empty resource_type. A
// retry cannot fix the row, so the delivery is not retried.
var errEmptyResourceType = fmt.Errorf("%w: resource type is empty", events.ErrPermanent)

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
	counts := make(map[string]float64, len(want))
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

// The event handlers use this. An event means something changed, so the
// record is written without checking the cluster first.
func (a *AssociationSubscriber) upsertAssociationsForMembership(ctx context.Context, membership models.ComputeAllocationMembership) error {
	return a.syncAssociationsForMembership(ctx, membership, nil)
}

// Writes the association of one member of one allocation. Every caller comes
// through here. It builds the full record, overrides included, and compares
// it with the existing one before writing, since Slurm overwrites the whole
// record on each write. With no existing map it writes every time.
func (a *AssociationSubscriber) syncAssociationsForMembership(ctx context.Context, membership models.ComputeAllocationMembership, existing map[assocKey]client.Association) error {
	association, err := a.desiredAssociationForMembership(ctx, membership)
	if err != nil || association == nil {
		return err
	}

	// Nothing changed on the cluster, so skip the write.
	if got, ok := existing[keyOf(*association)]; ok && sameLimits(*association, got) {
		return nil
	}
	if err := a.slurmClient.UpsertAssociation(*association); err != nil {
		return fmt.Errorf("upsert association: %w", err)
	}
	slog.InfoContext(ctx, "Upserted association", "association", *association)
	return nil
}

// Builds the member's association without writing it. It has no partition,
// so the cluster's own partition settings decide where the member can run.
// Returns nil when the allocation is not active.
func (a *AssociationSubscriber) desiredAssociationForMembership(ctx context.Context, membership models.ComputeAllocationMembership) (*client.Association, error) {
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

	overrides, err := a.coreService.ListOverridesForMembership(ctx, membership.ID)
	if err != nil {
		return nil, fmt.Errorf("list overrides for membership: %w", err)
	}
	limits, err := a.limitsFor(ctx, overrides)
	if err != nil {
		return nil, err
	}
	return &client.Association{
		Account: slurmAccount(allocation.Name),
		Cluster: cluster.Name,
		User:    csu.LocalUsername,
		QoS:     []string{"normal"},
		Limits:  limits,
	}, nil
}

// Builds the member's limits from their overrides. The association covers
// every partition, so overrides on two GPU resources become one GPU limit,
// their sum. CPU and GPU stay separate.
func (a *AssociationSubscriber) limitsFor(ctx context.Context, overrides []models.ComputeAllocationMembershipResourceOverride) (client.AssocLimits, error) {
	var limits client.AssocLimits
	for _, o := range overrides {
		resource, err := a.coreService.GetComputeAllocationResource(ctx, o.ComputeAllocationResourceID)
		if err != nil {
			return limits, fmt.Errorf("get compute allocation resource: %w", err)
		}
		if o.OverrideResourceAmount > 0 {
			limits.GrpTRES = addTRES(limits.GrpTRES, tresFor(resource.ResourceType, o.OverrideResourceAmount))
		}
		if o.OverrideResourceTime > 0 {
			limits.GrpTRESMins = addTRES(limits.GrpTRESMins, tresFor(resource.ResourceType, o.OverrideResourceTime))
		}
	}
	return limits, nil
}

func addTRES(list []client.TRES, t client.TRES) []client.TRES {
	for i := range list {
		if list[i].Type == t.Type && list[i].Name == t.Name {
			list[i].Count += t.Count
			return list
		}
	}
	return append(list, t)
}
