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

type Subscriber struct {
	ops             provisioner
	bus             *events.Bus
	core            service.CoreService
	custosClusterID string
}

// provisioner is what the subscriber needs from the orchestrator.
type provisioner interface {
	EnsureClusterAccount(ctx context.Context, cu *models.ComputeClusterUser) error
	RemoveClusterAccount(ctx context.Context, cu *models.ComputeClusterUser) error
	EnsureClusterAdminMembership(ctx context.Context, cu *models.ComputeClusterUser) error
	EnsureAllocationGroup(ctx context.Context, a *models.ComputeAllocation) error
	RemoveAllocationGroup(ctx context.Context, a *models.ComputeAllocation) error
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
	s.bus.SubscribeComputeClusterUserIdentityRequested(subscriber, s.handleClusterUserApproved)
	s.bus.SubscribeComputeClusterUserDeleted(subscriber, s.handleClusterUserDeleted)
	s.bus.SubscribeComputeAllocationGroupRequested(subscriber, s.handleComputeAllocationChanged)
	s.bus.SubscribeComputeAllocationDeleted(subscriber, s.handleComputeAllocationDeleted)
}

// handleClusterUserDeleted removes the account from the registry.
// An account that was never approved was never sent there, so it is skipped.
func (s *Subscriber) handleClusterUserDeleted(ctx context.Context, cu models.ComputeClusterUser) error {
	if cu.ComputeClusterID != s.custosClusterID || cu.ApprovalStatus != models.ClusterAccountApproved {
		return nil
	}
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.cluster_user_deleted")
	defer span.End()
	if err := s.ops.RemoveClusterAccount(ctx, &cu); err != nil {
		slog.ErrorContext(ctx, "comanage subscriber: RemovePOSIXAccount failed", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
		return err
	}
	return nil
}

func (s *Subscriber) handleClusterUserApproved(ctx context.Context, cu models.ComputeClusterUser) error {
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.cluster_user_approved")
	defer span.End()

	// A retried request can outlive its account, so it acts on the current row.
	row, err := s.core.GetComputeClusterUser(ctx, cu.ID)
	if errors.Is(err, service.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	cu = *row
	if cu.ComputeClusterID != s.custosClusterID {
		return s.core.MarkComputeClusterUserProvisioned(ctx, cu.ID)
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
		// Core requests it again at onboarding.
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
	if err := s.ops.EnsureClusterAccount(ctx, &cu); err != nil {
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
	return nil
}
