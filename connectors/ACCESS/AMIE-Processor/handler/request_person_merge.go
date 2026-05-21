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
	"log/slog"

	"github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/model"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type RequestPersonMergeHandler struct {
	svc        *service.Service
	amieClient AmieClient
	auditSvc   AuditService
}

func NewRequestPersonMergeHandler(svc *service.Service, amieClient AmieClient, auditSvc AuditService) *RequestPersonMergeHandler {
	return &RequestPersonMergeHandler{svc: svc, amieClient: amieClient, auditSvc: auditSvc}
}

func (h *RequestPersonMergeHandler) SupportsType() string { return "request_person_merge" }

func (h *RequestPersonMergeHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}
	keepGlobalID := getString(body, "KeepGlobalID")
	if err := requireText(keepGlobalID, "KeepGlobalID"); err != nil {
		return err
	}
	deleteGlobalID := getString(body, "DeleteGlobalID")
	if err := requireText(deleteGlobalID, "DeleteGlobalID"); err != nil {
		return err
	}
	if keepGlobalID == deleteGlobalID {
		return fmt.Errorf("request_person_merge: keep and delete global IDs are identical")
	}

	survivor, err := h.svc.GetUserByExternalIdentity(ctx, amieIdentitySource, keepGlobalID)
	if err != nil {
		return fmt.Errorf("request_person_merge: resolve surviving user: %w", err)
	}
	retiring, err := h.svc.GetUserByExternalIdentity(ctx, amieIdentitySource, deleteGlobalID)
	if err != nil {
		return fmt.Errorf("request_person_merge: resolve retiring user: %w", err)
	}

	memberships, err := h.svc.ListAllocationsForUser(ctx, retiring.ID)
	if err != nil {
		return fmt.Errorf("request_person_merge: list memberships: %w", err)
	}
	activeMemberships := make([]string, 0, len(memberships))
	for _, m := range memberships {
		if m.MembershipStatus == models.ACTIVE {
			activeMemberships = append(activeMemberships, m.ID)
		}
	}
	if len(activeMemberships) > 0 {
		slog.WarnContext(ctx, "merging user with active memberships",
			"keep_global_id", keepGlobalID,
			"delete_global_id", deleteGlobalID,
			"retiring_id", retiring.ID,
			"active_membership_ids", activeMemberships,
		)
	}

	slog.InfoContext(ctx, "executing user merge",
		"keep_global_id", keepGlobalID,
		"delete_global_id", deleteGlobalID,
		"survivor_id", survivor.ID,
		"retiring_id", retiring.ID,
	)

	reason := getString(body, "MergeReason")
	if reason == "" {
		reason = fmt.Sprintf("AMIE request_person_merge packet %d", packet.AmieID)
	}
	if _, err := h.svc.MergeUsers(ctx, survivor.ID, retiring.ID, reason); err != nil {
		return fmt.Errorf("request_person_merge: merge: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditMergePersons, "user", survivor.ID, fmt.Sprintf("merged %s into %s", retiring.ID, survivor.ID)); err != nil {
		return fmt.Errorf("request_person_merge: audit MERGE_PERSONS: %w", err)
	}

	reply := map[string]any{
		"type": "inform_transaction_complete",
		"body": map[string]any{
			"StatusCode": "Success",
			"DetailCode": float64(1),
			"Message":    "Person merge completed.",
		},
	}
	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("request_person_merge: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("request_person_merge: audit REPLY_SENT: %w", err)
	}
	return nil
}
