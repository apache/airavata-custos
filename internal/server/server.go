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

func (s *Server) routes() {
	s.router.Public("GET /healthz", s.healthz)

	s.router.RequirePrivilege("POST /organizations", models.OrganizationsWrite, s.createOrganization)
	s.router.RequirePrivilege("GET /organizations", models.OrganizationsRead, s.listOrganizations)
	s.router.RequirePrivilege("GET /organizations/{id}", models.OrganizationsRead, s.getOrganization)

	s.router.RequirePrivilege("POST /users", models.UsersWrite, s.createUser)
	s.router.RequirePrivilege("GET /users", models.UsersRead, s.listUsers)
	s.router.RequirePrivilege("GET /users/{id}", models.UsersRead, s.getUser)
	s.router.RequirePrivilege("PUT /users/{id}", models.UsersWrite, s.updateUser)
	s.router.RequirePrivilege("PUT /users/{id}/status", models.UsersWrite, s.updateUserStatus)
	s.router.RequirePrivilege("POST /users/merge", models.UsersWrite, s.mergeUsers)

	s.router.RequirePrivilege("GET /projects", models.ProjectsRead, s.listProjects)
	s.router.RequirePrivilege("POST /projects", models.ProjectsWrite, s.createProject)
	s.router.RequirePrivilege("GET /projects/{id}", models.ProjectsRead, s.getProject)
	s.router.RequirePrivilege("PUT /projects/{id}/status", models.ProjectsWrite, s.updateProjectStatus)
	s.router.RequirePrivilege("GET /projects/{id}/members", models.ProjectsRead, s.listProjectMembers)

	s.router.RequirePrivilege("POST /compute-clusters", models.ClustersWrite, s.createComputeCluster)
	s.router.RequirePrivilege("GET /compute-clusters", models.ClustersRead, s.listComputeClusters)
	s.router.RequirePrivilege("GET /compute-clusters/{id}", models.ClustersRead, s.getComputeCluster)

	s.router.RequirePrivilege("POST /compute-cluster-users", models.ClustersWrite, s.createComputeClusterUser)
	s.router.RequirePrivilege("GET /compute-cluster-users/{id}", models.ClustersRead, s.getComputeClusterUser)
	s.router.RequirePrivilege("PUT /compute-cluster-users/{id}", models.ClustersWrite, s.updateComputeClusterUser)
	s.router.RequirePrivilege("DELETE /compute-cluster-users/{id}", models.ClustersWrite, s.deleteComputeClusterUser)
	s.router.RequirePrivilege("GET /compute-clusters/{id}/users", models.ClustersRead, s.listComputeClusterUsersByCluster)
	s.router.RequirePrivilege("GET /compute-clusters/{id}/users/{userId}", models.ClustersRead, s.getComputeClusterUserByPair)
	s.router.RequirePrivilege("GET /users/{id}/compute-cluster-users", models.ClustersRead, s.listComputeClusterUsersByUser)

	s.router.RequirePrivilege("GET /compute-allocations", models.AllocationsRead, s.listComputeAllocations)
	s.router.RequirePrivilege("POST /compute-allocations", models.AllocationsWrite, s.createComputeAllocation)
	s.router.RequirePrivilege("GET /compute-allocations/{id}", models.AllocationsRead, s.getComputeAllocation)

	s.router.RequirePrivilege("POST /compute-allocation-resources", models.AllocationsWrite, s.createComputeAllocationResource)
	s.router.RequirePrivilege("GET /compute-allocation-resources", models.AllocationsRead, s.listComputeAllocationResources)
	s.router.RequirePrivilege("GET /compute-allocation-resources/summary", models.AllocationsRead, s.listComputeAllocationResourceSummaries)
	s.router.RequirePrivilege("GET /compute-allocation-resources/{id}", models.AllocationsRead, s.getComputeAllocationResource)

	s.router.RequirePrivilege("GET /compute-allocations/{id}/resources", models.AllocationsRead, s.listResourcesForAllocation)
	s.router.RequirePrivilege("POST /compute-allocations/{id}/resources", models.AllocationsWrite, s.attachResourceToAllocation)
	s.router.RequirePrivilege("PUT /compute-allocations/{id}/resources/{resourceId}", models.AllocationsWrite, s.updateAllocationResourceMapping)
	s.router.RequirePrivilege("DELETE /compute-allocations/{id}/resources/{resourceId}", models.AllocationsWrite, s.detachResourceFromAllocation)
	s.router.RequirePrivilege("GET /compute-allocation-resources/{id}/allocations", models.AllocationsRead, s.listAllocationsForResource)

	s.router.RequirePrivilege("POST /compute-allocation-resource-rates", models.AllocationsWrite, s.createComputeAllocationResourceRate)
	s.router.RequirePrivilege("GET /compute-allocation-resource-rates/{id}", models.AllocationsRead, s.getComputeAllocationResourceRate)
	s.router.RequirePrivilege("GET /compute-allocation-resources/{id}/rates", models.AllocationsRead, s.listRatesForResource)
	s.router.RequirePrivilege("GET /compute-allocation-resources/{id}/rates/effective", models.AllocationsRead, s.getEffectiveRateForResource)

	s.router.RequirePrivilege("POST /compute-allocation-diffs", models.AllocationsWrite, s.createComputeAllocationDiff)
	s.router.RequirePrivilege("GET /compute-allocation-diffs/{id}", models.AllocationsRead, s.getComputeAllocationDiff)
	s.router.RequirePrivilege("DELETE /compute-allocation-diffs/{id}", models.AllocationsWrite, s.deleteComputeAllocationDiff)
	s.router.RequirePrivilege("GET /compute-allocations/{id}/diffs", models.AllocationsRead, s.listDiffsForAllocation)
	s.router.RequirePrivilege("GET /compute-allocations/{id}/diffs/latest", models.AllocationsRead, s.getLatestDiffForAllocation)

	s.router.RequirePrivilege("GET /compute-allocation-change-requests", models.AllocationsRead, s.listChangeRequests)
	s.router.RequirePrivilege("POST /compute-allocation-change-requests", models.AllocationsWrite, s.createComputeAllocationChangeRequest)
	s.router.RequirePrivilege("GET /compute-allocation-change-requests/{id}", models.AllocationsRead, s.getComputeAllocationChangeRequest)
	s.router.RequirePrivilege("PUT /compute-allocation-change-requests/{id}", models.AllocationsWrite, s.updateComputeAllocationChangeRequest)
	s.router.RequirePrivilege("DELETE /compute-allocation-change-requests/{id}", models.AllocationsWrite, s.deleteComputeAllocationChangeRequest)
	s.router.RequirePrivilege("GET /compute-allocations/{id}/change-requests", models.AllocationsRead, s.listChangeRequestsForAllocation)
	s.router.RequirePrivilege("GET /users/{id}/change-requests", models.AllocationsRead, s.listChangeRequestsByRequester)

	s.router.RequirePrivilege("POST /compute-allocation-change-request-events", models.AllocationsWrite, s.createComputeAllocationChangeRequestEvent)
	s.router.RequirePrivilege("GET /compute-allocation-change-request-events/{id}", models.AllocationsRead, s.getComputeAllocationChangeRequestEvent)
	s.router.RequirePrivilege("DELETE /compute-allocation-change-request-events/{id}", models.AllocationsWrite, s.deleteComputeAllocationChangeRequestEvent)
	s.router.RequirePrivilege("GET /compute-allocation-change-requests/{id}/events", models.AllocationsRead, s.listEventsForChangeRequest)
	s.router.RequirePrivilege("GET /compute-allocation-change-requests/{id}/events/latest", models.AllocationsRead, s.getLatestEventForChangeRequest)

	s.router.RequirePrivilege("POST /compute-allocation-memberships", models.AllocationsWrite, s.createComputeAllocationMembership)
	s.router.RequirePrivilege("GET /compute-allocation-memberships/{id}", models.AllocationsRead, s.getComputeAllocationMembership)
	s.router.RequirePrivilege("PUT /compute-allocation-memberships/{id}", models.AllocationsWrite, s.updateComputeAllocationMembership)
	s.router.RequirePrivilege("PUT /compute-allocation-memberships/{id}/status", models.AllocationsWrite, s.updateMembershipStatus)
	s.router.RequirePrivilege("DELETE /compute-allocation-memberships/{id}", models.AllocationsWrite, s.deleteComputeAllocationMembership)
	s.router.RequirePrivilege("GET /compute-allocations/{id}/memberships", models.AllocationsRead, s.listMembersForAllocation)
	s.router.RequirePrivilege("GET /users/{id}/compute-allocation-memberships", models.AllocationsRead, s.listAllocationsForUser)
	s.router.RequirePrivilege("GET /compute-allocation-memberships/{id}/resource-overrides", models.AllocationsRead, s.listOverridesForMembership)

	s.router.RequirePrivilege("POST /compute-allocation-membership-resource-overrides", models.AllocationsWrite, s.createComputeAllocationMembershipResourceOverride)
	s.router.RequirePrivilege("GET /compute-allocation-membership-resource-overrides/{id}", models.AllocationsRead, s.getComputeAllocationMembershipResourceOverride)
	s.router.RequirePrivilege("PUT /compute-allocation-membership-resource-overrides/{id}", models.AllocationsWrite, s.updateComputeAllocationMembershipResourceOverride)
	s.router.RequirePrivilege("DELETE /compute-allocation-membership-resource-overrides/{id}", models.AllocationsWrite, s.deleteComputeAllocationMembershipResourceOverride)
	s.router.RequirePrivilege("GET /compute-allocation-resources/{id}/membership-overrides", models.AllocationsRead, s.listOverridesForResource)

	s.router.RequirePrivilege("POST /compute-allocation-usages", models.AllocationsWrite, s.createComputeAllocationUsage)
	s.router.RequirePrivilege("GET /compute-allocation-usages/{id}", models.AllocationsRead, s.getComputeAllocationUsage)
	s.router.RequirePrivilege("DELETE /compute-allocation-usages/{id}", models.AllocationsWrite, s.deleteComputeAllocationUsage)
	s.router.RequirePrivilege("GET /compute-allocations/{id}/usages", models.AllocationsRead, s.listUsagesForAllocation)
	s.router.RequirePrivilege("GET /compute-allocations/{id}/usages/total", models.AllocationsRead, s.getTotalSUUsageForAllocation)
	s.router.RequirePrivilege("GET /compute-allocations/{id}/users/{userId}/usages/total", models.AllocationsRead, s.getTotalSUUsageForUserInAllocation)
	s.router.RequirePrivilege("GET /users/{id}/compute-allocation-usages", models.AllocationsRead, s.listUsagesByUser)

	s.router.RequirePrivilege("POST /user-identities", models.UsersWrite, s.createUserIdentity)
	s.router.RequirePrivilege("GET /user-identities/{id}", models.UsersRead, s.getUserIdentity)
	s.router.RequirePrivilege("PUT /user-identities/{id}", models.UsersWrite, s.updateUserIdentity)
	s.router.RequirePrivilege("DELETE /user-identities/{id}", models.UsersWrite, s.deleteUserIdentity)
	s.router.RequirePrivilege("GET /user-identities/sources/{source}/external/{externalId}", models.UsersRead, s.getUserIdentityBySourceAndExternalID)
	s.router.RequirePrivilege("GET /user-identities/oidc-subjects/{oidcSub}", models.UsersRead, s.getUserIdentityByOIDCSub)
	s.router.RequirePrivilege("GET /users/{id}/user-identities", models.UsersRead, s.listUserIdentitiesForUser)

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

	s.router.RequirePrivilege("GET /audit/traces", models.TracesRead, s.handleListTraces)
	s.router.RequirePrivilege("GET /audit/traces/{trace_id}", models.TracesRead, s.handleGetTrace)
	s.router.RequirePrivilege("GET /audit/events", models.TracesRead, s.handleListEvents)
	s.router.RequirePrivilege("GET /audit/sources", models.TracesRead, s.handleListSources)
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
