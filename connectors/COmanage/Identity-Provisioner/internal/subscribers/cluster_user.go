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
// ComputeClusterUserApproveEvent, ComputeClusterUserDeleteEvent, and UserIdentityCreateEvent.
// Both are ignored if the user doesn't have a cluster account on `CustosClusterID`.
type ClusterUserSubscriber struct {
	ops             *operations.Orchestrator
	bus             *events.Bus
	core            *service.Service
	custosClusterID string
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
	// Accounts are created on the cluster only after an admin approves them,
	// so the `approve event` is what starts provisioning, not the `create event`.
	s.bus.SubscribeComputeClusterUserApproved(subscriber, s.handleClusterUserApproved)
	s.bus.SubscribeComputeClusterUserDeleted(subscriber, s.handleClusterUserDeleted)
	s.bus.SubscribeUserIdentityCreated(subscriber, s.handleUserIdentityCreate)

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
	if err := s.ops.RemovePOSIXAccount(ctx, &cu); err != nil {
		slog.Error("comanage subscriber: RemovePOSIXAccount failed", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
		return err
	}
	return s.syncUserAllocations(ctx, cu.UserID)
}

// handleUserIdentityCreate re-runs provisioning once the `User` has a `sub`.
// A user provisioned before their first sign-in has none in the registry, so
// the cluster cannot match the person to their ssh login.
func (s *ClusterUserSubscriber) handleUserIdentityCreate(ctx context.Context, identity models.UserIdentity) error {
	// Provisioning stores a COmanage identity of its own, which comes back with
	// the `identity`. The event is fired for every `UserIdentity` source, and only the `oidc` one
	// carries the sub that needs to be updated in the COmanage registry.
	if identity.Source != "oidc" || identity.OIDCSub == "" {
		return nil
	}

	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.user_identity_create")
	defer span.End()
	span.SetAttributes(attribute.String("comanage.user_id", identity.UserID))

	cu, err := s.core.GetComputeClusterUserByPair(ctx, s.custosClusterID, identity.UserID)
	if errors.Is(err, service.ErrNotFound) {
		// Expected for portal admins, system users, and temp users, which have no compute cluster account.
		slog.Debug("comanage subscriber: no cluster account to link the sub to", "user_id", identity.UserID, "cluster_id", s.custosClusterID)
		return nil
	}
	if err != nil {
		slog.Error("comanage subscriber: cluster user lookup failed", "user_id", identity.UserID, "err", err)
		return err
	}
	if err := s.ops.EnsurePOSIXAccount(ctx, cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("comanage subscriber: EnsurePOSIXAccount failed after identity link", "compute_cluster_user_id", cu.ID, "user_id", identity.UserID, "err", err)
		return err
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
		slog.Error("comanage subscriber: EnsurePOSIXAccount failed", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
		return err
	}
	// Even if this fails, the cluster account is usable without sudo, so a
	// failed group join is logged and left for a retry.
	if cu.AccessLevel == models.ClusterAccessAdmin {
		if err := s.ops.EnsureClusterAdminMembership(ctx, &cu); err != nil {
			span.RecordError(err)
			slog.Error("comanage subscriber: EnsureClusterAdminMembership failed", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
			return err
		}
	}
	return s.syncUserAllocations(ctx, cu.UserID)
}
