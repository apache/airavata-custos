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

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

func (a *AssociationSubscriber) SubscribeToComputeAllocationChange(ctx context.Context, allocation models.ComputeAllocation) error {
	return a.handleEvent(ctx, "compute_allocation_change", allocation.ID, func(ctx context.Context) error { return a.syncAccount(ctx, allocation) })
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationResourceMappingChange(ctx context.Context, mapping models.ComputeAllocationResourceMapping) error {
	return a.handleEvent(ctx, "compute_allocation_resource_mapping_change", mapping.ID, func(ctx context.Context) error {
		return a.syncAccount(ctx, models.ComputeAllocation{ID: mapping.ComputeAllocationID})
	})
}

func (a *AssociationSubscriber) syncAccount(ctx context.Context, allocation models.ComputeAllocation) error {
	current, err := a.coreService.GetComputeAllocation(ctx, allocation.ID)
	if err == nil {
		allocation = *current
	} else if !errors.Is(err, service.ErrNotFound) || allocation.Name == "" {
		return err
	}
	cluster, err := a.slurmCluster(ctx, allocation.ComputeClusterID)
	if err != nil || cluster == "" {
		return err
	}
	var wants []client.Association
	var settling error
	if current != nil && allocation.Status == models.ACTIVE {
		if wants, err = a.accountAssociations(ctx, cluster, allocation); err != nil && !errors.Is(err, errNotProvisioned) {
			return err
		}
		settling = err
	}
	if err := a.syncAssociations(ctx, client.AssocFilter{Cluster: cluster, Account: slurmAccount(allocation.Name)}, wants); err != nil {
		return err
	}
	return settling
}

// accountAssociations puts every mapping in one record, since Slurm replaces the whole record on each write.
// It returns errNotProvisioned beside the rest while a member's account settles, so the event is retried.
func (a *AssociationSubscriber) accountAssociations(ctx context.Context, cluster string, allocation models.ComputeAllocation) ([]client.Association, error) {
	resources, err := a.coreService.ListResourcesForAllocation(ctx, allocation.ID)
	if err != nil {
		return nil, err
	}
	var mappings []models.ComputeAllocationResourceMapping
	for _, r := range resources {
		m, err := a.mappings.FindByPair(ctx, allocation.ID, r.ID)
		if err != nil {
			return nil, err
		}
		if m != nil {
			mappings = append(mappings, *m)
		}
	}
	limits, err := tresLimits(ctx, a, mappings, func(m models.ComputeAllocationResourceMapping) tresGrant {
		return tresGrant{m.ComputeAllocationResourceID, m.ResourceAmount, m.ResourceTime}
	})
	if err != nil {
		return nil, err
	}
	wants := []client.Association{{Account: slurmAccount(allocation.Name), Cluster: cluster, Limits: limits}}
	members, err := a.coreService.ListMembersForAllocation(ctx, allocation.ID)
	if err != nil {
		return nil, err
	}
	var settling error
	for _, m := range members {
		if m.MembershipStatus != models.ACTIVE || m.ProvisionedAt == nil {
			continue
		}
		if !a.readyForAssociation(*m.ProvisionedAt) {
			settling = errNotProvisioned
			continue
		}
		assoc, err := a.userAssociation(ctx, cluster, allocation, m.LocalUsername, m.ID)
		if err != nil {
			return nil, err
		}
		wants = append(wants, assoc)
	}
	return wants, settling
}
