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
	"time"

	"github.com/apache/airavata-custos/pkg/models"
)

// DefaultJobCheckInterval is how often the job-submission health probe runs
// when the connector config does not set job_check_interval.
const DefaultJobCheckInterval = 30 * time.Second

const probeSweepTimeout = time.Minute

// StartStatusProbe reports job-submission health for every active membership
// until ctx is cancelled. It only reads from the cluster; writing and pruning
// stay with the reconciler, so a probe bug can never change cluster state.
func (a *AssociationSubscriber) StartStatusProbe(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = DefaultJobCheckInterval
	}
	slog.Info("Starting SLURM access status probe", "interval", interval)

	a.probeOnce(ctx)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping SLURM access status probe", "reason", ctx.Err())
			return
		case <-ticker.C:
			a.probeOnce(ctx)
		}
	}
}

// probeOnce never lets one bad pass kill the loop; the reconciler must keep
// running regardless of what happens in here.
func (a *AssociationSubscriber) probeOnce(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Access status probe: pass panicked", "panic", r)
		}
	}()
	probeCtx, cancel := context.WithTimeout(ctx, probeSweepTimeout)
	defer cancel()

	clusters, err := a.coreService.ListComputeClusters(probeCtx)
	if err != nil {
		slog.Error("Access status probe: failed to list compute clusters", "error", err)
		return
	}
	for _, cluster := range clusters {
		a.probeCluster(probeCtx, cluster)
	}
}

func (a *AssociationSubscriber) probeCluster(ctx context.Context, cluster models.ComputeCluster) {
	clusterUsers, err := a.coreService.ListComputeClusterUsersByCluster(ctx, cluster.ID)
	if err != nil {
		slog.Error("Access status probe: failed to list cluster users", "cluster_id", cluster.ID, "error", err)
		return
	}

	existing, listErr := a.existingAssociations(ctx, cluster.Name)

	for _, csu := range clusterUsers {
		memberships, err := a.coreService.ListAllocationsForUser(ctx, csu.UserID)
		if err != nil {
			slog.Error("Access status probe: failed to list memberships", "user_id", csu.UserID, "error", err)
			continue
		}
		for _, membership := range memberships {
			if membership.MembershipStatus != models.ACTIVE {
				continue
			}
			desired, err := a.desiredAssociationsForMembership(ctx, membership)
			switch {
			case errors.Is(err, errNotProvisioned):
				// Only the cluster the membership lives on reports it, so a
				// user on several clusters is not re-recorded by each sweep.
				alloc, aerr := a.coreService.GetComputeAllocation(ctx, membership.ComputeAllocationID)
				if aerr != nil || alloc.ComputeClusterID != cluster.ID {
					continue
				}
				a.recordJobCheck(ctx, membership, csu.UserID, false, "waiting for the cluster account")
				continue
			case err != nil:
				// A failed core lookup says nothing about the cluster; skip
				// rather than record a result that could be wrong.
				slog.Warn("Access status probe: could not resolve expected associations",
					"membership_id", membership.ID, "error", err)
				continue
			}
			relevant := 0
			missing := false
			for _, record := range desired {
				if record.Cluster != cluster.Name {
					continue
				}
				relevant++
				if _, ok := existing[keyOf(record)]; !ok {
					missing = true
				}
			}
			// Memberships on other clusters, inactive allocations, and
			// allocations without resources leave nothing to verify here.
			if relevant == 0 {
				continue
			}
			switch {
			case listErr != nil:
				a.recordJobCheck(ctx, membership, csu.UserID, false, "cluster API unreachable")
			case missing:
				a.recordJobCheck(ctx, membership, csu.UserID, false, "association missing")
			default:
				a.recordJobCheck(ctx, membership, csu.UserID, true, "")
			}
		}
	}
}

func (a *AssociationSubscriber) recordJobCheck(ctx context.Context, membership models.ComputeAllocationMembership, userID string, ok bool, detail string) {
	err := a.coreService.RecordAccessCheckResult(ctx, membership.ComputeAllocationID, userID, models.AccessCheckJobSubmission, ok, detail)
	if err != nil {
		slog.Error("Access status probe: failed to record result",
			"membership_id", membership.ID, "error", err)
	}
}
