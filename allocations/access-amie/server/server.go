// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HealthChecker interface {
	CheckHealth(ctx context.Context) error
}

func New(port int, db *sqlx.DB, healthChecker HealthChecker) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler(db, healthChecker))
	mux.HandleFunc("GET /ready", readyHandler(db))
	mux.Handle("GET /metrics", promhttp.Handler())

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func healthHandler(db *sqlx.DB, checker HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"status": "UP",
		}

		// Check DB
		if err := db.PingContext(r.Context()); err != nil {
			resp["status"] = "DOWN"
			resp["database"] = map[string]string{"status": "DOWN", "error": err.Error()}
			w.WriteHeader(http.StatusServiceUnavailable)
			writeJSON(w, resp)
			return
		}
		resp["database"] = map[string]string{"status": "UP"}

		// Check AMIE API
		if checker != nil {
			if err := checker.CheckHealth(r.Context()); err != nil {
				resp["amie"] = map[string]string{"status": "DOWN", "error": err.Error()}
				// AMIE being down doesn't make the service unhealthy
			} else {
				resp["amie"] = map[string]string{"status": "UP"}
			}
		}

		writeJSON(w, resp)
	}
}

func readyHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			writeJSON(w, map[string]string{"status": "NOT_READY", "error": err.Error()})
			return
		}
		writeJSON(w, map[string]string{"status": "READY"})
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}
