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
	"strings"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/pkg/service"
)

type RequestUserModifyHandler struct {
	svc        *service.Service
	amieClient AmieClient
	auditSvc   AuditService
}

func NewRequestUserModifyHandler(svc *service.Service, amieClient AmieClient, auditSvc AuditService) *RequestUserModifyHandler {
	return &RequestUserModifyHandler{svc: svc, amieClient: amieClient, auditSvc: auditSvc}
}

func (h *RequestUserModifyHandler) SupportsType() string { return "request_user_modify" }

func (h *RequestUserModifyHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}
	actionType := getString(body, "ActionType")
	if err := requireText(actionType, "ActionType"); err != nil {
		return err
	}
	userGlobalID := getString(body, "UserGlobalID")
	if err := requireText(userGlobalID, "UserGlobalID"); err != nil {
		return err
	}

	user, err := h.svc.GetUserByExternalIdentity(ctx, amieIdentitySource, userGlobalID)
	if err != nil && !errors.Is(err, service.ErrNotFound) {
		return fmt.Errorf("request_user_modify: resolve user: %w", err)
	}
	if errors.Is(err, service.ErrNotFound) {
		user = nil
	}

	switch {
	case strings.EqualFold(actionType, "replace"):
		if user != nil {
			if v := getString(body, "UserFirstName"); v != "" {
				user.FirstName = v
			}
			if v := getString(body, "UserLastName"); v != "" {
				user.LastName = v
			}
			if v := getString(body, "UserEmail"); v != "" {
				user.Email = v
			}
			if err := h.svc.UpdateUser(ctx, user); err != nil {
				return fmt.Errorf("request_user_modify: update user: %w", err)
			}
			if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditUpdatePerson, "user", user.ID, ""); err != nil {
				return fmt.Errorf("request_user_modify: audit UPDATE_PERSON: %w", err)
			}
		}
	case strings.EqualFold(actionType, "delete"):
		if user != nil {
			if err := h.svc.DeleteUser(ctx, user.ID); err != nil {
				return fmt.Errorf("request_user_modify: delete user: %w", err)
			}
			if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditDeletePerson, "user", user.ID, ""); err != nil {
				return fmt.Errorf("request_user_modify: audit DELETE_PERSON: %w", err)
			}
		}
	default:
		return fmt.Errorf("unsupported ActionType: %s", actionType)
	}

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
