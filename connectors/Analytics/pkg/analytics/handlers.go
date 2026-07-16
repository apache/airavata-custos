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

package analytics

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

// Handlers serves the analytics endpoints on top of the connector's service.
type Handlers struct {
	svc *Service
}

// NewHandlers returns the connector's HTTP handlers.
func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// RegisterRoutes attaches the analytics endpoints to router. Access is
// authenticated and membership-scoped inside each handler.
func (h *Handlers) RegisterRoutes(router *identity.Router) {
	router.RequireAuth("GET /connectors/analytics/contexts", h.getCallerAnalyticsContexts)
	router.RequireAuth("GET /connectors/analytics/allocations/{id}/usage-summary", h.getAllocationUsageSummary)
	router.RequireAuth("GET /connectors/analytics/allocations/{id}/jobs", h.getAllocationJobs)
}

// @Summary	List the caller's analytics contexts
// @Description	Projects the caller belongs to (through a role or allocation membership), each with the caller's role and its allocations. No privilege required; scoped to the caller's own memberships.
// @Tags	Analytics
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	ProjectContext
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Router	/connectors/analytics/contexts [get]
func (h *Handlers) getCallerAnalyticsContexts(w http.ResponseWriter, r *http.Request) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	contexts, err := h.svc.AnalyticsContexts(r.Context(), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, contexts)
}

// @Summary	Get an allocation's usage summary
// @Description	Aggregated usage for one allocation. Access is membership-scoped: the caller must hold a membership on the allocation or a governance role on its project. The per-member breakdown is populated only for project managers. Callers without access receive 404 so allocation existence is not leaked.
// @Tags	Analytics
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{object}	UsageSummary
// @Failure	400	{object}	object{error=string}	"Missing allocation id"
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Failure	404	{object}	object{error=string}	"Allocation not found or caller lacks access"
// @Router	/connectors/analytics/allocations/{id}/usage-summary [get]
func (h *Handlers) getAllocationUsageSummary(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("allocation id is required"))
		return
	}
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	alloc, err := h.svc.GetAllocation(r.Context(), id)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}

	canManage, hasAccess, err := h.allocationAccess(r, alloc, caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	if !hasAccess {
		// Same 404 as a missing allocation, so a non-member cannot probe which
		// allocation ids exist.
		common.WriteServiceError(w, coreservice.ErrNotFound)
		return
	}

	summary, err := h.svc.AllocationUsageSummary(r.Context(), alloc, caller.UserID, canManage)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, summary)
}

// @Summary	List an allocation's jobs (usage records)
// @Description	A page of usage records (a job's charge) the newest first, with the total count. Access is membership-scoped like the usage summary. Plain members always see only their own records; project managers see everyone's by default and can narrow to their own with mine=true. Non-members receive 404.
// @Tags	Analytics
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Param	limit	query	integer	false	"Page size (default 20, max 100)"
// @Param	offset	query	integer	false	"Page offset"
// @Param	mine	query	boolean	false	"Restrict to the caller's own jobs"
// @Success	200	{object}	AllocationJobs
// @Failure	400	{object}	object{error=string}	"Missing allocation id"
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Failure	404	{object}	object{error=string}	"Allocation not found or caller lacks access"
// @Router	/connectors/analytics/allocations/{id}/jobs [get]
func (h *Handlers) getAllocationJobs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("allocation id is required"))
		return
	}
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	alloc, err := h.svc.GetAllocation(r.Context(), id)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	canManage, hasAccess, err := h.allocationAccess(r, alloc, caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	if !hasAccess {
		common.WriteServiceError(w, coreservice.ErrNotFound)
		return
	}

	q := r.URL.Query()
	// Members may only see their own records; managers default to everyone and
	// opt into their own with mine=true.
	mineOnly := !canManage || q.Get("mine") == "true"
	page, err := h.svc.AllocationJobsPage(r.Context(), id, caller.UserID, mineOnly, atoiOr(q.Get("limit"), 20), atoiOr(q.Get("offset"), 0))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, page)
}

// allocationAccess resolves the caller's standing on an allocation. canManage
// is true for project managers (PI/co-PI/allocation manager) and gates the
// per-member breakdown; hasAccess additionally admits plain allocation members
// (their own scoped view). Access comes from membership alone, never a
// site-wide privilege.
func (h *Handlers) allocationAccess(r *http.Request, alloc *models.ComputeAllocation, userID string) (canManage, hasAccess bool, err error) {
	role, err := h.svc.ProjectRoleForUser(r.Context(), alloc.ProjectID, userID)
	if err != nil {
		return false, false, err
	}
	switch role {
	case models.ProjectRolePI, models.ProjectRoleCoPI, models.ProjectRoleAllocationManager:
		return true, true, nil
	}
	member, err := h.svc.IsAllocationMember(r.Context(), alloc.ID, userID)
	if err != nil {
		return false, false, err
	}
	return false, member, nil
}

func requireCaller(w http.ResponseWriter, r *http.Request) *identity.Caller {
	c := identity.CallerFromContext(r.Context())
	if c == nil {
		common.WriteError(w, http.StatusUnauthorized, errors.New("missing authenticated caller"))
		return nil
	}
	return c
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
