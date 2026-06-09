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

// getCallerPrivileges handles GET /user/privileges. Returns the
// authenticated caller's effective privilege set.
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

// getPrivilegeCatalog handles GET /privileges/catalog. Returns the static
// catalog of declared privilege keys. Gated on privileges:grant.
func (s *Server) getPrivilegeCatalog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.svc.PrivilegeCatalog())
}

// listUserPrivileges handles GET /users/{id}/privileges. Returns the active
// privileges of the target user. Gated on privileges:grant.
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

// listPrivilegeHolders handles GET /privileges/{key}/holders. Gated on
// privileges:grant.
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

// grantPrivilege handles POST /users/{id}/privileges. Gated on privileges:grant.
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

// revokePrivilege handles DELETE /users/{id}/privileges/{key}. Gated on
// privileges:grant.
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
