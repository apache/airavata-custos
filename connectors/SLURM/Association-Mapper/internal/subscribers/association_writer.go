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

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
)

// errNotProvisioned means the account does not exist on the cluster yet, so
// writing the association now would make slurmctld cache a failed uid lookup
// and reject the user's jobs. The reconciler retries these.
var errNotProvisioned = errors.New("cluster account not provisioned yet")

// upsertAssociationsForMembership writes the Slurm association for every
// resource on the membership's allocation.
//
// Every caller goes through here on purpose. Slurm keys an association by
// (cluster, account, user, partition) and its upsert is last-write-wins, so a
// caller that built the record without limits would wipe limits another caller
// had set. Folding the per-member overrides in here keeps all writers
// producing the same record for the same key.
func (a *AssociationSubscriber) upsertAssociationsForMembership(ctx context.Context, membership models.ComputeAllocationMembership) error {
	allocation, err := a.coreService.GetComputeAllocation(ctx, membership.ComputeAllocationID)
	if err != nil {
		return fmt.Errorf("get compute allocation: %w", err)
	}
	cluster, err := a.coreService.GetComputeCluster(ctx, allocation.ComputeClusterID)
	if err != nil {
		return fmt.Errorf("get compute cluster: %w", err)
	}
	csu, err := a.coreService.GetComputeClusterUserByPair(ctx, cluster.ID, membership.UserID)
	if err != nil {
		return fmt.Errorf("get compute cluster user: %w", err)
	}
	if csu.ProvisionedAt == nil {
		return errNotProvisioned
	}

	resources, err := a.coreService.ListResourcesForAllocation(ctx, allocation.ID)
	if err != nil {
		return fmt.Errorf("list resources for allocation: %w", err)
	}
	if len(resources) == 0 {
		// Nothing to map onto a partition. Skipping beats guessing a
		// partition name the cluster may not have.
		slog.Warn("Allocation has no resources, no association written",
			"allocation_id", allocation.ID, "user_id", membership.UserID)
		return nil
	}

	overrides, err := a.coreService.ListOverridesForMembership(ctx, membership.ID)
	if err != nil {
		return fmt.Errorf("list overrides for membership: %w", err)
	}
	overrideByResource := make(map[string]models.ComputeAllocationMembershipResourceOverride, len(overrides))
	for _, o := range overrides {
		overrideByResource[o.ComputeAllocationResourceID] = o
	}

	for _, resource := range resources {
		association := client.Association{
			Account:   allocation.Name,
			Cluster:   cluster.Name,
			User:      csu.LocalUsername,
			Partition: resource.Name,
			QoS:       []string{"normal"},
			Limits:    limitsFor(resource, overrideByResource[resource.ID]),
		}
		if err := a.slurmClient.UpsertAssociation(association); err != nil {
			return fmt.Errorf("upsert association for partition %s: %w", resource.Name, err)
		}
		slog.Info("Upserted association", "association", association)
	}
	return nil
}

// limitsFor turns a per-member override into association limits. A zero
// override means the member draws on the allocation-wide limits, so the
// association carries none of its own.
func limitsFor(resource models.ComputeAllocationResource, override models.ComputeAllocationMembershipResourceOverride) client.AssocLimits {
	var limits client.AssocLimits
	if override.OverrideResourceAmount > 0 {
		limits.GrpTRES = []client.TRES{{
			Type:  resource.ResourceType,
			Count: override.OverrideResourceAmount,
		}}
	}
	if override.OverrideResourceTime > 0 {
		limits.GrpTRESMins = []client.TRES{{
			Type:  resource.ResourceType,
			Count: override.OverrideResourceTime,
		}}
	}
	return limits
}
