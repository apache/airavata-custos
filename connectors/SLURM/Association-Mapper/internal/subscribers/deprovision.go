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

// syncAllocationCluster syncs the cluster the allocation lives on.
func (a *AssociationSubscriber) syncAllocationCluster(ctx context.Context, allocationID string) error {
	allocation, err := a.coreService.GetComputeAllocation(ctx, allocationID)
	if errors.Is(err, service.ErrNotFound) {
		slog.InfoContext(ctx, "Allocation is gone, nothing to sync", "allocation_id", allocationID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("get compute allocation: %w", err)
	}
	return a.syncClusterByID(ctx, allocation.ComputeClusterID)
}

// syncClusterByID syncs a whole cluster, which takes as long as a sweep, so
// it gets the sweep's budget rather than the event's fetch deadline.
func (a *AssociationSubscriber) syncClusterByID(ctx context.Context, clusterID string) error {
	cluster, err := a.coreService.GetComputeCluster(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("get compute cluster: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), reconcileSweepTimeout)
	defer cancel()
	_, err = a.syncCluster(ctx, *cluster)
	return err
}
