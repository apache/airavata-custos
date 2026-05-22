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
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type RequestAccountInactivateHandler struct {
	svc        *service.Service
	amieClient AmieClient
	auditSvc   AuditService
}

func NewRequestAccountInactivateHandler(svc *service.Service, amieClient AmieClient, auditSvc AuditService) *RequestAccountInactivateHandler {
	return &RequestAccountInactivateHandler{svc: svc, amieClient: amieClient, auditSvc: auditSvc}
}

func (h *RequestAccountInactivateHandler) SupportsType() string { return "request_account_inactivate" }

func (h *RequestAccountInactivateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
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

	flipped, err := flipUserMemberships(ctx, h.svc, projectID, personID, models.INACTIVE)
	if err != nil {
		return fmt.Errorf("request_account_inactivate: flip memberships: %w", err)
	}
	for _, m := range flipped {
		if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditInactivateMembership, "compute_allocation_membership", m.ID,
			fmt.Sprintf("user=%s allocation=%s", m.UserID, m.ComputeAllocationID)); err != nil {
			return fmt.Errorf("request_account_inactivate: audit INACTIVATE_MEMBERSHIP: %w", err)
		}
	}

	reply := map[string]any{
		"type": "notify_account_inactivate",
		"body": map[string]any{"ProjectID": projectID, "PersonID": personID},
	}
	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_account_inactivate: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_account_inactivate: audit REPLY_SENT: %w", err)
	}
	return nil
}
