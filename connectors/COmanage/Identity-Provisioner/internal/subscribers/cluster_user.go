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

// ClusterUserSubscriber drives the orchestrator from two events,
// ComputeClusterUserCreateEvent and UserIdentityCreateEvent.
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

func (s *ClusterUserSubscriber) RegisterSubscribers() {
	s.bus.SubscribeComputeClusterUserCreated(s.handleClusterUserCreate)
	s.bus.SubscribeUserIdentityCreated(s.handleUserIdentityCreate)
}

// handleUserIdentityCreate re-runs provisioning once the `User` has a `sub`.
// A user provisioned before their first sign-in has none in the registry, so
// the cluster cannot match the person to their ssh login.
func (s *ClusterUserSubscriber) handleUserIdentityCreate(ctx context.Context, identity models.UserIdentity) {
	// Provisioning stores a COmanage identity of its own, which comes back with
	// the `identity`. The event is fired for every `UserIdentity` source, and only the `oidc` one
	// carries the sub that needs to be updated in the COmanage registry.
	if identity.Source != "oidc" || identity.OIDCSub == "" {
		return
	}

	ctx = audit.WithSource(ctx, "comanage")
	ctx, span := tracing.Start(ctx, "comanage.user_identity_create")
	defer span.End()
	span.SetAttributes(attribute.String("comanage.user_id", identity.UserID))

	cu, err := s.core.GetComputeClusterUserByPair(ctx, s.custosClusterID, identity.UserID)
	if errors.Is(err, service.ErrNotFound) {
		// Expected for portal admins, system users, and temp users, which have no compute cluster account.
		slog.Debug("comanage subscriber: no cluster account to link the sub to", "user_id", identity.UserID, "cluster_id", s.custosClusterID)
		return
	}
	if err != nil {
		slog.Error("comanage subscriber: cluster user lookup failed", "user_id", identity.UserID, "err", err)
		return
	}
	if err := s.ops.EnsurePOSIXAccount(ctx, cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.Error("comanage subscriber: EnsurePOSIXAccount failed after identity link", "compute_cluster_user_id", cu.ID, "user_id", identity.UserID, "err", err)
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
		slog.Error("comanage subscriber: EnsurePOSIXAccount failed", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID, "err", err)
		return
	}
	if cu.AccessLevel == models.ClusterAccessAdmin {
		s.provisionClusterAdmin(ctx, &cu)
	}
}

// TODO - add ADMIN cluster users to the cluster admin group.
func (s *ClusterUserSubscriber) provisionClusterAdmin(_ context.Context, cu *models.ComputeClusterUser) {
	slog.Warn("comanage subscriber: cluster admin group provisioning not implemented", "compute_cluster_user_id", cu.ID, "user_id", cu.UserID)
}
