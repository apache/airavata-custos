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

// @Summary	Create a user identity
// @Tags	User Identities
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.UserIdentity	true	"Identity payload"
// @Success	201	{object}	models.UserIdentity
// @Failure	400	{object}	object{error=string}
// @Router	/user-identities [post]
func (s *Server) createUserIdentity(w http.ResponseWriter, r *http.Request) {
	var e models.UserIdentity
	if err := common.DecodeJSON(r, &e); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateUserIdentity(r.Context(), &e)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
}

// @Summary	Get a user identity
// @Tags	User Identities
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Identity ID"
// @Success	200	{object}	models.UserIdentity
// @Failure	404	{object}	object{error=string}
// @Router	/user-identities/{id} [get]
func (s *Server) getUserIdentity(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetUserIdentity(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, e)
}

// @Summary	Get a user identity by source and external ID
// @Tags	User Identities
// @Security	BearerAuth
// @Produce	json
// @Param	source	path	string	true	"Identity source"
// @Param	externalId	path	string	true	"Identity's external ID at that source"
// @Success	200	{object}	models.UserIdentity
// @Failure	404	{object}	object{error=string}
// @Router	/user-identities/sources/{source}/external/{externalId} [get]
func (s *Server) getUserIdentityBySourceAndExternalID(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetUserIdentityBySourceAndExternalID(r.Context(), r.PathValue("source"), r.PathValue("externalId"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, e)
}

// @Summary	Get a user identity by its OIDC subject claim
// @Tags	User Identities
// @Security	BearerAuth
// @Produce	json
// @Param	oidcSub	path	string	true	"OIDC subject claim"
// @Success	200	{object}	models.UserIdentity
// @Failure	404	{object}	object{error=string}
// @Router	/user-identities/oidc-subjects/{oidcSub} [get]
func (s *Server) getUserIdentityByOIDCSub(w http.ResponseWriter, r *http.Request) {
	e, err := s.svc.GetUserIdentityByOIDCSub(r.Context(), r.PathValue("oidcSub"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, e)
}

// @Summary	List a user's identities
// @Tags	User Identities
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{array}	models.UserIdentity
// @Failure	404	{object}	object{error=string}
// @Router	/users/{id}/user-identities [get]
func (s *Server) listUserIdentitiesForUser(w http.ResponseWriter, r *http.Request) {
	out, err := s.svc.ListUserIdentitiesForUser(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, out)
}

// @Summary	Update a user identity
// @Tags	User Identities
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Identity ID"
// @Param	request	body	models.UserIdentity	true	"Identity payload"
// @Success	200	{object}	models.UserIdentity
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/user-identities/{id} [put]
func (s *Server) updateUserIdentity(w http.ResponseWriter, r *http.Request) {
	var e models.UserIdentity
	if err := common.DecodeJSON(r, &e); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	e.ID = r.PathValue("id")
	if err := s.svc.UpdateUserIdentity(r.Context(), &e); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, &e)
}

// @Summary	Delete a user identity
// @Tags	User Identities
// @Security	BearerAuth
// @Param	id	path	string	true	"Identity ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/user-identities/{id} [delete]
func (s *Server) deleteUserIdentity(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteUserIdentity(r.Context(), r.PathValue("id")); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
