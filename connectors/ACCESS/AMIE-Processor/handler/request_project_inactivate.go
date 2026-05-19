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

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
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

func (h *RequestProjectInactivateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}
	originatedID := getString(body, "ProjectID")
	if err := requireText(originatedID, "ProjectID"); err != nil {
		return err
	}

	project, err := h.svc.GetProjectByOriginatedID(ctx, originatedID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			slog.WarnContext(ctx, "request_project_inactivate: project not found in core; skipping status flip",
				"originatedID", originatedID)
		} else {
			return fmt.Errorf("request_project_inactivate: lookup project: %w", err)
		}
	} else {
		if _, err := h.svc.UpdateProjectStatus(ctx, project.ID, models.INACTIVE); err != nil {
			return fmt.Errorf("request_project_inactivate: update status: %w", err)
		}
		if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditInactivateProject, "project", project.ID, ""); err != nil {
			return fmt.Errorf("request_project_inactivate: audit INACTIVATE_PROJECT: %w", err)
		}
	}

	reply := map[string]any{
		"type": "notify_project_inactivate",
		"body": map[string]any{"ProjectID": originatedID},
	}
	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_project_inactivate: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_project_inactivate: audit REPLY_SENT: %w", err)
	}
	return nil
}
