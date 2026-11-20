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

// Subscriber drives the orchestrator from four events, ComputeClusterUserApproveEvent,
// ComputeClusterUserUpdateEvent, ComputeClusterUserDeleteEvent, and UserOnboardedEvent.
// Each is ignored if the user doesn't have a compute cluster account on `ComputeClusterID`.
type Subscriber struct {
	ops              provisioner
	bus              *events.Bus
	core             service.CoreService
	computeClusterID string
}

// provisioner is what the subscriber needs from the orchestrator.
type provisioner interface {
	ProvisionUnixClusterAccount(ctx context.Context, cu *models.ComputeClusterUser) error
	DeprovisionUnixClusterAccount(ctx context.Context, cu *models.ComputeClusterUser) error
	DeactivateUnixClusterAccount(ctx context.Context, cu *models.ComputeClusterUser) error
	ProvisionAdminGroupMembership(ctx context.Context, cu *models.ComputeClusterUser) error
	ProvisionAllocationGroup(ctx context.Context, a *models.ComputeAllocation) error
	DeprovisionAllocationGroup(ctx context.Context, a *models.ComputeAllocation) error
}

func NewSubscriber(c *client.Client, bus *events.Bus, core *service.Service, computeClusterID string) *Subscriber {
	return &Subscriber{
		ops:              operations.New(c, core),
		bus:              bus,
		core:             core,
		computeClusterID: computeClusterID,
	}
}

// RegisterSubscribers subscribes the handlers under the given subscriber name.
func (s *Subscriber) RegisterSubscribers(subscriber string) {
	// Provisioning needs an approved account and an onboarded user approving the user agreement.
	// Whichever of the `approve event` and the `onboarded event` comes second starts it.
	s.bus.SubscribeComputeClusterUserApproved(subscriber, s.handleComputeClusterUserApproved)
	s.bus.SubscribeComputeClusterUserDeleted(subscriber, s.handleComputeClusterUserDeleted)
	s.bus.SubscribeComputeClusterUserUpdated(subscriber, s.handleComputeClusterUserUpdated)
	s.bus.SubscribeUserOnboarded(subscriber, s.handleUserOnboarded)

	s.bus.SubscribeComputeAllocationCreated(subscriber, s.handleComputeAllocationChanged)
	s.bus.SubscribeComputeAllocationUpdated(subscriber, s.handleComputeAllocationChanged)
	s.bus.SubscribeComputeAllocationDeleted(subscriber, s.handleComputeAllocationDeleted)
	s.bus.SubscribeComputeAllocationMembershipCreated(subscriber, s.handleMembershipChanged)
	s.bus.SubscribeComputeAllocationMembershipUpdated(subscriber, s.handleMembershipChanged)
	s.bus.SubscribeComputeAllocationMembershipDeleted(subscriber, s.handleMembershipChanged)
}

// handleComputeClusterUserUpdated provisions the account of an ACTIVE user, and otherwise updates its status and leaves its allocation groups.
func (s *Subscriber) handleComputeClusterUserUpdated(ctx context.Context, req models.ComputeClusterUser) error {
	if req.ComputeClusterID != s.computeClusterID {
		return nil
	}
	cu, err := s.core.GetComputeClusterUser(ctx, req.ID)
	if errors.Is(err, service.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	active, err := s.isUserActive(ctx, cu.UserID)
	if err != nil {
		return err
	}
	if active {
		return s.handleComputeClusterUserApproved(ctx, *cu)
	}
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.cluster_user_updated")
	defer span.End()
	if err := s.ops.DeactivateUnixClusterAccount(ctx, cu); err != nil {
		return err
	}
	return s.syncUserAllocations(ctx, cu.UserID)
}

// provisionUnixClusterAccount provisions the account, then suspends it if its user left ACTIVE meanwhile.
func (s *Subscriber) provisionUnixClusterAccount(ctx context.Context, cu *models.ComputeClusterUser) error {
	if err := s.ops.ProvisionUnixClusterAccount(ctx, cu); err != nil {
		return err
	}
	if active, err := s.isUserActive(ctx, cu.UserID); err != nil || active {
		return err
	}
	return s.ops.DeactivateUnixClusterAccount(ctx, cu)
}

// isUserActive re-reads the user, so a retried event sees a status change made since it was published.
func (s *Subscriber) isUserActive(ctx context.Context, userID string) (bool, error) {
	u, err := s.core.GetUser(ctx, userID)
	if errors.Is(err, service.ErrNotFound) {
		return false, nil
	}
	return err == nil && u.Status == models.UserActive, err
}

// handleComputeClusterUserDeleted removes the account from the registry.
// An account that was never approved was never sent there, so it is skipped.
func (s *Subscriber) handleComputeClusterUserDeleted(ctx context.Context, cu models.ComputeClusterUser) error {
	if cu.ComputeClusterID != s.computeClusterID || cu.ApprovalStatus != models.ComputeClusterUserApproved {
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

// handleUserOnboarded provisions the account once the user signs in, which makes them ACTIVE.
// An account approved earlier is held until then.
func (s *Subscriber) handleUserOnboarded(ctx context.Context, user models.User) error {
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.user_onboarded")
	defer span.End()
	span.SetAttributes(attribute.String("comanage.user_id", user.ID))

	cu, err := s.core.GetComputeClusterUserByPair(ctx, s.computeClusterID, user.ID)
	if errors.Is(err, service.ErrNotFound) {
		// Expected for portal admins, system users, and temp users, which have no compute cluster account.
		slog.DebugContext(ctx, "comanage subscriber: no compute cluster account to link the sub to", "user_id", user.ID, "cluster_id", s.computeClusterID)
		return nil
	}
	if err != nil {
		slog.ErrorContext(ctx, "comanage subscriber: compute cluster user lookup failed", "user_id", user.ID, "err", err)
		return err
	}
	if active, err := s.isUserActive(ctx, user.ID); err != nil || !active {
		return err
	}
	if err := s.provisionUnixClusterAccount(ctx, cu); err != nil {
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

	if cu.ComputeClusterID != s.computeClusterID {
		return nil
	}
	// A retry of a suspend that failed after provisioning lands here too.
	if active, err := s.isUserActive(ctx, cu.UserID); err != nil {
		return err
	} else if !active {
		return s.ops.DeactivateUnixClusterAccount(ctx, &cu)
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
	if err := s.provisionUnixClusterAccount(ctx, &cu); err != nil {
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
