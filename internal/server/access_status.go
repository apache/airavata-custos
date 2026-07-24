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
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

// UI states derived from a check's raw status plus how long it has been
// failing.
const (
	accessUIStateSettingUp = "setting_up"
	accessUIStateOK        = "ok"
	accessUIStateRetrying  = "retrying"
	accessUIStateStuck     = "stuck"
)

type accessCheckResponse struct {
	Type          models.AccessCheckType   `json:"type"`
	Status        models.AccessCheckStatus `json:"status"`
	UIState       string                   `json:"ui_state"`
	LastCheckedAt time.Time                `json:"last_checked_at"`
	LastOKAt      *time.Time               `json:"last_ok_at,omitempty"`
	FailingSince  *time.Time               `json:"failing_since,omitempty"`
}

type accessCheckEventResponse struct {
	CheckType  models.AccessCheckType `json:"check_type"`
	EventType  string                 `json:"event_type"`
	OccurredAt time.Time              `json:"occurred_at"`
}

type accessStatusResponse struct {
	Checks []accessCheckResponse      `json:"checks"`
	Events []accessCheckEventResponse `json:"events"`
}

type memberAccessStatusResponse struct {
	UserID      string                `json:"user_id"`
	DisplayName string                `json:"display_name"`
	Email       string                `json:"email"`
	Checks      []accessCheckResponse `json:"checks"`
}

type allocationAccessMembersResponse struct {
	Members []memberAccessStatusResponse `json:"members"`
}

// getAllocationAccessStatus serves the caller's own checks; with
// members=true, managers and read-privileged staff get every ACTIVE
// member's checks instead. A plain member asking for members=true silently
// gets the self view, mirroring the analytics scope-narrowing convention.
func (s *Server) getAllocationAccessStatus(w http.ResponseWriter, r *http.Request) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	allocationID := r.PathValue("id")

	if r.URL.Query().Get("members") == "true" && s.canViewMemberAccess(r, allocationID, caller.UserID) {
		rows, err := s.svc.AccessStatusForAllocation(r.Context(), allocationID)
		if err != nil {
			common.WriteServiceError(w, err)
			return
		}
		resp := allocationAccessMembersResponse{Members: make([]memberAccessStatusResponse, 0, len(rows))}
		now := time.Now().UTC()
		for _, m := range rows {
			resp.Members = append(resp.Members, memberAccessStatusResponse{
				UserID:      m.UserID,
				DisplayName: m.DisplayName,
				Email:       m.Email,
				Checks:      s.checkResponses(m.Checks, now),
			})
		}
		common.WriteJSON(w, http.StatusOK, resp)
		return
	}

	status, err := s.svc.AccessStatusForUser(r.Context(), allocationID, caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	now := time.Now().UTC()
	resp := accessStatusResponse{
		Checks: s.checkResponses(status.Checks, now),
		Events: make([]accessCheckEventResponse, 0, len(status.Events)),
	}
	typeByCheckID := make(map[string]models.AccessCheckType, len(status.Checks))
	for _, c := range status.Checks {
		if c.ID != "" {
			typeByCheckID[c.ID] = c.CheckType
		}
	}
	for _, e := range status.Events {
		resp.Events = append(resp.Events, accessCheckEventResponse{
			CheckType:  typeByCheckID[e.AccessCheckID],
			EventType:  e.EventType,
			OccurredAt: e.OccurredAt,
		})
	}
	common.WriteJSON(w, http.StatusOK, resp)
}

func (s *Server) canViewMemberAccess(r *http.Request, allocationID, userID string) bool {
	if identity.HasPrivilege(r.Context(), models.AllocationsRead) {
		return true
	}
	alloc, err := s.svc.GetComputeAllocation(r.Context(), allocationID)
	if err != nil {
		return false
	}
	role, err := s.svc.ProjectRoleForUser(r.Context(), alloc.ProjectID, userID)
	if err != nil {
		return false
	}
	return models.IsGovernanceRole(role)
}

func (s *Server) checkResponses(checks []models.AccessCheck, now time.Time) []accessCheckResponse {
	out := make([]accessCheckResponse, 0, len(checks))
	for _, c := range checks {
		out = append(out, accessCheckResponse{
			Type:          c.CheckType,
			Status:        c.Status,
			UIState:       s.accessUIState(c, now),
			LastCheckedAt: c.LastCheckedAt,
			LastOKAt:      c.LastOKAt,
			FailingSince:  c.FailingSince,
		})
	}
	return out
}

func (s *Server) accessUIState(c models.AccessCheck, now time.Time) string {
	switch c.Status {
	case models.AccessCheckOK:
		return accessUIStateOK
	case models.AccessCheckFailing:
		if c.FailingSince != nil && now.Sub(*c.FailingSince) >= s.svc.AccessCheckStuckAfter() {
			return accessUIStateStuck
		}
		return accessUIStateRetrying
	default:
		return accessUIStateSettingUp
	}
}
