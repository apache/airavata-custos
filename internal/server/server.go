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

// Package server exposes the pkg/service API over HTTP/JSON.
package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// Server is an HTTP handler that exposes the service API.
type Server struct {
	svc *service.Service
	mux *http.ServeMux
}

// New builds an HTTP handler wired to the supplied service.
func New(svc *service.Service) *Server {
	s := &Server{svc: svc, mux: http.NewServeMux()}
	s.routes()
	return s
}

// ServeHTTP satisfies http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.healthz)

	s.mux.HandleFunc("POST /organizations", s.createOrganization)
	s.mux.HandleFunc("GET /organizations/{id}", s.getOrganization)

	s.mux.HandleFunc("POST /users", s.createUser)
	s.mux.HandleFunc("GET /users/{id}", s.getUser)

	s.mux.HandleFunc("POST /projects", s.createProject)
	s.mux.HandleFunc("GET /projects/{id}", s.getProject)

	s.mux.HandleFunc("POST /compute-clusters", s.createComputeCluster)
	s.mux.HandleFunc("GET /compute-clusters", s.listComputeClusters)
	s.mux.HandleFunc("GET /compute-clusters/{id}", s.getComputeCluster)

	s.mux.HandleFunc("POST /compute-allocations", s.createComputeAllocation)
	s.mux.HandleFunc("GET /compute-allocations/{id}", s.getComputeAllocation)

	s.mux.HandleFunc("POST /compute-allocation-resources", s.createComputeAllocationResource)
	s.mux.HandleFunc("GET /compute-allocation-resources", s.listComputeAllocationResources)
	s.mux.HandleFunc("GET /compute-allocation-resources/{id}", s.getComputeAllocationResource)

	s.mux.HandleFunc("GET /compute-allocations/{id}/resources", s.listResourcesForAllocation)
	s.mux.HandleFunc("POST /compute-allocations/{id}/resources", s.attachResourceToAllocation)
	s.mux.HandleFunc("DELETE /compute-allocations/{id}/resources/{resourceId}", s.detachResourceFromAllocation)
	s.mux.HandleFunc("GET /compute-allocation-resources/{id}/allocations", s.listAllocationsForResource)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) createOrganization(w http.ResponseWriter, r *http.Request) {
	var org models.Organization
	if err := decodeJSON(r, &org); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateOrganization(r.Context(), &org)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getOrganization(w http.ResponseWriter, r *http.Request) {
	org, err := s.svc.GetOrganization(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, org)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := decodeJSON(r, &u); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateUser(r.Context(), &u)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	u, err := s.svc.GetUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) createProject(w http.ResponseWriter, r *http.Request) {
	var p models.Project
	if err := decodeJSON(r, &p); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateProject(r.Context(), &p)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	p, err := s.svc.GetProject(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) createComputeCluster(w http.ResponseWriter, r *http.Request) {
	var c models.ComputeCluster
	if err := decodeJSON(r, &c); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeCluster(r.Context(), &c)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getComputeCluster(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.GetComputeCluster(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) listComputeClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := s.svc.ListComputeClusters(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, clusters)
}

func (s *Server) createComputeAllocation(w http.ResponseWriter, r *http.Request) {
	var a models.ComputeAllocation
	if err := decodeJSON(r, &a); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocation(r.Context(), &a)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getComputeAllocation(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetComputeAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func (s *Server) createComputeAllocationResource(w http.ResponseWriter, r *http.Request) {
	var res models.ComputeAllocationResource
	if err := decodeJSON(r, &res); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationResource(r.Context(), &res)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getComputeAllocationResource(w http.ResponseWriter, r *http.Request) {
	res, err := s.svc.GetComputeAllocationResource(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (s *Server) listComputeAllocationResources(w http.ResponseWriter, r *http.Request) {
	resources, err := s.svc.ListComputeAllocationResources(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resources)
}

type attachResourceRequest struct {
	ComputeAllocationResourceID string `json:"compute_allocation_resource_id"`
}

func (s *Server) attachResourceToAllocation(w http.ResponseWriter, r *http.Request) {
	var body attachResourceRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	mapping, err := s.svc.AttachResourceToAllocation(r.Context(), r.PathValue("id"), body.ComputeAllocationResourceID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapping)
}

func (s *Server) detachResourceFromAllocation(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DetachResourceFromAllocation(r.Context(), r.PathValue("id"), r.PathValue("resourceId")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listResourcesForAllocation(w http.ResponseWriter, r *http.Request) {
	resources, err := s.svc.ListResourcesForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resources)
}

func (s *Server) listAllocationsForResource(w http.ResponseWriter, r *http.Request) {
	allocs, err := s.svc.ListAllocationsForResource(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, allocs)
}

// LoggingMiddleware logs every request once it completes.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		slog.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", time.Since(start).String(),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.Is(err, service.ErrAlreadyExists):
		writeError(w, http.StatusConflict, err)
	case errors.Is(err, service.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err)
	default:
		// Avoid leaking driver messages to clients; log the full error.
		slog.Error("internal server error", "error", err.Error())
		writeError(w, http.StatusInternalServerError, errors.New(strings.TrimSpace("internal server error")))
	}
}
