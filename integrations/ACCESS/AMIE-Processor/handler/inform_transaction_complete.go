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

	"github.com/apache/airavata-custos/allocations/access-amie/model"
)

type informTransactionCompleteAuditService interface {
	Log(ctx context.Context, tx *sql.Tx, packetID, eventID string, action model.AuditAction, entityType, entityID, summary string) error
}

type InformTransactionCompleteHandler struct {
	auditSvc informTransactionCompleteAuditService
}

func NewInformTransactionCompleteHandler(
	auditSvc informTransactionCompleteAuditService,
) *InformTransactionCompleteHandler {
	return &InformTransactionCompleteHandler{
		auditSvc: auditSvc,
	}
}

func (h *InformTransactionCompleteHandler) SupportsType() string {
	return "inform_transaction_complete"
}

func (h *InformTransactionCompleteHandler) Handle(ctx context.Context, tx *sql.Tx, packetJSON map[string]any, packet *model.Packet, eventID string) error {
	body, err := getBody(packetJSON)
	if err != nil {
		return err
	}

	statusCode := getString(body, "StatusCode")
	if statusCode == "" {
		statusCode = "Unknown"
	}
	message := getString(body, "Message")

	slog.InfoContext(ctx, "transaction complete",
		"packetID", packet.ID,
		"statusCode", statusCode,
		"message", message,
	)

	summary := fmt.Sprintf("Transaction complete: statusCode=%s, message=%s", statusCode, message)
	if err := h.auditSvc.Log(ctx, tx, packet.ID, eventID, model.AuditTransactionComplete, "transaction", "", summary); err != nil {
		return fmt.Errorf("inform_transaction_complete: audit TRANSACTION_COMPLETE: %w", err)
	}

	return nil
}
