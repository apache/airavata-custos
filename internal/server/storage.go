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
	"time"

	"github.com/apache/airavata-custos/pkg/common"
)

// @Summary	List storages
// @Tags	Storages
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	models.Storage
// @Router	/storages [get]
func (s *Server) listStorages(w http.ResponseWriter, r *http.Request) {
	items, err := s.svc.ListStorages(r.Context())
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, items)
}

type updateStorageRequest struct {
	QuotaBytes int64      `json:"quota_bytes"`
	QuotaUntil *time.Time `json:"quota_until"`
	Reason     string     `json:"reason"`
}

// @Summary	Update a storage
// @Description	Sets its temporary quota, which lasts until a date at most 90 days ahead.
// @Tags	Storages
// @Security	BearerAuth
// @Accept	json
// @Param	id	path	string	true	"Storage ID"
// @Param	request	body	updateStorageRequest	true	"Storage update"
// @Success	204	"No Content"
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/storages/{id} [put]
func (s *Server) updateStorage(w http.ResponseWriter, r *http.Request) {
	var req updateStorageRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	caller := common.RequireCaller(w, r)
	if caller == nil {
		return
	}
	if err := s.svc.UpdateStorage(r.Context(), r.PathValue("id"), req.QuotaBytes, req.QuotaUntil, req.Reason, caller.UserID); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	Delete a pending storage
// @Description	It goes with the storages under it, once all are pending.
// @Tags	Storages
// @Security	BearerAuth
// @Param	id	path	string	true	"Storage ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/storages/{id} [delete]
func (s *Server) deleteStorage(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteStorage(r.Context(), r.PathValue("id")); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
