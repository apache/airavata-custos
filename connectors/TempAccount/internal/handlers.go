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

package internal

import (
	"errors"
	"net/http"

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type Handlers struct {
	coreService *service.Service
}

func NewHandlers(coreService *service.Service) *Handlers {
	return &Handlers{coreService: coreService}
}

// RegisterRoutes attaches the TempAccount connector's HTTP endpoints to mux.
func (h *Handlers) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/connectors/temp-account/create", h.createTempAccount)
	mux.HandleFunc("/connectors/temp-account/assign-allocation", h.assignAllocationToTempAccount)
	mux.HandleFunc("/connectors/temp-account/update-allocation", h.updateAllocationToTempAccount)
	mux.HandleFunc("/connectors/temp-account/remove/{user_id}", h.removeTempAccount)
	mux.HandleFunc("/connectors/temp-account/membership/{user_id}", h.getAllocationMembershipForTempUser)
}

func (h *Handlers) createTempAccount(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := common.DecodeJSON(r, &u); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if u.Type != models.UserTypeVirtual {
		common.WriteError(w, http.StatusBadRequest, errors.New("User type must be VIRTUAL"))
		return
	}

	created, err := h.coreService.CreateUser(r.Context(), &u)

	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
}

func (h *Handlers) removeTempAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("user_id query parameter is required"))
		return
	}

	user, err := h.coreService.GetUser(r.Context(), userID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	if user == nil {
		common.WriteError(w, http.StatusNotFound, errors.New("user not found"))
		return
	}

	if user.Type != models.UserTypeVirtual {
		common.WriteError(w, http.StatusBadRequest, errors.New("User type must be VIRTUAL"))
		return
	}

	err = h.coreService.DeleteUser(r.Context(), userID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) assignAllocationToTempAccount(w http.ResponseWriter, r *http.Request) {
	var m models.ComputeAllocationMembership
	if err := common.DecodeJSON(r, &m); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if m.UserID == "" || m.ComputeAllocationID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("UserID and ComputeAllocationID must be provided"))
		return
	}

	// In future, we may want to check that the user already has a membership to an allocation
	// If that's the case, prevent user from getting membershup to multiple allocations, which TempAccount connector doesn't support.

	user, err := h.coreService.GetUser(r.Context(), m.UserID)

	if err != nil {
		common.WriteServiceError(w, err)
		return
	}

	if user.Type != models.UserTypeVirtual {
		common.WriteError(w, http.StatusBadRequest, errors.New("User type must be VIRTUAL"))
		return
	}

	assigned, err := h.coreService.CreateComputeAllocationMembership(r.Context(), &m)

	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, assigned)
}

func (h *Handlers) updateAllocationToTempAccount(w http.ResponseWriter, r *http.Request) {
	var m models.ComputeAllocationMembership
	if err := common.DecodeJSON(r, &m); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if m.UserID == "" || m.ComputeAllocationID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("UserID and ComputeAllocationID must be provided"))
		return
	}

	updated, err := h.coreService.UpdateComputeAllocationMembership(r.Context(), &m)

	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, updated)
}

func (h *Handlers) getAllocationMembershipForTempUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("user_id query parameter is required"))
		return
	}

	user, err := h.coreService.GetUser(r.Context(), userID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	if user == nil {
		common.WriteError(w, http.StatusNotFound, errors.New("user not found"))
		return
	}

	if user.Type != models.UserTypeVirtual {
		common.WriteError(w, http.StatusBadRequest, errors.New("User type must be VIRTUAL"))
		return
	}

	memberships, err := h.coreService.ListAllocationsForUser(r.Context(), userID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	if memberships == nil {
		common.WriteError(w, http.StatusNotFound, errors.New("membership not found for user"))
		return
	}

	common.WriteJSON(w, http.StatusOK, memberships)
}
