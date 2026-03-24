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
	"encoding/json"
	"net/http"
	"time"

	"github.com/apache/airavata-custos/signer/internal/store"
	vaultpkg "github.com/apache/airavata-custos/signer/internal/vault"
)

type HealthResponse struct {
	Status string                        `json:"status"`
	Checks map[string]*HealthCheckResult `json:"checks"`
}

type HealthCheckResult struct {
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
	Error     string `json:"error,omitempty"`
}

type HealthHandler struct {
	db          *store.DB
	vaultClient *vaultpkg.Client
}

func NewHealthHandler(db *store.DB, vaultClient *vaultpkg.Client) *HealthHandler {
	return &HealthHandler{
		db:          db,
		vaultClient: vaultClient,
	}
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	checks := make(map[string]*HealthCheckResult)
	healthy := true

	dbResult := h.checkDatabase(ctx)
	checks["database"] = dbResult
	if dbResult.Status != "up" {
		healthy = false
	}

	vaultResult := h.checkVault(ctx)
	checks["vault"] = vaultResult
	if vaultResult.Status != "up" {
		healthy = false
	}

	status := "healthy"
	httpStatus := http.StatusOK
	if !healthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	resp := HealthResponse{
		Status: status,
		Checks: checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resp)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) *HealthCheckResult {
	start := time.Now()
	err := h.db.Ping()
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &HealthCheckResult{
			Status:    "down",
			LatencyMs: latency,
			Error:     "database connection failed",
		}
	}
	return &HealthCheckResult{
		Status:    "up",
		LatencyMs: latency,
	}
}

func (h *HealthHandler) checkVault(ctx context.Context) *HealthCheckResult {
	start := time.Now()
	err := h.vaultClient.Healthy(ctx)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &HealthCheckResult{
			Status:    "down",
			LatencyMs: latency,
			Error:     "vault connection failed",
		}
	}
	return &HealthCheckResult{
		Status:    "up",
		LatencyMs: latency,
	}
}
