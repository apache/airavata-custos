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
	"github.com/apache/airavata-custos/pkg/models"
)

// @Summary	List all roles
// @Tags	Roles
// @Security	CustosUserHeader
// @Produce	json
// @Success	200	{array}	models.Role
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks roles:manage"
// @Router	/roles [get]
func (s *Server) listRoles(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListRoles(r.Context())
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
}

// @Summary	Get a role with its privilege bundle
// @Tags	Roles
// @Security	CustosUserHeader
// @Produce	json
// @Param	id	path	string	true	"Role ID"
// @Success	200	{object}	object{role=models.Role,privileges=[]models.PrivilegeKey}
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/roles/{id} [get]
func (s *Server) getRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("role id is required"))
		return
	}
	role, err := s.svc.GetRole(r.Context(), roleID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	keys, err := s.svc.ListRolePrivileges(r.Context(), roleID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{
		"role":       role,
		"privileges": keys,
	})
}

type createRoleRequest struct {
	Name        string `JSON:"name"`
	Description string `JSON:"description"`
}

// @Summary	Create a role
// @Tags	Roles
// @Security	CustosUserHeader
// @Accept	json
// @Produce	json
// @Param	request	body	createRoleRequest	true	"Role payload"
// @Success	201	{object}	models.Role
// @Failure	400	{object}	object{error=string}
// @Failure	401	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}	"Role name collides"
// @Router	/roles [post]
func (s *Server) createRole(w http.ResponseWriter, r *http.Request) {
	actorID := r.Header.Get(callerHeader)
	if actorID == "" {
		common.WriteError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	var req createRoleRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	role, err := s.svc.CreateRole(r.Context(), req.Name, req.Description, actorID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, role)
}

type updateRoleRequest struct {
	Name        string `JSON:"name"`
	Description string `JSON:"description"`
}

// @Summary	Update role name / description
// @Description	System roles cannot be renamed.
// @Tags	Roles
// @Security	CustosUserHeader
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Role ID"
// @Param	request	body	updateRoleRequest	true	"Role patch"
// @Success	200	{object}	models.Role
// @Failure	400	{object}	object{error=string}
// @Failure	401	{object}	object{error=string}
// @Router	/roles/{id} [put]
func (s *Server) updateRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("role id is required"))
		return
	}
	actorID := r.Header.Get(callerHeader)
	if actorID == "" {
		common.WriteError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	var req updateRoleRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	role, err := s.svc.UpdateRole(r.Context(), roleID, req.Name, req.Description, actorID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.authCache.invalidateAll()
	common.WriteJSON(w, http.StatusOK, role)
}

// @Summary	Delete a role
// @Description	System roles cannot be deleted. CASCADE drops every assignment of this role.
// @Tags	Roles
// @Security	CustosUserHeader
// @Param	id	path	string	true	"Role ID"
// @Success	204	"No Content"
// @Failure	400	{object}	object{error=string}	"System role / unknown role"
// @Failure	401	{object}	object{error=string}
// @Router	/roles/{id} [delete]
func (s *Server) deleteRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("role id is required"))
		return
	}
	actorID := r.Header.Get(callerHeader)
	if actorID == "" {
		common.WriteError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	if err := s.svc.DeleteRole(r.Context(), roleID, actorID); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.authCache.invalidateAll()
	w.WriteHeader(http.StatusNoContent)
}

type rolePrivilegeRequest struct {
	Privilege models.PrivilegeKey `JSON:"privilege"`
}

// @Summary	Add a privilege to a role
// @Description	The new privilege propagates to every current holder.
// @Tags	Roles
// @Security	CustosUserHeader
// @Accept	json
// @Param	id	path	string	true	"Role ID"
// @Param	request	body	rolePrivilegeRequest	true	"Privilege key"
// @Success	204	"No Content"
// @Failure	400	{object}	object{error=string}	"Unknown key"
// @Failure	401	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}	"Role already carries that privilege"
// @Router	/roles/{id}/privileges [post]
func (s *Server) addRolePrivilege(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("role id is required"))
		return
	}
	actorID := r.Header.Get(callerHeader)
	if actorID == "" {
		common.WriteError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	var req rolePrivilegeRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.svc.AddPrivilegeToRole(r.Context(), roleID, req.Privilege, actorID); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.authCache.invalidateAll()
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	Remove a privilege from a role
// @Description	Removal propagates to every holder. Refuses to remove `privileges:grant` or `roles:manage` if that would leave no role anywhere carrying it.
// @Tags	Roles
// @Security	CustosUserHeader
// @Param	id	path	string	true	"Role ID"
// @Param	key	path	models.PrivilegeKey	true	"Privilege key"
// @Success	204	"No Content"
// @Failure	400	{object}	object{error=string}	"Last-source guard or unknown key"
// @Failure	401	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}	"Role does not carry that privilege"
// @Router	/roles/{id}/privileges/{key} [delete]
func (s *Server) removeRolePrivilege(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	key := models.PrivilegeKey(r.PathValue("key"))
	if roleID == "" || key == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("role id and privilege key are required"))
		return
	}
	actorID := r.Header.Get(callerHeader)
	if actorID == "" {
		common.WriteError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	if err := s.svc.RemovePrivilegeFromRole(r.Context(), roleID, key, actorID); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.authCache.invalidateAll()
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List roles a user holds
// @Tags	Role Assignments
// @Security	CustosUserHeader
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{array}	models.UserRole
// @Failure	400	{object}	object{error=string}
// @Router	/users/{id}/roles [get]
func (s *Server) listUserRoles(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("user id is required"))
		return
	}
	rows, err := s.svc.ListUserRoles(r.Context(), userID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
}

// @Summary	List users holding the role
// @Tags	Roles
// @Security	CustosUserHeader
// @Produce	json
// @Param	id	path	string	true	"Role ID"
// @Success	200	{array}	models.UserRole
// @Failure	400	{object}	object{error=string}
// @Router	/roles/{id}/holders [get]
func (s *Server) listRoleHolders(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("role id is required"))
		return
	}
	rows, err := s.svc.ListRoleHolders(r.Context(), roleID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
}

type grantRoleRequest struct {
	RoleID string `JSON:"role_id"`
	Reason string `JSON:"reason"`
}

// @Summary	Grant a role to a user
// @Tags	Role Assignments
// @Security	CustosUserHeader
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Param	request	body	grantRoleRequest	true	"Role assignment payload"
// @Success	201	{object}	models.UserRole
// @Failure	400	{object}	object{error=string}
// @Failure	401	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}	"User already holds that role"
// @Router	/users/{id}/roles [post]
func (s *Server) grantRoleToUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("user id is required"))
		return
	}
	granterID := r.Header.Get(callerHeader)
	if granterID == "" {
		common.WriteError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	var req grantRoleRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	assignment, err := s.svc.GrantRoleToUser(r.Context(), userID, req.RoleID, granterID, req.Reason)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.authCache.invalidate(userID)
	common.WriteJSON(w, http.StatusCreated, assignment)
}

type revokeRoleRequest struct {
	Reason string `JSON:"reason"`
}

// @Summary	Revoke a role from a user
// @Description	Refuses if revoking would leave no holder of `privileges:grant` or `roles:manage` anywhere (last-meta-holder guard).
// @Tags	Role Assignments
// @Security	CustosUserHeader
// @Accept	json
// @Param	id	path	string	true	"User ID"
// @Param	roleId	path	string	true	"Role ID"
// @Param	request	body	revokeRoleRequest	false	"Optional reason"
// @Success	204	"No Content"
// @Failure	400	{object}	object{error=string}	"Last-meta-holder guard"
// @Failure	401	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}	"User does not hold that role"
// @Router	/users/{id}/roles/{roleId} [delete]
func (s *Server) revokeRoleFromUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	roleID := r.PathValue("roleId")
	if userID == "" || roleID == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("user id and role id are required"))
		return
	}
	revokerID := r.Header.Get(callerHeader)
	if revokerID == "" {
		common.WriteError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	var req revokeRoleRequest
	_ = common.DecodeJSON(r, &req)
	if err := s.svc.RevokeRoleFromUser(r.Context(), userID, roleID, revokerID, req.Reason); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.authCache.invalidate(userID)
	s.authCache.invalidate(revokerID)
	w.WriteHeader(http.StatusNoContent)
}
