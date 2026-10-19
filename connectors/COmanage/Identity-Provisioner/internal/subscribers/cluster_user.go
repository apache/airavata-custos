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

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/operations"
	"github.com/apache/airavata-custos/internal/audit"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// ClusterUserSubscriber drives the orchestrator from three events,
// ComputeClusterUserApproveEvent, ComputeClusterUserDeleteEvent, and UserOnboardedEvent.
// Both are ignored if the user doesn't have a cluster account on `CustosClusterID`.
type ClusterUserSubscriber struct {
	ops             provisioner
	bus             *events.Bus
	core            service.CoreService
	custosClusterID string
}

// provisioner is what the subscriber needs from the orchestrator.
type provisioner interface {
	EnsurePOSIXAccount(ctx context.Context, cu *models.ComputeClusterUser) error
	RemovePOSIXAccount(ctx context.Context, cu *models.ComputeClusterUser) error
	EnsureClusterAdminMembership(ctx context.Context, cu *models.ComputeClusterUser) error
	EnsureAllocationGroup(ctx context.Context, a *models.ComputeAllocation) error
	RemoveAllocationGroup(ctx context.Context, a *models.ComputeAllocation) error
}

func NewClusterUserSubscriber(c *client.Client, bus *events.Bus, core *service.Service, custosClusterID string) *ClusterUserSubscriber {
	return &ClusterUserSubscriber{
		ops:             operations.New(c, core),
		bus:             bus,
		core:            core,
		custosClusterID: custosClusterID,
	}
}

// RegisterSubscribers subscribes the handlers under the given subscriber name.
func (s *ClusterUserSubscriber) RegisterSubscribers(subscriber string) {
	// Provisioning needs an approved account and an onboarded user approving the user agreement.
	// Whichever of the `approve event` and the `onboarded event` comes second starts it.
	s.bus.SubscribeComputeClusterUserApproved(subscriber, s.handleClusterUserApproved)
	s.bus.SubscribeComputeClusterUserDeleted(subscriber, s.handleClusterUserDeleted)
	s.bus.SubscribeUserOnboarded(subscriber, s.handleUserOnboarded)

	s.bus.SubscribeComputeAllocationCreated(subscriber, s.handleComputeAllocationChanged)
	s.bus.SubscribeComputeAllocationUpdated(subscriber, s.handleComputeAllocationChanged)
	s.bus.SubscribeComputeAllocationDeleted(subscriber, s.handleComputeAllocationDeleted)
	s.bus.SubscribeComputeAllocationMembershipCreated(subscriber, s.handleMembershipChanged)
	s.bus.SubscribeComputeAllocationMembershipUpdated(subscriber, s.handleMembershipChanged)
	s.bus.SubscribeComputeAllocationMembershipDeleted(subscriber, s.handleMembershipChanged)
}

// handleClusterUserDeleted removes the account from the registry.
// An account that was never approved was never sent there, so it is skipped.
func (s *ClusterUserSubscriber) handleClusterUserDeleted(ctx context.Context, cu models.ComputeClusterUser) error {
	if cu.ComputeClusterID != s.custosClusterID || cu.ApprovalStatus != models.ClusterAccountApproved {
		return nil
	}
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.cluster_user_deleted")
	defer span.End()
	if err := s.ops.RemovePOSIXAccount(ctx, &cu); err != nil {
		slog.ErrorContext(ctx, "comanage subscriber: RemovePOSIXAccount failed", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
		return err
	}
	return s.syncUserAllocations(ctx, cu.UserID)
}

// handleUserOnboarded re-runs provisioning once the `User` has a `sub`.
// A user provisioned before their first sign-in has none in the registry, so
// the cluster cannot match the person to their ssh login.
func (s *ClusterUserSubscriber) handleUserOnboarded(ctx context.Context, user models.User) error {
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.user_onboarded")
	defer span.End()
	span.SetAttributes(attribute.String("comanage.user_id", user.ID))

	cu, err := s.core.GetComputeClusterUserByPair(ctx, s.custosClusterID, user.ID)
	if errors.Is(err, service.ErrNotFound) {
		// Expected for portal admins, system users, and temp users, which have no compute cluster account.
		slog.DebugContext(ctx, "comanage subscriber: no cluster account to link the sub to", "user_id", user.ID, "cluster_id", s.custosClusterID)
		return nil
	}
	if err != nil {
		slog.ErrorContext(ctx, "comanage subscriber: cluster user lookup failed", "user_id", user.ID, "err", err)
		return err
	}
	if err := s.ops.EnsurePOSIXAccount(ctx, cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "comanage subscriber: EnsurePOSIXAccount failed after onboarding", "compute_cluster_user_id", cu.ID, "user_id", user.ID, "err", err)
		return err
	}
	if cu.AccessLevel == models.ClusterAccessAdmin && cu.ApprovalStatus == models.ClusterAccountApproved {
		if err := s.ops.EnsureClusterAdminMembership(ctx, cu); err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, "comanage subscriber: EnsureClusterAdminMembership failed after onboarding", "compute_cluster_user_id", cu.ID, "user_id", user.ID, "err", err)
			return err
		}
	}
	return s.syncUserAllocations(ctx, cu.UserID)
}

func (s *ClusterUserSubscriber) handleClusterUserApproved(ctx context.Context, cu models.ComputeClusterUser) error {
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.cluster_user_approved")
	defer span.End()

	if cu.ComputeClusterID != s.custosClusterID {
		return nil
	}
	span.SetAttributes(
		attribute.String("comanage.cluster_user_id", cu.ID),
		attribute.String("comanage.user_id", cu.UserID),
	)
	oidc, err := s.core.GetOIDCIdentity(ctx, cu.UserID)
	if err != nil {
		return err
	}
	if oidc == nil {
		// The registry needs the sign-in sub, so provisioning waits for the first sign-in.
		// The onboarded event runs it then.
		_, _ = s.core.CreateAuditEvent(ctx, &models.AuditEvent{
			EventType:  "ComanageProvisioningDeferred",
			EntityID:   cu.ID,
			EntityType: "compute_cluster_user",
			Details:    "cluster_user_id=" + cu.ID + " reason=no_oidc_identity",
		})
		return nil
	}
	// Subscription marker so downstream audits have a parent in the table.
	_, _ = s.core.CreateAuditEvent(ctx, &models.AuditEvent{
		EventType:  "ComanageProvisioningStarted",
		EntityID:   cu.ID,
		EntityType: "compute_cluster_user",
		Details:    "cluster_user_id=" + cu.ID + " user_id=" + cu.UserID,
	})
	// TODO: move to a transactional scope. In-process delivery loses events
	// if the process crashes between the core commit and subscriber pickup.
	if err := s.ops.EnsurePOSIXAccount(ctx, &cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "comanage subscriber: EnsurePOSIXAccount failed", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
		return err
	}
	// A failed group join returns the error so the bus retries. The account
	// already works without sudo in the meantime.
	if cu.AccessLevel == models.ClusterAccessAdmin {
		if err := s.ops.EnsureClusterAdminMembership(ctx, &cu); err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, "comanage subscriber: EnsureClusterAdminMembership failed", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
			return err
		}
	}
	return s.syncUserAllocations(ctx, cu.UserID)
}
