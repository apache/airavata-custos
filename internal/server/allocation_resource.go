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
	"errors"
	"net/http"
	"time"

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/models"
)

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
	if err := common.DecodeJSON(r, &res); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationResource(r.Context(), &res)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, res)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, resources)
}

// @Summary	List compute allocation resources with aggregate figures
// @Description	Each row is a resource plus its allocation count, total allocated amount, total used SUs, and rate count.
// @Tags	Compute Allocation Resources
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	store.ComputeAllocationResourceSummary
// @Failure	401	{object}	object{error=string}
// @Router	/compute-allocation-resources/summary [get]
func (s *Server) listComputeAllocationResourceSummaries(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListComputeAllocationResourceSummaries(r.Context())
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
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
	if err := common.DecodeJSON(r, &body); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	mapping, err := s.svc.AttachResourceToAllocation(r.Context(), r.PathValue("id"), body.ComputeAllocationResourceID, body.ResourceAmount, body.ResourceTime)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, mapping)
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
	if err := common.DecodeJSON(r, &body); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	mapping, err := s.svc.UpdateAllocationResourceMapping(r.Context(), r.PathValue("id"), r.PathValue("resourceId"), body.ResourceAmount, body.ResourceTime)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, mapping)
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
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List resources attached to a compute allocation
// @Description	Callers without the allocations read privilege must hold a membership on the allocation or a governance role on its project; others return 404.
// @Tags	Compute Allocation Resources
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute allocation ID"
// @Success	200	{array}	models.ComputeAllocationResourceMapping
// @Failure	401	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-allocations/{id}/resources [get]
func (s *Server) listResourcesForAllocation(w http.ResponseWriter, r *http.Request) {
	resources, err := s.svc.ListResourcesForAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, resources)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, allocs)
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
	if err := common.DecodeJSON(r, &rate); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeAllocationResourceRate(r.Context(), &rate)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rate)
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rates)
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
			common.WriteError(w, http.StatusBadRequest, errors.New("invalid 'at' query parameter; expected RFC 3339"))
			return
		}
		at = parsed
	}
	rate, err := s.svc.GetEffectiveRateForResource(r.Context(), r.PathValue("id"), at)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rate)
}
