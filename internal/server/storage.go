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
	"github.com/apache/airavata-custos/pkg/models"
)

// @Summary	Search storages (paginated)
// @Tags	Storages
// @Security	BearerAuth
// @Produce	json
// @Param	cluster_id	query	string	false	"Only this compute cluster's"
// @Param	limit	query	integer	false	"Page size"
// @Param	offset	query	integer	false	"Page offset"
// @Success	200	{object}	object{items=[]models.Storage,total=int}
// @Router	/storages [get]
func (s *Server) listStorages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, offset := common.Paged(q)
	items, total, err := s.svc.ListStorages(r.Context(), q.Get("cluster_id"), limit, offset)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// @Summary	Create a compute cluster's storages
// @Description	Makes the cluster's storages match its storage policies and owners, as the storage connector's sweep does, and returns those made or changed.
// @Tags	Storages
// @Security	BearerAuth
// @Produce	json
// @Param	cluster_id	query	string	false	"Compute cluster ID; every cluster without one"
// @Success	200	{array}	models.Storage
// @Router	/storages [post]
func (s *Server) createStorage(w http.ResponseWriter, r *http.Request) {
	made, err := s.svc.EnsureStorages(r.Context(), r.URL.Query().Get("cluster_id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, made)
}

type updateStorageRequest struct {
	QuotaBytes int64      `json:"quota_bytes"`
	QuotaUntil *time.Time `json:"quota_until"`
	Reason     string     `json:"reason"`
}

// @Summary	Set a storage's temporary quota
// @Description	Lasts until a date at most 90 days ahead.
// @Tags	Storages
// @Security	BearerAuth
// @Accept	json
// @Param	id	path	string	true	"Storage ID"
// @Param	request	body	updateStorageRequest	true	"Quota payload"
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
	if err := s.svc.SetStorageQuota(r.Context(), r.PathValue("id"), req.QuotaBytes, req.QuotaUntil, req.Reason, caller.UserID); err != nil {
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

// @Summary	List a compute cluster's storage policies
// @Tags	Storages
// @Security	BearerAuth
// @Produce	json
// @Param	cluster_id	query	string	true	"Compute cluster ID"
// @Success	200	{array}	models.StoragePolicy
// @Router	/storage-policies [get]
func (s *Server) listStoragePolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := s.svc.ListStoragePolicies(r.Context(), r.URL.Query().Get("cluster_id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, policies)
}

// @Summary	Add a storage policy to a compute cluster
// @Description	Says where the cluster's storages of one type go and how they are made.
// @Tags	Storages
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.StoragePolicy	true	"Storage policy payload"
// @Success	201	{object}	models.StoragePolicy
// @Failure	400	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}
// @Router	/storage-policies [post]
func (s *Server) createStoragePolicy(w http.ResponseWriter, r *http.Request) {
	var p models.StoragePolicy
	if err := common.DecodeJSON(r, &p); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateStoragePolicy(r.Context(), &p)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
}

// @Summary	Update a storage policy
// @Description	Its cluster and type stay. A storage takes the change on its next create; a new path_pattern makes new storages.
// @Tags	Storages
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Storage policy ID"
// @Param	request	body	models.StoragePolicy	true	"Storage policy payload"
// @Success	200	{object}	models.StoragePolicy
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}
// @Router	/storage-policies/{id} [put]
func (s *Server) updateStoragePolicy(w http.ResponseWriter, r *http.Request) {
	var p models.StoragePolicy
	if err := common.DecodeJSON(r, &p); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	p.ID = r.PathValue("id")
	updated, err := s.svc.UpdateStoragePolicy(r.Context(), &p)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, updated)
}

// @Summary	Delete a storage policy
// @Description	Storages already made from it stay.
// @Tags	Storages
// @Security	BearerAuth
// @Param	id	path	string	true	"Storage policy ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/storage-policies/{id} [delete]
func (s *Server) deleteStoragePolicy(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteStoragePolicy(r.Context(), r.PathValue("id")); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
