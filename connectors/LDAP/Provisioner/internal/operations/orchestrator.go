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

// Package operations provisions POSIX identity directly in LDAP for a
// (User, UnixCluster) pair. Parallels the COmanage Identity-Provisioner's
// operations package; direct-to-LDAP path for sites that don't run a
// COmanage Registry.
package operations

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/connectors/LDAP/Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// CoreService is the subset of pkg/service.Service the orchestrator needs.
// Declared here so tests can substitute a fake without standing up a real DB.
type CoreService interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
	ListUserIdentitiesForUser(ctx context.Context, userID string) ([]models.UserIdentity, error)
	CreateUserIdentity(ctx context.Context, ui *models.UserIdentity) (*models.UserIdentity, error)
	CreateAuditEvent(ctx context.Context, e *models.AuditEvent) (*models.AuditEvent, error)
}

// UIDAllocator hands out fresh uidNumbers from a persistent, monotonic
// counter. Never regresses — deleting an LDAP entry does not free the
// number for reuse, which is what closes the "new user inherits deleted
// user's files" hole. Backed by internal/store.UIDSequence in
// production; interfaced here so tests can substitute an in-memory
// counter.
type UIDAllocator interface {
	Allocate(ctx context.Context, clusterID string) (int64, error)
}

// Orchestrator wraps the LDAP client, the core service, and the UID
// allocator. Exposes a single EnsurePOSIXAccount call that the
// subscriber invokes on every ComputeClusterUserCreateEvent it accepts.
type Orchestrator struct {
	c    *client.Client
	core CoreService
	uids UIDAllocator
}

// New builds an Orchestrator. Takes *service.Service directly so the
// loader wiring stays symmetrical with the COmanage connector; the
// exported CoreService interface is what the orchestrator's methods
// actually depend on.
func New(c *client.Client, core *service.Service, uids UIDAllocator) *Orchestrator {
	return &Orchestrator{c: c, core: core, uids: uids}
}

// EnsurePOSIXAccount is the entry point invoked by the subscriber. It
// wraps the real flow in a tracing span so failures surface with the
// step name attached.
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

// audit is the success-side helper: emits an event to the core audit log
// tagged with the ComputeClusterUser row so downstream tooling can trace
// the provisioning history for that user.
func (o *Orchestrator) audit(ctx context.Context, cu *models.ComputeClusterUser, eventType, details string) {
	_, _ = o.core.CreateAuditEvent(ctx, &models.AuditEvent{
		EventType:  eventType,
		EntityID:   cu.ID,
		EntityType: "compute_cluster_user",
		Details:    details,
	})
}

// dlq (dead-letter) is the failure-side helper: emits a ProvisioningFailed
// event with the step name so operators can grep the audit table for
// stuck users without correlating logs.
func (o *Orchestrator) dlq(ctx context.Context, cu *models.ComputeClusterUser, step string, err error) {
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
