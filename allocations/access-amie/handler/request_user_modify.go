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
	"strings"

	"github.com/apache/airavata-custos/allocations/access-amie/model"
)

type requestUserModifyPersonService interface {
	ReplaceFromModifyPacket(ctx context.Context, tx *sql.Tx, body map[string]any) error
	DeleteFromModifyPacket(ctx context.Context, tx *sql.Tx, body map[string]any) error
}

type requestUserModifyAmieClient interface {
	ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error
}

type requestUserModifyAuditService interface {
	Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error
}

type RequestUserModifyHandler struct {
	personSvc  requestUserModifyPersonService
	amieClient requestUserModifyAmieClient
	auditSvc   requestUserModifyAuditService
}

func NewRequestUserModifyHandler(
	personSvc requestUserModifyPersonService,
	amieClient requestUserModifyAmieClient,
	auditSvc requestUserModifyAuditService,
) *RequestUserModifyHandler {
	return &RequestUserModifyHandler{
		personSvc:  personSvc,
		amieClient: amieClient,
		auditSvc:   auditSvc,
	}
}

func (h *RequestUserModifyHandler) SupportsType() string {
	return "request_user_modify"
}

func (h *RequestUserModifyHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}

	// Extract and validate ActionType.
	actionType := getString(body, "ActionType")
	if err := requireText(actionType, "ActionType"); err != nil {
		return err
	}

	if strings.EqualFold(actionType, "replace") {
		if err := h.personSvc.ReplaceFromModifyPacket(ctx, tx, body); err != nil {
			return fmt.Errorf("request_user_modify: replacing person fields: %w", err)
		}
		if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditUpdatePerson, "person", "", ""); err != nil {
			return fmt.Errorf("request_user_modify: audit UPDATE_PERSON: %w", err)
		}
	} else if strings.EqualFold(actionType, "delete") {
		if err := h.personSvc.DeleteFromModifyPacket(ctx, tx, body); err != nil {
			return fmt.Errorf("request_user_modify: deleting person: %w", err)
		}
		if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditDeletePerson, "person", "", ""); err != nil {
			return fmt.Errorf("request_user_modify: audit DELETE_PERSON: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported ActionType: %s", actionType)
	}

	// Build and send the reply.
	reply := map[string]any{
		"type": "inform_transaction_complete",
		"body": map[string]any{
			"StatusCode": "Success",
			"DetailCode": float64(1),
			"Message":    "Transaction completed successfully",
		},
	}

	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_user_modify: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_user_modify: audit REPLY_SENT: %w", err)
	}

	return nil
}
