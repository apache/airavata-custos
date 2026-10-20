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

	"github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client"
	"github.com/apache/airavata-custos/internal/connectors/health"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// DefaultReconcileInterval is how often the sweep runs when the connector
// config does not set association_reconcile_interval.
const DefaultReconcileInterval = 5 * time.Minute

// DefaultProvisionGrace holds a just-provisioned account's associations back.
// The stamp means the registry write finished, but the account still has to
// reach the directory and the cluster's name cache, which Custos cannot
// observe. How long that takes is site-specific, hence configurable.
const DefaultProvisionGrace = 30 * time.Second

// StartReconciler runs the sweep until ctx is cancelled, beginning with an
// immediate pass so a restart backfills what it missed.
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

// reconcile compares what the cluster has against what the allocations say it
// should have, and applies the difference in both directions: it writes what
// is missing or drifted, and removes access that should no longer exist.
//
// Removal matters more than creation here. A lost creation event only delays
// access, but a lost deactivation leaves a member able to submit indefinitely
// with nothing reporting it.
func (a *AssociationSubscriber) reconcile(ctx context.Context) {
	ctx, span := tracing.Start(ctx, "slurm.reconcile")
	defer span.End()
	sweepCtx, cancel := context.WithTimeout(ctx, reconcileSweepTimeout)
	defer cancel()

	clusters, err := a.coreService.ListComputeClusters(sweepCtx)
	if err != nil {
		slog.ErrorContext(ctx, "Association reconciler: failed to list compute clusters", "error", err)
		return
	}

	for _, cluster := range clusters {
		a.reconcileCluster(sweepCtx, cluster)
	}
}

func (a *AssociationSubscriber) reconcileCluster(ctx context.Context, cluster models.ComputeCluster) {
	err := a.syncCluster(ctx, cluster)
	health.Record(a.name, cluster.ID, err)
	if err != nil {
		slog.ErrorContext(ctx, "Association reconciler", "cluster", cluster.Name, "error", err)
	}
}

func (a *AssociationSubscriber) syncCluster(ctx context.Context, cluster models.ComputeCluster) error {
	name, err := a.slurmCluster(ctx, cluster.ID)
	if err != nil || name == "" {
		return err
	}
	allocations, err := a.coreService.ListComputeAllocationsByCluster(ctx, cluster.ID)
	if err != nil {
		return err
	}
	users, err := a.coreService.ListComputeClusterUsersByCluster(ctx, cluster.ID)
	if err != nil {
		return err
	}
	accounts, err := a.slurmClient.ListAccounts()
	if err != nil {
		return err
	}
	defaults, err := a.slurmClient.ListAssociations(client.AssocFilter{Cluster: name, Account: noAllocation})
	if err != nil {
		return err
	}

	var errs []error
	collect := func(err error) {
		if !errors.Is(err, service.ErrNotFound) && !errors.Is(err, errNotProvisioned) {
			errs = append(errs, err)
		}
	}
	named := map[string]bool{}
	for _, allocation := range allocations {
		named[slurmAccount(allocation.Name)] = true
		collect(a.syncAccount(ctx, allocation))
	}
	for _, account := range accounts {
		if account.Organization == custos && !named[account.Name] {
			collect(a.syncAssociations(ctx, client.AssocFilter{Cluster: name, Account: account.Name}, nil))
		}
	}
	named = map[string]bool{}
	for _, cu := range users {
		named[cu.LocalUsername] = cu.ProvisionedAt != nil
	}
	for _, d := range defaults {
		if d.User != "" && !named[d.User] {
			collect(a.removeUser(ctx, name, d.User))
		}
	}
	return errors.Join(errs...)
}

// readyForAssociation reports whether the account has been provisioned long
// enough for the cluster to resolve it.
func (a *AssociationSubscriber) readyForAssociation(provisionedAt time.Time) bool {
	grace := a.provisionGrace
	if grace <= 0 {
		grace = DefaultProvisionGrace
	}
	return time.Since(provisionedAt) >= grace
}
