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
	"time"

	"github.com/apache/airavata-custos/signer/internal/cert"
	"github.com/apache/airavata-custos/signer/internal/httputil"
	vaultpkg "github.com/apache/airavata-custos/signer/internal/vault"
)

type RotateCAResponse struct {
	Success              bool   `json:"success"`
	CurrentCAFingerprint string `json:"current_ca_fingerprint"`
	NextCAFingerprint    string `json:"next_ca_fingerprint"`
	RotationTimestamp    int64  `json:"rotation_timestamp"`
}

type AdminHandler struct {
	vaultClient *vaultpkg.Client
	logger      *slog.Logger
}

func NewAdminHandler(vaultClient *vaultpkg.Client, logger *slog.Logger) *AdminHandler {
	return &AdminHandler{
		vaultClient: vaultClient,
		logger:      logger,
	}
}

func (h *AdminHandler) Handle(w http.ResponseWriter, r *http.Request) {
	clientCfg := httputil.ClientConfigFromContext(r.Context())
	if clientCfg == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Missing client config")
		return
	}

	tenantID := clientCfg.TenantID
	clientID := clientCfg.ClientID

	_, _, err := h.vaultClient.RotateCA(r.Context(), tenantID, clientID)
	if err != nil {
		h.logger.Error("failed to rotate CA", "error", err, "tenant_id", tenantID, "client_id", clientID)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to rotate CA keys")
		return
	}

	currentKey, err := h.vaultClient.GetCurrentCAKey(r.Context(), tenantID, clientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to read rotated keys")
		return
	}
	currentFP, _ := cert.CAFingerprint(currentKey.PublicKey)

	nextKey, err := h.vaultClient.GetNextCAKey(r.Context(), tenantID, clientID)
	nextFP := ""
	if err == nil && nextKey != nil {
		nextFP, _ = cert.CAFingerprint(nextKey.PublicKey)
	}

	resp := RotateCAResponse{
		Success:              true,
		CurrentCAFingerprint: currentFP,
		NextCAFingerprint:    nextFP,
		RotationTimestamp:    time.Now().UTC().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
