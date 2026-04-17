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

type dataAccountCreatePersonService interface {
	PersistDNsForPerson(ctx context.Context, tx *sql.Tx, personID string, dnList []string) error
}

type dataAccountCreateAmieClient interface {
	ReplyToPacket(ctx context.Context, packetRecID int64, reply map[string]any) error
}

type dataAccountCreateAuditService interface {
	Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error
}

type DataAccountCreateHandler struct {
	personSvc  dataAccountCreatePersonService
	amieClient dataAccountCreateAmieClient
	auditSvc   dataAccountCreateAuditService
}

func NewDataAccountCreateHandler(
	personSvc dataAccountCreatePersonService,
	amieClient dataAccountCreateAmieClient,
	auditSvc dataAccountCreateAuditService,
) *DataAccountCreateHandler {
	return &DataAccountCreateHandler{
		personSvc:  personSvc,
		amieClient: amieClient,
		auditSvc:   auditSvc,
	}
}

func (h *DataAccountCreateHandler) SupportsType() string {
	return "data_account_create"
}

func (h *DataAccountCreateHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}

	// Validate required fields.
	if err := requireText(getString(body, "ProjectID"), "ProjectID"); err != nil {
		return err
	}
	personID := getString(body, "PersonID")
	if err := requireText(personID, "PersonID"); err != nil {
		return err
	}

	// Persist DNs if present.
	if rawDNs, ok := body["DnList"]; ok {
		if arr, ok := rawDNs.([]any); ok && len(arr) > 0 {
			var dnStrings []string
			for _, item := range arr {
				if s, ok := item.(string); ok {
					dnStrings = append(dnStrings, s)
				}
			}
			if len(dnStrings) > 0 {
				if err := h.personSvc.PersistDNsForPerson(ctx, tx, personID, dnStrings); err != nil {
					return fmt.Errorf("data_account_create: persisting DNs: %w", err)
				}
				summary := fmt.Sprintf("Persisted %d DNs", len(dnStrings))
				if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditPersistDNs, "person", personID, summary); err != nil {
					return fmt.Errorf("data_account_create: audit PERSIST_DNS: %w", err)
				}
			}
		}
	}

	// Build and send the reply.
	reply := map[string]any{
		"type": "inform_transaction_complete",
		"body": map[string]any{
			"StatusCode": "Success",
			"DetailCode": float64(1),
			"Message":    "Transaction completed successfully by handler.",
		},
	}

	if err := h.amieClient.ReplyToPacket(ctx, packet.AmieID, reply); err != nil {
		return fmt.Errorf("data_account_create: sending reply: %w", err)
	}
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditReplySent, "reply", "", ""); err != nil {
		return fmt.Errorf("data_account_create: audit REPLY_SENT: %w", err)
	}

	return nil
}
