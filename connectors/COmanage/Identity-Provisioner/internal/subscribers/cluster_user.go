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

// ClusterUserSubscriber listens for ComputeClusterUserCreateEvent and drives
// the orchestrator. Events whose ComputeClusterID does not match
// CustosClusterID are dropped.
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

func (s *ClusterUserSubscriber) RegisterSubscribers() {
	s.bus.SubscribeComputeClusterUserCreated(s.handleClusterUserCreate)
	s.bus.SubscribeUserIdentityCreated(s.handleUserIdentityCreate)
}

// handleUserIdentityCreate re-runs provisioning once the user has a `sub`.
// A user provisioned before their first sign-in has none in the registry, so
// the cluster cannot match the person to their ssh login.
func (s *ClusterUserSubscriber) handleUserIdentityCreate(ctx context.Context, ident models.UserIdentity) {
	// Provisioning stores a comanage identity of its own, which lands back
	// here. Without this the handler would call itself without end.
	if ident.Source != "oidc" || ident.OIDCSub == "" {
		return
	}

	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.user_identity_create")
	defer span.End()
	span.SetAttributes(attribute.String("comanage.user_id", ident.UserID))

	cu, err := s.core.GetComputeClusterUserByPair(ctx, s.custosClusterID, ident.UserID)
	if errors.Is(err, service.ErrNotFound) {
		// No account on this cluster, so there is nothing to link.
		return
	}
	if err != nil {
		slog.Error("comanage subscriber: cluster user lookup failed", "user_id", ident.UserID, "err", err)
		return
	}
	if err := s.ops.EnsurePOSIXAccount(ctx, cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("comanage subscriber: EnsurePOSIXAccount failed after identity link", "compute_cluster_user_id", cu.ID, "user_id", ident.UserID, "err", err)
	}
}

func (s *ClusterUserSubscriber) handleClusterUserCreate(ctx context.Context, cu models.ComputeClusterUser) {
	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.cluster_user_create")
	defer span.End()

	if cu.ComputeClusterID != s.custosClusterID {
		return
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
		slog.Error("comanage subscriber: EnsurePOSIXAccount failed",
			"compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
		return
	}
	if cu.AccessLevel == models.ClusterAccessAdmin {
		s.provisionClusterAdmin(ctx, &cu)
	}
}

// TODO - add ADMIN cluster users to the cluster admin group.
func (s *ClusterUserSubscriber) provisionClusterAdmin(_ context.Context, cu *models.ComputeClusterUser) {
	slog.Warn("comanage subscriber: cluster admin group provisioning not implemented",
		"compute_cluster_user_id", cu.ID, "user_id", cu.UserID)
}
