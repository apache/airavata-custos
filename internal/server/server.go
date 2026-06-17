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
	"strconv"
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

	s.mux.HandleFunc("GET /projects", s.listProjects)
	s.mux.HandleFunc("POST /projects", s.createProject)
	s.mux.HandleFunc("GET /projects/{id}", s.getProject)
	s.mux.HandleFunc("PUT /projects/{id}/status", s.updateProjectStatus)
	s.mux.HandleFunc("GET /projects/{id}/members", s.listProjectMembers)

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

	s.mux.HandleFunc("GET /compute-allocations", s.listComputeAllocations)
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

	s.mux.HandleFunc("GET /compute-allocation-change-requests", s.listChangeRequests)
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

// @Summary	Create an organization
// @Tags	Organizations
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.Organization	true	"Organization payload"
// @Success	201	{object}	models.Organization
// @Failure	400	{object}	object{error=string}
// @Router	/organizations [post]
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

// @Summary	Get an organization by ID
// @Tags	Organizations
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Organization ID"
// @Success	200	{object}	models.Organization
// @Failure	404	{object}	object{error=string}
// @Router	/organizations/{id} [get]
func (s *Server) getOrganization(w http.ResponseWriter, r *http.Request) {
	org, err := s.svc.GetOrganization(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, org)
}

// @Summary	Create a user
// @Tags	Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.User	true	"User payload"
// @Success	201	{object}	models.User
// @Failure	400	{object}	object{error=string}
// @Router	/users [post]
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

// @Summary	Get a user by ID
// @Tags	Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{object}	models.User
// @Failure	404	{object}	object{error=string}
// @Router	/users/{id} [get]
func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	u, err := s.svc.GetUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// @Summary	Create a project
// @Tags	Projects
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.Project	true	"Project payload"
// @Success	201	{object}	models.Project
// @Failure	400	{object}	object{error=string}
// @Router	/projects [post]
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

// @Summary	Get a project by ID
// @Tags	Projects
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Project ID"
// @Success	200	{object}	ProjectResponse
// @Failure	404	{object}	object{error=string}
// @Router	/projects/{id} [get]
func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	p, err := s.svc.GetProjectWithPI(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, projectResponseFrom(p))
}

// projectResponseFrom builds the API response shape from a JOIN-fetched row.
func projectResponseFrom(p *store.ProjectWithPI) ProjectResponse {
	v := ProjectResponse{Project: p.Project, ProjectPIEmail: p.PIEmail}
	if full := strings.TrimSpace(p.PIFirstName + " " + p.PILastName); full != "" {
		v.ProjectPIDisplayName = full
	} else if p.PIEmail != "" {
		v.ProjectPIDisplayName = p.PIEmail
	}
	return v
}

// @Summary	Create a compute cluster
// @Tags	Compute Clusters
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeCluster	true	"Cluster payload"
// @Success	201	{object}	models.ComputeCluster
// @Failure	400	{object}	object{error=string}
// @Router	/compute-clusters [post]
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

// @Summary	Get a compute cluster by ID
// @Tags	Compute Clusters
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute cluster ID"
// @Success	200	{object}	models.ComputeCluster
// @Failure	404	{object}	object{error=string}
// @Router	/compute-clusters/{id} [get]
func (s *Server) getComputeCluster(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.GetComputeCluster(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// @Summary	List compute clusters
// @Tags	Compute Clusters
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	models.ComputeCluster
// @Failure	500	{object}	object{error=string}
// @Router	/compute-clusters [get]
func (s *Server) listComputeClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := s.svc.ListComputeClusters(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, clusters)
}

// @Summary	Create a compute cluster user
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeClusterUser	true	"Cluster user payload"
// @Success	201	{object}	models.ComputeClusterUser
// @Failure	400	{object}	object{error=string}
// @Router	/compute-cluster-users [post]
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

// @Summary	Get a compute cluster user by ID
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute cluster user ID"
// @Success	200	{object}	models.ComputeClusterUser
// @Failure	404	{object}	object{error=string}
// @Router	/compute-cluster-users/{id} [get]
func (s *Server) getComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	cu, err := s.svc.GetComputeClusterUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cu)
}

// @Summary	Update a compute cluster user
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Compute cluster user ID"
// @Param	request	body	models.ComputeClusterUser	true	"Cluster user payload"
// @Success	200	{object}	models.ComputeClusterUser
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-cluster-users/{id} [put]
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

// @Summary	Delete a compute cluster user
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Param	id	path	string	true	"Compute cluster user ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/compute-cluster-users/{id} [delete]
func (s *Server) deleteComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeClusterUser(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List users on a compute cluster
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute cluster ID"
// @Success	200	{array}	models.ComputeClusterUser
// @Failure	404	{object}	object{error=string}
// @Router	/compute-clusters/{id}/users [get]
func (s *Server) listComputeClusterUsersByCluster(w http.ResponseWriter, r *http.Request) {
	users, err := s.svc.ListComputeClusterUsersByCluster(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// @Summary	Get a compute cluster user by (cluster, user) pair
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute cluster ID"
// @Param	userId	path	string	true	"User ID"
// @Success	200	{object}	models.ComputeClusterUser
// @Failure	404	{object}	object{error=string}
// @Router	/compute-clusters/{id}/users/{userId} [get]
func (s *Server) getComputeClusterUserByPair(w http.ResponseWriter, r *http.Request) {
	cu, err := s.svc.GetComputeClusterUserByPair(r.Context(), r.PathValue("id"), r.PathValue("userId"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cu)
}

// @Summary	List compute cluster users for a user
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{array}	models.ComputeClusterUser
// @Failure	404	{object}	object{error=string}
// @Router	/users/{id}/compute-cluster-users [get]
func (s *Server) listComputeClusterUsersByUser(w http.ResponseWriter, r *http.Request) {
	users, err := s.svc.ListComputeClusterUsersByUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// @Summary	Create a compute allocation
// @Tags	Compute Allocations
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeAllocation	true	"Compute allocation payload"
// @Success	201	{object}	models.ComputeAllocation
// @Failure	400	{object}	object{error=string}
// @Router	/compute-allocations [post]
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

// @Summary	Get a compute allocation by ID
// @Tags	Compute Allocations
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{object}	models.ComputeAllocation
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id} [get]
func (s *Server) getComputeAllocation(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetComputeAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, a)
}

// @Summary	Create a compute allocation resource
// @Description	Defines a resource (partition) — e.g. a CPU or GPU partition that allocations can be attached to.
// @Tags	Compute Allocation Resources
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeAllocationResource	true	"Resource payload"
// @Success	201	{object}	models.ComputeAllocationResource
// @Failure	400	{object}	object{error=string}
// @Failure	401	{object}	object{error=string}
// @Router	/compute-allocation-resources [post]
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

// @Summary	Get a compute allocation resource
// @Tags	Compute Allocation Resources
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Resource ID"
// @Success	200	{object}	models.ComputeAllocationResource
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-resources/{id} [get]
func (s *Server) getComputeAllocationResource(w http.ResponseWriter, r *http.Request) {
	res, err := s.svc.GetComputeAllocationResource(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// @Summary	List all compute allocation resources
// @Tags	Compute Allocation Resources
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	models.ComputeAllocationResource
// @Failure	401	{object}	object{error=string}
// @Router	/compute-allocation-resources [get]
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

// @Summary	Attach a resource to a compute allocation
// @Description	Creates a mapping between a compute allocation and a resource (partition) with a specific amount and wall-clock time grant.
// @Tags	Compute Allocation Resources
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Param	request	body	attachResourceRequest	true	"Attach payload"
// @Success	201	{object}	models.ComputeAllocationResourceMapping
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/resources [post]
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

// @Summary	Update a compute allocation -> resource mapping
// @Tags	Compute Allocation Resources
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Param	resourceId	path	string	true	"Compute allocation resource ID"
// @Param	request	body	updateAllocationResourceMappingRequest	true	"Mapping update payload"
// @Success	200	{object}	models.ComputeAllocationResourceMapping
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/resources/{resourceId} [put]
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

// @Summary	Detach a resource from a compute allocation
// @Tags	Compute Allocation Resources
// @Security	BearerAuth
// @Param	id	path	string	true	"Compute allocation ID"
// @Param	resourceId	path	string	true	"Compute allocation resource ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/resources/{resourceId} [delete]
func (s *Server) detachResourceFromAllocation(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DetachResourceFromAllocation(r.Context(), r.PathValue("id"), r.PathValue("resourceId")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List resources attached to a compute allocation
// @Tags	Compute Allocation Resources
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{array}	models.ComputeAllocationResourceMapping
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/resources [get]
func (s *Server) listResourcesForAllocation(w http.ResponseWriter, r *http.Request) {
	resources, err := s.svc.ListResourcesForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resources)
}

// @Summary	List compute allocations attached to a resource
// @Tags	Compute Allocation Resources
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation resource ID"
// @Success	200	{array}	models.ComputeAllocationResourceMapping
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-resources/{id}/allocations [get]
func (s *Server) listAllocationsForResource(w http.ResponseWriter, r *http.Request) {
	allocs, err := s.svc.ListAllocationsForResource(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, allocs)
}

// @Summary	Create a compute allocation resource rate
// @Description	Records the SU rate (e.g. SU per CPU-hour) for a resource over a time window.
// @Tags	Compute Allocation Resource Rates
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeAllocationResourceRate	true	"Rate payload"
// @Success	201	{object}	models.ComputeAllocationResourceRate
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-resource-rates [post]
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

// @Summary	Get a compute allocation resource rate
// @Tags	Compute Allocation Resource Rates
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Rate ID"
// @Success	200	{object}	models.ComputeAllocationResourceRate
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-resource-rates/{id} [get]
func (s *Server) getComputeAllocationResourceRate(w http.ResponseWriter, r *http.Request) {
	rate, err := s.svc.GetComputeAllocationResourceRate(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rate)
}

// @Summary	List rate history for a resource
// @Tags	Compute Allocation Resource Rates
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation resource ID"
// @Success	200	{array}	models.ComputeAllocationResourceRate
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-resources/{id}/rates [get]
func (s *Server) listRatesForResource(w http.ResponseWriter, r *http.Request) {
	rates, err := s.svc.ListRatesForResource(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rates)
}

// @Summary	Get the effective rate for a resource at a given time
// @Description	Returns the rate whose `[start_time, end_time)` window contains the `at` timestamp. Defaults to now when `at` is omitted.
// @Tags	Compute Allocation Resource Rates
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation resource ID"
// @Param	at	query	string	false	"RFC3339 time; defaults to now"
// @Success	200	{object}	models.ComputeAllocationResourceRate
// @Failure	400	{object}	object{error=string}	"Invalid 'at' query parameter"
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-resources/{id}/rates/effective [get]
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

// @Summary	Create a compute allocation diff
// @Description	Records a discrete change against a compute allocation (e.g. USAGE_UPDATE or ALLOCATION_STATUS_CHANGE).
// @Tags	Compute Allocation Diffs
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeAllocationDiff	true	"Diff payload"
// @Success	201	{object}	models.ComputeAllocationDiff
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-diffs [post]
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

// @Summary	Get a compute allocation diff
// @Tags	Compute Allocation Diffs
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Diff ID"
// @Success	200	{object}	models.ComputeAllocationDiff
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-diffs/{id} [get]
func (s *Server) getComputeAllocationDiff(w http.ResponseWriter, r *http.Request) {
	diff, err := s.svc.GetComputeAllocationDiff(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, diff)
}

// @Summary	Delete a compute allocation diff
// @Tags	Compute Allocation Diffs
// @Security	BearerAuth
// @Param	id	path	string	true	"Diff ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-diffs/{id} [delete]
func (s *Server) deleteComputeAllocationDiff(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationDiff(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List diffs for a compute allocation
// @Tags	Compute Allocation Diffs
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{array}	models.ComputeAllocationDiff
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/diffs [get]
func (s *Server) listDiffsForAllocation(w http.ResponseWriter, r *http.Request) {
	diffs, err := s.svc.ListDiffsForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, diffs)
}

// @Summary	Get the most recent diff for a compute allocation
// @Tags	Compute Allocation Diffs
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{object}	models.ComputeAllocationDiff
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/diffs/latest [get]
func (s *Server) getLatestDiffForAllocation(w http.ResponseWriter, r *http.Request) {
	diff, err := s.svc.GetLatestDiffForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, diff)
}

// @Summary	Create a compute allocation change request
// @Tags	Compute Allocation Change Requests
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeAllocationChangeRequest	true	"Change request payload"
// @Success	201	{object}	models.ComputeAllocationChangeRequest
// @Failure	400	{object}	object{error=string}
// @Router	/compute-allocation-change-requests [post]
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

// @Summary	Get a compute allocation change request
// @Tags	Compute Allocation Change Requests
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Change request ID"
// @Success	200	{object}	models.ComputeAllocationChangeRequest
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-change-requests/{id} [get]
func (s *Server) getComputeAllocationChangeRequest(w http.ResponseWriter, r *http.Request) {
	req, err := s.svc.GetComputeAllocationChangeRequest(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, req)
}

// @Summary	Update a compute allocation change request
// @Tags	Compute Allocation Change Requests
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Change request ID"
// @Param	request	body	models.ComputeAllocationChangeRequest	true	"Change request payload"
// @Success	200	{object}	models.ComputeAllocationChangeRequest
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-change-requests/{id} [put]
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

// @Summary	Delete a compute allocation change request
// @Tags	Compute Allocation Change Requests
// @Security	BearerAuth
// @Param	id	path	string	true	"Change request ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-change-requests/{id} [delete]
func (s *Server) deleteComputeAllocationChangeRequest(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationChangeRequest(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List change requests for a compute allocation
// @Tags	Compute Allocation Change Requests
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{array}	models.ComputeAllocationChangeRequest
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/change-requests [get]
func (s *Server) listChangeRequestsForAllocation(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListChangeRequestsForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// @Summary	List change requests submitted by a user
// @Tags	Compute Allocation Change Requests
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{array}	models.ComputeAllocationChangeRequest
// @Failure	404	{object}	object{error=string}
// @Router	/users/{id}/change-requests [get]
func (s *Server) listChangeRequestsByRequester(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListChangeRequestsByRequester(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// @Summary	Create a change request event
// @Tags	Change Request Events
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeAllocationChangeRequestEvent	true	"Event payload"
// @Success	201	{object}	models.ComputeAllocationChangeRequestEvent
// @Failure	400	{object}	object{error=string}
// @Router	/compute-allocation-change-request-events [post]
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

// @Summary	Get a change request event
// @Tags	Change Request Events
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Event ID"
// @Success	200	{object}	models.ComputeAllocationChangeRequestEvent
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-change-request-events/{id} [get]
func (s *Server) getComputeAllocationChangeRequestEvent(w http.ResponseWriter, r *http.Request) {
	evt, err := s.svc.GetComputeAllocationChangeRequestEvent(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, evt)
}

// @Summary	Delete a change request event
// @Tags	Change Request Events
// @Security	BearerAuth
// @Param	id	path	string	true	"Event ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-change-request-events/{id} [delete]
func (s *Server) deleteComputeAllocationChangeRequestEvent(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationChangeRequestEvent(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List events for a change request
// @Tags	Change Request Events
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Change request ID"
// @Success	200	{array}	models.ComputeAllocationChangeRequestEvent
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-change-requests/{id}/events [get]
func (s *Server) listEventsForChangeRequest(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListEventsForChangeRequest(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// @Summary	Get the most recent event for a change request
// @Tags	Change Request Events
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Change request ID"
// @Success	200	{object}	models.ComputeAllocationChangeRequestEvent
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-change-requests/{id}/events/latest [get]
func (s *Server) getLatestEventForChangeRequest(w http.ResponseWriter, r *http.Request) {
	evt, err := s.svc.GetLatestEventForChangeRequest(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, evt)
}

// @Summary	Create a compute allocation membership
// @Tags	Compute Allocation Memberships
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeAllocationMembership	true	"Membership payload"
// @Success	201	{object}	models.ComputeAllocationMembership
// @Failure	400	{object}	object{error=string}
// @Router	/compute-allocation-memberships [post]
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

// @Summary	Get a compute allocation membership
// @Tags	Compute Allocation Memberships
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Membership ID"
// @Success	200	{object}	models.ComputeAllocationMembership
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-memberships/{id} [get]
func (s *Server) getComputeAllocationMembership(w http.ResponseWriter, r *http.Request) {
	m, err := s.svc.GetComputeAllocationMembership(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

// @Summary	Update a compute allocation membership
// @Tags	Compute Allocation Memberships
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Membership ID"
// @Param	request	body	models.ComputeAllocationMembership	true	"Membership payload"
// @Success	200	{object}	models.ComputeAllocationMembership
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-memberships/{id} [put]
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

// @Summary	Update a membership's status
// @Tags	Compute Allocation Memberships
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Membership ID"
// @Param	request	body	object{membership_status=models.AllocationStatus}	true	"Status patch"
// @Success	200	{object}	models.ComputeAllocationMembership
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-memberships/{id}/status [put]
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

// @Summary	Delete a compute allocation membership
// @Tags	Compute Allocation Memberships
// @Security	BearerAuth
// @Param	id	path	string	true	"Membership ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-memberships/{id} [delete]
func (s *Server) deleteComputeAllocationMembership(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationMembership(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List members of a compute allocation
// @Tags	Compute Allocation Memberships
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{array}	AllocationMembershipResponse
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/memberships [get]
func (s *Server) listMembersForAllocation(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListMembersForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	out := make([]AllocationMembershipResponse, 0, len(rows))
	for _, m := range rows {
		out = append(out, AllocationMembershipResponse{
			ComputeAllocationMembership: m.ComputeAllocationMembership,
			Role:                        m.Role,
			DisplayName:                 m.DisplayName,
			Email:                       m.Email,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// @Summary	List a user's compute allocation memberships
// @Tags	Compute Allocation Memberships
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{array}	models.ComputeAllocationMembership
// @Failure	404	{object}	object{error=string}
// @Router	/users/{id}/compute-allocation-memberships [get]
func (s *Server) listAllocationsForUser(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListAllocationsForUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// @Summary	Create a compute allocation usage record
// @Tags	Compute Allocation Usages
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeAllocationUsage	true	"Usage payload"
// @Success	201	{object}	models.ComputeAllocationUsage
// @Failure	400	{object}	object{error=string}
// @Router	/compute-allocation-usages [post]
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

// @Summary	Get a compute allocation usage record
// @Tags	Compute Allocation Usages
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Usage ID"
// @Success	200	{object}	models.ComputeAllocationUsage
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-usages/{id} [get]
func (s *Server) getComputeAllocationUsage(w http.ResponseWriter, r *http.Request) {
	u, err := s.svc.GetComputeAllocationUsage(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// @Summary	Delete a compute allocation usage record
// @Tags	Compute Allocation Usages
// @Security	BearerAuth
// @Param	id	path	string	true	"Usage ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-usages/{id} [delete]
func (s *Server) deleteComputeAllocationUsage(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationUsage(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List usages for a compute allocation
// @Tags	Compute Allocation Usages
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{array}	models.ComputeAllocationUsage
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/usages [get]
func (s *Server) listUsagesForAllocation(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListUsagesForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// @Summary	List usages submitted by a user
// @Tags	Compute Allocation Usages
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{array}	models.ComputeAllocationUsage
// @Failure	404	{object}	object{error=string}
// @Router	/users/{id}/compute-allocation-usages [get]
func (s *Server) listUsagesByUser(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListUsagesByUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// @Summary	Create a membership resource override
// @Tags	Membership Resource Overrides
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeAllocationMembershipResourceOverride	true	"Override payload"
// @Success	201	{object}	models.ComputeAllocationMembershipResourceOverride
// @Failure	400	{object}	object{error=string}
// @Router	/compute-allocation-membership-resource-overrides [post]
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

// @Summary	Get a membership resource override
// @Tags	Membership Resource Overrides
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Override ID"
// @Success	200	{object}	models.ComputeAllocationMembershipResourceOverride
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-membership-resource-overrides/{id} [get]
func (s *Server) getComputeAllocationMembershipResourceOverride(w http.ResponseWriter, r *http.Request) {
	o, err := s.svc.GetComputeAllocationMembershipResourceOverride(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

// @Summary	Update a membership resource override
// @Tags	Membership Resource Overrides
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Override ID"
// @Param	request	body	models.ComputeAllocationMembershipResourceOverride	true	"Override payload"
// @Success	200	{object}	models.ComputeAllocationMembershipResourceOverride
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-membership-resource-overrides/{id} [put]
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

// @Summary	Delete a membership resource override
// @Tags	Membership Resource Overrides
// @Security	BearerAuth
// @Param	id	path	string	true	"Override ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-membership-resource-overrides/{id} [delete]
func (s *Server) deleteComputeAllocationMembershipResourceOverride(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeAllocationMembershipResourceOverride(r.Context(), r.PathValue("id")); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List resource overrides for a membership
// @Tags	Membership Resource Overrides
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Membership ID"
// @Success	200	{array}	models.ComputeAllocationMembershipResourceOverride
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-memberships/{id}/resource-overrides [get]
func (s *Server) listOverridesForMembership(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListOverridesForMembership(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// @Summary	List membership overrides referencing a resource
// @Tags	Membership Resource Overrides
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation resource ID"
// @Success	200	{array}	models.ComputeAllocationMembershipResourceOverride
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocation-resources/{id}/membership-overrides [get]
func (s *Server) listOverridesForResource(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListOverridesForResource(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// @Summary	Get total SU usage for a compute allocation
// @Tags	Compute Allocation Usages
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{object}	AllocationSUTotalResponse
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/usages/total [get]
func (s *Server) getTotalSUUsageForAllocation(w http.ResponseWriter, r *http.Request) {
	total, err := s.svc.GetTotalSUUsageForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, AllocationSUTotalResponse{
		ComputeAllocationID: r.PathValue("id"),
		TotalSUAmount:       total,
	})
}

// @Summary	Get total SU usage for a user within a compute allocation
// @Tags	Compute Allocation Usages
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Param	userId	path	string	true	"User ID"
// @Success	200	{object}	UserAllocationSUTotalResponse
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/users/{userId}/usages/total [get]
func (s *Server) getTotalSUUsageForUserInAllocation(w http.ResponseWriter, r *http.Request) {
	allocationID := r.PathValue("id")
	userID := r.PathValue("userId")
	total, err := s.svc.GetTotalSUUsageForUserInAllocation(r.Context(), allocationID, userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, UserAllocationSUTotalResponse{
		ComputeAllocationID: allocationID,
		UserID:              userID,
		TotalSUAmount:       total,
	})
}

type statusUpdateRequest struct {
	Status string `json:"status"`
}

// @Summary	Update a user's status
// @Tags	Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Param	request	body	object{status=models.UserStatus}	true	"Status patch"
// @Success	200	{object}	models.User
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/users/{id}/status [put]
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

// @Summary	Update a project's status
// @Tags	Projects
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Project ID"
// @Param	request	body	object{status=models.ProjectStatus}	true	"Status patch"
// @Success	200	{object}	models.Project
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/projects/{id}/status [put]
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

// @Summary	Create a user identity
// @Tags	User Identities
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.UserIdentity	true	"Identity payload"
// @Success	201	{object}	models.UserIdentity
// @Failure	400	{object}	object{error=string}
// @Router	/user-identities [post]
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

// @Summary	Get a user identity
// @Tags	User Identities
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Identity ID"
// @Success	200	{object}	models.UserIdentity
// @Failure	404	{object}	object{error=string}
// @Router	/user-identities/{id} [get]
func (s *Server) getUserIdentity(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetUserIdentity(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// @Summary	Get a user identity by source and external ID
// @Tags	User Identities
// @Security	BearerAuth
// @Produce	json
// @Param	source	path	string	true	"Identity source"
// @Param	externalId	path	string	true	"Identity's external ID at that source"
// @Success	200	{object}	models.UserIdentity
// @Failure	404	{object}	object{error=string}
// @Router	/user-identities/sources/{source}/external/{externalId} [get]
func (s *Server) getUserIdentityBySourceAndExternalID(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetUserIdentityBySourceAndExternalID(r.Context(), r.PathValue("source"), r.PathValue("externalId"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// @Summary	Get a user identity by its OIDC subject claim
// @Tags	User Identities
// @Security	BearerAuth
// @Produce	json
// @Param	oidcSub	path	string	true	"OIDC subject claim"
// @Success	200	{object}	models.UserIdentity
// @Failure	404	{object}	object{error=string}
// @Router	/user-identities/oidc-subjects/{oidcSub} [get]
func (s *Server) getUserIdentityByOIDCSub(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetUserIdentityByOIDCSub(r.Context(), r.PathValue("oidcSub"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// @Summary	List a user's identities
// @Tags	User Identities
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{array}	models.UserIdentity
// @Failure	404	{object}	object{error=string}
// @Router	/users/{id}/user-identities [get]
func (s *Server) listUserIdentitiesForUser(w http.ResponseWriter, r *http.Request) {
	out, err := s.svc.ListUserIdentitiesForUser(r.Context(), r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// @Summary	Update a user identity
// @Tags	User Identities
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Identity ID"
// @Param	request	body	models.UserIdentity	true	"Identity payload"
// @Success	200	{object}	models.UserIdentity
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/user-identities/{id} [put]
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

// @Summary	Delete a user identity
// @Tags	User Identities
// @Security	BearerAuth
// @Param	id	path	string	true	"Identity ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/user-identities/{id} [delete]
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

// @Summary	Merge two users
// @Description	Merges the retiring user into the surviving user; the surviving record is returned.
// @Tags	Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	mergeUsersRequest	true	"Merge payload"
// @Success	200	{object}	models.User
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/users/merge [post]
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

// @Summary	List projects (filtered + paginated, PI joined)
// @Tags	Projects
// @Security	BearerAuth
// @Produce	json
// @Param	pi_id	query	string	false	"Filter by PI user ID"
// @Param	status	query	string	false	"Filter by status"
// @Param	q	query	string	false	"Search title / originated_id"
// @Param	limit	query	integer	false	"Page size"
// @Param	offset	query	integer	false	"Page offset"
// @Success	200	{object}	ProjectListResponse
// @Router	/projects [get]
func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.ProjectListFilter{
		PIID:   q.Get("pi_id"),
		Status: q.Get("status"),
		Query:  q.Get("q"),
		Limit:  atoiOr(q.Get("limit"), 50),
		Offset: atoiOr(q.Get("offset"), 0),
	}
	rows, total, err := s.svc.ListProjectsWithPI(r.Context(), f)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	items := make([]ProjectResponse, 0, len(rows))
	for i := range rows {
		items = append(items, projectResponseFrom(&rows[i]))
	}
	writeJSON(w, http.StatusOK, ProjectListResponse{Items: items, Total: total})
}

// @Summary	List project members (one row per distinct user, with allocations)
// @Tags	Projects
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Project ID"
// @Success	200	{array}	ProjectMemberResponse
// @Failure	404	{object}	object{error=string}
// @Router	/projects/{id}/members [get]
func (s *Server) listProjectMembers(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	rows, err := s.svc.ListMembersForProject(r.Context(), projectID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	// Role is project-level (single value per user), so the dedup just
	// collects the user's allocations.
	type aggregate struct {
		base        store.MembershipWithUser
		allocations []ProjectMemberAllocationRef
	}
	byUser := map[string]*aggregate{}
	order := []string{}
	for _, m := range rows {
		agg, ok := byUser[m.UserID]
		if !ok {
			agg = &aggregate{base: m}
			byUser[m.UserID] = agg
			order = append(order, m.UserID)
		}
		agg.allocations = append(agg.allocations, ProjectMemberAllocationRef{
			ID:   m.ComputeAllocationID,
			Name: m.AllocationName,
			Role: m.Role,
		})
	}

	out := make([]ProjectMemberResponse, 0, len(order))
	for _, uid := range order {
		agg := byUser[uid]
		out = append(out, ProjectMemberResponse{
			ID:          agg.base.ID,
			ProjectID:   projectID,
			UserID:      agg.base.UserID,
			Email:       agg.base.Email,
			DisplayName: agg.base.DisplayName,
			Role:        agg.base.Role,
			Status:      string(agg.base.MembershipStatus),
			AddedTime:   agg.base.StartTime.UTC(),
			Allocations: agg.allocations,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// @Summary	List compute allocations (filtered + paginated)
// @Tags	Compute Allocations
// @Security	BearerAuth
// @Produce	json
// @Param	project_id	query	string	false	"Filter by project ID"
// @Param	status	query	string	false	"Filter by status"
// @Param	q	query	string	false	"Search name"
// @Param	limit	query	integer	false	"Page size"
// @Param	offset	query	integer	false	"Page offset"
// @Success	200	{object}	ComputeAllocationListResponse
// @Router	/compute-allocations [get]
func (s *Server) listComputeAllocations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.AllocationListFilter{
		ProjectID: q.Get("project_id"),
		Status:    q.Get("status"),
		Query:     q.Get("q"),
		Limit:     atoiOr(q.Get("limit"), 50),
		Offset:    atoiOr(q.Get("offset"), 0),
	}
	rows, total, err := s.svc.ListComputeAllocations(r.Context(), f)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if rows == nil {
		rows = []models.ComputeAllocation{}
	}
	writeJSON(w, http.StatusOK, ComputeAllocationListResponse{Items: rows, Total: total})
}

func (s *Server) listChangeRequests(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.ChangeRequestListFilter{
		Status: q.Get("status"),
		Limit:  atoiOr(q.Get("limit"), 50),
	}
	rows, err := s.svc.ListChangeRequests(r.Context(), f)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	if rows == nil {
		rows = []models.ComputeAllocationChangeRequest{}
	}
	writeJSON(w, http.StatusOK, rows)
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
