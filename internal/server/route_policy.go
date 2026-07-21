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
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// canReadAllocation is the RequireScoped policy for allocation reads keyed by
// the {id} path value. It admits the read privilege, a governance role on the
// parent project, or an active membership on the allocation. Everyone else
// gets the same 404 as a missing id, so ids cannot be probed.
func (s *Server) canReadAllocation(w http.ResponseWriter, r *http.Request) bool {
	caller := requireCaller(w, r)
	if caller == nil {
		return false
	}
	if identity.HasPrivilege(r.Context(), models.AllocationsRead) {
		return true
	}
	alloc, err := s.svc.GetComputeAllocation(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return false
	}
	role, err := s.svc.ProjectRoleForUser(r.Context(), alloc.ProjectID, caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return false
	}
	if models.IsGovernanceRole(role) {
		return true
	}
	member, err := s.svc.IsAllocationMember(r.Context(), alloc.ID, caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return false
	}
	if !member {
		common.WriteServiceError(w, service.ErrNotFound)
		return false
	}
	return true
}

// canReadProject is the RequireScoped policy for project reads keyed by the
// {id} path value. It admits the read privilege or a project participant (a
// governance role or an active allocation membership). Everyone else gets the
// same 404 as a missing id.
func (s *Server) canReadProject(w http.ResponseWriter, r *http.Request) bool {
	caller := requireCaller(w, r)
	if caller == nil {
		return false
	}
	if identity.HasPrivilege(r.Context(), models.ProjectsRead) {
		return true
	}
	ok, err := s.svc.IsProjectParticipant(r.Context(), r.PathValue("id"), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return false
	}
	if !ok {
		common.WriteServiceError(w, service.ErrNotFound)
		return false
	}
	return true
}
