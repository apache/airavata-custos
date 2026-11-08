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
	"context"
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
)

const Type = "signer"

// The raw error has the signer's internal address, so callers get this instead.
var errSignerUnavailable = errors.New("signer unavailable")

type Handlers struct {
	client  *http.Client
	baseURL string
}

func NewHandlers(baseURL string, timeout time.Duration) *Handlers {
	return &Handlers{
		client:  &http.Client{Timeout: timeout},
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

func (h *Handlers) RegisterRoutes(router *identity.Router) {
	router.RequirePrivilege("GET /connectors/signer/certificates", CertificatesRead, h.forwardTo("/api/v1/admin/certificates"))
	router.RequirePrivilege("GET /connectors/signer/certificates/{serial}", CertificatesRead, h.forwardTo("/api/v1/admin/certificates/{serial}"))
	router.RequirePrivilege("POST /connectors/signer/certificates/{serial}/revoke", CertificatesWrite, h.forwardTo("/api/v1/certificates/{serial}/revoke"))
	router.RequireAuth("GET /connectors/signer/me/certificates", h.forwardTo("/api/v1/certificates"))
	router.RequireAuth("GET /connectors/signer/me/certificates/{serial}", h.forwardTo("/api/v1/certificates/{serial}"))
	router.RequirePrivilege("GET /connectors/signer/status", CertificatesRead, h.status)
}

// The caller's own token goes to the signer, so the signer checks the real user.
func (h *Handlers) forwardTo(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		target := h.baseURL + strings.Replace(path, "{serial}", url.PathEscape(r.PathValue("serial")), 1)
		if r.URL.RawQuery != "" {
			target += "?" + r.URL.RawQuery
		}
		req, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
		if err != nil {
			common.WriteError(w, http.StatusInternalServerError, errSignerUnavailable)
			return
		}
		req.Header.Set("Authorization", r.Header.Get("Authorization"))
		if ct := r.Header.Get("Content-Type"); ct != "" {
			req.Header.Set("Content-Type", ct)
		}
		resp, err := h.client.Do(req)
		health.Record(Type, Type, callResult(resp, err))
		if err != nil {
			slog.Warn("signer request failed", "path", path, "error", err)
			common.WriteError(w, http.StatusBadGateway, errSignerUnavailable)
			return
		}
		defer resp.Body.Close()
		w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

// A 5xx counts as a failure. A 4xx is the signer working and refusing the caller.
func callResult(resp *http.Response, err error) error {
	if err == nil && resp.StatusCode >= http.StatusInternalServerError {
		return fmt.Errorf("signer returned %d", resp.StatusCode)
	}
	return err
}

type statusResponse struct {
	Reachable bool   `json:"reachable"`
	LatencyMs int64  `json:"latency_ms"`
	BaseURL   string `json:"base_url"`
}

func (h *Handlers) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.baseURL+"/api/v1/health", nil)
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
	common.WriteJSON(w, http.StatusOK, statusResponse{Reachable: err == nil, LatencyMs: time.Since(start).Milliseconds(), BaseURL: h.baseURL})
}
