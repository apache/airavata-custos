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

// Subscriber drives the orchestrator from cluster user approve and delete, user onboarded and user update events.
// It ignores users without an account on `CustosClusterID`.
type Subscriber struct {
	ops             provisioner
	bus             *events.Bus
	core            service.CoreService
	custosClusterID string
}

// provisioner is what the subscriber needs from the orchestrator.
type provisioner interface {
	ProvisionUnixClusterAccount(ctx context.Context, cu *models.ComputeClusterUser) error
	DeprovisionUnixClusterAccount(ctx context.Context, cu *models.ComputeClusterUser) error
	ProvisionAdminGroupMembership(ctx context.Context, cu *models.ComputeClusterUser) error
	ProvisionAllocationGroup(ctx context.Context, a *models.ComputeAllocation) error
	DeprovisionAllocationGroup(ctx context.Context, a *models.ComputeAllocation) error
}

func NewSubscriber(c *client.Client, bus *events.Bus, core *service.Service, custosClusterID string) *Subscriber {
	return &Subscriber{
		ops:             operations.New(c, core),
		bus:             bus,
		core:            core,
		custosClusterID: custosClusterID,
	}
}

// RegisterSubscribers subscribes the handlers under the given subscriber name.
func (s *Subscriber) RegisterSubscribers(subscriber string) {
	// Provisioning needs an approved account and an ACTIVE user.
	// The later of the approve event and the onboarded or update event starts it.
	s.bus.SubscribeComputeClusterUserApproved(subscriber, s.handleComputeClusterUserApproved)
	s.bus.SubscribeComputeClusterUserDeleted(subscriber, s.handleComputeClusterUserDeleted)
	s.bus.SubscribeUserOnboarded(subscriber, s.handleUserOnboarded)
	s.bus.SubscribeUserUpdated(subscriber, s.handleUserUpdated)

	s.bus.SubscribeComputeAllocationCreated(subscriber, s.handleComputeAllocationChanged)
	s.bus.SubscribeComputeAllocationUpdated(subscriber, s.handleComputeAllocationChanged)
	s.bus.SubscribeComputeAllocationDeleted(subscriber, s.handleComputeAllocationDeleted)
	s.bus.SubscribeComputeAllocationMembershipCreated(subscriber, s.handleMembershipChanged)
	s.bus.SubscribeComputeAllocationMembershipUpdated(subscriber, s.handleMembershipChanged)
	s.bus.SubscribeComputeAllocationMembershipDeleted(subscriber, s.handleMembershipChanged)
}

// handleComputeClusterUserDeleted removes the account from the registry.
// An account that was never approved was never sent there, so it is skipped.
func (s *Subscriber) handleComputeClusterUserDeleted(ctx context.Context, cu models.ComputeClusterUser) error {
	if cu.ComputeClusterID != s.custosClusterID || cu.ApprovalStatus != models.ComputeClusterUserApproved {
		return nil
	}
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.cluster_user_deleted")
	defer span.End()
	if err := s.ops.DeprovisionUnixClusterAccount(ctx, &cu); err != nil {
		slog.ErrorContext(ctx, "comanage subscriber: DeprovisionUnixClusterAccount failed", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
		return err
	}
	return s.syncUserAllocations(ctx, cu.UserID)
}

// handleUserUpdated provisions an approved account once its user is ACTIVE.
// It resends a provisioned account so its registry status follows the user's.
func (s *Subscriber) handleUserUpdated(ctx context.Context, user models.User) error {
	cu, err := s.core.GetComputeClusterUserByPair(ctx, s.custosClusterID, user.ID)
	if errors.Is(err, service.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if cu.ProvisionedAt == nil {
		if cu.ApprovalStatus != models.ComputeClusterUserApproved || user.Status != models.UserActive {
			return nil
		}
		return s.handleComputeClusterUserApproved(ctx, *cu)
	}
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.user_updated")
	defer span.End()
	return s.ops.ProvisionUnixClusterAccount(ctx, cu)
}

// handleUserOnboarded re-runs provisioning once the `User` has a `sub`.
// A user provisioned before their first sign-in has none in the registry, so
// the cluster cannot match the person to their ssh login.
func (s *Subscriber) handleUserOnboarded(ctx context.Context, user models.User) error {
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
	if err := s.ops.ProvisionUnixClusterAccount(ctx, cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "comanage subscriber: ProvisionUnixClusterAccount failed after onboarding", "compute_cluster_user_id", cu.ID, "user_id", user.ID, "err", err)
		return err
	}
	if cu.AccessLevel == models.ClusterAccessAdmin && cu.ApprovalStatus == models.ComputeClusterUserApproved {
		if err := s.ops.ProvisionAdminGroupMembership(ctx, cu); err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, "comanage subscriber: ProvisionAdminGroupMembership failed after onboarding", "compute_cluster_user_id", cu.ID, "user_id", user.ID, "err", err)
			return err
		}
	}
	return s.syncUserAllocations(ctx, cu.UserID)
}

func (s *Subscriber) handleComputeClusterUserApproved(ctx context.Context, cu models.ComputeClusterUser) error {
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
	user, err := s.core.GetUser(ctx, cu.UserID)
	if err != nil {
		return err
	}
	if user.Status != models.UserActive {
		// The user update event runs provisioning once the user is ACTIVE.
		_, _ = s.core.CreateAuditEvent(ctx, &models.AuditEvent{
			EventType:  "ComanageProvisioningDeferred",
			EntityID:   cu.ID,
			EntityType: "compute_cluster_user",
			Details:    "cluster_user_id=" + cu.ID + " reason=user_not_active",
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
	if err := s.ops.ProvisionUnixClusterAccount(ctx, &cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "comanage subscriber: ProvisionUnixClusterAccount failed", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
		return err
	}
	// A failed group join returns the error so the bus retries. The account
	// already works without sudo in the meantime.
	if cu.AccessLevel == models.ClusterAccessAdmin {
		if err := s.ops.ProvisionAdminGroupMembership(ctx, &cu); err != nil {
			span.RecordError(err)
			slog.ErrorContext(ctx, "comanage subscriber: ProvisionAdminGroupMembership failed", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
			return err
		}
	}
	return s.syncUserAllocations(ctx, cu.UserID)
}
