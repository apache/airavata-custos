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

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/pkg/service"
)

type RequestAccountReactivateHandler struct {
	svc        *service.Service
	amieClient AmieClient
	auditSvc   AuditService
}

func NewRequestAccountReactivateHandler(svc *service.Service, amieClient AmieClient, auditSvc AuditService) *RequestAccountReactivateHandler {
	return &RequestAccountReactivateHandler{svc: svc, amieClient: amieClient, auditSvc: auditSvc}
}

func (h *RequestAccountReactivateHandler) SupportsType() string { return "request_account_reactivate" }

func (h *RequestAccountReactivateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}
	projectID := getString(body, "ProjectID")
	if err := requireText(projectID, "ProjectID"); err != nil {
		return err
	}
	personID := getString(body, "PersonID")
	if err := requireText(personID, "PersonID"); err != nil {
		return err
	}

	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReactivateMembership, "membership_request", personID, fmt.Sprintf("project=%s person=%s (membership persistence pending allocation mapping)", projectID, personID)); err != nil {
		return fmt.Errorf("request_account_reactivate: audit REACTIVATE_MEMBERSHIP: %w", err)
	}

	reply := map[string]any{
		"type": "notify_account_reactivate",
		"body": map[string]any{"ProjectID": projectID, "PersonID": personID},
	}
	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_account_reactivate: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_account_reactivate: audit REPLY_SENT: %w", err)
	}
	return nil
}
