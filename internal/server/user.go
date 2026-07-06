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

// @Summary	Create a user
// @Tags	Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.User	true	"User payload"
// @Success	201	{object}	models.User
// @Failure	400	{object}	object{error=string}
// @Router	/users [post]
func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var u models.User
	if err := common.DecodeJSON(r, &u); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateUser(r.Context(), &u)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
}

// @Summary	List users (paginated)
// @Tags	Users
// @Security	BearerAuth
// @Produce	json
// @Param	limit	query	integer	false	"Page size"
// @Param	offset	query	integer	false	"Page offset"
// @Success	200	{object}	UserListResponse
// @Router	/users [get]
func (s *Server) listUsers(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows, total, err := s.svc.ListUsers(r.Context(), atoiOr(q.Get("limit"), 50), atoiOr(q.Get("offset"), 0))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	if rows == nil {
		rows = []models.User{}
	}
	common.WriteJSON(w, http.StatusOK, UserListResponse{Items: rows, Total: total})
}

// @Summary	Get a user by ID
// @Tags	Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{object}	models.User
// @Failure	404	{object}	object{error=string}
// @Router	/users/{id} [get]
func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	u, err := s.svc.GetUser(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, u)
}

// @Summary	Update a user's status
// @Tags	Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Param	request	body	object{status=models.UserStatus}	true	"Status patch"
// @Success	200	{object}	models.User
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/users/{id}/status [put]
func (s *Server) updateUserStatus(w http.ResponseWriter, r *http.Request) {
	var req statusUpdateRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	u, err := s.svc.UpdateUserStatus(r.Context(), r.PathValue("id"), models.UserStatus(req.Status))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, u)
}

type mergeUsersRequest struct {
	SurvivingUserID string `json:"surviving_user_id"`
	RetiringUserID  string `json:"retiring_user_id"`
}

// @Summary	Merge two users
// @Description	Merges the retiring user into the surviving user; the surviving record is returned.
// @Tags	Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	mergeUsersRequest	true	"Merge payload"
// @Success	200	{object}	models.User
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/users/merge [post]
func (s *Server) mergeUsers(w http.ResponseWriter, r *http.Request) {
	var req mergeUsersRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	survivor, err := s.svc.MergeUsers(r.Context(), req.SurvivingUserID, req.RetiringUserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, survivor)
}
