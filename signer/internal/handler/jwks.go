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

type JWKSResponse struct {
	CurrentKey          interface{} `json:"current_key"`
	NextKey             interface{} `json:"next_key"`
	CurrentFingerprint  string      `json:"current_fingerprint"`
	NextFingerprint     string      `json:"next_fingerprint"`
	RotationScheduledAt int64       `json:"rotation_scheduled_at"`
}

type JWKSHandler struct {
	vaultClient *vaultpkg.Client
	logger      *slog.Logger
}

func NewJWKSHandler(vaultClient *vaultpkg.Client, logger *slog.Logger) *JWKSHandler {
	return &JWKSHandler{
		vaultClient: vaultClient,
		logger:      logger,
	}
}

func (h *JWKSHandler) Handle(w http.ResponseWriter, r *http.Request) {
	clientCfg := httputil.ClientConfigFromContext(r.Context())
	if clientCfg == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Missing client config")
		return
	}

	tenantID := clientCfg.TenantID
	clientID := clientCfg.ClientID

	currentKey, err := h.vaultClient.GetCurrentCAKey(r.Context(), tenantID, clientID)
	if err != nil {
		h.logger.Error("failed to get current CA key", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve CA keys")
		return
	}

	currentJWK, err := cert.PublicKeyToJWK(currentKey.PublicKey)
	if err != nil {
		h.logger.Error("failed to convert current key to JWK", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to convert CA key to JWK")
		return
	}

	currentFP, _ := cert.CAFingerprint(currentKey.PublicKey)

	resp := JWKSResponse{
		CurrentKey:         currentJWK,
		NextKey:            "",
		CurrentFingerprint: currentFP,
		NextFingerprint:    "",
	}

	nextKey, err := h.vaultClient.GetNextCAKey(r.Context(), tenantID, clientID)
	if err == nil && nextKey != nil {
		nextJWK, err := cert.PublicKeyToJWK(nextKey.PublicKey)
		if err == nil {
			resp.NextKey = nextJWK
			nextFP, _ := cert.CAFingerprint(nextKey.PublicKey)
			resp.NextFingerprint = nextFP
		}
	}

	meta, err := h.vaultClient.GetMetadata(r.Context(), tenantID, clientID)
	if err == nil && meta != nil && meta.NextRotationAt != "" {
		if t, err := time.Parse(time.RFC3339, meta.NextRotationAt); err == nil {
			resp.RotationScheduledAt = t.Unix()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
