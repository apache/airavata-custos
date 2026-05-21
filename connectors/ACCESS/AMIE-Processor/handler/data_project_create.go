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
	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/store"
	"github.com/apache/airavata-custos/pkg/service"
)

type DataProjectCreateHandler struct {
	svc         *service.Service
	userDNStore store.UserDNStore
	amieClient  AmieClient
	auditSvc    AuditService
}

func NewDataProjectCreateHandler(svc *service.Service, userDNStore store.UserDNStore, amieClient AmieClient, auditSvc AuditService) *DataProjectCreateHandler {
	return &DataProjectCreateHandler{svc: svc, userDNStore: userDNStore, amieClient: amieClient, auditSvc: auditSvc}
}

func (h *DataProjectCreateHandler) SupportsType() string { return "data_project_create" }

func (h *DataProjectCreateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}
	if err := requireText(getString(body, "ProjectID"), "ProjectID"); err != nil {
		return err
	}
	piGlobalID := getString(body, "GlobalID")
	if err := requireText(piGlobalID, "GlobalID"); err != nil {
		return err
	}

	user, err := h.svc.GetUserByUserIdentity(ctx, amieIdentitySource, piGlobalID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			slog.WarnContext(ctx, "data_project_create: PI user not found; skipping DN persistence and UserIdentity upsert",
				"piGlobalID", piGlobalID)
		} else {
			return fmt.Errorf("data_project_create: resolve PI user: %w", err)
		}
	}
	if user != nil {
		if err := ensureUserIdentity(ctx, h.svc, user.ID, piGlobalID); err != nil {
			return fmt.Errorf("data_project_create: ensure user identity: %w", err)
		}
		dns := getDNList(body)
		if len(dns) > 0 {
			for _, dn := range dns {
				if err := h.userDNStore.Add(ctx, tx, &model.UserDN{UserID: user.ID, DN: dn}); err != nil {
					return fmt.Errorf("data_project_create: add DN %q: %w", dn, err)
				}
			}
			if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditPersistDNs, "user", user.ID, fmt.Sprintf("Persisted %d PI DNs", len(dns))); err != nil {
				return fmt.Errorf("data_project_create: audit PERSIST_DNS: %w", err)
			}
		}
	}

	reply := map[string]any{
		"type": "inform_transaction_complete",
		"body": map[string]any{
			"StatusCode": "Success",
			"DetailCode": float64(1),
			"Message":    "Transaction completed successfully by handler.",
		},
	}
	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("data_project_create: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("data_project_create: audit REPLY_SENT: %w", err)
	}
	return nil
}
