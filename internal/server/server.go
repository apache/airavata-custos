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
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/apache/airavata-custos/internal/httputil"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// Server is an HTTP handler that exposes the service API.
type Server struct {
	svc    *service.Service
	router *identity.Router
}

// New builds an HTTP handler wired to the supplied service. The router owns the mux
// and gates every authenticated route.
func New(svc *service.Service, router *identity.Router) *Server {
	s := &Server{svc: svc, router: router}
	s.routes()
	return s
}

// ServeHTTP satisfies http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// requireCaller pulls the verified caller off the request context. Returns
// nil and writes a 401 when the middleware did not attach one.
func requireCaller(w http.ResponseWriter, r *http.Request) *identity.Caller {
	c := identity.CallerFromContext(r.Context())
	if c == nil {
		common.WriteError(w, http.StatusUnauthorized, errors.New("missing authenticated caller"))
		return nil
	}
	return c
}

// TODO(auth): tighten RequireAuth routes to RequirePrivilege as new keys land.
func (s *Server) routes() {
	s.router.Public("GET /healthz", s.healthz)

	s.router.RequireAuth("POST /organizations", s.createOrganization)
	s.router.RequireAuth("GET /organizations/{id}", s.getOrganization)

	s.router.RequireAuth("POST /users", s.createUser)
	s.router.RequireAuth("GET /users/{id}", s.getUser)
	s.router.RequireAuth("PUT /users/{id}/status", s.updateUserStatus)
	s.router.RequireAuth("POST /users/merge", s.mergeUsers)

	s.router.RequireAuth("GET /projects", s.listProjects)
	s.router.RequireAuth("POST /projects", s.createProject)
	s.router.RequireAuth("GET /projects/{id}", s.getProject)
	s.router.RequireAuth("PUT /projects/{id}/status", s.updateProjectStatus)
	s.router.RequireAuth("GET /projects/{id}/members", s.listProjectMembers)

	s.router.RequirePrivilege("POST /compute-clusters", models.ClustersWrite, s.createComputeCluster)
	s.router.RequirePrivilege("GET /compute-clusters", models.ClustersRead, s.listComputeClusters)
	s.router.RequirePrivilege("GET /compute-clusters/{id}", models.ClustersRead, s.getComputeCluster)

	s.router.RequireAuth("POST /compute-cluster-users", s.createComputeClusterUser)
	s.router.RequireAuth("GET /compute-cluster-users/{id}", s.getComputeClusterUser)
	s.router.RequireAuth("PUT /compute-cluster-users/{id}", s.updateComputeClusterUser)
	s.router.RequireAuth("DELETE /compute-cluster-users/{id}", s.deleteComputeClusterUser)
	s.router.RequirePrivilege("GET /compute-clusters/{id}/users", models.ClustersRead, s.listComputeClusterUsersByCluster)
	s.router.RequirePrivilege("GET /compute-clusters/{id}/users/{userId}", models.ClustersRead, s.getComputeClusterUserByPair)
	s.router.RequireAuth("GET /users/{id}/compute-cluster-users", s.listComputeClusterUsersByUser)

	s.router.RequireAuth("GET /compute-allocations", s.listComputeAllocations)
	s.router.RequireAuth("POST /compute-allocations", s.createComputeAllocation)
	s.router.RequireAuth("GET /compute-allocations/{id}", s.getComputeAllocation)

	s.router.RequireAuth("POST /compute-allocation-resources", s.createComputeAllocationResource)
	s.router.RequireAuth("GET /compute-allocation-resources", s.listComputeAllocationResources)
	s.router.RequireAuth("GET /compute-allocation-resources/{id}", s.getComputeAllocationResource)

	s.router.RequireAuth("GET /compute-allocations/{id}/resources", s.listResourcesForAllocation)
	s.router.RequireAuth("POST /compute-allocations/{id}/resources", s.attachResourceToAllocation)
	s.router.RequireAuth("PUT /compute-allocations/{id}/resources/{resourceId}", s.updateAllocationResourceMapping)
	s.router.RequireAuth("DELETE /compute-allocations/{id}/resources/{resourceId}", s.detachResourceFromAllocation)
	s.router.RequireAuth("GET /compute-allocation-resources/{id}/allocations", s.listAllocationsForResource)

	s.router.RequireAuth("POST /compute-allocation-resource-rates", s.createComputeAllocationResourceRate)
	s.router.RequireAuth("GET /compute-allocation-resource-rates/{id}", s.getComputeAllocationResourceRate)
	s.router.RequireAuth("GET /compute-allocation-resources/{id}/rates", s.listRatesForResource)
	s.router.RequireAuth("GET /compute-allocation-resources/{id}/rates/effective", s.getEffectiveRateForResource)

	s.router.RequireAuth("POST /compute-allocation-diffs", s.createComputeAllocationDiff)
	s.router.RequireAuth("GET /compute-allocation-diffs/{id}", s.getComputeAllocationDiff)
	s.router.RequireAuth("DELETE /compute-allocation-diffs/{id}", s.deleteComputeAllocationDiff)
	s.router.RequireAuth("GET /compute-allocations/{id}/diffs", s.listDiffsForAllocation)
	s.router.RequireAuth("GET /compute-allocations/{id}/diffs/latest", s.getLatestDiffForAllocation)

	s.router.RequireAuth("GET /compute-allocation-change-requests", s.listChangeRequests)
	s.router.RequireAuth("POST /compute-allocation-change-requests", s.createComputeAllocationChangeRequest)
	s.router.RequireAuth("GET /compute-allocation-change-requests/{id}", s.getComputeAllocationChangeRequest)
	s.router.RequireAuth("PUT /compute-allocation-change-requests/{id}", s.updateComputeAllocationChangeRequest)
	s.router.RequireAuth("DELETE /compute-allocation-change-requests/{id}", s.deleteComputeAllocationChangeRequest)
	s.router.RequireAuth("GET /compute-allocations/{id}/change-requests", s.listChangeRequestsForAllocation)
	s.router.RequireAuth("GET /users/{id}/change-requests", s.listChangeRequestsByRequester)

	s.router.RequireAuth("POST /compute-allocation-change-request-events", s.createComputeAllocationChangeRequestEvent)
	s.router.RequireAuth("GET /compute-allocation-change-request-events/{id}", s.getComputeAllocationChangeRequestEvent)
	s.router.RequireAuth("DELETE /compute-allocation-change-request-events/{id}", s.deleteComputeAllocationChangeRequestEvent)
	s.router.RequireAuth("GET /compute-allocation-change-requests/{id}/events", s.listEventsForChangeRequest)
	s.router.RequireAuth("GET /compute-allocation-change-requests/{id}/events/latest", s.getLatestEventForChangeRequest)

	s.router.RequireAuth("POST /compute-allocation-memberships", s.createComputeAllocationMembership)
	s.router.RequireAuth("GET /compute-allocation-memberships/{id}", s.getComputeAllocationMembership)
	s.router.RequireAuth("PUT /compute-allocation-memberships/{id}", s.updateComputeAllocationMembership)
	s.router.RequireAuth("PUT /compute-allocation-memberships/{id}/status", s.updateMembershipStatus)
	s.router.RequireAuth("DELETE /compute-allocation-memberships/{id}", s.deleteComputeAllocationMembership)
	s.router.RequireAuth("GET /compute-allocations/{id}/memberships", s.listMembersForAllocation)
	s.router.RequireAuth("GET /users/{id}/compute-allocation-memberships", s.listAllocationsForUser)
	s.router.RequireAuth("GET /compute-allocation-memberships/{id}/resource-overrides", s.listOverridesForMembership)

	s.router.RequireAuth("POST /compute-allocation-membership-resource-overrides", s.createComputeAllocationMembershipResourceOverride)
	s.router.RequireAuth("GET /compute-allocation-membership-resource-overrides/{id}", s.getComputeAllocationMembershipResourceOverride)
	s.router.RequireAuth("PUT /compute-allocation-membership-resource-overrides/{id}", s.updateComputeAllocationMembershipResourceOverride)
	s.router.RequireAuth("DELETE /compute-allocation-membership-resource-overrides/{id}", s.deleteComputeAllocationMembershipResourceOverride)
	s.router.RequireAuth("GET /compute-allocation-resources/{id}/membership-overrides", s.listOverridesForResource)

	s.router.RequireAuth("POST /compute-allocation-usages", s.createComputeAllocationUsage)
	s.router.RequireAuth("GET /compute-allocation-usages/{id}", s.getComputeAllocationUsage)
	s.router.RequireAuth("DELETE /compute-allocation-usages/{id}", s.deleteComputeAllocationUsage)
	s.router.RequireAuth("GET /compute-allocations/{id}/usages", s.listUsagesForAllocation)
	s.router.RequireAuth("GET /compute-allocations/{id}/usages/total", s.getTotalSUUsageForAllocation)
	s.router.RequireAuth("GET /compute-allocations/{id}/users/{userId}/usages/total", s.getTotalSUUsageForUserInAllocation)
	s.router.RequireAuth("GET /users/{id}/compute-allocation-usages", s.listUsagesByUser)

	s.router.RequireAuth("POST /user-identities", s.createUserIdentity)
	s.router.RequireAuth("GET /user-identities/{id}", s.getUserIdentity)
	s.router.RequireAuth("PUT /user-identities/{id}", s.updateUserIdentity)
	s.router.RequireAuth("DELETE /user-identities/{id}", s.deleteUserIdentity)
	s.router.RequireAuth("GET /user-identities/sources/{source}/external/{externalId}", s.getUserIdentityBySourceAndExternalID)
	s.router.RequireAuth("GET /user-identities/oidc-subjects/{oidcSub}", s.getUserIdentityByOIDCSub)
	s.router.RequireAuth("GET /users/{id}/user-identities", s.listUserIdentitiesForUser)

	// Any authenticated caller may read their own profile and effective privilege set with no privilege check.
	s.router.RequireAuth("GET /user/privileges", s.getCallerPrivileges)
	s.router.RequireAuth("GET /me", s.getCallerProfile)
	s.router.RequirePrivilege("GET /privileges/catalog", models.PrivilegesGrant, s.getPrivilegeCatalog)
	s.router.RequirePrivilege("GET /users/{id}/privileges", models.PrivilegesGrant, s.listUserPrivileges)
	s.router.RequirePrivilege("GET /privileges/{key}/holders", models.PrivilegesGrant, s.listPrivilegeHolders)
	s.router.RequirePrivilege("POST /users/{id}/privileges", models.PrivilegesGrant, s.grantPrivilege)
	s.router.RequirePrivilege("DELETE /users/{id}/privileges/{key}", models.PrivilegesGrant, s.revokePrivilege)

	s.router.RequirePrivilege("GET /roles", models.RolesManage, s.listRoles)
	s.router.RequirePrivilege("POST /roles", models.RolesManage, s.createRole)
	s.router.RequirePrivilege("GET /roles/{id}", models.RolesManage, s.getRole)
	s.router.RequirePrivilege("PUT /roles/{id}", models.RolesManage, s.updateRole)
	s.router.RequirePrivilege("DELETE /roles/{id}", models.RolesManage, s.deleteRole)
	s.router.RequirePrivilege("POST /roles/{id}/privileges", models.RolesManage, s.addRolePrivilege)
	s.router.RequirePrivilege("DELETE /roles/{id}/privileges/{key}", models.RolesManage, s.removeRolePrivilege)
	s.router.RequirePrivilege("GET /roles/{id}/holders", models.RolesManage, s.listRoleHolders)
	s.router.RequirePrivilege("GET /users/{id}/roles", models.RolesManage, s.listUserRoles)
	s.router.RequirePrivilege("POST /users/{id}/roles", models.RolesManage, s.grantRoleToUser)
	s.router.RequirePrivilege("DELETE /users/{id}/roles/{roleId}", models.RolesManage, s.revokeRoleFromUser)

	// TODO(auth): introduce audit-read and switch to router.RequirePrivilege.
	s.router.RequireAuth("GET /audit/traces", s.handleListTraces)
	s.router.RequireAuth("GET /audit/traces/{trace_id}", s.handleGetTrace)
	s.router.RequireAuth("GET /audit/events", s.handleListEvents)
	s.router.RequireAuth("GET /audit/sources", s.handleListSources)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	common.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

type statusUpdateRequest struct {
	Status string `json:"status"`
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

func atoiOr(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
