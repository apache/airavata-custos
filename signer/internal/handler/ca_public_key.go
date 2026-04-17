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
	"strings"

	"github.com/apache/airavata-custos/signer/internal/cert"
	"github.com/apache/airavata-custos/signer/internal/httputil"
	vaultpkg "github.com/apache/airavata-custos/signer/internal/vault"
)

type CAPublicKeyJSONResponse struct {
	PublicKey   string `json:"public_key"`
	Fingerprint string `json:"fingerprint"`
	Algorithm   string `json:"algorithm"`
}

type CAPublicKeyHandler struct {
	vaultClient *vaultpkg.Client
	logger      *slog.Logger
}

func NewCAPublicKeyHandler(vaultClient *vaultpkg.Client, logger *slog.Logger) *CAPublicKeyHandler {
	return &CAPublicKeyHandler{
		vaultClient: vaultClient,
		logger:      logger,
	}
}

func (h *CAPublicKeyHandler) Handle(w http.ResponseWriter, r *http.Request) {
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
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve CA key")
		return
	}

	opensshKey, err := cert.PEMToOpenSSH(currentKey.PublicKey)
	if err != nil {
		h.logger.Error("failed to convert CA key to OpenSSH format", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to convert CA key")
		return
	}

	// If the client accepts JSON, return structured response
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		fingerprint, _ := cert.CAFingerprint(currentKey.PublicKey)
		algorithm := cert.NormalizeKeyType(strings.SplitN(opensshKey, " ", 2)[0])

		resp := CAPublicKeyJSONResponse{
			PublicKey:   opensshKey,
			Fingerprint: fingerprint,
			Algorithm:   algorithm,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Default: plain text for direct piping to trusted-user-ca-keys
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(opensshKey + "\n"))
}
