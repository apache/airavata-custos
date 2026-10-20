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
	"context"
	"errors"
	"net/http"

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// deletionPrivileges is what reviewing a deletion of each kind needs, the same as deleting it did.
var deletionPrivileges = map[string]models.PrivilegeKey{
	service.DeletionOfClusterUser: models.ClustersWrite,
	service.DeletionOfAllocation:  models.AllocationsWrite,
}

// @Summary	List pending deletions
// @Description	Cluster accounts and allocations whose deletion waits for an admin's approval, oldest first.
// @Tags	Deletion Requests
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	models.DeletionRequest
// @Router	/deletion-requests [get]
func (s *Server) listDeletionRequests(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListDeletionRequests(r.Context())
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, rows)
}

// @Summary	Approve a deletion
// @Description	Deletes the cluster account or allocation, which removes it from the cluster.
// @Tags	Deletion Requests
// @Security	BearerAuth
// @Param	entity_type	path	string	true	"compute_cluster_user | compute_allocation"
// @Param	id	path	string	true	"Cluster account or allocation ID"
// @Success	204	"No Content"
// @Failure	403	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}	"No pending deletion"
// @Router	/deletion-requests/{entity_type}/{id}/approve [post]
func (s *Server) approveDeletion(w http.ResponseWriter, r *http.Request) {
	s.reviewDeletion(w, r, s.svc.ApproveDeletion)
}

// @Summary	Deny a deletion
// @Description	Drops the request and keeps the cluster account or allocation.
// @Tags	Deletion Requests
// @Security	BearerAuth
// @Param	entity_type	path	string	true	"compute_cluster_user | compute_allocation"
// @Param	id	path	string	true	"Cluster account or allocation ID"
// @Success	204	"No Content"
// @Failure	403	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}	"No pending deletion"
// @Router	/deletion-requests/{entity_type}/{id}/deny [post]
func (s *Server) denyDeletion(w http.ResponseWriter, r *http.Request) {
	s.reviewDeletion(w, r, s.svc.DenyDeletion)
}

// canReviewDeletion allows a caller who may delete the entity's kind.
func canReviewDeletion(w http.ResponseWriter, r *http.Request) bool {
	privilege, ok := deletionPrivileges[r.PathValue("entity_type")]
	if !ok || !identity.HasPrivilege(r.Context(), privilege) {
		common.WriteError(w, http.StatusForbidden, errors.New("caller may not review this deletion"))
		return false
	}
	return true
}

func (s *Server) reviewDeletion(w http.ResponseWriter, r *http.Request, review func(ctx context.Context, entityType, entityID, actorID string) error) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	if err := review(r.Context(), r.PathValue("entity_type"), r.PathValue("id"), caller.UserID); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// requestDeletion records the caller's request to delete the entity, which an admin then approves or denies.
func (s *Server) requestDeletion(w http.ResponseWriter, r *http.Request, entityType string) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	if err := s.svc.RequestDeletion(r.Context(), entityType, r.PathValue("id"), caller.UserID); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
