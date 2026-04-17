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

type requestProjectInactivateProjectService interface {
	InactivateProject(ctx context.Context, tx *sql.Tx, projectID string) error
}

type requestProjectInactivateMembershipService interface {
	InactivateAllForProject(ctx context.Context, tx *sql.Tx, projectID string) error
}

type requestProjectInactivateAmieClient interface {
	ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error
}

type requestProjectInactivateAuditService interface {
	Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error
}

type RequestProjectInactivateHandler struct {
	projectSvc    requestProjectInactivateProjectService
	membershipSvc requestProjectInactivateMembershipService
	amieClient    requestProjectInactivateAmieClient
	auditSvc      requestProjectInactivateAuditService
}

func NewRequestProjectInactivateHandler(
	projectSvc requestProjectInactivateProjectService,
	membershipSvc requestProjectInactivateMembershipService,
	amieClient requestProjectInactivateAmieClient,
	auditSvc requestProjectInactivateAuditService,
) *RequestProjectInactivateHandler {
	return &RequestProjectInactivateHandler{
		projectSvc:    projectSvc,
		membershipSvc: membershipSvc,
		amieClient:    amieClient,
		auditSvc:      auditSvc,
	}
}

func (h *RequestProjectInactivateHandler) SupportsType() string {
	return "request_project_inactivate"
}

func (h *RequestProjectInactivateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}

	// Validate required fields.
	projectID := getString(body, "ProjectID")
	if err := requireText(projectID, "ProjectID"); err != nil {
		return err
	}

	// Inactivate the project.
	if err := h.projectSvc.InactivateProject(ctx, tx, projectID); err != nil {
		return fmt.Errorf("request_project_inactivate: inactivating project: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditInactivateProject, "project", projectID, ""); err != nil {
		return fmt.Errorf("request_project_inactivate: audit INACTIVATE_PROJECT: %w", err)
	}

	// Inactivate all memberships for the project.
	if err := h.membershipSvc.InactivateAllForProject(ctx, tx, projectID); err != nil {
		return fmt.Errorf("request_project_inactivate: inactivating memberships: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditInactivateMembership, "membership", projectID, ""); err != nil {
		return fmt.Errorf("request_project_inactivate: audit INACTIVATE_MEMBERSHIP: %w", err)
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
		"type": "notify_project_inactivate",
		"body": replyBody,
	}

	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_project_inactivate: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_project_inactivate: audit REPLY_SENT: %w", err)
	}

	return nil
}
