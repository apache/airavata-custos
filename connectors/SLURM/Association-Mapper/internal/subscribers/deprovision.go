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
)

// removeAssociationsForMembership revokes one member's access to one
// allocation, for when their membership is deactivated or deleted. Their
// access to other allocations, and everyone else's access to this one, is
// untouched.
//
// A member holds one association per partition on the allocation. The filter
// names the cluster, account, and user but leaves the partition empty, which
// matches every partition, so revoke removes the member's whole set.
func (a *AssociationSubscriber) removeAssociationsForMembership(ctx context.Context, membership models.ComputeAllocationMembership) error {
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

	filter := client.AssocFilter{
		Cluster: cluster.Name,
		Account: allocation.Name,
		User:    csu.LocalUsername,
	}
	if err := a.revoke(filter); err != nil {
		return fmt.Errorf("delete associations for %s on %s: %w", csu.LocalUsername, allocation.Name, err)
	}
	slog.Info("Removed associations for membership",
		"user", csu.LocalUsername, "account", allocation.Name, "cluster", cluster.Name)
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
		Account: allocation.Name,
	}
	if err := a.revoke(filter); err != nil {
		return fmt.Errorf("delete associations for allocation %s: %w", allocation.Name, err)
	}
	slog.Info("Removed associations for allocation",
		"account", allocation.Name, "cluster", cluster.Name)
	return nil
}

// revoke removes the member associations f matches, one call each, since
// slurmrestd removes at most one per call. Slurm refuses to remove a user's
// default association while they hold another on the cluster, so the default
// first moves to another account they keep, or, when they keep only other
// partitions of the same account, those go too and the reconciler adds them back.
func (a *AssociationSubscriber) revoke(f client.AssocFilter) error {
	matched, err := a.slurmClient.ListAssociations(f)
	if err != nil {
		return err
	}
	byUser := map[string][]client.Association{}
	for _, m := range matched {
		if m.User != "" {
			byUser[m.User] = append(byUser[m.User], m)
		}
	}
	for user, drop := range byUser {
		held, err := a.slurmClient.ListAssociations(client.AssocFilter{Cluster: f.Cluster, User: user})
		if err != nil {
			return err
		}
		kept := slices.DeleteFunc(held, func(h client.Association) bool {
			return slices.ContainsFunc(drop, func(d client.Association) bool { return d.Account == h.Account && d.Partition == h.Partition })
		})
		if i := slices.IndexFunc(drop, func(d client.Association) bool { return d.IsDefault != nil && *d.IsDefault }); i >= 0 {
			def := drop[i]
			drop = append(slices.Delete(drop, i, i+1), def)
			if j := slices.IndexFunc(kept, func(k client.Association) bool { return k.Account != def.Account }); j >= 0 {
				next, yes := kept[j], true
				next.ID, next.IsDefault = 0, &yes
				if err := a.slurmClient.UpsertAssociation(next); err != nil {
					return fmt.Errorf("move default of %s to %s: %w", user, next.Account, err)
				}
			} else {
				drop = append(kept, drop...)
			}
		}
		for _, d := range drop {
			if err := a.slurmClient.DeleteAssociation(client.AssocFilter{Cluster: d.Cluster, Account: d.Account, User: d.User, Partition: d.Partition}); err != nil {
				return err
			}
		}
	}
	return nil
}

// activeAllocation reports whether an allocation should still grant access.
func activeAllocation(allocation models.ComputeAllocation) bool {
	return allocation.Status == models.ACTIVE
}

// restoreAssociationsForAllocation re-grants every active member of an
// allocation, for when the allocation becomes usable again. Members whose
// account is not provisioned yet are left to the reconciler, the same as on
// first membership.
func (a *AssociationSubscriber) restoreAssociationsForAllocation(ctx context.Context, allocation models.ComputeAllocation) error {
	members, err := a.coreService.ListMembersForAllocation(ctx, allocation.ID)
	if err != nil {
		return fmt.Errorf("list members for allocation: %w", err)
	}
	var failed int
	for _, member := range members {
		if member.MembershipStatus != models.ACTIVE {
			continue
		}
		if err := a.upsertAssociationsForMembership(ctx, member.ComputeAllocationMembership); err != nil {
			if errors.Is(err, errNotProvisioned) {
				continue
			}
			failed++
			slog.Error("Failed to restore association for member",
				"membership_id", member.ID, "allocation_id", allocation.ID, "error", err)
		}
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d members could not be restored", failed, len(members))
	}
	slog.Info("Restored associations for allocation", "account", allocation.Name, "members", len(members))
	return nil
}
