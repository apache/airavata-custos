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
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"

	"github.com/apache/airavata-custos/signer/internal/auth"
	"github.com/apache/airavata-custos/signer/internal/httputil"
	signerservice "github.com/apache/airavata-custos/signer/internal/service"
	"github.com/apache/airavata-custos/signer/internal/store"
)

type CertificateResponse struct {
	TenantID             string   `json:"tenant_id"`
	ClientID             string   `json:"client_id"`
	SerialNumber         int64    `json:"serial_number"`
	KeyID                string   `json:"key_id"`
	Principal            string   `json:"principal"`
	UserEmail            string   `json:"user_email"`
	PublicKeyFingerprint string   `json:"public_key_fingerprint"`
	CAFingerprint        string   `json:"ca_fingerprint"`
	ValidAfter           int64    `json:"valid_after"`
	ValidBefore          int64    `json:"valid_before"`
	IssuedAt             int64    `json:"issued_at"`
	SourceIP             string   `json:"source_ip,omitempty"`
	GrantedExtensions    []string `json:"granted_extensions,omitempty"`
	ForceCommand         *string  `json:"force_command,omitempty"`
	Revoked              bool     `json:"revoked"`
	RevokedAt            *int64   `json:"revoked_at,omitempty"`
	RevocationReason     string   `json:"revocation_reason,omitempty"`
	RevokedBy            string   `json:"revoked_by,omitempty"`
}

type CertificateListResponse struct {
	Certificates []CertificateResponse `json:"certificates"`
	Total        int                   `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

type AdminCertificateListResponse struct {
	Certificates []CertificateResponse `json:"certificates"`
	Limit        int                   `json:"limit"`
	NextCursor   string                `json:"next_cursor,omitempty"`
}

type CertificateRevokeRequest struct {
	Reason string `json:"reason"`
}

type CertificateRevokeResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	SerialNumber   int64  `json:"serial_number"`
	Revoked        bool   `json:"revoked"`
	RevokedAt      int64  `json:"revoked_at"`
	Reason         string `json:"reason"`
	AlreadyRevoked bool   `json:"already_revoked"`
}

type CertificateStore interface {
	ListCertificatesByEmail(context.Context, string, int, int) (*store.CertificateListResult, error)
	ListCertificates(context.Context, int, *store.CertificateCursor) (*store.CertificatePageResult, error)
	GetCertificateBySerial(context.Context, int64) (*store.CertificateWithStatus, error)
}

type CertificatesHandler struct {
	db             CertificateStore
	coreAuthorizer auth.CoreAuthorizer
	revoker        CertificateRevoker
	logger         *slog.Logger
}

func NewCertificatesHandler(db CertificateStore, logger *slog.Logger) *CertificatesHandler {
	return &CertificatesHandler{
		db:     db,
		logger: logger,
	}
}

func (h *CertificatesHandler) WithRevocation(authorizer auth.CoreAuthorizer, revoker CertificateRevoker) *CertificatesHandler {
	h.coreAuthorizer = authorizer
	h.revoker = revoker
	return h
}

func (h *CertificatesHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	identity := httputil.UserIdentityFromContext(r.Context())
	if identity == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Missing user identity")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit <= 0 {
		limit = 20
	}

	result, err := h.db.ListCertificatesByEmail(r.Context(), identity.Email, limit, offset)
	if err != nil {
		h.logger.Error("failed to list certificates", "error", err, "email", identity.Email)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list certificates")
		return
	}

	certs := make([]CertificateResponse, 0, len(result.Certificates))
	for _, c := range result.Certificates {
		cert := CertificateResponse{
			TenantID:             c.TenantID,
			ClientID:             c.ClientID,
			SerialNumber:         c.SerialNumber,
			KeyID:                c.KeyID,
			Principal:            c.Principal,
			UserEmail:            c.UserEmail,
			PublicKeyFingerprint: c.PublicKeyFingerprint,
			CAFingerprint:        c.CAFingerprint,
			ValidAfter:           c.ValidAfter.Unix(),
			ValidBefore:          c.ValidBefore.Unix(),
			IssuedAt:             c.IssuedAt.Unix(),
			SourceIP:             c.SourceIP,
			GrantedExtensions:    c.GrantedExtensions,
			ForceCommand:         c.ForceCommand,
			Revoked:              c.Revoked,
			RevocationReason:     c.RevocationReason,
			RevokedBy:            c.RevokedBy,
		}
		if c.RevokedAt != nil {
			ts := c.RevokedAt.Unix()
			cert.RevokedAt = &ts
		}
		certs = append(certs, cert)
	}

	resp := CertificateListResponse{
		Certificates: certs,
		Total:        result.Total,
		Limit:        limit,
		Offset:       offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// HandleGet expects the serial number as the last URL path segment.
func (h *CertificatesHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	identity := httputil.UserIdentityFromContext(r.Context())
	if identity == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Missing user identity")
		return
	}

	parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
	serialStr := parts[len(parts)-1]

	serial, err := strconv.ParseInt(serialStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid serial number")
		return
	}

	cert, err := h.db.GetCertificateBySerial(r.Context(), serial)
	if err != nil {
		if err == sql.ErrNoRows || strings.Contains(err.Error(), "no rows") {
			writeError(w, http.StatusNotFound, "not_found", "Certificate not found")
			return
		}
		h.logger.Error("failed to get certificate", "error", err, "serial", serial)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get certificate")
		return
	}

	// Authorization: user can only view their own certificates
	if cert.UserEmail != identity.Email {
		writeError(w, http.StatusForbidden, "forbidden", "You can only view your own certificates")
		return
	}

	resp := toCertificateResponse(cert)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func (h *CertificatesHandler) HandleAdminList(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	cursor, err := decodeCertificateCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_cursor", "Invalid certificate cursor")
		return
	}
	result, err := h.db.ListCertificates(r.Context(), limit, cursor)
	if err != nil {
		h.logger.Error("failed to list certificates for administrator", "error", err)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to list certificates")
		return
	}
	certificates := make([]CertificateResponse, 0, len(result.Certificates))
	for i := range result.Certificates {
		certificates = append(certificates, toCertificateResponse(&result.Certificates[i]))
	}
	response := AdminCertificateListResponse{Certificates: certificates, Limit: limit}
	if result.NextCursor != nil {
		response.NextCursor = encodeCertificateCursor(result.NextCursor)
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *CertificatesHandler) HandleAdminGet(w http.ResponseWriter, r *http.Request) {
	serial, ok := certificateSerial(w, r)
	if !ok {
		return
	}
	certificate, err := h.db.GetCertificateBySerial(r.Context(), serial)
	if err != nil {
		if errors.Is(err, store.ErrCertificateNotFound) || errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "Certificate not found")
			return
		}
		h.logger.Error("failed to get certificate for administrator", "error", err, "serial", serial)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to get certificate")
		return
	}
	writeJSON(w, http.StatusOK, toCertificateResponse(certificate))
}

func (h *CertificatesHandler) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	identity := httputil.UserIdentityFromContext(r.Context())
	if identity == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "Missing user identity")
		return
	}
	serial, ok := certificateSerial(w, r)
	if !ok {
		return
	}
	var request CertificateRevokeRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_request", "Request body must contain one JSON object")
		return
	}
	reason := strings.TrimSpace(request.Reason)
	if utf8.RuneCountInString(reason) == 0 || utf8.RuneCountInString(reason) > 255 {
		writeError(w, http.StatusBadRequest, "invalid_reason", "Reason must be between 1 and 255 characters")
		return
	}

	certificate, err := h.db.GetCertificateBySerial(r.Context(), serial)
	if err != nil {
		if errors.Is(err, store.ErrCertificateNotFound) || errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "not_found", "Certificate not found")
			return
		}
		h.logger.Error("failed to load certificate for revocation authorization", "error", err, "serial", serial)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to load certificate")
		return
	}

	revokedBy := identity.Subject
	if revokedBy == "" {
		revokedBy = identity.Email
	}
	canRevokeAny := false
	if certificate.UserEmail != identity.Email {
		if h.coreAuthorizer == nil {
			writeError(w, http.StatusServiceUnavailable, "authorization_unavailable", "Authorization service unavailable")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		caller, err := h.coreAuthorizer.ResolveCaller(r.Context(), token)
		if errors.Is(err, auth.ErrCoreUnauthorized) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid or expired bearer token")
			return
		}
		if err != nil || caller == nil {
			writeError(w, http.StatusServiceUnavailable, "authorization_unavailable", "Authorization service unavailable")
			return
		}
		if !caller.HasPrivilege(auth.SignerCertificatesWrite) {
			writeError(w, http.StatusForbidden, "insufficient_privilege", "Caller lacks required privilege")
			return
		}
		canRevokeAny = true
		revokedBy = caller.ID
		if revokedBy == "" {
			revokedBy = caller.Email
		}
	}
	if h.revoker == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "Revocation service unavailable")
		return
	}
	result, err := h.revoker.Revoke(r.Context(), signerservice.RevokeCommand{
		SerialNumber: serial, Reason: reason, RevokedBy: revokedBy,
		OwnerEmail: identity.Email, CanRevokeAny: canRevokeAny, RequireActive: true,
	})
	switch {
	case errors.Is(err, store.ErrCertificateNotFound):
		writeError(w, http.StatusNotFound, "not_found", "Certificate not found")
		return
	case errors.Is(err, signerservice.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden", "You cannot revoke this certificate")
		return
	case errors.Is(err, signerservice.ErrCertificateNotActive):
		writeError(w, http.StatusConflict, "certificate_not_active", "Certificate is not active")
		return
	case err != nil:
		h.logger.Error("failed to revoke certificate", "error", err, "serial", serial)
		writeError(w, http.StatusInternalServerError, "internal_error", "Failed to revoke certificate")
		return
	}
	message := "Certificate revoked successfully"
	if result.AlreadyRevoked {
		message = "Certificate was already revoked"
	}
	writeJSON(w, http.StatusOK, CertificateRevokeResponse{
		Success: true, Message: message, SerialNumber: result.SerialNumber, Revoked: true,
		RevokedAt: result.RevokedAt.Unix(), Reason: result.Reason, AlreadyRevoked: result.AlreadyRevoked,
	})
}

type certificateCursorPayload struct {
	Version  int   `json:"v"`
	IssuedAt int64 `json:"issued_at"`
	ID       int64 `json:"id"`
}

func encodeCertificateCursor(cursor *store.CertificateCursor) string {
	payload, _ := json.Marshal(certificateCursorPayload{Version: 1, IssuedAt: cursor.IssuedAt.UnixNano(), ID: cursor.ID})
	return base64.RawURLEncoding.EncodeToString(payload)
}

func decodeCertificateCursor(value string) (*store.CertificateCursor, error) {
	if value == "" {
		return nil, nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	var decoded certificateCursorPayload
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return nil, err
	}
	if decoded.Version != 1 || decoded.IssuedAt <= 0 || decoded.ID <= 0 {
		return nil, errors.New("invalid cursor payload")
	}
	return &store.CertificateCursor{IssuedAt: time.Unix(0, decoded.IssuedAt).UTC(), ID: decoded.ID}, nil
}

func certificateSerial(w http.ResponseWriter, r *http.Request) (int64, bool) {
	value := chi.URLParam(r, "serial")
	if value == "" {
		parts := strings.Split(strings.TrimRight(r.URL.Path, "/"), "/")
		last := len(parts) - 1
		if parts[last] == "revoke" && last > 0 {
			last--
		}
		value = parts[last]
	}
	serial, err := strconv.ParseInt(value, 10, 64)
	if err != nil || serial <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid serial number")
		return 0, false
	}
	return serial, true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func toCertificateResponse(c *store.CertificateWithStatus) CertificateResponse {
	resp := CertificateResponse{
		TenantID:             c.TenantID,
		ClientID:             c.ClientID,
		SerialNumber:         c.SerialNumber,
		KeyID:                c.KeyID,
		Principal:            c.Principal,
		UserEmail:            c.UserEmail,
		PublicKeyFingerprint: c.PublicKeyFingerprint,
		CAFingerprint:        c.CAFingerprint,
		ValidAfter:           c.ValidAfter.Unix(),
		ValidBefore:          c.ValidBefore.Unix(),
		IssuedAt:             c.IssuedAt.Unix(),
		SourceIP:             c.SourceIP,
		GrantedExtensions:    c.GrantedExtensions,
		ForceCommand:         c.ForceCommand,
		Revoked:              c.Revoked,
		RevocationReason:     c.RevocationReason,
		RevokedBy:            c.RevokedBy,
	}
	if c.RevokedAt != nil {
		ts := c.RevokedAt.Unix()
		resp.RevokedAt = &ts
	}
	return resp
}
