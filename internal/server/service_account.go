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

type serviceAccountWithTokenView struct {
	models.ServiceAccount
	Token string `json:"token"`
}

type createServiceAccountRequest struct {
	Key         string                `json:"key"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Privileges  []models.PrivilegeKey `json:"privileges"`
}

type updateServiceAccountRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func writeServiceAccount(w http.ResponseWriter, a *models.ServiceAccount, err error) {
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, a)
}

// @Summary	List service accounts
// @Tags	Service accounts
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	models.ServiceAccount
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks service-accounts:manage"
// @Router	/service-accounts [get]
func (s *Server) listServiceAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := s.svc.ListServiceAccounts(r.Context())
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, accounts)
}

// @Summary	List the privileges a service account may hold
// @Tags	Service accounts
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	string
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks service-accounts:manage"
// @Router	/service-accounts/privileges [get]
func (s *Server) listServiceAccountPrivileges(w http.ResponseWriter, r *http.Request) {
	common.WriteJSON(w, http.StatusOK, models.ServiceAccountPrivileges())
}

// @Summary	Create a service account
// @Description	The token is returned only here.
// @Tags	Service accounts
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	createServiceAccountRequest	true	"Service account"
// @Success	201	{object}	serviceAccountWithTokenView
// @Failure	400	{object}	object{error=string}
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks service-accounts:manage"
// @Failure	409	{object}	object{error=string}	"That key or name is taken"
// @Router	/service-accounts [post]
func (s *Server) createServiceAccount(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	var req createServiceAccountRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	a, token, err := s.svc.CreateServiceAccount(r.Context(), req.Key, req.Name, req.Description, req.Privileges, caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, serviceAccountWithTokenView{ServiceAccount: *a, Token: token})
}

// @Summary	Get a service account
// @Tags	Service accounts
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Service account ID"
// @Success	200	{object}	models.ServiceAccount
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks service-accounts:manage"
// @Failure	404	{object}	object{error=string}
// @Router	/service-accounts/{id} [get]
func (s *Server) getServiceAccount(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetServiceAccount(r.Context(), r.PathValue("id"))
	writeServiceAccount(w, a, err)
}

// @Summary	Delete a disabled service account
// @Description	The key can never be used again.
// @Tags	Service accounts
// @Security	BearerAuth
// @Param	id	path	string	true	"Service account ID"
// @Success	204
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks service-accounts:manage"
// @Failure	404	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}	"The service account is not disabled"
// @Router	/service-accounts/{id} [delete]
func (s *Server) deleteServiceAccount(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	writeNoContent(w, s.svc.DeleteServiceAccount(r.Context(), r.PathValue("id"), caller.UserID))
}

// @Summary	Edit a service account's name and description
// @Tags	Service accounts
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Service account ID"
// @Param	request	body	updateServiceAccountRequest	true	"Name and description"
// @Success	200	{object}	models.ServiceAccount
// @Failure	400	{object}	object{error=string}
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks service-accounts:manage"
// @Failure	404	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}	"That name is taken"
// @Router	/service-accounts/{id} [put]
func (s *Server) updateServiceAccount(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	var req updateServiceAccountRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	a, err := s.svc.UpdateServiceAccount(r.Context(), r.PathValue("id"), req.Name, req.Description, caller.UserID)
	writeServiceAccount(w, a, err)
}

// @Summary	Disable a service account
// @Tags	Service accounts
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Service account ID"
// @Success	200	{object}	models.ServiceAccount
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks service-accounts:manage"
// @Failure	404	{object}	object{error=string}
// @Router	/service-accounts/{id}/disable [post]
func (s *Server) disableServiceAccount(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	a, err := s.svc.DisableServiceAccount(r.Context(), r.PathValue("id"), caller.UserID)
	writeServiceAccount(w, a, err)
}

// @Summary	Enable a service account
// @Tags	Service accounts
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Service account ID"
// @Success	200	{object}	models.ServiceAccount
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks service-accounts:manage"
// @Failure	404	{object}	object{error=string}
// @Router	/service-accounts/{id}/enable [post]
func (s *Server) enableServiceAccount(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	a, err := s.svc.EnableServiceAccount(r.Context(), r.PathValue("id"), caller.UserID)
	writeServiceAccount(w, a, err)
}

// @Summary	Issue a new secret for a service account
// @Description	The old token stops working at once. The new token is returned only here.
// @Tags	Service accounts
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Service account ID"
// @Success	200	{object}	serviceAccountWithTokenView
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks service-accounts:manage"
// @Failure	404	{object}	object{error=string}
// @Router	/service-accounts/{id}/secret [post]
func (s *Server) issueServiceAccountSecret(w http.ResponseWriter, r *http.Request) {
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	a, token, err := s.svc.IssueServiceAccountSecret(r.Context(), r.PathValue("id"), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, serviceAccountWithTokenView{ServiceAccount: *a, Token: token})
}

func writeNoContent(w http.ResponseWriter, err error) {
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
