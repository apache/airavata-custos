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
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/apache/airavata-custos/signer/internal/audit"
	"github.com/apache/airavata-custos/signer/internal/auth"
	"github.com/apache/airavata-custos/signer/internal/cert"
	"github.com/apache/airavata-custos/signer/internal/httputil"
	"github.com/apache/airavata-custos/signer/internal/metrics"
	"github.com/apache/airavata-custos/signer/internal/policy"
	"github.com/apache/airavata-custos/signer/internal/validation"
	vaultpkg "github.com/apache/airavata-custos/signer/internal/vault"
)

var principalRegex = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)

type SignRequest struct {
	Principal       string `json:"principal"`
	TTLSeconds      int    `json:"ttl_seconds"`
	PublicKey       string `json:"public_key"`
	UserAccessToken string `json:"user_access_token"`
}

type SignResponse struct {
	Certificate    string `json:"certificate"`
	SerialNumber   int64  `json:"serial_number"`
	ValidAfter     int64  `json:"valid_after"`
	ValidBefore    int64  `json:"valid_before"`
	CAFingerprint  string `json:"ca_fingerprint"`
	TargetHost     string `json:"target_host"`
	TargetPort     int    `json:"target_port"`
	TargetUsername string `json:"target_username"`
}

type SignHandler struct {
	oidcValidator      *auth.OIDCValidator
	policyEnforcer     *policy.Enforcer
	principalValidator validation.PrincipalValidator
	vaultClient        *vaultpkg.Client
	auditLogger        *audit.Logger
	logger             *slog.Logger
}

func NewSignHandler(
	oidcValidator *auth.OIDCValidator,
	policyEnforcer *policy.Enforcer,
	principalValidator validation.PrincipalValidator,
	vaultClient *vaultpkg.Client,
	auditLogger *audit.Logger,
	logger *slog.Logger,
) *SignHandler {
	return &SignHandler{
		oidcValidator:      oidcValidator,
		policyEnforcer:     policyEnforcer,
		principalValidator: principalValidator,
		vaultClient:        vaultClient,
		auditLogger:        auditLogger,
		logger:             logger,
	}
}

func (h *SignHandler) Handle(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	clientCfg := httputil.ClientConfigFromContext(r.Context())
	if clientCfg == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Missing client config")
		return
	}

	tenantID := clientCfg.TenantID
	clientID := clientCfg.ClientID

	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.SignDurationSeconds.WithLabelValues(tenantID).Observe(duration)
	}()

	var req SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Principal == "" {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing required field: principal")
		return
	}
	if !principalRegex.MatchString(req.Principal) {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid principal format: must match ^[a-z_][a-z0-9_-]{0,31}$")
		return
	}

	if req.PublicKey == "" {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing required field: public_key")
		return
	}
	if req.UserAccessToken == "" {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing required field: user_access_token")
		return
	}

	sshPubKey, err := cert.ParseSSHPublicKey(req.PublicKey)
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid SSH public key format")
		return
	}

	identity, err := h.oidcValidator.ValidateAccessToken(r.Context(), req.UserAccessToken)
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		if authErr, ok := err.(*auth.AuthError); ok {
			writeError(w, authErr.Status, authErr.Code, authErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Token validation error")
		return
	}

	sourceIP := httputil.SourceIPFromContext(r.Context())
	if err := h.policyEnforcer.Enforce(req.TTLSeconds, sshPubKey.Type(), sourceIP, clientCfg); err != nil {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		if policyErr, ok := err.(*policy.PolicyError); ok {
			writeError(w, http.StatusBadRequest, "policy_violation", policyErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Policy enforcement error")
		return
	}

	valResult, err := h.principalValidator.Validate(tenantID, clientID, req.Principal, identity.Subject)
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		if valErr, ok := err.(*validation.ValidationError); ok {
			httputil.WriteJSONErrorWithExtra(w, http.StatusForbidden, "principal_denied", valErr.Message, map[string]string{
				"reason_code": valErr.ReasonCode,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Principal validation error")
		return
	}

	validatedPrincipal := valResult.ValidatedPrincipal

	caKeyPair, err := h.vaultClient.GetCurrentCAKey(r.Context(), tenantID, clientID)
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		metrics.VaultOperationsTotal.WithLabelValues("get_ca_key", "error").Inc()
		h.logger.Error("failed to get CA key", "error", err, "tenant_id", tenantID, "client_id", clientID)
		writeError(w, http.StatusServiceUnavailable, "internal_error", "CA key storage unavailable")
		return
	}
	metrics.VaultOperationsTotal.WithLabelValues("get_ca_key", "success").Inc()

	serial, err := h.vaultClient.IncrementSerialCounter(r.Context(), tenantID, clientID)
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		metrics.VaultOperationsTotal.WithLabelValues("increment_serial", "error").Inc()
		h.logger.Error("failed to increment serial counter", "error", err, "tenant_id", tenantID, "client_id", clientID)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to allocate serial number")
		return
	}
	metrics.VaultOperationsTotal.WithLabelValues("increment_serial", "success").Inc()

	criticalOpts := policy.GetCriticalOptions(clientCfg)

	signReq := &cert.SignRequest{
		PublicKey:       sshPubKey,
		CAPrivateKeyPEM: caKeyPair.PrivateKey,
		Serial:          uint64(serial),
		Principal:       validatedPrincipal,
		ClientID:        clientID,
		TTLSeconds:      uint64(req.TTLSeconds),
		CriticalOptions: criticalOpts,
	}

	signResult, err := cert.SignCertificate(signReq)
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
		h.logger.Error("failed to sign certificate", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to sign certificate")
		return
	}

	tokenHash := cert.HashToken(req.UserAccessToken)
	pubKeyFP := cert.SSHPublicKeyFingerprint(sshPubKey)

	auditEntry := &audit.IssuanceEntry{
		TenantID:             tenantID,
		ClientID:             clientID,
		SerialNumber:         serial,
		KeyID:                signResult.KeyID,
		Principal:            validatedPrincipal,
		UserEmail:            identity.Email,
		PublicKeyFingerprint: pubKeyFP,
		CAFingerprint:        signResult.CAFingerprint,
		ValidAfter:           time.Unix(int64(signResult.ValidAfter), 0).UTC(),
		ValidBefore:          time.Unix(int64(signResult.ValidBefore), 0).UTC(),
		SourceIP:             sourceIP,
		UserAccessTokenHash:  tokenHash,
	}

	if err := h.auditLogger.LogIssuance(r.Context(), auditEntry); err != nil {
		if !isDuplicateKeyError(err) {
			metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
			h.logger.Error("failed to write audit log", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to record issuance")
			return
		}

		// Duplicate serial — retry with a fresh serial
		h.logger.Warn("duplicate serial detected, retrying with fresh serial",
			"serial", serial, "tenant_id", tenantID, "client_id", clientID)

		retrySerial, err := h.vaultClient.IncrementSerialCounter(r.Context(), tenantID, clientID)
		if err != nil {
			metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
			metrics.VaultOperationsTotal.WithLabelValues("increment_serial", "error").Inc()
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to allocate serial number")
			return
		}
		metrics.VaultOperationsTotal.WithLabelValues("increment_serial", "success").Inc()

		signReq.Serial = uint64(retrySerial)
		retryResult, err := cert.SignCertificate(signReq)
		if err != nil {
			metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
			h.logger.Error("failed to sign certificate on retry", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to sign certificate")
			return
		}

		serial = retrySerial
		signResult = retryResult
		auditEntry.SerialNumber = retrySerial
		auditEntry.KeyID = retryResult.KeyID
		auditEntry.CAFingerprint = retryResult.CAFingerprint
		auditEntry.ValidAfter = time.Unix(int64(retryResult.ValidAfter), 0).UTC()
		auditEntry.ValidBefore = time.Unix(int64(retryResult.ValidBefore), 0).UTC()

		if err := h.auditLogger.LogIssuance(r.Context(), auditEntry); err != nil {
			metrics.SignRequestsTotal.WithLabelValues(tenantID, "error").Inc()
			h.logger.Error("failed to write audit log on retry", "error", err)
			writeError(w, http.StatusInternalServerError, "internal_error", "Failed to record issuance")
			return
		}
	}

	metrics.SignRequestsTotal.WithLabelValues(tenantID, "success").Inc()

	resp := SignResponse{
		Certificate:    base64.StdEncoding.EncodeToString(signResult.CertBytes),
		SerialNumber:   serial,
		ValidAfter:     int64(signResult.ValidAfter),
		ValidBefore:    int64(signResult.ValidBefore),
		CAFingerprint:  signResult.CAFingerprint,
		TargetHost:     clientCfg.TargetHost,
		TargetPort:     clientCfg.TargetPort,
		TargetUsername: validatedPrincipal,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// isDuplicateKeyError detects MySQL error 1062 (duplicate key constraint).
func isDuplicateKeyError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Duplicate entry")
}
