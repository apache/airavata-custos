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

type requestPersonMergePersonService interface {
	MergePersons(ctx context.Context, tx *sql.Tx, survivingID, retiringID string) error
}

type requestPersonMergeAmieClient interface {
	ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error
}

type requestPersonMergeAuditService interface {
	Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error
}

type RequestPersonMergeHandler struct {
	personSvc  requestPersonMergePersonService
	amieClient requestPersonMergeAmieClient
	auditSvc   requestPersonMergeAuditService
}

func NewRequestPersonMergeHandler(
	personSvc requestPersonMergePersonService,
	amieClient requestPersonMergeAmieClient,
	auditSvc requestPersonMergeAuditService,
) *RequestPersonMergeHandler {
	return &RequestPersonMergeHandler{
		personSvc:  personSvc,
		amieClient: amieClient,
		auditSvc:   auditSvc,
	}
}

func (h *RequestPersonMergeHandler) SupportsType() string {
	return "request_person_merge"
}

func (h *RequestPersonMergeHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}

	// Extract required fields.
	keepPersonID := getString(body, "KeepPersonID")
	if err := requireText(keepPersonID, "KeepPersonID"); err != nil {
		return err
	}
	deletePersonID := getString(body, "DeletePersonID")
	if err := requireText(deletePersonID, "DeletePersonID"); err != nil {
		return err
	}

	// Merge persons.
	if err := h.personSvc.MergePersons(ctx, tx, keepPersonID, deletePersonID); err != nil {
		return fmt.Errorf("request_person_merge: merging persons: %w", err)
	}
	summary := fmt.Sprintf("Merged person %s into %s", deletePersonID, keepPersonID)
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditMergePersons, "person", keepPersonID, summary); err != nil {
		return fmt.Errorf("request_person_merge: audit MERGE_PERSONS: %w", err)
	}

	// Build and send the reply.
	reply := map[string]any{
		"type": "inform_transaction_complete",
		"body": map[string]any{
			"StatusCode": "Success",
			"DetailCode": float64(1),
			"Message":    "Person merge transaction completed successfully.",
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
