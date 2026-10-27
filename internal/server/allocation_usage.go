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

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/models"
)

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
	if err := common.DecodeJSON(r, &u); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationUsage(r.Context(), &u)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, u)
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
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List usages for a compute allocation
// @Description	Callers without the allocations read privilege must hold a membership on the allocation or a governance role on its project; others return 404.
// @Tags	Compute Allocation Usages
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{array}	models.ComputeAllocationUsage
// @Failure	401	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/usages [get]
func (s *Server) listUsagesForAllocation(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListUsagesForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
}

// @Summary	Get total SU usage for a compute allocation
// @Description	Callers without the allocations read privilege must hold a membership on the allocation or a governance role on its project; others return 404.
// @Tags	Compute Allocation Usages
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{object}	AllocationSUTotalResponse
// @Failure	401	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/usages/total [get]
func (s *Server) getTotalSUUsageForAllocation(w http.ResponseWriter, r *http.Request) {
	total, err := s.svc.GetTotalSUUsageForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, AllocationSUTotalResponse{
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, UserAllocationSUTotalResponse{
		ComputeAllocationID: allocationID,
		UserID:              userID,
		TotalSUAmount:       total,
	})
}
