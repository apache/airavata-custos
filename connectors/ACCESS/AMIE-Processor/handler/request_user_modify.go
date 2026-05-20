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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/pkg/models"
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
			if err := h.applyReplace(ctx, tx, packet, eventID, body, user, userGlobalID); err != nil {
				return err
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

// applyReplace updates the User row (basic profile), the ExternalIdentity row
// (org / orgCode / NSF status carried as metadata), and reconciles the DN list.
func (h *RequestUserModifyHandler) applyReplace(ctx context.Context, tx *sql.Tx, packet *model.Packet, eventID string, body map[string]any, user *models.User, userGlobalID string) error {
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
		return fmt.Errorf("update user: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditUpdatePerson, "user", user.ID, ""); err != nil {
		return fmt.Errorf("audit UPDATE_PERSON: %w", err)
	}

	if err := h.updateExternalIdentity(ctx, body, user.ID, userGlobalID); err != nil {
		return fmt.Errorf("update external identity: %w", err)
	}
	if err := h.syncDNs(ctx, tx, packet, eventID, body, user.ID); err != nil {
		return fmt.Errorf("sync user DNs: %w", err)
	}
	return nil
}

// updateExternalIdentity refreshes the AMIE ExternalIdentity row's metadata
// (organization, org_code, NSF status) — these are AMIE-side attributes that
// may shift over a user's lifetime. The row's source / external_id are
// immutable identifiers.
func (h *RequestUserModifyHandler) updateExternalIdentity(ctx context.Context, body map[string]any, userID, userGlobalID string) error {
	ext, err := h.svc.GetExternalIdentityBySourceAndExternalID(ctx, amieIdentitySource, userGlobalID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil
		}
		return err
	}
	metadata := map[string]any{}
	if v := getString(body, "UserOrganization"); v != "" {
		metadata["organization"] = v
	}
	if v := getString(body, "UserOrgCode"); v != "" {
		metadata["org_code"] = v
	}
	if v := getString(body, "NsfStatusCode"); v != "" {
		metadata["nsf_status_code"] = v
	}
	if len(metadata) == 0 {
		return nil
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("encode metadata: %w", err)
	}
	ext.UserID = userID
	ext.Metadata = string(encoded)
	return h.svc.UpdateExternalIdentity(ctx, ext)
}

// syncDNs reconciles the user's DN list with the packet body's UserDnList:
// new DNs are added, DNs missing from the packet are removed. AMIE's
// request_user_modify with ActionType=replace is the authoritative source.
func (h *RequestUserModifyHandler) syncDNs(ctx context.Context, tx *sql.Tx, packet *model.Packet, eventID string, body map[string]any, userID string) error {
	incoming := getDNList(body)
	if incoming == nil {
		return nil
	}
	desired := make(map[string]struct{}, len(incoming))
	for _, dn := range incoming {
		desired[dn] = struct{}{}
	}
	existing, err := h.svc.ListUserDNs(ctx, userID)
	if err != nil {
		return fmt.Errorf("list user DNs: %w", err)
	}
	have := make(map[string]struct{}, len(existing))
	added := 0
	removed := 0
	for _, e := range existing {
		have[e.DN] = struct{}{}
		if _, keep := desired[e.DN]; !keep {
			if err := h.svc.RemoveUserDN(ctx, e.ID); err != nil {
				return fmt.Errorf("remove DN %s: %w", e.DN, err)
			}
			removed++
		}
	}
	for dn := range desired {
		if _, exists := have[dn]; exists {
			continue
		}
		if _, err := h.svc.AddUserDN(ctx, &models.UserDN{UserID: userID, DN: dn}); err != nil {
			if errors.Is(err, service.ErrAlreadyExists) {
				continue
			}
			return fmt.Errorf("add DN %s: %w", dn, err)
		}
		added++
	}
	if added > 0 || removed > 0 {
		if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditPersistDNs, "user", userID,
			fmt.Sprintf("DN sync: +%d -%d", added, removed)); err != nil {
			return fmt.Errorf("audit PERSIST_DNS: %w", err)
		}
	}
	return nil
}
