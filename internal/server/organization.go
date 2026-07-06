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
	if err := common.DecodeJSON(r, &org); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateOrganization(r.Context(), &org)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
}

// @Summary	List organizations (paginated)
// @Tags	Organizations
// @Security	BearerAuth
// @Produce	json
// @Param	limit	query	integer	false	"Page size"
// @Param	offset	query	integer	false	"Page offset"
// @Success	200	{object}	OrganizationListResponse
// @Router	/organizations [get]
func (s *Server) listOrganizations(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows, total, err := s.svc.ListOrganizations(r.Context(), atoiOr(q.Get("limit"), 50), atoiOr(q.Get("offset"), 0))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	if rows == nil {
		rows = []models.Organization{}
	}
	common.WriteJSON(w, http.StatusOK, OrganizationListResponse{Items: rows, Total: total})
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
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, org)
}
