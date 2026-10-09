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
	"strings"

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

type createComputeClusterUserRequest struct {
	ComputeClusterID string `json:"compute_cluster_id"`
	UserID           string `json:"user_id"`
	LocalUsername    string `json:"local_username"`
}

// @Summary	Create a compute cluster user
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	createComputeClusterUserRequest	true	"Cluster user payload"
// @Success	201	{object}	models.ComputeClusterUser
// @Failure	400	{object}	object{error=string}
// @Router	/compute-cluster-users [post]
func (s *Server) createComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	var req createComputeClusterUserRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	// Access level is left out on purpose. Granting cluster admin goes through
	// user onboarding, which checks roles:manage first.
	// TODO - accept an access level here once the caller can be checked for ADMIN on the same cluster
	// An admin creating the account by hand is the approval.
	created, err := s.svc.CreateComputeClusterUser(r.Context(), &models.ComputeClusterUser{
		ComputeClusterID: req.ComputeClusterID,
		UserID:           req.UserID,
		LocalUsername:    req.LocalUsername,
		ApprovalStatus:   models.ClusterAccountApproved,
		ReviewedBy:       &caller.UserID,
	})
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
// @Param	request	body	createComputeClusterUserRequest	true	"Cluster user payload"
// @Success	200	{object}	models.ComputeClusterUser
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/compute-cluster-users/{id} [put]
func (s *Server) updateComputeClusterUser(w http.ResponseWriter, r *http.Request) {
	var req createComputeClusterUserRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	// The store does not write access_level, so access_level is not expected here.
	// TODO - keep this behavior until there's a need to update the cluster access_level
	cu := models.ComputeClusterUser{
		ID:               r.PathValue("id"),
		ComputeClusterID: req.ComputeClusterID,
		UserID:           req.UserID,
		LocalUsername:    req.LocalUsername,
	}
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

// @Summary	List cluster accounts by approval status
// @Description	Cluster users with their user and cluster, newest first. Every status unless one is given.
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Produce	json
// @Param	approval_status	query	string	false	"PENDING | APPROVED | DENIED"
// @Param	limit	query	integer	false	"Page size"
// @Param	offset	query	integer	false	"Page offset"
// @Success	200	{object}	ClusterAccountListResponse
// @Failure	400	{object}	object{error=string}
// @Router	/compute-cluster-users [get]
func (s *Server) listClusterAccounts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	status := models.ClusterAccountApproval(q.Get("approval_status"))
	rows, total, err := s.svc.ListComputeClusterUsersByApproval(r.Context(), status, atoiOr(q.Get("limit"), 50), atoiOr(q.Get("offset"), 0))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	items := make([]ClusterAccountResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, ClusterAccountResponse{
			ComputeClusterUser: row.ComputeClusterUser,
			DisplayName:        row.DisplayName,
			Email:              row.Email,
			ClusterName:        row.ClusterName,
		})
	}
	common.WriteJSON(w, http.StatusOK, ClusterAccountListResponse{Items: items, Total: total})
}

// @Summary	Approve a cluster account
// @Description	Records the caller as the approver and starts creating the account on the cluster.
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Compute cluster user ID"
// @Success	200	{object}	models.ComputeClusterUser
// @Failure	404	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}	"Already approved"
// @Router	/compute-cluster-users/{id}/approve [post]
func (s *Server) approveClusterAccount(w http.ResponseWriter, r *http.Request) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	cu, err := s.svc.ApproveComputeClusterUser(r.Context(), r.PathValue("id"), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, cu)
}

type denyClusterAccountRequest struct {
	Note string `json:"note"`
}

// @Summary	Deny a cluster account
// @Description	Records the caller's decision and the reason. No account is created on the cluster, and the reason is passed on to whoever requested the account. The user keeps portal access.
// @Tags	Compute Cluster Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Compute cluster user ID"
// @Param	request	body	denyClusterAccountRequest	true	"Reason for the denial"
// @Success	200	{object}	models.ComputeClusterUser
// @Failure	400	{object}	object{error=string}	"Reason missing"
// @Failure	404	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}	"Already approved or denied"
// @Router	/compute-cluster-users/{id}/deny [post]
func (s *Server) denyClusterAccount(w http.ResponseWriter, r *http.Request) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	var req denyClusterAccountRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	cu, err := s.svc.DenyComputeClusterUser(r.Context(), r.PathValue("id"), caller.UserID, strings.TrimSpace(req.Note))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, cu)
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
