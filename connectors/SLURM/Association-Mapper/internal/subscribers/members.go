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

	"github.com/apache/airavata-custos/internal/audit"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	coreCallTimeout = 10 * time.Second
	// A sweep touches every provisioned member, so it gets a longer budget
	// than a single event handler.
	reconcileSweepTimeout = 5 * time.Minute
)

func (a *AssociationSubscriber) SubscribeToComputeAllocationMembershipCreation(ctx context.Context, membership models.ComputeAllocationMembership) {
	ctx = audit.WithSource(ctx, "slurm")
	ctx, span := tracing.Start(ctx, "slurm.compute_allocation_membership_create")
	defer span.End()
	span.SetAttributes(
		attribute.String("slurm.allocation_id", membership.ComputeAllocationID),
		attribute.String("slurm.user_id", membership.UserID),
	)

	slog.Info("Received compute allocation membership creation event", "membership", membership)

	// The short timeout bounds the core fetches only; the provisioning
	// wait below runs on its own deadline.
	fetchCtx, cancel := context.WithTimeout(ctx, coreCallTimeout)
	defer cancel()

	allocation, err := a.coreService.GetComputeAllocation(fetchCtx, membership.ComputeAllocationID)
	if err != nil {
		slog.Error("Failed to get compute allocation", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationMembershipCreationFailed", "compute_allocation_membership", membership.ID, "Failed to get compute allocation. Error: "+err.Error())
		return
	}

	cluster, err := a.coreService.GetComputeCluster(fetchCtx, allocation.ComputeClusterID)
	if err != nil {
		slog.Error("Failed to get compute cluster", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationMembershipCreationFailed", "compute_allocation_membership", membership.ID, "Failed to get compute cluster. Error: "+err.Error())
		return
	}
	span.SetAttributes(attribute.String("slurm.cluster_id", cluster.ID))

	// One writer for every path so the association record stays identical
	// for a given (cluster, account, user, partition) key.
	if err := a.upsertAssociationsForMembership(fetchCtx, membership); err != nil {
		if errors.Is(err, errNotProvisioned) {
			// Expected on a first-time account: provisioning is still in
			// flight. The reconciler writes it once the account lands.
			slog.Info("Cluster account not provisioned yet, leaving the association to the reconciler",
				"membership_id", membership.ID, "user_id", membership.UserID)
			return
		}
		slog.Error("Failed to upsert association for membership", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationMembershipCreationFailed", "compute_allocation_membership", membership.ID, "Failed to upsert association. Error: "+err.Error())
		return
	}
	a.recordAuditEvent(ctx, "ComputeAllocationMembershipCreationSucceeded", "compute_allocation_membership", membership.ID, "Successfully upserted association.")
}

func (a *AssociationSubscriber) SubscribeToComputeAllocationMembershipResourceOverrideCreation(ctx context.Context, override models.ComputeAllocationMembershipResourceOverride) {
	ctx = audit.WithSource(ctx, "slurm")
	ctx, span := tracing.Start(ctx, "slurm.compute_allocation_membership_resource_override_create")
	defer span.End()

	slog.Info("Received compute allocation membership resource override creation event", "override", override)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// TODO: read per-resource override via
	membership, err := a.coreService.GetComputeAllocationMembership(ctx, override.ComputeAllocationMembershipID)
	if err != nil {
		slog.Error("Failed to get compute allocation membership for resource override creation", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationMembershipResourceOverrideCreationFailed", "compute_allocation_membership_resource_override", override.ID, "Failed to get compute allocation membership. Error: "+err.Error())
		return
	}
	span.SetAttributes(
		attribute.String("slurm.allocation_id", membership.ComputeAllocationID),
		attribute.String("slurm.user_id", membership.UserID),
	)

	// Same single writer as the membership path: it re-reads every override
	// for this membership, so the record carries the full limit set rather
	// than only this override's slice.
	if err := a.upsertAssociationsForMembership(ctx, *membership); err != nil {
		if errors.Is(err, errNotProvisioned) {
			slog.Info("Cluster account not provisioned yet, leaving the association to the reconciler",
				"membership_id", membership.ID, "override_id", override.ID)
			return
		}
		slog.Error("Failed to upsert association for membership resource override", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationMembershipResourceOverrideCreationFailed", "compute_allocation_membership_resource_override", override.ID, "Failed to upsert association. Error: "+err.Error())
		return
	}
	a.recordAuditEvent(ctx, "ComputeAllocationMembershipResourceOverrideCreationSucceeded", "compute_allocation_membership_resource_override", override.ID, "Successfully upserted association.")
}

// SubscribeToComputeAllocationMembershipUpdate reacts to a membership-changing
// state. Anything other than ACTIVE means the member should no longer be able
// to submit, so the associations go; going back to ACTIVE re-grants them.
func (a *AssociationSubscriber) SubscribeToComputeAllocationMembershipUpdate(ctx context.Context, membership models.ComputeAllocationMembership) {
	ctx = audit.WithSource(ctx, "slurm")
	ctx, span := tracing.Start(ctx, "slurm.compute_allocation_membership_update")
	defer span.End()
	span.SetAttributes(
		attribute.String("slurm.allocation_id", membership.ComputeAllocationID),
		attribute.String("slurm.user_id", membership.UserID),
	)

	fetchCtx, cancel := context.WithTimeout(ctx, coreCallTimeout)
	defer cancel()

	if membership.MembershipStatus == models.ACTIVE {
		if err := a.upsertAssociationsForMembership(fetchCtx, membership); err != nil {
			if errors.Is(err, errNotProvisioned) {
				slog.Info("Cluster account not provisioned yet, leaving the association to the reconciler",
					"membership_id", membership.ID)
				return
			}
			slog.Error("Failed to upsert association for reactivated membership", "error", err)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			a.recordAuditEvent(ctx, "ComputeAllocationMembershipUpdateFailed", "compute_allocation_membership", membership.ID, "Failed to upsert association. Error: "+err.Error())
		}
		return
	}

	if err := a.removeAssociationsForMembership(fetchCtx, membership); err != nil {
		slog.Error("Failed to remove associations for deactivated membership", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationMembershipUpdateFailed", "compute_allocation_membership", membership.ID, "Failed to remove associations. Error: "+err.Error())
		return
	}
	a.recordAuditEvent(ctx, "ComputeAllocationMembershipDeprovisioned", "compute_allocation_membership", membership.ID, "Removed cluster associations for the deactivated membership.")
}

// SubscribeToComputeAllocationMembershipDeletion removes the member's access
// when the membership row itself goes away.
func (a *AssociationSubscriber) SubscribeToComputeAllocationMembershipDeletion(ctx context.Context, membership models.ComputeAllocationMembership) {
	ctx = audit.WithSource(ctx, "slurm")
	ctx, span := tracing.Start(ctx, "slurm.compute_allocation_membership_delete")
	defer span.End()
	span.SetAttributes(
		attribute.String("slurm.allocation_id", membership.ComputeAllocationID),
		attribute.String("slurm.user_id", membership.UserID),
	)

	fetchCtx, cancel := context.WithTimeout(ctx, coreCallTimeout)
	defer cancel()

	if err := a.removeAssociationsForMembership(fetchCtx, membership); err != nil {
		slog.Error("Failed to remove associations for deleted membership", "error", err)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, "ComputeAllocationMembershipDeletionFailed", "compute_allocation_membership", membership.ID, "Failed to remove associations. Error: "+err.Error())
		return
	}
	a.recordAuditEvent(ctx, "ComputeAllocationMembershipDeprovisioned", "compute_allocation_membership", membership.ID, "Removed cluster associations for the deleted membership.")
}
