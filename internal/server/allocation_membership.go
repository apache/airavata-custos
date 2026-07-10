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
	if err := common.DecodeJSON(r, &m); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationMembership(r.Context(), &m)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, m)
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
	if err := common.DecodeJSON(r, &m); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	m.ID = r.PathValue("id")
	updated, err := s.svc.UpdateComputeAllocationMembership(r.Context(), &m)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, updated)
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
	if err := common.DecodeJSON(r, &body); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	updated, err := s.svc.UpdateMembershipStatus(r.Context(), r.PathValue("id"), body.MembershipStatus)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, updated)
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
		common.WriteServiceError(w, err)
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
		common.WriteServiceError(w, err)
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
	common.WriteJSON(w, http.StatusOK, out)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
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
	if err := common.DecodeJSON(r, &o); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationMembershipResourceOverride(r.Context(), &o)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, o)
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
	if err := common.DecodeJSON(r, &o); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	o.ID = r.PathValue("id")
	updated, err := s.svc.UpdateComputeAllocationMembershipResourceOverride(r.Context(), &o)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, updated)
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
		common.WriteServiceError(w, err)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
}
