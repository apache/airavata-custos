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
	"fmt"

	"github.com/apache/airavata-custos/allocations/access-amie/model"
)

type requestProjectReactivateProjectService interface {
	ReactivateProject(ctx context.Context, tx *sql.Tx, projectID string) error
}

type requestProjectReactivateMembershipService interface {
	ReactivatePiMembership(ctx context.Context, tx *sql.Tx, projectID string) error
}

type requestProjectReactivateAmieClient interface {
	ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error
}

type requestProjectReactivateAuditService interface {
	Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error
}

type RequestProjectReactivateHandler struct {
	projectSvc    requestProjectReactivateProjectService
	membershipSvc requestProjectReactivateMembershipService
	amieClient    requestProjectReactivateAmieClient
	auditSvc      requestProjectReactivateAuditService
}

func NewRequestProjectReactivateHandler(
	projectSvc requestProjectReactivateProjectService,
	membershipSvc requestProjectReactivateMembershipService,
	amieClient requestProjectReactivateAmieClient,
	auditSvc requestProjectReactivateAuditService,
) *RequestProjectReactivateHandler {
	return &RequestProjectReactivateHandler{
		projectSvc:    projectSvc,
		membershipSvc: membershipSvc,
		amieClient:    amieClient,
		auditSvc:      auditSvc,
	}
}

func (h *RequestProjectReactivateHandler) SupportsType() string {
	return "request_project_reactivate"
}

func (h *RequestProjectReactivateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}

	// Validate required fields.
	projectID := getString(body, "ProjectID")
	if err := requireText(projectID, "ProjectID"); err != nil {
		return err
	}

	// Reactivate the project.
	if err := h.projectSvc.ReactivateProject(ctx, tx, projectID); err != nil {
		return fmt.Errorf("request_project_reactivate: reactivating project: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReactivateProject, "project", projectID, ""); err != nil {
		return fmt.Errorf("request_project_reactivate: audit REACTIVATE_PROJECT: %w", err)
	}

	// Reactivate PI memberships.
	if err := h.membershipSvc.ReactivatePiMembership(ctx, tx, projectID); err != nil {
		return fmt.Errorf("request_project_reactivate: reactivating PI membership: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReactivateMembership, "membership", projectID, ""); err != nil {
		return fmt.Errorf("request_project_reactivate: audit REACTIVATE_MEMBERSHIP: %w", err)
	}

	// Build and send the reply.
	replyBody := map[string]any{
		"ProjectID":    projectID,
		"ResourceList": getResourceList(body),
	}
	personID := getString(body, "PersonID")
	if personID != "" {
		replyBody["PersonID"] = personID
	}

	reply := map[string]any{
		"type": "notify_project_reactivate",
		"body": replyBody,
	}

	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_project_reactivate: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_project_reactivate: audit REPLY_SENT: %w", err)
	}

	return nil
}
