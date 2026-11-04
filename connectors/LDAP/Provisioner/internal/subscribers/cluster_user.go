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

// Package subscribers wires cluster account events to the LDAP
// orchestrator.
package subscribers

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/operations"
	"github.com/apache/airavata-custos/internal/audit"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// ClusterUserSubscriber provisions approved cluster accounts and
// deactivates deleted ones. Events for other clusters are dropped.
type ClusterUserSubscriber struct {
	ops             *operations.Orchestrator
	bus             *events.Bus
	core            *service.Service
	custosClusterID string
}

func NewClusterUserSubscriber(c *client.Client, accounts operations.Accounts, floor int64, bus *events.Bus, core *service.Service, custosClusterID string) *ClusterUserSubscriber {
	return &ClusterUserSubscriber{
		ops:             operations.New(c, core, accounts, floor),
		bus:             bus,
		core:            core,
		custosClusterID: custosClusterID,
	}
}

func (s *ClusterUserSubscriber) RegisterSubscribers(subscriber string) {
	s.bus.SubscribeComputeClusterUserApproved(subscriber, s.handleClusterUserApproved)
	s.bus.SubscribeComputeClusterUserDeleted(subscriber, s.handleClusterUserDeleted)
}

// handleClusterUserApproved returns errors so the bus retries the delivery.
func (s *ClusterUserSubscriber) handleClusterUserApproved(ctx context.Context, c models.ComputeClusterUser) error {
	if c.ComputeClusterID != s.custosClusterID {
		return nil
	}
	cu := &c
	ctx = audit.WithSource(ctx, "ldap")
	ctx, span := tracing.Start(ctx, "ldap.cluster_user_approved")
	defer span.End()
	span.SetAttributes(
		attribute.String("ldap.cluster_user_id", cu.ID),
		attribute.String("ldap.user_id", cu.UserID),
	)
	// Subscription marker so downstream audits have a parent in the table.
	_, _ = s.core.CreateAuditEvent(ctx, &models.AuditEvent{
		EventType:  "LDAPProvisioningStarted",
		EntityID:   cu.ID,
		EntityType: "compute_cluster_user",
		Details:    "cluster_user_id=" + cu.ID + " user_id=" + cu.UserID,
	})
	if err := s.ops.EnsurePOSIXAccount(ctx, cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "ldap subscriber: EnsurePOSIXAccount failed",
			"correlation_id", cu.ID, "user_id", cu.UserID, "err", err)
		return err
	}
	return nil
}

func (s *ClusterUserSubscriber) handleClusterUserDeleted(ctx context.Context, cu models.ComputeClusterUser) error {
	if cu.ComputeClusterID != s.custosClusterID || cu.ApprovalStatus != models.ClusterAccountApproved {
		return nil
	}
	ctx = audit.WithSource(ctx, "ldap")
	ctx, span := tracing.Start(ctx, "ldap.cluster_user_deleted")
	defer span.End()
	span.SetAttributes(attribute.String("ldap.cluster_user_id", cu.ID))
	if err := s.ops.DeactivatePOSIXAccount(ctx, &cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		slog.ErrorContext(ctx, "ldap subscriber: DeactivatePOSIXAccount failed",
			"correlation_id", cu.ID, "user_id", cu.UserID, "err", err)
		return err
	}
	return nil
}
