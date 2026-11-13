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

package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/apache/airavata-custos/internal/connectors/health"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

const Type = "signer"

const (
	AuditCertificateIssued  = "SSH_CERTIFICATE_ISSUED"
	AuditCertificateRevoked = "SSH_CERTIFICATE_REVOKED"
)

const maxRequestBytes = 16 << 10

var errServiceTokenRefused = errors.New("signer refused the service token")

// CoreAccounts is what the connector needs from core, who can generate, and the audit log.
type CoreAccounts interface {
	GetComputeClusterUserByPair(ctx context.Context, clusterID, userID string) (*models.ComputeClusterUser, error)
	ListAllocationsForUser(ctx context.Context, userID string) ([]models.ComputeAllocationMembership, error)
	GetComputeAllocation(ctx context.Context, id string) (*models.ComputeAllocation, error)
	GetUser(ctx context.Context, id string) (*models.User, error)
	CreateAuditEvent(ctx context.Context, e *models.AuditEvent) (*models.AuditEvent, error)
}

type Config struct {
	BaseURL      string
	ServiceToken string
	ClusterID    string
	TTLSeconds   int
	Timeout      time.Duration
}

type Handlers struct {
	client *http.Client
	cfg    Config
	core   CoreAccounts
}

func NewHandlers(cfg Config, core CoreAccounts) *Handlers {
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")
	return &Handlers{client: &http.Client{Timeout: cfg.Timeout}, cfg: cfg, core: core}
}

func (h *Handlers) RegisterRoutes(router *identity.Router) {
	router.RequirePrivilege("GET /connectors/signer/certificates", CertificatesRead, h.listAllCertificates)
	router.RequirePrivilege("GET /connectors/signer/certificates/{serial}", CertificatesRead, h.getAnyCertificate)
	router.RequirePrivilege("POST /connectors/signer/certificates/{serial}/revoke", CertificatesWrite, h.revoke)
	router.RequireAuth("GET /connectors/signer/me/certificates", h.listOwnCertificates)
	router.RequireAuth("POST /connectors/signer/me/certificates", h.generateCertificate)
	router.RequireAuth("GET /connectors/signer/me/certificates/{serial}", h.getOwnCertificate)
	router.RequireAuth("POST /connectors/signer/me/certificates/{serial}/revoke", h.revokeOwnCertificate)
	router.RequirePrivilege("GET /connectors/signer/status", CertificatesRead, h.status)
	router.RequirePrivilege("GET /connectors/signer/users/{id}", CertificatesRead, h.getUserEmail)
}

func refuse(w http.ResponseWriter, status int, code, message string) {
	common.WriteJSON(w, status, map[string]string{"error": code, "message": message})
}

func signerUnavailable(w http.ResponseWriter) {
	refuse(w, http.StatusBadGateway, "signer_unavailable", "The certificate service cannot be reached")
}

func certificatePath(r *http.Request) string {
	return "/api/v1/certificates/" + url.PathEscape(r.PathValue("serial"))
}

// requireClaims writes the refusal and returns nil if the token does not say who the caller is.
func requireClaims(w http.ResponseWriter, r *http.Request) *identity.Claims {
	claims := identity.ClaimsFromContext(r.Context())
	if claims == nil || claims.Sub == "" {
		refuse(w, http.StatusForbidden, "user_not_active", "Your account cannot do this")
		return nil
	}
	return claims
}

// The caller's own token never goes to the signer.
func (h *Handlers) callSigner(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		payload = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, h.cfg.BaseURL+path, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+h.cfg.ServiceToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := h.client.Do(req)
	if err == nil && resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		err = errServiceTokenRefused
	}
	health.Record(Type, Type, callResult(resp, err))
	if err != nil {
		slog.Warn("signer request failed", "path", path, "error", err)
		return nil, err
	}
	return resp, nil
}

func copyResponse(w http.ResponseWriter, resp *http.Response) {
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func (h *Handlers) relay(w http.ResponseWriter, r *http.Request, method, path string, body any) {
	resp, err := h.callSigner(r.Context(), method, path, body)
	if err != nil {
		signerUnavailable(w)
		return
	}
	defer resp.Body.Close()
	copyResponse(w, resp)
}

// relayAndAudit relays the call and records eventType in the audit log once the signer has done it.
func (h *Handlers) relayAndAudit(w http.ResponseWriter, r *http.Request, path string, body any, eventType string, details map[string]any) {
	resp, err := h.callSigner(r.Context(), http.MethodPost, path, body)
	if err != nil {
		signerUnavailable(w)
		return
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		signerUnavailable(w)
		return
	}
	if resp.StatusCode == http.StatusOK {
		// The certificate is already issued or revoked by now, so a failed write is only logged, and the caller still gets the result.
		var result struct {
			SerialNumber int64 `json:"serial_number"`
		}
		err := json.Unmarshal(respBody, &result)
		if err == nil {
			payload, _ := json.Marshal(details)
			_, err = h.core.CreateAuditEvent(r.Context(), &models.AuditEvent{EventType: eventType, EntityType: "ssh_certificate",
				EntityID: strconv.FormatInt(result.SerialNumber, 10), Source: Type, Details: string(payload)})
		}
		if err != nil {
			slog.Error("signer: writing the audit event failed", "event_type", eventType, "error", err)
		}
	}
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

func (h *Handlers) listAllCertificates(w http.ResponseWriter, r *http.Request) {
	h.relay(w, r, http.MethodGet, "/api/v1/certificates?"+r.URL.RawQuery, nil)
}

func (h *Handlers) getAnyCertificate(w http.ResponseWriter, r *http.Request) {
	h.relay(w, r, http.MethodGet, certificatePath(r), nil)
}

func (h *Handlers) revoke(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBytes)).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		refuse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}
	revokedBy := identity.CallerFromContext(r.Context()).Requester()
	h.relayAndAudit(w, r, certificatePath(r)+"/revoke", map[string]string{"reason": req.Reason, "revoked_by": revokedBy},
		AuditCertificateRevoked, map[string]any{"revoked_by": revokedBy, "reason": req.Reason})
}

func (h *Handlers) listOwnCertificates(w http.ResponseWriter, r *http.Request) {
	claims := requireClaims(w, r)
	if claims == nil {
		return
	}
	query := r.URL.Query()
	query.Set("user_sub", claims.Sub)
	h.relay(w, r, http.MethodGet, "/api/v1/certificates?"+query.Encode(), nil)
}

// fetchUsersCertificate gets the certificate from the signer and answers 404 when it is not the caller's.
func (h *Handlers) fetchUsersCertificate(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	claims := requireClaims(w, r)
	if claims == nil {
		return nil, false
	}
	resp, err := h.callSigner(r.Context(), http.MethodGet, certificatePath(r), nil)
	if err != nil {
		signerUnavailable(w)
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		copyResponse(w, resp)
		return nil, false
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		signerUnavailable(w)
		return nil, false
	}
	var certificate struct {
		UserSub string `json:"user_sub"`
	}
	if err := json.Unmarshal(body, &certificate); err != nil {
		signerUnavailable(w)
		return nil, false
	}
	if certificate.UserSub != claims.Sub {
		refuse(w, http.StatusNotFound, "not_found", "Certificate not found")
		return nil, false
	}
	return body, true
}

func writeCertificate(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func (h *Handlers) getOwnCertificate(w http.ResponseWriter, r *http.Request) {
	if body, ok := h.fetchUsersCertificate(w, r); ok {
		writeCertificate(w, body)
	}
}

func (h *Handlers) revokeOwnCertificate(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.fetchUsersCertificate(w, r); ok {
		h.revoke(w, r)
	}
}

func (h *Handlers) generateCertificate(w http.ResponseWriter, r *http.Request) {
	claims := requireClaims(w, r)
	if claims == nil {
		return
	}
	caller := identity.CallerFromContext(r.Context())
	var req struct {
		PublicKey         string   `json:"public_key"`
		TTLSeconds        int      `json:"ttl_seconds"`
		SourceAddress     string   `json:"source_address"`
		ForceCommand      string   `json:"force_command"`
		ExcludeExtensions []string `json:"exclude_extensions"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBytes)).Decode(&req); err != nil {
		refuse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}
	if req.PublicKey == "" {
		refuse(w, http.StatusBadRequest, "invalid_public_key", "A public key is required")
		return
	}
	// The config value is the default and the longest a caller may ask for.
	if req.TTLSeconds == 0 {
		req.TTLSeconds = h.cfg.TTLSeconds
	}
	if req.TTLSeconds < 0 || req.TTLSeconds > h.cfg.TTLSeconds {
		refuse(w, http.StatusBadRequest, "invalid_ttl", fmt.Sprintf("Lifetime must be between 1 and %d seconds", h.cfg.TTLSeconds))
		return
	}
	h.issueCertificate(w, r, caller.UserID, map[string]any{
		"public_key":         req.PublicKey,
		"ttl_seconds":        req.TTLSeconds,
		"user_email":         caller.Email,
		"user_sub":           claims.Sub,
		"requested_by":       caller.Requester(),
		"source_address":     req.SourceAddress,
		"force_command":      req.ForceCommand,
		"exclude_extensions": req.ExcludeExtensions,
	})
}

func accountReadFailed(w http.ResponseWriter, err error) {
	slog.Error("signer: reading the account failed", "error", err)
	refuse(w, http.StatusInternalServerError, "internal_error", "Could not read the account")
}

func (h *Handlers) issueCertificate(w http.ResponseWriter, r *http.Request, userID string, sign map[string]any) {
	username, code, err := h.loginName(r.Context(), userID)
	if err != nil {
		accountReadFailed(w, err)
		return
	}
	if code != "" {
		slog.Info("signer: certificate refused", "code", code, "user_id", userID, "requested_by", sign["requested_by"])
		refuse(w, http.StatusForbidden, code, "This account cannot generate a certificate")
		return
	}
	sign["principal"] = username
	h.relayAndAudit(w, r, "/api/v1/sign", sign, AuditCertificateIssued, map[string]any{
		"user_id": userID, "principal": username, "requested_by": sign["requested_by"], "ttl_seconds": sign["ttl_seconds"],
		"source_address": sign["source_address"], "force_command": sign["force_command"], "exclude_extensions": sign["exclude_extensions"],
	})
}

// loginName returns the username to sign for.
func (h *Handlers) loginName(ctx context.Context, userID string) (username, code string, err error) {
	account, err := h.core.GetComputeClusterUserByPair(ctx, h.cfg.ClusterID, userID)
	if errors.Is(err, service.ErrNotFound) {
		return "", "no_cluster_account", nil
	}
	if err != nil {
		return "", "", err
	}
	if account.ApprovalStatus != models.ClusterAccountApproved || account.ProvisionedAt == nil {
		return "", "cluster_account_not_ready", nil
	}
	memberships, err := h.core.ListAllocationsForUser(ctx, userID)
	if err != nil {
		return "", "", err
	}
	now := time.Now()
	for _, membership := range memberships {
		if !activeAt(now, membership.MembershipStatus, membership.StartTime, membership.EndTime) {
			continue
		}
		allocation, err := h.core.GetComputeAllocation(ctx, membership.ComputeAllocationID)
		if err != nil {
			return "", "", err
		}
		if allocation.ComputeClusterID == h.cfg.ClusterID && activeAt(now, allocation.Status, allocation.StartTime, allocation.EndTime) {
			return account.LocalUsername, "", nil
		}
	}
	return "", "no_active_allocation", nil
}

func activeAt(now time.Time, status models.AllocationStatus, start, end time.Time) bool {
	return status == models.ACTIVE && !now.Before(start) && now.Before(end)
}

// A 5xx counts as a failure. A 4xx is the signer working and refusing the request.
func callResult(resp *http.Response, err error) error {
	if err == nil && resp.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("signer returned %d", resp.StatusCode)
	}
	return err
}

type statusResponse struct {
	Reachable bool  `json:"reachable"`
	LatencyMs int64 `json:"latency_ms"`
}

func (h *Handlers) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.cfg.BaseURL+"/api/v1/health", nil)
	if err != nil {
		return err
	}
	resp, err := h.client.Do(req)
	if err == nil {
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("signer health returned %d", resp.StatusCode)
		}
	}
	health.Record(Type, Type, err)
	return err
}

// An unreachable signer is still a 200 here, so the caller can show it.
func (h *Handlers) status(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	err := h.CheckHealth(r.Context())
	common.WriteJSON(w, http.StatusOK, statusResponse{Reachable: err == nil, LatencyMs: time.Since(start).Milliseconds()})
}

// getUserEmail lets a certificate reader see who generated or revoked a certificate.
// Only the email, since the reader may not be allowed to read users.
func (h *Handlers) getUserEmail(w http.ResponseWriter, r *http.Request) {
	user, err := h.core.GetUser(r.Context(), r.PathValue("id"))
	if errors.Is(err, service.ErrNotFound) {
		refuse(w, http.StatusNotFound, "not_found", "User not found")
		return
	}
	if err != nil {
		accountReadFailed(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]string{"id": user.ID, "email": user.Email})
}
