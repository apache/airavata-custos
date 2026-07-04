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
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationChangeRequest(r.Context(), &req)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, req)
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
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	req.ID = r.PathValue("id")
	updated, err := s.svc.UpdateComputeAllocationChangeRequest(r.Context(), &req)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, updated)
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
		common.WriteServiceError(w, err)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
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
	if err := common.DecodeJSON(r, &evt); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationChangeRequestEvent(r.Context(), &evt)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, evt)
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
		common.WriteServiceError(w, err)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, evt)
}

func (s *Server) listChangeRequests(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := store.ChangeRequestListFilter{
		Status: q.Get("status"),
		Limit:  atoiOr(q.Get("limit"), 50),
	}
	rows, err := s.svc.ListChangeRequests(r.Context(), f)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	if rows == nil {
		rows = []models.ComputeAllocationChangeRequest{}
	}
	common.WriteJSON(w, http.StatusOK, rows)
}
