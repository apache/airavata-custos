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

package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel/codes"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/internal/tracing"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type RequestProjectReactivateHandler struct {
	svc        *service.Service
	amieClient AmieClient
	auditSvc   AuditService
}

func NewRequestProjectReactivateHandler(svc *service.Service, amieClient AmieClient, auditSvc AuditService) *RequestProjectReactivateHandler {
	return &RequestProjectReactivateHandler{svc: svc, amieClient: amieClient, auditSvc: auditSvc}
}

func (h *RequestProjectReactivateHandler) SupportsType() string { return "request_project_reactivate" }

// Handle flips the Project and all of its ComputeAllocations back to ACTIVE.
// Per AMIE protocol, only the PI's membership is reactivated automatically;
// other members must request reactivation via request_account_reactivate.
func (h *RequestProjectReactivateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) (err error) {
	ctx, span := tracing.Start(ctx, "amie.handle:"+packet.Type)
	defer span.End()
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
	}()

	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}
	projectID := getString(body, "ProjectID")
	if err := requireText(projectID, "ProjectID"); err != nil {
		return err
	}

	// AMIE carries the Custos project.id we returned in notify_project_create.
	project, err := h.svc.GetProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			slog.WarnContext(ctx, "request_project_reactivate: project not found in core; skipping",
				"projectID", projectID)
			return h.reply(ctx, tx, packet, eventID, projectID)
		}
		return fmt.Errorf("request_project_reactivate: lookup project: %w", err)
	}

	if _, err := h.svc.UpdateProjectStatus(ctx, project.ID, models.ProjectActive); err != nil {
		return fmt.Errorf("request_project_reactivate: update project status: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReactivateProject, "project", project.ID, ""); err != nil {
		return fmt.Errorf("request_project_reactivate: audit REACTIVATE_PROJECT: %w", err)
	}

	if err := h.reactivateAllocationsAndPI(ctx, tx, packet, eventID, project); err != nil {
		return err
	}

	return h.reply(ctx, tx, packet, eventID, projectID)
}

// reactivateAllocationsAndPI flips every ComputeAllocation under the project
// back to ACTIVE, writes a status-change Diff per allocation, and reactivates
// only the PI's membership on each allocation. Other members stay INACTIVE
// until their own request_account_reactivate arrives.
func (h *RequestProjectReactivateHandler) reactivateAllocationsAndPI(ctx context.Context, tx *sql.Tx, packet *model.Packet, eventID string, project *models.Project) error {
	allocations, err := h.svc.ListComputeAllocationsByProject(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("list allocations: %w", err)
	}
	for _, a := range allocations {
		a.Status = models.ACTIVE
		if err := h.svc.UpdateComputeAllocation(ctx, &a); err != nil {
			return fmt.Errorf("reactivate allocation %s: %w", a.ID, err)
		}
		if _, err := h.svc.CreateComputeAllocationDiff(ctx, &models.ComputeAllocationDiff{
			ComputeAllocationID: a.ID,
			DiffType:            "ALLOCATION_STATUS_CHANGE",
			Status:              models.ACTIVE,
			Description:         "Reactivated by AMIE request_project_reactivate",
		}); err != nil {
			return fmt.Errorf("record reactivate diff for %s: %w", a.ID, err)
		}
		if project.ProjectPIID == "" {
			continue
		}
		members, err := h.svc.ListMembersForAllocation(ctx, a.ID)
		if err != nil {
			return fmt.Errorf("list memberships for allocation %s: %w", a.ID, err)
		}
		for _, m := range members {
			if m.UserID != project.ProjectPIID {
				continue
			}
			if _, err := h.svc.UpdateMembershipStatus(ctx, m.ID, models.ACTIVE); err != nil {
				return fmt.Errorf("reactivate PI membership %s: %w", m.ID, err)
			}
			if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReactivateMembership, "compute_allocation_membership", m.ID,
				fmt.Sprintf("PI user=%s allocation=%s", m.UserID, a.ID)); err != nil {
				return fmt.Errorf("audit REACTIVATE_MEMBERSHIP for %s: %w", m.ID, err)
			}
		}
	}
	return nil
}

func (h *RequestProjectReactivateHandler) reply(ctx context.Context, tx *sql.Tx, packet *model.Packet, eventID, projectID string) error {
	reply := map[string]any{
		"type": "notify_project_reactivate",
		"body": map[string]any{"ProjectID": projectID},
	}
	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_project_reactivate: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_project_reactivate: audit REPLY_SENT: %w", err)
	}
	return nil
}
