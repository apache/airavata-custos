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

// Package operations provisions POSIX identity in COmanage for a
// (CoPerson, UnixCluster) pair.
package operations

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/connectors/COmanage/Identity-Provisioner/internal/client"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// CoreService is the subset of pkg/service.Service the orchestrator depends
// on. Defined here so tests can stub without standing up a real DB.
type CoreService interface {
	GetUser(ctx context.Context, id string) (*models.User, error)
	ListUserIdentitiesForUser(ctx context.Context, userID string) ([]models.UserIdentity, error)
	CreateUserIdentity(ctx context.Context, ui *models.UserIdentity) (*models.UserIdentity, error)
	CreateAuditEvent(ctx context.Context, e *models.AuditEvent) (*models.AuditEvent, error)
}

type Orchestrator struct {
	c    *client.Client
	core CoreService
}

func New(c *client.Client, core *service.Service) *Orchestrator {
	return &Orchestrator{c: c, core: core}
}

func (o *Orchestrator) EnsurePOSIXAccount(ctx context.Context, cu *models.ComputeClusterUser) error {
	ctx, span := tracing.Start(ctx, "comanage.ensure_posix_account")
	defer span.End()
	span.SetAttributes(
		attribute.String("comanage.cluster_user_id", cu.ID),
		attribute.String("comanage.user_id", cu.UserID),
		attribute.Int("comanage.co_id", o.c.Config().COID),
	)
	if err := o.ensurePOSIXAccountImpl(ctx, cu); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}
