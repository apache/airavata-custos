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

// Package operations provisions and removes POSIX accounts in LDAP for
// cluster accounts.
package operations

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/client"
	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/store"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// CoreService is the subset of pkg/service.Service the orchestrator needs.
type CoreService interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
	CreateAuditEvent(ctx context.Context, e *models.AuditEvent) (*models.AuditEvent, error)
	MarkComputeClusterUserProvisioned(ctx context.Context, id string) error
}

// Accounts is the connector's own table of assigned POSIX attributes.
type Accounts interface {
	Get(ctx context.Context, computeClusterUserID string) (*store.Account, error)
	GetByUsername(ctx context.Context, username string) (*store.Account, error)
	Create(ctx context.Context, a store.Account, floor int64) (*store.Account, error)
	Renumber(ctx context.Context, computeClusterUserID string, floor int64) (*store.Account, error)
	SetStatus(ctx context.Context, computeClusterUserID, status string) error
}

type Orchestrator struct {
	c        *client.Client
	core     CoreService
	accounts Accounts
	// floor is the lowest uidNumber to assign.
	floor int64
}

func New(c *client.Client, core *service.Service, accounts Accounts, floor int64) *Orchestrator {
	return &Orchestrator{c: c, core: core, accounts: accounts, floor: floor}
}

func (o *Orchestrator) EnsurePOSIXAccount(ctx context.Context, cu *models.ComputeClusterUser) error {
	ctx, span := tracing.Start(ctx, "ldap.ensure_posix_account")
	defer span.End()
	span.SetAttributes(
		attribute.String("ldap.cluster_user_id", cu.ID),
		attribute.String("ldap.user_id", cu.UserID),
		attribute.String("ldap.base_dn", o.c.Config().BaseDN),
	)
	if err := o.ensurePOSIXAccountImpl(ctx, cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (o *Orchestrator) DeactivatePOSIXAccount(ctx context.Context, cu *models.ComputeClusterUser) error {
	ctx, span := tracing.Start(ctx, "ldap.deactivate_posix_account")
	defer span.End()
	span.SetAttributes(attribute.String("ldap.cluster_user_id", cu.ID))
	if err := o.deactivatePOSIXAccountImpl(ctx, cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (o *Orchestrator) audit(ctx context.Context, cu *models.ComputeClusterUser, eventType, details string) {
	_, _ = o.core.CreateAuditEvent(ctx, &models.AuditEvent{
		EventType:  eventType,
		EntityID:   cu.ID,
		EntityType: "compute_cluster_user",
		Details:    details,
	})
}

// auditFailure records an LDAPProvisioningFailed event naming the failed step.
func (o *Orchestrator) auditFailure(ctx context.Context, cu *models.ComputeClusterUser, step string, err error) {
	details := "step=" + step + " err="
	if err != nil {
		details += err.Error()
	}
	_, _ = o.core.CreateAuditEvent(ctx, &models.AuditEvent{
		EventType:  "LDAPProvisioningFailed",
		EntityID:   cu.ID,
		EntityType: "compute_cluster_user",
		Details:    details,
	})
}
