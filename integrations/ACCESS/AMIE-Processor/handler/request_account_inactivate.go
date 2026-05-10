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

type requestAccountInactivateMembershipService interface {
	InactivateMembershipsByPersonAndProject(ctx context.Context, tx *sql.Tx, projectID, personID string) (int, error)
}

type requestAccountInactivateAmieClient interface {
	ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error
}

type requestAccountInactivateAuditService interface {
	Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error
}

type RequestAccountInactivateHandler struct {
	membershipSvc requestAccountInactivateMembershipService
	amieClient    requestAccountInactivateAmieClient
	auditSvc      requestAccountInactivateAuditService
}

func NewRequestAccountInactivateHandler(
	membershipSvc requestAccountInactivateMembershipService,
	amieClient requestAccountInactivateAmieClient,
	auditSvc requestAccountInactivateAuditService,
) *RequestAccountInactivateHandler {
	return &RequestAccountInactivateHandler{
		membershipSvc: membershipSvc,
		amieClient:    amieClient,
		auditSvc:      auditSvc,
	}
}

func (h *RequestAccountInactivateHandler) SupportsType() string {
	return "request_account_inactivate"
}

func (h *RequestAccountInactivateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}

	// Validate required fields.
	projectID := getString(body, "ProjectID")
	if err := requireText(projectID, "ProjectID"); err != nil {
		return err
	}
	personID := getString(body, "PersonID")
	if err := requireText(personID, "PersonID"); err != nil {
		return err
	}

	// Inactivate memberships.
	if _, err := h.membershipSvc.InactivateMembershipsByPersonAndProject(ctx, tx, projectID, personID); err != nil {
		return fmt.Errorf("request_account_inactivate: inactivating memberships: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditInactivateMembership, "membership", "", ""); err != nil {
		return fmt.Errorf("request_account_inactivate: audit INACTIVATE_MEMBERSHIP: %w", err)
	}

	// Build and send the reply.
	reply := map[string]any{
		"type": "notify_account_inactivate",
		"body": map[string]any{
			"ProjectID":    projectID,
			"PersonID":     personID,
			"ResourceList": getResourceList(body),
		},
	}

	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_account_inactivate: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_account_inactivate: audit REPLY_SENT: %w", err)
	}

	return nil
}
