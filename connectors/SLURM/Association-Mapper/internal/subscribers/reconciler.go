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

// reconcile re-declares the association for every active membership whose
// account is provisioned. Upserts are idempotent, so this repairs both a
// missing association and one whose limits were overwritten.
func (a *AssociationSubscriber) reconcile(ctx context.Context) {
	sweepCtx, cancel := context.WithTimeout(ctx, reconcileSweepTimeout)
	defer cancel()

	clusters, err := a.coreService.ListComputeClusters(sweepCtx)
	if err != nil {
		slog.Error("Association reconciler: failed to list compute clusters", "error", err)
		return
	}

	written, skipped := 0, 0
	for _, cluster := range clusters {
		clusterUsers, err := a.coreService.ListComputeClusterUsersByCluster(sweepCtx, cluster.ID)
		if err != nil {
			slog.Error("Association reconciler: failed to list cluster users", "cluster_id", cluster.ID, "error", err)
			continue
		}
		for _, csu := range clusterUsers {
			if !a.readyForAssociation(csu) {
				skipped++
				continue
			}
			memberships, err := a.coreService.ListAllocationsForUser(sweepCtx, csu.UserID)
			if err != nil {
				slog.Error("Association reconciler: failed to list memberships", "user_id", csu.UserID, "error", err)
				continue
			}
			for _, membership := range memberships {
				if membership.MembershipStatus != models.ACTIVE {
					continue
				}
				if err := a.upsertAssociationsForMembership(sweepCtx, membership); err != nil {
					slog.Error("Association reconciler: failed to upsert association",
						"membership_id", membership.ID, "user_id", csu.UserID, "error", err)
					continue
				}
				written++
			}
		}
	}
	slog.Debug("Association reconciler pass complete", "written", written, "skipped_unprovisioned", skipped)
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
