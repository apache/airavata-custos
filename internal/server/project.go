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

package server

import (
	"net/http"
	"strings"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

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
	if err := common.DecodeJSON(r, &p); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateProject(r.Context(), &p)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
}

// @Summary	Get a project by ID
// @Description	Callers without the projects read privilege only see projects they participate in; others return 404.
// @Tags	Projects
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Project ID"
// @Success	200	{object}	ProjectResponse
// @Failure	401	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/projects/{id} [get]
func (s *Server) getProject(w http.ResponseWriter, r *http.Request) {
	p, err := s.svc.GetProjectWithPI(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, projectResponseFrom(p))
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
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	p, err := s.svc.UpdateProjectStatus(r.Context(), r.PathValue("id"), models.ProjectStatus(req.Status))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, p)
}

// @Summary	List projects (filtered + paginated, PI joined)
// @Description	Callers without the projects read privilege get only the projects they participate in; filters and pagination apply to privileged callers only.
// @Tags	Projects
// @Security	BearerAuth
// @Produce	json
// @Param	pi_id	query	string	false	"Filter by PI user ID"
// @Param	status	query	string	false	"Filter by status"
// @Param	q	query	string	false	"Search title / originated_id"
// @Param	limit	query	integer	false	"Page size"
// @Param	offset	query	integer	false	"Page offset"
// @Success	200	{object}	ProjectListResponse
// @Failure	401	{object}	object{error=string}
// @Router	/projects [get]
func (s *Server) listProjects(w http.ResponseWriter, r *http.Request) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	if !identity.HasPrivilege(r.Context(), models.ProjectsRead) {
		rows, err := s.svc.ListProjectsForParticipant(r.Context(), caller.UserID)
		if err != nil {
			common.WriteServiceError(w, err)
			return
		}
		items := make([]ProjectResponse, 0, len(rows))
		for i := range rows {
			items = append(items, projectResponseFrom(&rows[i]))
		}
		common.WriteJSON(w, http.StatusOK, ProjectListResponse{Items: items, Total: len(items)})
		return
	}
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
		common.WriteServiceError(w, err)
		return
	}
	items := make([]ProjectResponse, 0, len(rows))
	for i := range rows {
		items = append(items, projectResponseFrom(&rows[i]))
	}
	common.WriteJSON(w, http.StatusOK, ProjectListResponse{Items: items, Total: total})
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
		common.WriteServiceError(w, err)
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
	common.WriteJSON(w, http.StatusOK, out)
}
