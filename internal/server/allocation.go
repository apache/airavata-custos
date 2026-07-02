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

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/models"
)

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
	if err := common.DecodeJSON(r, &a); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocation(r.Context(), &a)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, a)
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
	if err := common.DecodeJSON(r, &diff); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationDiff(r.Context(), &diff)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, diff)
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
		common.WriteServiceError(w, err)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, diffs)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, diff)
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
		common.WriteServiceError(w, err)
		return
	}
	if rows == nil {
		rows = []models.ComputeAllocation{}
	}
	common.WriteJSON(w, http.StatusOK, ComputeAllocationListResponse{Items: rows, Total: total})
}
