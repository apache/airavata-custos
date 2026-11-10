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
	"strings"
	"time"

	"github.com/apache/airavata-custos/internal/connectors/health"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

const Type = "signer"

const maxRequestBytes = 16 << 10

var errServiceTokenRefused = errors.New("signer refused the service token")

// CoreAccounts is what the connector reads from core to decide who can generate.
type CoreAccounts interface {
	GetComputeClusterUserByPair(ctx context.Context, clusterID, userID string) (*models.ComputeClusterUser, error)
	ListAllocationsForUser(ctx context.Context, userID string) ([]models.ComputeAllocationMembership, error)
	GetComputeAllocation(ctx context.Context, id string) (*models.ComputeAllocation, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	PendingUserAgreements(ctx context.Context, userID string) ([]models.UserAgreement, error)
	ListUserIdentitiesForUser(ctx context.Context, userID string) ([]models.UserIdentity, error)
	IsAppAuthorized(ctx context.Context, appID, userID string) (bool, error)
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
	router.RequirePrivilege("POST /connectors/signer/certificates/{serial}/revoke", CertificatesWrite, h.revokeAnyCertificate)
	router.RequireAuth("GET /connectors/signer/me/certificates", h.listOwnCertificates)
	router.RequireAuth("POST /connectors/signer/me/certificates", h.generateCertificate)
	router.RequireAuth("GET /connectors/signer/me/certificates/{serial}", h.getOwnCertificate)
	router.RequireAuth("POST /connectors/signer/me/certificates/{serial}/revoke", h.revokeOwnCertificate)
	router.RequirePrivilege("GET /connectors/signer/status", CertificatesRead, h.status)
	router.RequireApp("POST /connectors/signer/app/certificates", AppActionGenerateCertificate, h.generateCertificateForApp)
	router.RequireApp("GET /connectors/signer/app/certificates", AppActionGenerateCertificate, h.listCertificatesForApp)
	router.RequireApp("GET /connectors/signer/app/certificates/{serial}", AppActionGenerateCertificate, h.getCertificateForApp)
	router.RequireApp("POST /connectors/signer/app/certificates/{serial}/revoke", AppActionRevokeCertificate, h.revokeCertificateForApp)
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
	if claims == nil || claims.Sub == "" || claims.Email == "" {
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

func (h *Handlers) listAllCertificates(w http.ResponseWriter, r *http.Request) {
	h.relay(w, r, http.MethodGet, "/api/v1/certificates?"+r.URL.RawQuery, nil)
}

func (h *Handlers) getAnyCertificate(w http.ResponseWriter, r *http.Request) {
	h.relay(w, r, http.MethodGet, certificatePath(r), nil)
}

func (h *Handlers) revokeAnyCertificate(w http.ResponseWriter, r *http.Request) {
	if claims := requireClaims(w, r); claims != nil {
		h.revoke(w, r, claims.Email)
	}
}

// revokedBy is the admin's email, the app, or "self" when users revoke their own.
func (h *Handlers) revoke(w http.ResponseWriter, r *http.Request, revokedBy string) {
	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBytes)).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		refuse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}
	h.relay(w, r, http.MethodPost, certificatePath(r)+"/revoke", map[string]string{"reason": req.Reason, "revoked_by": revokedBy})
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

type certificateOwner struct {
	UserSub     string `json:"user_sub"`
	RequestedBy string `json:"requested_by"`
}

// fetchIfOwned extracts own certificate from the Signer and returns false if it is not the caller's.
func (h *Handlers) fetchIfOwned(w http.ResponseWriter, r *http.Request, owned func(certificateOwner) bool) ([]byte, bool) {
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
	var certificate certificateOwner
	if err := json.Unmarshal(body, &certificate); err != nil {
		signerUnavailable(w)
		return nil, false
	}
	if !owned(certificate) {
		refuse(w, http.StatusNotFound, "not_found", "Certificate not found")
		return nil, false
	}
	return body, true
}

func (h *Handlers) fetchUsersCertificate(w http.ResponseWriter, r *http.Request) ([]byte, bool) {
	claims := requireClaims(w, r)
	if claims == nil {
		return nil, false
	}
	return h.fetchIfOwned(w, r, func(c certificateOwner) bool { return c.UserSub == claims.Sub })
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
		h.revoke(w, r, "self")
	}
}

func (h *Handlers) generateCertificate(w http.ResponseWriter, r *http.Request) {
	claims := requireClaims(w, r)
	if claims == nil {
		return
	}
	userID := identity.CallerFromContext(r.Context()).UserID
	var req struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBytes)).Decode(&req); err != nil || req.PublicKey == "" {
		refuse(w, http.StatusBadRequest, "invalid_public_key", "A public key is required")
		return
	}
	h.issueCertificate(w, r, userID, map[string]any{
		"public_key":   req.PublicKey,
		"user_email":   claims.Email,
		"user_sub":     claims.Sub,
		"requested_by": "user/" + userID,
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
		refuse(w, http.StatusForbidden, code, "This account cannot generate a certificate")
		return
	}
	sign["principal"] = username
	sign["ttl_seconds"] = h.cfg.TTLSeconds
	h.relay(w, r, http.MethodPost, "/api/v1/sign", sign)
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
