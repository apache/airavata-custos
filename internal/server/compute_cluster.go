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

// @Summary	Create a compute cluster
// @Tags	Compute Clusters
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeCluster	true	"Cluster payload"
// @Success	201	{object}	models.ComputeCluster
// @Failure	400	{object}	object{error=string}
// @Router	/compute-clusters [post]
func (s *Server) createComputeCluster(w http.ResponseWriter, r *http.Request) {
	var c models.ComputeCluster
	if err := common.DecodeJSON(r, &c); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeCluster(r.Context(), &c)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
}

// @Summary	Get a compute cluster by ID
// @Tags	Compute Clusters
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute cluster ID"
// @Success	200	{object}	models.ComputeCluster
// @Failure	404	{object}	object{error=string}
// @Router	/compute-clusters/{id} [get]
func (s *Server) getComputeCluster(w http.ResponseWriter, r *http.Request) {
	c, err := s.svc.GetComputeCluster(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, c)
}

// @Summary	List compute clusters
// @Tags	Compute Clusters
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	models.ComputeCluster
// @Failure	500	{object}	object{error=string}
// @Router	/compute-clusters [get]
func (s *Server) listComputeClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := s.svc.ListComputeClusters(r.Context())
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, clusters)
}

// @Summary	Create a compute cluster user
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	models.ComputeClusterUser	true	"Cluster user payload"
// @Success	201	{object}	models.ComputeClusterUser
// @Failure	400	{object}	object{error=string}
// @Router	/compute-cluster-users [post]
func (s *Server) createComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	var cu models.ComputeClusterUser
	if err := common.DecodeJSON(r, &cu); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	created, err := s.svc.CreateComputeClusterUser(r.Context(), &cu)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
}

// @Summary	Get a compute cluster user by ID
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute cluster user ID"
// @Success	200	{object}	models.ComputeClusterUser
// @Failure	404	{object}	object{error=string}
// @Router	/compute-cluster-users/{id} [get]
func (s *Server) getComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	cu, err := s.svc.GetComputeClusterUser(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, cu)
}

// @Summary	Update a compute cluster user
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Compute cluster user ID"
// @Param	request	body	models.ComputeClusterUser	true	"Cluster user payload"
// @Success	200	{object}	models.ComputeClusterUser
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-cluster-users/{id} [put]
func (s *Server) updateComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	var cu models.ComputeClusterUser
	if err := common.DecodeJSON(r, &cu); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	cu.ID = r.PathValue("id")
	if err := s.svc.UpdateComputeClusterUser(r.Context(), &cu); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, &cu)
}

// @Summary	Delete a compute cluster user
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Param	id	path	string	true	"Compute cluster user ID"
// @Success	204	"No Content"
// @Failure	404	{object}	object{error=string}
// @Router	/compute-cluster-users/{id} [delete]
func (s *Server) deleteComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.DeleteComputeClusterUser(r.Context(), r.PathValue("id")); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	List users on a compute cluster
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute cluster ID"
// @Success	200	{array}	models.ComputeClusterUser
// @Failure	404	{object}	object{error=string}
// @Router	/compute-clusters/{id}/users [get]
func (s *Server) listComputeClusterUsersByCluster(w http.ResponseWriter, r *http.Request) {
	users, err := s.svc.ListComputeClusterUsersByCluster(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, users)
}

// @Summary	Get a compute cluster user by (cluster, user) pair
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute cluster ID"
// @Param	userId	path	string	true	"User ID"
// @Success	200	{object}	models.ComputeClusterUser
// @Failure	404	{object}	object{error=string}
// @Router	/compute-clusters/{id}/users/{userId} [get]
func (s *Server) getComputeClusterUserByPair(w http.ResponseWriter, r *http.Request) {
	cu, err := s.svc.GetComputeClusterUserByPair(r.Context(), r.PathValue("id"), r.PathValue("userId"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, cu)
}

// @Summary	List compute cluster users for a user
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Success	200	{array}	models.ComputeClusterUser
// @Failure	404	{object}	object{error=string}
// @Router	/users/{id}/compute-cluster-users [get]
func (s *Server) listComputeClusterUsersByUser(w http.ResponseWriter, r *http.Request) {
	users, err := s.svc.ListComputeClusterUsersByUser(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, users)
}
