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
	"log/slog"
	"slices"
	"time"

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

const (
	// A sweep touches every provisioned member, so it gets a longer budget
	// than a single event handler.
	reconcileSweepTimeout = 5 * time.Minute
)

func (a *AssociationSubscriber) SubscribeToComputeAllocationMembershipChange(ctx context.Context, m models.ComputeAllocationMembership) error {
	return a.handleEvent(ctx, "compute_allocation_membership_change", m.ID, func(ctx context.Context) error {
		return a.syncAccount(ctx, models.ComputeAllocation{ID: m.ComputeAllocationID})
	})
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationMembershipResourceOverrideChange(ctx context.Context, o models.ComputeAllocationMembershipResourceOverride) error {
	return a.handleEvent(ctx, "compute_allocation_membership_resource_override_change", o.ID, func(ctx context.Context) error {
		m, err := a.coreService.GetComputeAllocationMembership(ctx, o.ComputeAllocationMembershipID)
		if err != nil {
			return err
		}
		return a.syncAccount(ctx, models.ComputeAllocation{ID: m.ComputeAllocationID})
	})
}

func (a *AssociationSubscriber) SubscribeToComputeClusterUserChange(ctx context.Context, cu models.ComputeClusterUser) error {
	return a.handleEvent(ctx, "compute_cluster_user_change", cu.ID, func(ctx context.Context) error { return a.syncUser(ctx, cu) })
}

// syncUser syncs the accounts of a provisioned user and removes any other.
func (a *AssociationSubscriber) syncUser(ctx context.Context, cu models.ComputeClusterUser) error {
	current, err := a.coreService.GetComputeClusterUserByPair(ctx, cu.ComputeClusterID, cu.UserID)
	if err != nil && !errors.Is(err, service.ErrNotFound) {
		return err
	}
	if current == nil || current.ProvisionedAt == nil || current.LocalUsername != cu.LocalUsername {
		cluster, err := a.slurmCluster(ctx, cu.ComputeClusterID)
		if err != nil || cluster == "" {
			return err
		}
		return a.removeUser(ctx, cluster, cu.LocalUsername)
	}
	memberships, err := a.coreService.ListAllocationsForUser(ctx, cu.UserID)
	if err != nil {
		return err
	}
	var errs []error
	for _, m := range memberships {
		if err := a.syncAccount(ctx, models.ComputeAllocation{ID: m.ComputeAllocationID}); !errors.Is(err, service.ErrNotFound) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// removeUser only removes users holding the no-jobs default, and keeps a user with associations on another cluster.
func (a *AssociationSubscriber) removeUser(ctx context.Context, cluster, user string) error {
	others, err := a.slurmClient.ListAssociations(client.AssocFilter{User: user})
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(others, func(x client.Association) bool { return x.Account == client.DefaultAccount }) {
		return nil
	}
	if !slices.ContainsFunc(others, func(x client.Association) bool { return x.Cluster != cluster }) && a.slurmClient.DeleteUser(user) == nil {
		slog.InfoContext(ctx, "Deleted Slurm user", "user", user)
		return nil
	}
	return a.syncAssociations(ctx, client.AssocFilter{Cluster: cluster, User: user}, nil)
}
