// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/apache/airavata-custos/signer/internal/audit"
	"github.com/apache/airavata-custos/signer/internal/httputil"
	"github.com/apache/airavata-custos/signer/internal/metrics"
)

type RevokeRequest struct {
	SerialNumber  *int64  `json:"serial_number,omitempty"`
	KeyID         *string `json:"key_id,omitempty"`
	CAFingerprint *string `json:"ca_fingerprint,omitempty"`
	Reason        string  `json:"reason"`
}

type RevokeResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	RevokedCount int    `json:"revoked_count"`
}

type RevokeHandler struct {
	auditLogger *audit.Logger
	logger      *slog.Logger
}

func NewRevokeHandler(auditLogger *audit.Logger, logger *slog.Logger) *RevokeHandler {
	return &RevokeHandler{
		auditLogger: auditLogger,
		logger:      logger,
	}
}

func (h *RevokeHandler) Handle(w http.ResponseWriter, r *http.Request) {
	clientCfg := httputil.ClientConfigFromContext(r.Context())
	if clientCfg == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Missing client config")
		return
	}

	tenantID := clientCfg.TenantID
	clientID := clientCfg.ClientID

	var req RevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		metrics.RevokeRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.SerialNumber == nil && req.KeyID == nil && req.CAFingerprint == nil {
		metrics.RevokeRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "At least one of serial_number, key_id, or ca_fingerprint is required")
		return
	}

	if req.Reason == "" {
		metrics.RevokeRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing required field: reason")
		return
	}

	entry := &audit.RevocationEntry{
		TenantID:      tenantID,
		ClientID:      clientID,
		SerialNumber:  req.SerialNumber,
		KeyID:         req.KeyID,
		CAFingerprint: req.CAFingerprint,
		Reason:        req.Reason,
		RevokedBy:     tenantID + ":" + clientID,
	}

	if err := h.auditLogger.LogRevocation(r.Context(), entry); err != nil {
		metrics.RevokeRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		h.logger.Error("failed to record revocation", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to record revocation")
		return
	}

	metrics.RevokeRequestsTotal.WithLabelValues(tenantID, "success").Inc()

	resp := RevokeResponse{
		Success:      true,
		Message:      "Certificate(s) revoked successfully",
		RevokedCount: 1,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
