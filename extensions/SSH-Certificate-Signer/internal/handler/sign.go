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
	"net/netip"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/apache/airavata-custos/signer/internal/audit"
	"github.com/apache/airavata-custos/signer/internal/cert"
	"github.com/apache/airavata-custos/signer/internal/httputil"
	"github.com/apache/airavata-custos/signer/internal/metrics"
	"github.com/apache/airavata-custos/signer/internal/policy"
	"github.com/apache/airavata-custos/signer/internal/store"
	"github.com/apache/airavata-custos/signer/internal/validation"
	vaultpkg "github.com/apache/airavata-custos/signer/internal/vault"
)

var principalRegex = regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)

// The user_email, user_sub and requested_by columns are VARCHAR(255). The key ID
// joins two of these with ":", so it always fits its VARCHAR(512) column.
const maxAttributeLength = 255

type SignRequest struct {
	Principal         string   `json:"principal"`
	TTLSeconds        int      `json:"ttl_seconds"`
	PublicKey         string   `json:"public_key"`
	UserEmail         string   `json:"user_email"`
	UserSub           string   `json:"user_sub"`
	RequestedBy       string   `json:"requested_by"`
	ForceCommand      string   `json:"force_command,omitempty"`
	SourceAddress     string   `json:"source_address,omitempty"`
	ExcludeExtensions []string `json:"exclude_extensions,omitempty"`
}

type SignResponse struct {
	Certificate       string           `json:"certificate"`
	SerialNumber      int64            `json:"serial_number"`
	ValidAfter        int64            `json:"valid_after"`
	ValidBefore       int64            `json:"valid_before"`
	CAFingerprint     string           `json:"ca_fingerprint"`
	LoginHost         string           `json:"login_host"`
	LoginPort         int              `json:"login_port"`
	Principal         string           `json:"principal"`
	ForceCommand      string           `json:"force_command,omitempty"`
	GrantedExtensions []string         `json:"granted_extensions"`
	TimingMicros      map[string]int64 `json:"timing_micros,omitempty"`
}

type SignSettings struct {
	ReservedUsernames []string
	DeniedExtensions  []string
	LoginHost         string
	LoginPort         int
}

type SignHandler struct {
	policyEnforcer     *policy.Enforcer
	principalValidator *validation.ValidatorDispatcher
	vaultClient        *vaultpkg.Client
	db                 *store.DB
	auditLogger        *audit.Logger
	settings           SignSettings
	logger             *slog.Logger
}

func NewSignHandler(
	policyEnforcer *policy.Enforcer,
	principalValidator *validation.ValidatorDispatcher,
	vaultClient *vaultpkg.Client,
	db *store.DB,
	auditLogger *audit.Logger,
	settings SignSettings,
	logger *slog.Logger,
) *SignHandler {
	return &SignHandler{
		policyEnforcer:     policyEnforcer,
		principalValidator: principalValidator,
		vaultClient:        vaultClient,
		db:                 db,
		auditLogger:        auditLogger,
		settings:           settings,
		logger:             logger,
	}
}

func (h *SignHandler) Handle(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	// Per-component timing for the response.
	timings := map[string]int64{}
	stageStart := startTime
	mark := func(stage string) {
		now := time.Now()
		timings[stage] += now.Sub(stageStart).Microseconds()
		stageStart = now
	}

	defer func() {
		metrics.SignDurationSeconds.Observe(time.Since(startTime).Seconds())
	}()

	var req SignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if !principalRegex.MatchString(req.Principal) {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid principal format: must match ^[a-z_][a-z0-9_-]{0,31}$")
		return
	}
	// Checked before the validator, so it holds whichever validator is on.
	if slices.ContainsFunc(h.settings.ReservedUsernames, func(name string) bool { return strings.EqualFold(name, req.Principal) }) {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		writeError(w, http.StatusForbidden, "reserved_username", "This username cannot be signed for")
		return
	}

	if req.UserEmail == "" || req.UserSub == "" || req.RequestedBy == "" {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing required field: user_email, user_sub and requested_by")
		return
	}
	if !validAttributes(&req) {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "user_email, user_sub, requested_by or force_command is not valid")
		return
	}
	if req.SourceAddress != "" && !validSourceAddress(req.SourceAddress) {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", "source_address must be a comma-separated list of IP addresses or CIDR blocks")
		return
	}

	sshPubKey, err := cert.ParseSSHPublicKey(req.PublicKey)
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_public_key", "Invalid SSH public key format")
		return
	}
	mark("parse_request_us")

	if err := h.policyEnforcer.Enforce(req.TTLSeconds, sshPubKey.Type()); err != nil {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		if policyErr, ok := err.(*policy.PolicyError); ok {
			writeError(w, http.StatusBadRequest, policyErr.Code, policyErr.Message)
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Policy enforcement error")
		return
	}
	mark("policy_enforcement_us")

	grantedExts, err := cert.ResolveExtensions(h.settings.DeniedExtensions, req.ExcludeExtensions)
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	extensionsMap := cert.ExtensionsToMap(grantedExts)
	grantedExtNames := cert.ExtensionNames(grantedExts)

	valResult, err := h.principalValidator.Validate(r.Context(), req.Principal, req.UserSub)
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		if valErr, ok := err.(*validation.ValidationError); ok {
			httputil.WriteJSONErrorWithExtra(w, http.StatusForbidden, "principal_denied", valErr.Message, map[string]string{
				"reason_code": valErr.ReasonCode,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", "Principal validation error")
		return
	}
	mark("principal_validation_us")

	validatedPrincipal := valResult.ValidatedPrincipal

	caKeyPair, err := h.vaultClient.GetCurrentCAKey(r.Context())
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		metrics.VaultOperationsTotal.WithLabelValues("get_ca_key", "error").Inc()
		h.logger.Error("failed to get CA key", "error", err)
		writeError(w, http.StatusServiceUnavailable, "internal_error", "CA key storage unavailable")
		return
	}
	metrics.VaultOperationsTotal.WithLabelValues("get_ca_key", "success").Inc()
	mark("vault_key_fetch_us")

	serial, err := h.db.NextSerial(r.Context())
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		h.logger.Error("failed to allocate serial number", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to allocate serial number")
		return
	}
	mark("serial_us")

	criticalOpts := map[string]string{}
	if req.ForceCommand != "" {
		criticalOpts["force-command"] = req.ForceCommand
	}
	if req.SourceAddress != "" {
		criticalOpts["source-address"] = req.SourceAddress
	}

	signReq := &cert.SignRequest{
		PublicKey:       sshPubKey,
		CAPrivateKeyPEM: caKeyPair.PrivateKey,
		Serial:          uint64(serial),
		Principal:       validatedPrincipal,
		UserEmail:       req.UserEmail,
		UserSub:         req.UserSub,
		RequestedBy:     req.RequestedBy,
		TTLSeconds:      uint64(req.TTLSeconds),
		CriticalOptions: criticalOpts,
		Extensions:      extensionsMap,
	}

	signResult, err := cert.SignCertificate(signReq)
	if err != nil {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		h.logger.Error("failed to sign certificate", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to sign certificate")
		return
	}
	mark("cert_signing_us")

	pubKeyFP := cert.SSHPublicKeyFingerprint(sshPubKey)

	auditEntry := &audit.IssuanceEntry{
		RequestedBy:          req.RequestedBy,
		SerialNumber:         serial,
		KeyID:                signResult.KeyID,
		Principal:            validatedPrincipal,
		UserEmail:            req.UserEmail,
		UserSub:              req.UserSub,
		PublicKeyFingerprint: pubKeyFP,
		CAFingerprint:        signResult.CAFingerprint,
		ValidAfter:           time.Unix(int64(signResult.ValidAfter), 0).UTC(),
		ValidBefore:          time.Unix(int64(signResult.ValidBefore), 0).UTC(),
		GrantedExtensions:    grantedExtNames,
		ForceCommand:         stringPtrIfNonEmpty(req.ForceCommand),
		SourceAddress:        stringPtrIfNonEmpty(req.SourceAddress),
	}

	if err := h.auditLogger.LogIssuance(r.Context(), auditEntry); err != nil {
		metrics.SignRequestsTotal.WithLabelValues("error").Inc()
		h.logger.Error("failed to write audit log", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to record issuance")
		return
	}
	mark("audit_log_us")

	metrics.SignRequestsTotal.WithLabelValues("success").Inc()
	timings["total_us"] = time.Since(startTime).Microseconds()

	resp := SignResponse{
		Certificate:       signResult.AuthorizedKey,
		SerialNumber:      serial,
		ValidAfter:        int64(signResult.ValidAfter),
		ValidBefore:       int64(signResult.ValidBefore),
		CAFingerprint:     signResult.CAFingerprint,
		LoginHost:         h.settings.LoginHost,
		LoginPort:         h.settings.LoginPort,
		Principal:         validatedPrincipal,
		ForceCommand:      req.ForceCommand,
		GrantedExtensions: grantedExtNames,
		TimingMicros:      timings,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// These values reach the key ID in the login log. A control character could forge
// a log line, and ":" could make two key IDs the same.
func validAttributes(req *SignRequest) bool {
	for _, value := range []string{req.UserEmail, req.UserSub, req.RequestedBy} {
		if len(value) > maxAttributeLength || strings.ContainsFunc(value, unicode.IsControl) {
			return false
		}
	}
	return !strings.ContainsFunc(req.ForceCommand, unicode.IsControl) &&
		!strings.Contains(req.UserEmail, ":") && !strings.Contains(req.RequestedBy, ":")
}

// validSourceAddress validate the ip address that is going be restricted on the certificate
func validSourceAddress(list string) bool {
	for entry := range strings.SplitSeq(list, ",") {
		if addr, err := netip.ParseAddr(entry); err == nil && addr.Zone() == "" && !addr.Is4In6() {
			continue
		}
		prefix, err := netip.ParsePrefix(entry)
		if err != nil || prefix != prefix.Masked() || prefix.Addr().Is4In6() {
			return false
		}
	}
	return true
}

func stringPtrIfNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
