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

	"github.com/apache/airavata-custos/internal/httputil"
	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// AdminDeps wires the /audit/* endpoints. A nil field makes its routes return 503.
// Connector endpoints (/connectors/{name}/...) are registered by the connector.
type AdminDeps struct {
	AuditTraces store.AuditTraceStore
}

// Server is an HTTP handler that exposes the service API.
type Server struct {
	svc       *service.Service
	mux       *http.ServeMux
	authCache *authProfileCache
	admin     *AdminDeps
}

// New builds an HTTP handler wired to the supplied service. admin may be nil
// to disable the admin routes.
func New(svc *service.Service, admin *AdminDeps) *Server {
	s := &Server{
		svc:       svc,
		mux:       http.NewServeMux(),
		authCache: newAuthProfileCache(authProfileTTL),
		admin:     admin,
	}
	s.routes()
	return s
}

// ServeHTTP satisfies http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// Mux exposes the underlying mux so connectors can register their own routes.
func (s *Server) Mux() *http.ServeMux { return s.mux }

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.healthz)

	s.mux.HandleFunc("POST /organizations", s.createOrganization)
	s.mux.HandleFunc("GET /organizations/{id}", s.getOrganization)

	s.mux.HandleFunc("POST /users", s.createUser)
	s.mux.HandleFunc("GET /users/{id}", s.getUser)
	s.mux.HandleFunc("PUT /users/{id}/status", s.updateUserStatus)
	s.mux.HandleFunc("POST /users/merge", s.mergeUsers)

	s.mux.HandleFunc("POST /projects", s.createProject)
	s.mux.HandleFunc("GET /projects/{id}", s.getProject)
	s.mux.HandleFunc("PUT /projects/{id}/status", s.updateProjectStatus)

	s.mux.HandleFunc("POST /compute-clusters", s.createComputeCluster)
	s.mux.HandleFunc("GET /compute-clusters", s.listComputeClusters)
	s.mux.HandleFunc("GET /compute-clusters/{id}", s.getComputeCluster)

	s.mux.HandleFunc("POST /compute-cluster-users", s.createComputeClusterUser)
	s.mux.HandleFunc("GET /compute-cluster-users/{id}", s.getComputeClusterUser)
	s.mux.HandleFunc("PUT /compute-cluster-users/{id}", s.updateComputeClusterUser)
	s.mux.HandleFunc("DELETE /compute-cluster-users/{id}", s.deleteComputeClusterUser)
	s.mux.HandleFunc("GET /compute-clusters/{id}/users", s.listComputeClusterUsersByCluster)
	s.mux.HandleFunc("GET /compute-clusters/{id}/users/{userId}", s.getComputeClusterUserByPair)
	s.mux.HandleFunc("GET /users/{id}/compute-cluster-users", s.listComputeClusterUsersByUser)

	s.mux.HandleFunc("POST /compute-allocations", s.createComputeAllocation)
	s.mux.HandleFunc("GET /compute-allocations/{id}", s.getComputeAllocation)

	s.mux.HandleFunc("POST /compute-allocation-resources", s.createComputeAllocationResource)
	s.mux.HandleFunc("GET /compute-allocation-resources", s.listComputeAllocationResources)
	s.mux.HandleFunc("GET /compute-allocation-resources/{id}", s.getComputeAllocationResource)

	s.mux.HandleFunc("GET /compute-allocations/{id}/resources", s.listResourcesForAllocation)
	s.mux.HandleFunc("POST /compute-allocations/{id}/resources", s.attachResourceToAllocation)
	s.mux.HandleFunc("PUT /compute-allocations/{id}/resources/{resourceId}", s.updateAllocationResourceMapping)
	s.mux.HandleFunc("DELETE /compute-allocations/{id}/resources/{resourceId}", s.detachResourceFromAllocation)
	s.mux.HandleFunc("GET /compute-allocation-resources/{id}/allocations", s.listAllocationsForResource)

	s.mux.HandleFunc("POST /compute-allocation-resource-rates", s.createComputeAllocationResourceRate)
	s.mux.HandleFunc("GET /compute-allocation-resource-rates/{id}", s.getComputeAllocationResourceRate)
	s.mux.HandleFunc("GET /compute-allocation-resources/{id}/rates", s.listRatesForResource)
	s.mux.HandleFunc("GET /compute-allocation-resources/{id}/rates/effective", s.getEffectiveRateForResource)

	s.mux.HandleFunc("POST /compute-allocation-diffs", s.createComputeAllocationDiff)
	s.mux.HandleFunc("GET /compute-allocation-diffs/{id}", s.getComputeAllocationDiff)
	s.mux.HandleFunc("DELETE /compute-allocation-diffs/{id}", s.deleteComputeAllocationDiff)
	s.mux.HandleFunc("GET /compute-allocations/{id}/diffs", s.listDiffsForAllocation)
	s.mux.HandleFunc("GET /compute-allocations/{id}/diffs/latest", s.getLatestDiffForAllocation)

	s.mux.HandleFunc("POST /compute-allocation-change-requests", s.createComputeAllocationChangeRequest)
	s.mux.HandleFunc("GET /compute-allocation-change-requests/{id}", s.getComputeAllocationChangeRequest)
	s.mux.HandleFunc("PUT /compute-allocation-change-requests/{id}", s.updateComputeAllocationChangeRequest)
	s.mux.HandleFunc("DELETE /compute-allocation-change-requests/{id}", s.deleteComputeAllocationChangeRequest)
	s.mux.HandleFunc("GET /compute-allocations/{id}/change-requests", s.listChangeRequestsForAllocation)
	s.mux.HandleFunc("GET /users/{id}/change-requests", s.listChangeRequestsByRequester)

	s.mux.HandleFunc("POST /compute-allocation-change-request-events", s.createComputeAllocationChangeRequestEvent)
	s.mux.HandleFunc("GET /compute-allocation-change-request-events/{id}", s.getComputeAllocationChangeRequestEvent)
	s.mux.HandleFunc("DELETE /compute-allocation-change-request-events/{id}", s.deleteComputeAllocationChangeRequestEvent)
	s.mux.HandleFunc("GET /compute-allocation-change-requests/{id}/events", s.listEventsForChangeRequest)
	s.mux.HandleFunc("GET /compute-allocation-change-requests/{id}/events/latest", s.getLatestEventForChangeRequest)

	s.mux.HandleFunc("POST /compute-allocation-memberships", s.createComputeAllocationMembership)
	s.mux.HandleFunc("GET /compute-allocation-memberships/{id}", s.getComputeAllocationMembership)
	s.mux.HandleFunc("PUT /compute-allocation-memberships/{id}", s.updateComputeAllocationMembership)
	s.mux.HandleFunc("PUT /compute-allocation-memberships/{id}/status", s.updateMembershipStatus)
	s.mux.HandleFunc("DELETE /compute-allocation-memberships/{id}", s.deleteComputeAllocationMembership)
	s.mux.HandleFunc("GET /compute-allocations/{id}/memberships", s.listMembersForAllocation)
	s.mux.HandleFunc("GET /users/{id}/compute-allocation-memberships", s.listAllocationsForUser)
	s.mux.HandleFunc("GET /compute-allocation-memberships/{id}/resource-overrides", s.listOverridesForMembership)

	s.mux.HandleFunc("POST /compute-allocation-membership-resource-overrides", s.createComputeAllocationMembershipResourceOverride)
	s.mux.HandleFunc("GET /compute-allocation-membership-resource-overrides/{id}", s.getComputeAllocationMembershipResourceOverride)
	s.mux.HandleFunc("PUT /compute-allocation-membership-resource-overrides/{id}", s.updateComputeAllocationMembershipResourceOverride)
	s.mux.HandleFunc("DELETE /compute-allocation-membership-resource-overrides/{id}", s.deleteComputeAllocationMembershipResourceOverride)
	s.mux.HandleFunc("GET /compute-allocation-resources/{id}/membership-overrides", s.listOverridesForResource)

	s.mux.HandleFunc("POST /compute-allocation-usages", s.createComputeAllocationUsage)
	s.mux.HandleFunc("GET /compute-allocation-usages/{id}", s.getComputeAllocationUsage)
	s.mux.HandleFunc("DELETE /compute-allocation-usages/{id}", s.deleteComputeAllocationUsage)
	s.mux.HandleFunc("GET /compute-allocations/{id}/usages", s.listUsagesForAllocation)
	s.mux.HandleFunc("GET /compute-allocations/{id}/usages/total", s.getTotalSUUsageForAllocation)
	s.mux.HandleFunc("GET /compute-allocations/{id}/users/{userId}/usages/total", s.getTotalSUUsageForUserInAllocation)
	s.mux.HandleFunc("GET /users/{id}/compute-allocation-usages", s.listUsagesByUser)

	s.mux.HandleFunc("POST /user-identities", s.createUserIdentity)
	s.mux.HandleFunc("GET /user-identities/{id}", s.getUserIdentity)
	s.mux.HandleFunc("PUT /user-identities/{id}", s.updateUserIdentity)
	s.mux.HandleFunc("DELETE /user-identities/{id}", s.deleteUserIdentity)
	s.mux.HandleFunc("GET /user-identities/sources/{source}/external/{externalId}", s.getUserIdentityBySourceAndExternalID)
	s.mux.HandleFunc("GET /user-identities/oidc-subjects/{oidcSub}", s.getUserIdentityByOIDCSub)
	s.mux.HandleFunc("GET /users/{id}/user-identities", s.listUserIdentitiesForUser)

	s.mux.HandleFunc("GET /user/privileges", s.getCallerPrivileges)
	s.mux.HandleFunc("GET /privileges/catalog", s.requirePrivilege(models.PrivilegeGrant, s.getPrivilegeCatalog))
	s.mux.HandleFunc("GET /users/{id}/privileges", s.requirePrivilege(models.PrivilegeGrant, s.listUserPrivileges))
	s.mux.HandleFunc("GET /privileges/{key}/holders", s.requirePrivilege(models.PrivilegeGrant, s.listPrivilegeHolders))
	s.mux.HandleFunc("POST /users/{id}/privileges", s.requirePrivilege(models.PrivilegeGrant, s.grantPrivilege))
	s.mux.HandleFunc("DELETE /users/{id}/privileges/{key}", s.requirePrivilege(models.PrivilegeGrant, s.revokePrivilege))

	s.mux.HandleFunc("GET /roles", s.requirePrivilege(models.PrivilegeRolesManage, s.listRoles))
	s.mux.HandleFunc("POST /roles", s.requirePrivilege(models.PrivilegeRolesManage, s.createRole))
	s.mux.HandleFunc("GET /roles/{id}", s.requirePrivilege(models.PrivilegeRolesManage, s.getRole))
	s.mux.HandleFunc("PUT /roles/{id}", s.requirePrivilege(models.PrivilegeRolesManage, s.updateRole))
	s.mux.HandleFunc("DELETE /roles/{id}", s.requirePrivilege(models.PrivilegeRolesManage, s.deleteRole))
	s.mux.HandleFunc("POST /roles/{id}/privileges", s.requirePrivilege(models.PrivilegeRolesManage, s.addRolePrivilege))
	s.mux.HandleFunc("DELETE /roles/{id}/privileges/{key}", s.requirePrivilege(models.PrivilegeRolesManage, s.removeRolePrivilege))
	s.mux.HandleFunc("GET /roles/{id}/holders", s.requirePrivilege(models.PrivilegeRolesManage, s.listRoleHolders))
	s.mux.HandleFunc("GET /users/{id}/roles", s.requirePrivilege(models.PrivilegeRolesManage, s.listUserRoles))
	s.mux.HandleFunc("POST /users/{id}/roles", s.requirePrivilege(models.PrivilegeRolesManage, s.grantRoleToUser))
	s.mux.HandleFunc("DELETE /users/{id}/roles/{roleId}", s.requirePrivilege(models.PrivilegeRolesManage, s.revokeRoleFromUser))

	s.mux.HandleFunc("GET /audit/traces", s.handleListTraces)
	s.mux.HandleFunc("GET /audit/traces/{trace_id}", s.handleGetTrace)
	s.mux.HandleFunc("GET /audit/events", s.handleListEvents)
	s.mux.HandleFunc("GET /audit/sources", s.handleListSources)
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

func (s *Server) createComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	var cu models.ComputeClusterUser
	if err := decodeJSON(r, &cu); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeClusterUser(r.Context(), &cu)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	cu, err := s.svc.GetComputeClusterUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cu)
}

func (s *Server) updateComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	var cu models.ComputeClusterUser
	if err := decodeJSON(r, &cu); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	cu.ID = r.PathValue("id")
	if err := s.svc.UpdateComputeClusterUser(r.Context(), &cu); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, &cu)
}

func (s *Server) deleteComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeClusterUser(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listComputeClusterUsersByCluster(w http.ResponseWriter, r *http.Request) {
	users, err := s.svc.ListComputeClusterUsersByCluster(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (s *Server) getComputeClusterUserByPair(w http.ResponseWriter, r *http.Request) {
	cu, err := s.svc.GetComputeClusterUserByPair(r.Context(), r.PathValue("id"), r.PathValue("userId"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cu)
}

func (s *Server) listComputeClusterUsersByUser(w http.ResponseWriter, r *http.Request) {
	users, err := s.svc.ListComputeClusterUsersByUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
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
	ResourceAmount              int64  `json:"resource_amount"`
	ResourceTime                int64  `json:"resource_time"`
}

func (s *Server) attachResourceToAllocation(w http.ResponseWriter, r *http.Request) {
	var body attachResourceRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	mapping, err := s.svc.AttachResourceToAllocation(r.Context(), r.PathValue("id"), body.ComputeAllocationResourceID, body.ResourceAmount, body.ResourceTime)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, mapping)
}

type updateAllocationResourceMappingRequest struct {
	ResourceAmount int64 `json:"resource_amount"`
	ResourceTime   int64 `json:"resource_time"`
}

func (s *Server) updateAllocationResourceMapping(w http.ResponseWriter, r *http.Request) {
	var body updateAllocationResourceMappingRequest
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	mapping, err := s.svc.UpdateAllocationResourceMapping(r.Context(), r.PathValue("id"), r.PathValue("resourceId"), body.ResourceAmount, body.ResourceTime)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mapping)
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

func (s *Server) createComputeAllocationResourceRate(w http.ResponseWriter, r *http.Request) {
	var rate models.ComputeAllocationResourceRate
	if err := decodeJSON(r, &rate); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationResourceRate(r.Context(), &rate)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getComputeAllocationResourceRate(w http.ResponseWriter, r *http.Request) {
	rate, err := s.svc.GetComputeAllocationResourceRate(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rate)
}

func (s *Server) listRatesForResource(w http.ResponseWriter, r *http.Request) {
	rates, err := s.svc.ListRatesForResource(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rates)
}

func (s *Server) getEffectiveRateForResource(w http.ResponseWriter, r *http.Request) {
	var at time.Time
	if raw := r.URL.Query().Get("at"); raw != "" {
		parsed, err := time.Parse(time.RFC3339Nano, raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, errors.New("invalid 'at' query parameter; expected RFC 3339"))
			return
		}
		at = parsed
	}
	rate, err := s.svc.GetEffectiveRateForResource(r.Context(), r.PathValue("id"), at)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rate)
}

func (s *Server) createComputeAllocationDiff(w http.ResponseWriter, r *http.Request) {
	var diff models.ComputeAllocationDiff
	if err := decodeJSON(r, &diff); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationDiff(r.Context(), &diff)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getComputeAllocationDiff(w http.ResponseWriter, r *http.Request) {
	diff, err := s.svc.GetComputeAllocationDiff(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, diff)
}

func (s *Server) deleteComputeAllocationDiff(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationDiff(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listDiffsForAllocation(w http.ResponseWriter, r *http.Request) {
	diffs, err := s.svc.ListDiffsForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, diffs)
}

func (s *Server) getLatestDiffForAllocation(w http.ResponseWriter, r *http.Request) {
	diff, err := s.svc.GetLatestDiffForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, diff)
}

func (s *Server) createComputeAllocationChangeRequest(w http.ResponseWriter, r *http.Request) {
	var req models.ComputeAllocationChangeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationChangeRequest(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getComputeAllocationChangeRequest(w http.ResponseWriter, r *http.Request) {
	req, err := s.svc.GetComputeAllocationChangeRequest(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, req)
}

func (s *Server) updateComputeAllocationChangeRequest(w http.ResponseWriter, r *http.Request) {
	var req models.ComputeAllocationChangeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	req.ID = r.PathValue("id")
	updated, err := s.svc.UpdateComputeAllocationChangeRequest(r.Context(), &req)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteComputeAllocationChangeRequest(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationChangeRequest(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listChangeRequestsForAllocation(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListChangeRequestsForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) listChangeRequestsByRequester(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListChangeRequestsByRequester(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) createComputeAllocationChangeRequestEvent(w http.ResponseWriter, r *http.Request) {
	var evt models.ComputeAllocationChangeRequestEvent
	if err := decodeJSON(r, &evt); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationChangeRequestEvent(r.Context(), &evt)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getComputeAllocationChangeRequestEvent(w http.ResponseWriter, r *http.Request) {
	evt, err := s.svc.GetComputeAllocationChangeRequestEvent(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, evt)
}

func (s *Server) deleteComputeAllocationChangeRequestEvent(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationChangeRequestEvent(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listEventsForChangeRequest(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListEventsForChangeRequest(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) getLatestEventForChangeRequest(w http.ResponseWriter, r *http.Request) {
	evt, err := s.svc.GetLatestEventForChangeRequest(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, evt)
}

func (s *Server) createComputeAllocationMembership(w http.ResponseWriter, r *http.Request) {
	var m models.ComputeAllocationMembership
	if err := decodeJSON(r, &m); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationMembership(r.Context(), &m)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getComputeAllocationMembership(w http.ResponseWriter, r *http.Request) {
	m, err := s.svc.GetComputeAllocationMembership(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (s *Server) updateComputeAllocationMembership(w http.ResponseWriter, r *http.Request) {
	var m models.ComputeAllocationMembership
	if err := decodeJSON(r, &m); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	m.ID = r.PathValue("id")
	updated, err := s.svc.UpdateComputeAllocationMembership(r.Context(), &m)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) updateMembershipStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MembershipStatus models.AllocationStatus `json:"membership_status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	updated, err := s.svc.UpdateMembershipStatus(r.Context(), r.PathValue("id"), body.MembershipStatus)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteComputeAllocationMembership(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationMembership(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listMembersForAllocation(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListMembersForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) listAllocationsForUser(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListAllocationsForUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) createComputeAllocationUsage(w http.ResponseWriter, r *http.Request) {
	var u models.ComputeAllocationUsage
	if err := decodeJSON(r, &u); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationUsage(r.Context(), &u)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getComputeAllocationUsage(w http.ResponseWriter, r *http.Request) {
	u, err := s.svc.GetComputeAllocationUsage(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) deleteComputeAllocationUsage(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationUsage(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listUsagesForAllocation(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListUsagesForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) listUsagesByUser(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListUsagesByUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) createComputeAllocationMembershipResourceOverride(w http.ResponseWriter, r *http.Request) {
	var o models.ComputeAllocationMembershipResourceOverride
	if err := decodeJSON(r, &o); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationMembershipResourceOverride(r.Context(), &o)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getComputeAllocationMembershipResourceOverride(w http.ResponseWriter, r *http.Request) {
	o, err := s.svc.GetComputeAllocationMembershipResourceOverride(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (s *Server) updateComputeAllocationMembershipResourceOverride(w http.ResponseWriter, r *http.Request) {
	var o models.ComputeAllocationMembershipResourceOverride
	if err := decodeJSON(r, &o); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	o.ID = r.PathValue("id")
	updated, err := s.svc.UpdateComputeAllocationMembershipResourceOverride(r.Context(), &o)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteComputeAllocationMembershipResourceOverride(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationMembershipResourceOverride(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listOverridesForMembership(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListOverridesForMembership(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) listOverridesForResource(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListOverridesForResource(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) getTotalSUUsageForAllocation(w http.ResponseWriter, r *http.Request) {
	total, err := s.svc.GetTotalSUUsageForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"compute_allocation_id": r.PathValue("id"),
		"total_su_amount":       total,
	})
}

func (s *Server) getTotalSUUsageForUserInAllocation(w http.ResponseWriter, r *http.Request) {
	allocationID := r.PathValue("id")
	userID := r.PathValue("userId")
	total, err := s.svc.GetTotalSUUsageForUserInAllocation(r.Context(), allocationID, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"compute_allocation_id": allocationID,
		"user_id":               userID,
		"total_su_amount":       total,
	})
}

type statusUpdateRequest struct {
	Status string `json:"status"`
}

func (s *Server) updateUserStatus(w http.ResponseWriter, r *http.Request) {
	var req statusUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	u, err := s.svc.UpdateUserStatus(r.Context(), r.PathValue("id"), models.UserStatus(req.Status))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) updateProjectStatus(w http.ResponseWriter, r *http.Request) {
	var req statusUpdateRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	p, err := s.svc.UpdateProjectStatus(r.Context(), r.PathValue("id"), models.ProjectStatus(req.Status))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *Server) createUserIdentity(w http.ResponseWriter, r *http.Request) {
	var e models.UserIdentity
	if err := decodeJSON(r, &e); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateUserIdentity(r.Context(), &e)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) getUserIdentity(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetUserIdentity(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) getUserIdentityBySourceAndExternalID(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetUserIdentityBySourceAndExternalID(r.Context(), r.PathValue("source"), r.PathValue("externalId"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) getUserIdentityByOIDCSub(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetUserIdentityByOIDCSub(r.Context(), r.PathValue("oidcSub"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) listUserIdentitiesForUser(w http.ResponseWriter, r *http.Request) {
	out, err := s.svc.ListUserIdentitiesForUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) updateUserIdentity(w http.ResponseWriter, r *http.Request) {
	var e models.UserIdentity
	if err := decodeJSON(r, &e); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	e.ID = r.PathValue("id")
	if err := s.svc.UpdateUserIdentity(r.Context(), &e); err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, &e)
}

func (s *Server) deleteUserIdentity(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteUserIdentity(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type mergeUsersRequest struct {
	SurvivingUserID string `json:"surviving_user_id"`
	RetiringUserID  string `json:"retiring_user_id"`
}

func (s *Server) mergeUsers(w http.ResponseWriter, r *http.Request) {
	var req mergeUsersRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	survivor, err := s.svc.MergeUsers(r.Context(), req.SurvivingUserID, req.RetiringUserID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, survivor)
}

// LoggingMiddleware logs every request once it completes. It wraps
// tracing.Middleware (logging outer, tracing inner) and reads the trace_id
// from the X-Trace-Id response header the inner span set, since the inner
// request ctx is not visible at this scope after ServeHTTP returns.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &httputil.StatusRecorder{ResponseWriter: w, Status: http.StatusOK}
		next.ServeHTTP(rw, r)
		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.Status,
			"duration", time.Since(start).String(),
		}
		if tid := rw.Header().Get("X-Trace-Id"); tid != "" {
			attrs = append(attrs, "trace_id", tid)
		}
		slog.InfoContext(r.Context(), "http request", attrs...)
	})
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
