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

type RequestProjectInactivateHandler struct {
	svc        *service.Service
	amieClient AmieClient
	auditSvc   AuditService
}

func NewRequestProjectInactivateHandler(svc *service.Service, amieClient AmieClient, auditSvc AuditService) *RequestProjectInactivateHandler {
	return &RequestProjectInactivateHandler{svc: svc, amieClient: amieClient, auditSvc: auditSvc}
}

func (h *RequestProjectInactivateHandler) SupportsType() string { return "request_project_inactivate" }

func (h *RequestProjectInactivateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) (err error) {
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
			slog.WarnContext(ctx, "request_project_inactivate: project not found in core; skipping",
				"projectID", projectID)
			return h.reply(ctx, tx, packet, eventID, projectID)
		}
		return fmt.Errorf("request_project_inactivate: lookup project: %w", err)
	}

	if _, err := h.svc.UpdateProjectStatus(ctx, project.ID, models.ProjectInactive); err != nil {
		return fmt.Errorf("request_project_inactivate: update project status: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditInactivateProject, "project", project.ID, ""); err != nil {
		return fmt.Errorf("request_project_inactivate: audit INACTIVATE_PROJECT: %w", err)
	}

	if err := h.deactivateAllocations(ctx, tx, packet, eventID, project.ID, getString(body, "Comment")); err != nil {
		return err
	}

	return h.reply(ctx, tx, packet, eventID, projectID)
}

// deactivateAllocations flips every ComputeAllocation under the project to
// INACTIVE, writes a status-change Diff per allocation, and flips every
// ComputeAllocationMembership under those allocations to INACTIVE.
func (h *RequestProjectInactivateHandler) deactivateAllocations(ctx context.Context, tx *sql.Tx, packet *model.Packet, eventID, projectID, comment string) error {
	allocations, err := h.svc.ListComputeAllocationsByProject(ctx, projectID)
	if err != nil {
		return fmt.Errorf("list allocations: %w", err)
	}
	for _, a := range allocations {
		a.Status = models.INACTIVE
		if err := h.svc.UpdateComputeAllocation(ctx, &a); err != nil {
			return fmt.Errorf("inactivate allocation %s: %w", a.ID, err)
		}
		if _, err := h.svc.CreateComputeAllocationDiff(ctx, &models.ComputeAllocationDiff{
			ComputeAllocationID: a.ID,
			DiffType:            "ALLOCATION_STATUS_CHANGE",
			Status:              models.INACTIVE,
			Description:         describeStatusChange("Inactivated by AMIE request_project_inactivate", comment),
		}); err != nil {
			return fmt.Errorf("record inactivate diff for %s: %w", a.ID, err)
		}
		members, err := h.svc.ListMembersForAllocation(ctx, a.ID)
		if err != nil {
			return fmt.Errorf("list memberships for allocation %s: %w", a.ID, err)
		}
		for _, m := range members {
			if _, err := h.svc.UpdateMembershipStatus(ctx, m.ID, models.INACTIVE); err != nil {
				return fmt.Errorf("inactivate membership %s: %w", m.ID, err)
			}
			if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditInactivateMembership, "compute_allocation_membership", m.ID,
				fmt.Sprintf("user=%s allocation=%s", m.UserID, a.ID)); err != nil {
				return fmt.Errorf("audit INACTIVATE_MEMBERSHIP for %s: %w", m.ID, err)
			}
		}
	}
	return nil
}

func (h *RequestProjectInactivateHandler) reply(ctx context.Context, tx *sql.Tx, packet *model.Packet, eventID, projectID string) error {
	reply := map[string]any{
		"type": "notify_project_inactivate",
		"body": map[string]any{"ProjectID": projectID},
	}
	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_project_inactivate: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_project_inactivate: audit REPLY_SENT: %w", err)
	}
	return nil
}

func describeStatusChange(primary, comment string) string {
	if comment == "" {
		return primary
	}
	return primary + " (" + comment + ")"
}
