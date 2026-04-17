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
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/apache/airavata-custos/signer/internal/httputil"
	"github.com/apache/airavata-custos/signer/internal/store"
)

type CertificateResponse struct {
	SerialNumber         int64    `json:"serial_number"`
	KeyID                string   `json:"key_id"`
	Principal            string   `json:"principal"`
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
}

type CertificateListResponse struct {
	Certificates []CertificateResponse `json:"certificates"`
	Total        int                   `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

type CertificatesHandler struct {
	db     *store.DB
	logger *slog.Logger
}

func NewCertificatesHandler(db *store.DB, logger *slog.Logger) *CertificatesHandler {
	return &CertificatesHandler{
		db:     db,
		logger: logger,
	}
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
			SerialNumber:         c.SerialNumber,
			KeyID:                c.KeyID,
			Principal:            c.Principal,
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

func toCertificateResponse(c *store.CertificateWithStatus) CertificateResponse {
	resp := CertificateResponse{
		SerialNumber:         c.SerialNumber,
		KeyID:                c.KeyID,
		Principal:            c.Principal,
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
	}
	if c.RevokedAt != nil {
		ts := c.RevokedAt.Unix()
		resp.RevokedAt = &ts
	}
	return resp
}
