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
	"slices"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// removeAssociationsForMembership revokes one member's access to one
// allocation, for when their membership is deactivated or deleted. Their
// access to other allocations, and everyone else's access to this one, is
// untouched.
func (a *AssociationSubscriber) removeAssociationsForMembership(ctx context.Context, membership models.ComputeAllocationMembership) error {
	allocation, err := a.coreService.GetComputeAllocation(ctx, membership.ComputeAllocationID)
	if errors.Is(err, service.ErrNotFound) {
		return leftToCleanup(membership, "allocation")
	}
	if err != nil {
		return fmt.Errorf("get compute allocation: %w", err)
	}
	cluster, err := a.coreService.GetComputeCluster(ctx, allocation.ComputeClusterID)
	if err != nil {
		return fmt.Errorf("get compute cluster: %w", err)
	}
	csu, err := a.coreService.GetComputeClusterUserByPair(ctx, cluster.ID, membership.UserID)
	if errors.Is(err, service.ErrNotFound) {
		return leftToCleanup(membership, "cluster account")
	}
	if err != nil {
		return fmt.Errorf("get compute cluster user: %w", err)
	}

	filter := client.AssocFilter{
		Cluster: cluster.Name,
		Account: slurmAccount(allocation.Name),
		User:    csu.LocalUsername,
	}
	if err := a.slurmClient.DeleteAssociation(filter); err != nil {
		return fmt.Errorf("delete associations for %s on %s: %w", csu.LocalUsername, allocation.Name, err)
	}
	slog.InfoContext(ctx, "Removed associations for membership",
		"user", csu.LocalUsername, "account", allocation.Name, "cluster", cluster.Name)
	return nil
}

// leftToCleanup ends a membership revoke whose allocation or cluster account is
// already gone. Retrying cannot bring it back; the allocation delete or the
// reconciler removes whatever associations remain.
func leftToCleanup(membership models.ComputeAllocationMembership, missing string) error {
	slog.Info("Nothing to revoke for membership, its "+missing+" is gone",
		"membership_id", membership.ID, "user_id", membership.UserID)
	return nil
}

// removeAssociationsForAllocation drops every member's access to an allocation
// at once, for when the allocation itself stops being usable.
func (a *AssociationSubscriber) removeAssociationsForAllocation(ctx context.Context, allocation models.ComputeAllocation) error {
	cluster, err := a.coreService.GetComputeCluster(ctx, allocation.ComputeClusterID)
	if err != nil {
		return fmt.Errorf("get compute cluster: %w", err)
	}
	filter := client.AssocFilter{
		Cluster: cluster.Name,
		Account: slurmAccount(allocation.Name),
	}
	if err := a.slurmClient.DeleteAssociation(filter); err != nil {
		return fmt.Errorf("delete associations for allocation %s: %w", allocation.Name, err)
	}
	slog.InfoContext(ctx, "Removed associations for allocation",
		"account", allocation.Name, "cluster", cluster.Name)
	return nil
}

// removeUserFromCluster deletes a deleted account's Slurm user, or only this cluster's associations if the user is on others.
func (a *AssociationSubscriber) removeUserFromCluster(ctx context.Context, csu models.ComputeClusterUser) error {
	cluster, err := a.coreService.GetComputeCluster(ctx, csu.ComputeClusterID)
	if err != nil {
		return fmt.Errorf("get compute cluster: %w", err)
	}
	user := csu.LocalUsername
	assocs, err := a.slurmClient.ListAssociations(client.AssocFilter{User: user})
	if err != nil || len(assocs) == 0 {
		return err
	}
	if !slices.ContainsFunc(assocs, func(x client.Association) bool { return x.Cluster != cluster.Name }) {
		return a.slurmClient.DeleteUser(user)
	}
	// Slurm refuses the default account while others remain, so it goes last.
	for _, x := range assocs {
		if x.Cluster == cluster.Name && x.Account != noAllocation {
			if err := a.slurmClient.DeleteAssociation(client.AssocFilter{Cluster: cluster.Name, Account: x.Account, User: user}); err != nil {
				return err
			}
		}
	}
	return a.slurmClient.DeleteAssociation(client.AssocFilter{Cluster: cluster.Name, Account: noAllocation, User: user})
}

// activeAllocation reports whether an allocation should still grant access.
func activeAllocation(allocation models.ComputeAllocation) bool {
	return allocation.Status == models.ACTIVE
}
