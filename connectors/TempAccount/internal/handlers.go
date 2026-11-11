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
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type Handlers struct {
	coreService *service.Service
}

func NewHandlers(coreService *service.Service) *Handlers {
	return &Handlers{coreService: coreService}
}

// RegisterRoutes attaches the TempAccount connector's HTTP endpoints to router.
func (h *Handlers) RegisterRoutes(router *identity.Router) {
	router.RequireServiceAccount("POST /connectors/temp-account/users", UsersCreate, h.createTempAccount)
	router.RequireServiceAccount("GET /connectors/temp-account/users", UsersRead, h.getUser)
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
	// Only a PENDING user binds a first sign-in.
	u.ID, u.Status = "", ""

	status := http.StatusCreated
	user, err := h.coreService.CreateUser(r.Context(), &u)
	if errors.Is(err, service.ErrAlreadyExists) {
		status = http.StatusOK
		if user, err = h.coreService.GetUserByEmail(r.Context(), u.Email); err == nil && user.Type != models.UserTypeVirtual {
			err = service.ErrAlreadyExists
		}
	}
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, status, user)
}

// getUser returns the user holding ?email=, of any type.
func (h *Handlers) getUser(w http.ResponseWriter, r *http.Request) {
	user, err := h.coreService.GetUserByEmail(r.Context(), r.URL.Query().Get("email"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, user)
}
