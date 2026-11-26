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
	"fmt"
	"log/slog"
	"time"

	"github.com/apache/airavata-custos/internal/audit"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	coreCallTimeout = 10 * time.Second
	// A sweep touches every provisioned member, so it gets a longer budget
	// than a single event handler.
	reconcileSweepTimeout = 5 * time.Minute
)

// eventSync is a core event that syncs SLURM, audited on the event's entity.
type eventSync[T any] struct {
	span, failedEvent, okEvent, entityType string
	id                                     func(T) string
	sync                                   func(*AssociationSubscriber, context.Context, T) error
}

func allocationID(x models.ComputeAllocation) string                         { return x.ID }
func membershipID(x models.ComputeAllocationMembership) string               { return x.ID }
func mappingID(x models.ComputeAllocationResourceMapping) string             { return x.ID }
func overrideID(x models.ComputeAllocationMembershipResourceOverride) string { return x.ID }
func clusterUserID(x models.ComputeClusterUser) string                       { return x.ID }

var (
	allocationCreated = eventSync[models.ComputeAllocation]{"slurm.compute_allocation_create",
		"ComputeAllocationCreationFailed", "ComputeAllocationCreationSucceeded", "compute_allocation", allocationID, (*AssociationSubscriber).syncAllocation}
	allocationUpdated = eventSync[models.ComputeAllocation]{"slurm.compute_allocation_update",
		"ComputeAllocationUpdateFailed", "ComputeAllocationUpdateSucceeded", "compute_allocation", allocationID, (*AssociationSubscriber).syncAllocation}
	allocationDeleted = eventSync[models.ComputeAllocation]{"slurm.compute_allocation_delete",
		"ComputeAllocationDeletionFailed", "ComputeAllocationDeleted", "compute_allocation", allocationID, (*AssociationSubscriber).syncAllocation}
	membershipCreated = eventSync[models.ComputeAllocationMembership]{"slurm.compute_allocation_membership_create",
		"ComputeAllocationMembershipCreationFailed", "ComputeAllocationMembershipCreationSucceeded", "compute_allocation_membership", membershipID, (*AssociationSubscriber).syncMembership}
	membershipUpdated = eventSync[models.ComputeAllocationMembership]{"slurm.compute_allocation_membership_update",
		"ComputeAllocationMembershipUpdateFailed", "ComputeAllocationMembershipUpdateSucceeded", "compute_allocation_membership", membershipID, (*AssociationSubscriber).syncMembership}
	// The delete guards have already revoked access.
	membershipDeleted = eventSync[models.ComputeAllocationMembership]{"slurm.compute_allocation_membership_delete",
		"ComputeAllocationMembershipDeletionFailed", "ComputeAllocationMembershipDeleted", "compute_allocation_membership", membershipID, (*AssociationSubscriber).syncMembership}
	overrideCreated = eventSync[models.ComputeAllocationMembershipResourceOverride]{"slurm.compute_allocation_membership_resource_override_create",
		"ComputeAllocationMembershipResourceOverrideCreationFailed", "ComputeAllocationMembershipResourceOverrideCreationSucceeded", "compute_allocation_membership_resource_override", overrideID, (*AssociationSubscriber).syncOverride}
	mappingCreated = eventSync[models.ComputeAllocationResourceMapping]{"slurm.compute_allocation_resource_mapping_create",
		"ComputeAllocationResourceMappingCreationFailed", "ComputeAllocationResourceMappingCreationSucceeded", "compute_allocation_resource_mapping", mappingID, (*AssociationSubscriber).syncMapping}
	clusterUserDeleted = eventSync[models.ComputeClusterUser]{"slurm.compute_cluster_user_delete",
		"ComputeClusterUserDeletionFailed", "ComputeClusterUserDeprovisioned", "compute_cluster_user", clusterUserID, (*AssociationSubscriber).deprovisionSlurmUser}
)

// run syncs the event in its own span and audits the outcome on the event's entity.
func (e eventSync[T]) run(ctx context.Context, a *AssociationSubscriber, payload T) error {
	id := e.id(payload)
	ctx, span := tracing.Start(ctx, e.span, trace.WithAttributes(attribute.String("slurm.entity_id", id)))
	defer span.End()
	ctx = audit.WithSource(ctx, "slurm")
	slog.InfoContext(ctx, "Received SLURM sync event", "entity_type", e.entityType, "entity_id", id)
	fetchCtx, cancel := context.WithTimeout(ctx, coreCallTimeout)
	defer cancel()
	if err := e.sync(a, fetchCtx, payload); err != nil {
		slog.ErrorContext(ctx, "Failed to sync SLURM", "error", err, "entity_type", e.entityType, "entity_id", id)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		a.recordAuditEvent(ctx, e.failedEvent, e.entityType, id, "Failed to sync SLURM. Error: "+err.Error())
		return err
	}
	a.recordAuditEvent(ctx, e.okEvent, e.entityType, id, "Synced SLURM.")
	return nil
}

func (a *AssociationSubscriber) syncMembership(ctx context.Context, membership models.ComputeAllocationMembership) error {
	return a.syncAllocationCluster(ctx, membership.ComputeAllocationID)
}

func (a *AssociationSubscriber) syncOverride(ctx context.Context, override models.ComputeAllocationMembershipResourceOverride) error {
	membership, err := a.coreService.GetComputeAllocationMembership(ctx, override.ComputeAllocationMembershipID)
	if err != nil {
		return fmt.Errorf("get compute allocation membership: %w", err)
	}
	return a.syncAllocationCluster(ctx, membership.ComputeAllocationID)
}
