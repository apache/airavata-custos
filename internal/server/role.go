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

func (s *Server) listRoles(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListRoles(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// getRole returns the role plus its privilege bundle in one response.
func (s *Server) getRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		writeError(w, http.StatusBadRequest, errors.New("role id is required"))
		return
	}
	role, err := s.svc.GetRole(r.Context(), roleID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	keys, err := s.svc.ListRolePrivileges(r.Context(), roleID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"role":       role,
		"privileges": keys,
	})
}

type createRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Server) createRole(w http.ResponseWriter, r *http.Request) {
	actorID := r.Header.Get(callerHeader)
	if actorID == "" {
		writeError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	var req createRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	role, err := s.svc.CreateRole(r.Context(), req.Name, req.Description, actorID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, role)
}

type updateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (s *Server) updateRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		writeError(w, http.StatusBadRequest, errors.New("role id is required"))
		return
	}
	actorID := r.Header.Get(callerHeader)
	if actorID == "" {
		writeError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	var req updateRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	role, err := s.svc.UpdateRole(r.Context(), roleID, req.Name, req.Description, actorID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	s.authCache.invalidateAll()
	writeJSON(w, http.StatusOK, role)
}

func (s *Server) deleteRole(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		writeError(w, http.StatusBadRequest, errors.New("role id is required"))
		return
	}
	actorID := r.Header.Get(callerHeader)
	if actorID == "" {
		writeError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	if err := s.svc.DeleteRole(r.Context(), roleID, actorID); err != nil {
		writeServiceError(w, err)
		return
	}
	s.authCache.invalidateAll()
	w.WriteHeader(http.StatusNoContent)
}

type rolePrivilegeRequest struct {
	Privilege models.PrivilegeKey `json:"privilege"`
}

func (s *Server) addRolePrivilege(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		writeError(w, http.StatusBadRequest, errors.New("role id is required"))
		return
	}
	actorID := r.Header.Get(callerHeader)
	if actorID == "" {
		writeError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	var req rolePrivilegeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.svc.AddPrivilegeToRole(r.Context(), roleID, req.Privilege, actorID); err != nil {
		writeServiceError(w, err)
		return
	}
	s.authCache.invalidateAll()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) removeRolePrivilege(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	key := models.PrivilegeKey(r.PathValue("key"))
	if roleID == "" || key == "" {
		writeError(w, http.StatusBadRequest, errors.New("role id and privilege key are required"))
		return
	}
	actorID := r.Header.Get(callerHeader)
	if actorID == "" {
		writeError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	if err := s.svc.RemovePrivilegeFromRole(r.Context(), roleID, key, actorID); err != nil {
		writeServiceError(w, err)
		return
	}
	s.authCache.invalidateAll()
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listUserRoles(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, errors.New("user id is required"))
		return
	}
	rows, err := s.svc.ListUserRoles(r.Context(), userID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func (s *Server) listRoleHolders(w http.ResponseWriter, r *http.Request) {
	roleID := r.PathValue("id")
	if roleID == "" {
		writeError(w, http.StatusBadRequest, errors.New("role id is required"))
		return
	}
	rows, err := s.svc.ListRoleHolders(r.Context(), roleID)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

type grantRoleRequest struct {
	RoleID string `json:"role_id"`
	Reason string `json:"reason"`
}

func (s *Server) grantRoleToUser(w http.ResponseWriter, r *http.Request) {
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
	var req grantRoleRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	assignment, err := s.svc.GrantRoleToUser(r.Context(), userID, req.RoleID, granterID, req.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	s.authCache.invalidate(userID)
	writeJSON(w, http.StatusCreated, assignment)
}

type revokeRoleRequest struct {
	Reason string `json:"reason"`
}

func (s *Server) revokeRoleFromUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	roleID := r.PathValue("roleId")
	if userID == "" || roleID == "" {
		writeError(w, http.StatusBadRequest, errors.New("user id and role id are required"))
		return
	}
	revokerID := r.Header.Get(callerHeader)
	if revokerID == "" {
		writeError(w, http.StatusUnauthorized, errors.New("missing "+callerHeader+" header"))
		return
	}
	var req revokeRoleRequest
	_ = decodeJSON(r, &req)
	if err := s.svc.RevokeRoleFromUser(r.Context(), userID, roleID, revokerID, req.Reason); err != nil {
		writeServiceError(w, err)
		return
	}
	s.authCache.invalidate(userID)
	s.authCache.invalidate(revokerID)
	w.WriteHeader(http.StatusNoContent)
}
