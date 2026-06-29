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

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

// @Summary	Get caller's effective privileges
// @Description	Effective set is direct grants UNION every privilege carried by every role the caller holds.
// @Tags	Caller
// @Security	BearerAuth
// @Produce	json
// @Success	200	{object}	object{privileges=[]models.PrivilegeKey}	"Effective privilege set"
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Router	/user/privileges [get]
func (s *Server) getCallerPrivileges(w http.ResponseWriter, r *http.Request) {
	if requireCaller(w, r) == nil {
		return
	}
	keys := identity.PrivilegesFromContext(r.Context())
	if keys == nil {
		keys = []models.PrivilegeKey{}
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"privileges": keys})
}

// CallerProfileResponse bundles the caller's user row with the effective
// privilege set.
type CallerProfileResponse struct {
	User       *models.User          `json:"user"`
	Privileges []models.PrivilegeKey `json:"privileges"`
}

// @Summary	Get the caller's user record and effective privileges
// @Tags	Caller
// @Security	BearerAuth
// @Produce	json
// @Success	200	{object}	CallerProfileResponse
// @Failure	401	{object}	object{error=string}	"Missing authenticated caller"
// @Router	/me [get]
func (s *Server) getCallerProfile(w http.ResponseWriter, r *http.Request) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	user, err := s.svc.GetUser(r.Context(), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	keys := identity.PrivilegesFromContext(r.Context())
	if keys == nil {
		keys = []models.PrivilegeKey{}
	}
	common.WriteJSON(w, http.StatusOK, CallerProfileResponse{User: user, Privileges: keys})
}

// @Summary	List the declared privilege catalog
// @Tags	Privileges
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	models.PrivilegeKey
// @Failure	401	{object}	object{error=string}	"Missing or invalid bearer token"
// @Failure	403	{object}	object{error=string}	"Caller lacks privileges:grant"
// @Router	/privileges/catalog [get]
func (s *Server) getPrivilegeCatalog(w http.ResponseWriter, _ *http.Request) {
	common.WriteJSON(w, http.StatusOK, s.svc.PrivilegeCatalog())
}

// @Summary	List a user's direct privilege grants
// @Description	Returns DIRECT grants only (not role-derived). Combine with `GET /users/{id}/roles` for the full picture.
// @Tags	Privileges
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{array}	models.UserPrivilege
// @Failure	400	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks privileges:grant"
// @Router	/users/{id}/privileges [get]
func (s *Server) listUserPrivileges(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("user id is required"))
		return
	}
	rows, err := s.svc.ListUserPrivileges(r.Context(), userID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
}

// @Summary	List direct holders of a privilege
// @Description	Role-derived holders are NOT listed here.
// @Tags	Privileges
// @Security	BearerAuth
// @Produce	json
// @Param	key	path	models.PrivilegeKey	true	"Privilege key"
// @Success	200	{array}	models.UserPrivilege
// @Failure	400	{object}	object{error=string}	"Unknown privilege key"
// @Router	/privileges/{key}/holders [get]
func (s *Server) listPrivilegeHolders(w http.ResponseWriter, r *http.Request) {
	key := models.PrivilegeKey(r.PathValue("key"))
	if !models.IsKnownPrivilege(key) {
		common.WriteError(w, http.StatusBadRequest, errors.New("unknown privilege key"))
		return
	}
	rows, err := s.svc.ListPrivilegeHolders(r.Context(), key)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
}

type grantPrivilegeRequest struct {
	Privilege models.PrivilegeKey `json:"privilege"`
	Reason    string              `json:"reason"`
}

// @Summary	Grant a direct privilege to a user
// @Tags	Privileges
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Param	request	body	grantPrivilegeRequest	true	"Grant payload"
// @Success	201	{object}	models.UserPrivilege
// @Failure	400	{object}	object{error=string}	"Bad request"
// @Failure	401	{object}	object{error=string}	"Missing or invalid bearer token"
// @Failure	403	{object}	object{error=string}	"Caller lacks privileges:grant"
// @Failure	409	{object}	object{error=string}	"Privilege already active for user"
// @Router	/users/{id}/privileges [post]
func (s *Server) grantPrivilege(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("user id is required"))
		return
	}
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	var req grantPrivilegeRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	grant, err := s.svc.GrantPrivilege(r.Context(), userID, req.Privilege, caller.UserID, req.Reason)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.svc.InvalidateAllIdentities()
	common.WriteJSON(w, http.StatusCreated, grant)
}

type revokePrivilegeRequest struct {
	Reason string `json:"reason"`
}

// @Summary	Revoke a direct privilege from a user
// @Description	`privileges:grant` itself cannot be self-revoked or revoked from the last holder.
// @Tags	Privileges
// @Security	BearerAuth
// @Accept	json
// @Param	id	path	string	true	"User ID"
// @Param	key	path	models.PrivilegeKey	true	"Privilege key"
// @Param	request	body	revokePrivilegeRequest	false	"Optional reason"
// @Success	204	"No Content"
// @Failure	400	{object}	object{error=string}	"Self-revoke meta, last-holder, or unknown key"
// @Failure	401	{object}	object{error=string}	"Missing or invalid bearer token"
// @Failure	403	{object}	object{error=string}	"Caller lacks privileges:grant"
// @Failure	404	{object}	object{error=string}	"No active grant for that key"
// @Router	/users/{id}/privileges/{key} [delete]
func (s *Server) revokePrivilege(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	key := models.PrivilegeKey(r.PathValue("key"))
	if userID == "" || key == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("user id and privilege key are required"))
		return
	}
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	var req revokePrivilegeRequest
	_ = common.DecodeJSON(r, &req)
	if err := s.svc.RevokePrivilege(r.Context(), userID, key, caller.UserID, req.Reason); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.svc.InvalidateAllIdentities()
	w.WriteHeader(http.StatusNoContent)
}
