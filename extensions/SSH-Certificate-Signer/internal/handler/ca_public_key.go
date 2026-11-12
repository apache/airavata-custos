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

	"golang.org/x/crypto/ssh"

	"github.com/apache/airavata-custos/signer/internal/cert"
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
	currentKey, err := h.vaultClient.GetCurrentCAKey(r.Context())
	if err != nil {
		h.logger.Error("failed to get current CA key", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to retrieve CA key")
		return
	}

	caPublicKey, err := cert.ParseCAPublicKey(currentKey.PublicKey)
	if err != nil {
		h.logger.Error("failed to parse CA public key", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to read CA key")
		return
	}
	opensshKey := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(caPublicKey)))

	// If the client accepts JSON, return structured response
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		resp := CAPublicKeyJSONResponse{
			PublicKey:   opensshKey,
			Fingerprint: ssh.FingerprintSHA256(caPublicKey),
			Algorithm:   cert.NormalizeKeyType(caPublicKey.Type()),
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
