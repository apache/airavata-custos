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

	"github.com/apache/airavata-custos/pkg/models"
)

// @Summary	Get caller's effective privileges
// @Description	Effective set is direct grants UNION every privilege carried by every role the caller holds.
// @Tags	Caller
// @Security	CustosUserHeader
// @Produce	json
// @Success	200	{object}	object{privileges=[]models.PrivilegeKey}	"Effective privilege set"
// @Failure	401	{object}	object{error=string}	"Missing X-Custos-User-Id header"
// @Failure	503	{object}	object{error=string}	"Auth lookup failed"
// @Router	/user/privileges [get]
func (s *Server) getCallerPrivileges(w http.ResponseWriter, r *http.Request) {
	callerID := r.Header.Get(callerHeader)
	if callerID == "" {
		writeError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	profile, err := s.lookupAuthProfile(r.Context(), callerID)
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, errors.New("auth lookup failed"))
		return
	}
	keys := make([]models.PrivilegeKey, 0, len(profile.privileges))
	for k := range profile.privileges {
		keys = append(keys, k)
	}
	writeJSON(w, http.StatusOK, map[string]any{"privileges": keys})
}

// @Summary	List the declared privilege catalog
// @Tags	Privileges
// @Security	CustosUserHeader
// @Produce	json
// @Success	200	{array}	models.PrivilegeKey
// @Failure	401	{object}	object{error=string}	"Missing caller header"
// @Failure	403	{object}	object{error=string}	"Caller lacks privileges:grant"
// @Router	/privileges/catalog [get]
func (s *Server) getPrivilegeCatalog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.svc.PrivilegeCatalog())
}

// @Summary	List a user's direct privilege grants
// @Description	Returns DIRECT grants only (not role-derived). Combine with `GET /users/{id}/roles` for the full picture.
// @Tags	Privileges
// @Security	CustosUserHeader
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{array}	models.UserPrivilege
// @Failure	400	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks privileges:grant"
// @Router	/users/{id}/privileges [get]
func (s *Server) listUserPrivileges(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, errors.New("user id is required"))
		return
	}
	rows, err := s.svc.ListUserPrivileges(r.Context(), userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// @Summary	List direct holders of a privilege
// @Description	Role-derived holders are NOT listed here.
// @Tags	Privileges
// @Security	CustosUserHeader
// @Produce	json
// @Param	key	path	models.PrivilegeKey	true	"Privilege key"
// @Success	200	{array}	models.UserPrivilege
// @Failure	400	{object}	object{error=string}	"Unknown privilege key"
// @Router	/privileges/{key}/holders [get]
func (s *Server) listPrivilegeHolders(w http.ResponseWriter, r *http.Request) {
	key := models.PrivilegeKey(r.PathValue("key"))
	if !models.IsKnownPrivilege(key) {
		writeError(w, http.StatusBadRequest, errors.New("unknown privilege key"))
		return
	}
	rows, err := s.svc.ListPrivilegeHolders(r.Context(), key)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

type grantPrivilegeRequest struct {
	Privilege models.PrivilegeKey `json:"privilege"`
	Reason    string              `json:"reason"`
}

// @Summary	Grant a direct privilege to a user
// @Tags	Privileges
// @Security	CustosUserHeader
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Param	request	body	grantPrivilegeRequest	true	"Grant payload"
// @Success	201	{object}	models.UserPrivilege
// @Failure	400	{object}	object{error=string}	"Bad request"
// @Failure	401	{object}	object{error=string}	"Missing caller header"
// @Failure	403	{object}	object{error=string}	"Caller lacks privileges:grant"
// @Failure	409	{object}	object{error=string}	"Privilege already active for user"
// @Router	/users/{id}/privileges [post]
func (s *Server) grantPrivilege(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, errors.New("user id is required"))
		return
	}
	granterID := r.Header.Get(callerHeader)
	if granterID == "" {
		writeError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	var req grantPrivilegeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	grant, err := s.svc.GrantPrivilege(r.Context(), userID, req.Privilege, granterID, req.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	s.authCache.invalidate(userID)
	writeJSON(w, http.StatusCreated, grant)
}

type revokePrivilegeRequest struct {
	Reason string `json:"reason"`
}

// @Summary	Revoke a direct privilege from a user
// @Description	`privileges:grant` itself cannot be self-revoked or revoked from the last holder.
// @Tags	Privileges
// @Security	CustosUserHeader
// @Accept	json
// @Param	id	path	string	true	"User ID"
// @Param	key	path	models.PrivilegeKey	true	"Privilege key"
// @Param	request	body	revokePrivilegeRequest	false	"Optional reason"
// @Success	204	"No Content"
// @Failure	400	{object}	object{error=string}	"Self-revoke meta, last-holder, or unknown key"
// @Failure	401	{object}	object{error=string}	"Missing caller header"
// @Failure	403	{object}	object{error=string}	"Caller lacks privileges:grant"
// @Failure	404	{object}	object{error=string}	"No active grant for that key"
// @Router	/users/{id}/privileges/{key} [delete]
func (s *Server) revokePrivilege(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	key := models.PrivilegeKey(r.PathValue("key"))
	if userID == "" || key == "" {
		writeError(w, http.StatusBadRequest, errors.New("user id and privilege key are required"))
		return
	}
	revokerID := r.Header.Get(callerHeader)
	if revokerID == "" {
		writeError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	var req revokePrivilegeRequest
	_ = decodeJSON(r, &req)
	if err := s.svc.RevokePrivilege(r.Context(), userID, key, revokerID, req.Reason); err != nil {
		writeServiceError(w, err)
		return
	}
	s.authCache.invalidate(userID)
	s.authCache.invalidate(revokerID)
	w.WriteHeader(http.StatusNoContent)
}
