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
	"strings"
	"time"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

type createAccessRequestBody struct {
	Institution     string `json:"institution"`
	DesiredUsername string `json:"desired_username"`
	EventCode       string `json:"event_code"`
	Reason          string `json:"reason"`
}

type decideAccessRequestBody struct {
	Status     string     `json:"status"`
	ExpiresAt  *time.Time `json:"expires_at"`
	DenyReason string     `json:"deny_reason"`
}

// @Summary	Resolve an access event code
// @Tags	Access Requests
// @Security	BearerAuth
// @Produce	json
// @Param	code	path	string	true	"Event code"
// @Success	200	{object}	object{code=string,name=string}
// @Failure	404	{object}	object{error=string}
// @Router	/access-requests/events/{code} [get]
func (s *Server) getAccessEventByCode(w http.ResponseWriter, r *http.Request) {
	ev, err := s.svc.GetAccessEvent(r.Context(), r.PathValue("code"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	alloc, err := s.svc.GetComputeAllocation(r.Context(), ev.ComputeAllocationID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	// Only the code and allocation name go out; the caller is not a user yet.
	common.WriteJSON(w, http.StatusOK, map[string]string{"code": ev.Code, "name": alloc.Name})
}

// @Summary	Check a requested cluster username
// @Description	Returns an available suggested login and, when a candidate is given, whether it is well-formed and free on the event's cluster.
// @Tags	Access Requests
// @Security	BearerAuth
// @Produce	json
// @Param	event_code	query	string	true	"Event code"
// @Param	username	query	string	false	"Candidate username"
// @Success	200	{object}	object{suggestion=string,valid=bool,available=bool}
// @Failure	404	{object}	object{error=string}
// @Router	/access-requests/username [get]
func (s *Server) checkAccessRequestUsername(w http.ResponseWriter, r *http.Request) {
	claims := identity.ClaimsFromContext(r.Context())
	if claims == nil {
		common.WriteError(w, http.StatusUnauthorized, errors.New("verified token required"))
		return
	}
	q := r.URL.Query()
	name := claims.Name
	if name == "" {
		name = strings.TrimSpace(claims.GivenName + " " + claims.FamilyName)
	}
	res, err := s.svc.CheckAccessRequestUsername(r.Context(), q.Get("event_code"), strings.TrimSpace(q.Get("username")), name, claims.Email)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{
		"suggestion": res.Suggestion,
		"valid":      res.Valid,
		"available":  res.Available,
	})
}

// @Summary	Submit an access request
// @Tags	Access Requests
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	object{institution=string,event_code=string,reason=string}	true	"Access request payload"
// @Success	201	{object}	models.AccessRequest
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}
// @Router	/access-requests [post]
func (s *Server) createAccessRequest(w http.ResponseWriter, r *http.Request) {
	claims := identity.ClaimsFromContext(r.Context())
	if claims == nil {
		common.WriteError(w, http.StatusUnauthorized, errors.New("verified token required"))
		return
	}
	var body createAccessRequestBody
	if err := common.DecodeJSON(r, &body); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	if body.Institution == "" || body.EventCode == "" {
		common.WriteError(w, http.StatusBadRequest, errors.New("institution and event_code are required"))
		return
	}
	name := claims.Name
	if name == "" {
		name = strings.TrimSpace(claims.GivenName + " " + claims.FamilyName)
	}
	created, err := s.svc.CreateAccessRequest(r.Context(), &models.AccessRequest{
		OIDCSub:         claims.Sub,
		Email:           claims.Email,
		Name:            name,
		Institution:     body.Institution,
		DesiredUsername: strings.TrimSpace(body.DesiredUsername),
		EventCode:       body.EventCode,
		Reason:          body.Reason,
	})
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusCreated, created)
}

// @Summary	Get the caller's latest access request
// @Tags	Access Requests
// @Security	BearerAuth
// @Produce	json
// @Success	200	{object}	models.AccessRequest
// @Failure	404	{object}	object{error=string}
// @Router	/access-requests/me [get]
func (s *Server) getOwnAccessRequest(w http.ResponseWriter, r *http.Request) {
	claims := identity.ClaimsFromContext(r.Context())
	if claims == nil {
		common.WriteError(w, http.StatusUnauthorized, errors.New("verified token required"))
		return
	}
	req, err := s.svc.GetLatestAccessRequestBySub(r.Context(), claims.Sub)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, req)
}

// @Summary	List access requests
// @Tags	Access Requests
// @Security	BearerAuth
// @Produce	json
// @Param	status	query	string	false	"Filter by status"
// @Param	event	query	string	false	"Filter by event code"
// @Param	limit	query	integer	false	"Maximum rows (default 50)"
// @Success	200	{array}	AccessRequestListItem
// @Router	/access-requests [get]
func (s *Server) listAccessRequests(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	rows, err := s.svc.ListAccessRequests(r.Context(), store.AccessRequestListFilter{
		Status:    q.Get("status"),
		EventCode: q.Get("event"),
		Limit:     atoiOr(q.Get("limit"), 50),
	})
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	items, err := s.enrichAccessRequests(r.Context(), rows)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, items)
}

// enrichAccessRequests joins decision timestamps and the event's allocation
// onto the listed rows: one batch query for decision events plus one event
// lookup per distinct code.
func (s *Server) enrichAccessRequests(ctx context.Context, rows []models.AccessRequest) ([]AccessRequestListItem, error) {
	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	events, err := s.svc.ListAccessRequestDecisionEvents(ctx, ids)
	if err != nil {
		return nil, err
	}
	decidedAt := make(map[string]time.Time, len(events))
	for _, e := range events {
		// Events arrive timestamp-ascending, so the last write is the newest.
		decidedAt[e.AccessRequestID] = e.Timestamp
	}
	allocByCode := make(map[string]string)
	items := make([]AccessRequestListItem, 0, len(rows))
	for _, row := range rows {
		allocID, ok := allocByCode[row.EventCode]
		if !ok {
			ev, err := s.svc.GetAccessEvent(ctx, row.EventCode)
			if err != nil {
				return nil, err
			}
			allocID = ev.ComputeAllocationID
			allocByCode[row.EventCode] = allocID
		}
		item := AccessRequestListItem{AccessRequest: row, AllocationID: allocID}
		if ts, ok := decidedAt[row.ID]; ok {
			item.DecidedAt = &ts
		}
		items = append(items, item)
	}
	return items, nil
}

// @Summary	Approve or deny an access request
// @Tags	Access Requests
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Access request ID"
// @Param	request	body	object{status=string,expires_at=string,deny_reason=string}	true	"Decision payload"
// @Success	200	{object}	models.AccessRequest
// @Failure	400	{object}	object{error=string}
// @Failure	404	{object}	object{error=string}
// @Router	/access-requests/{id} [put]
func (s *Server) decideAccessRequest(w http.ResponseWriter, r *http.Request) {
	caller := identity.CallerFromContext(r.Context())
	if caller == nil {
		common.WriteError(w, http.StatusUnauthorized, errors.New("linked caller required"))
		return
	}
	var body decideAccessRequestBody
	if err := common.DecodeJSON(r, &body); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	var (
		updated *models.AccessRequest
		err     error
	)
	switch models.AccessRequestStatus(body.Status) {
	case models.AccessRequestApproved:
		updated, err = s.svc.ApproveAccessRequest(r.Context(), r.PathValue("id"), caller.UserID, body.ExpiresAt)
	case models.AccessRequestDenied:
		updated, err = s.svc.DenyAccessRequest(r.Context(), r.PathValue("id"), caller.UserID, body.DenyReason)
	default:
		common.WriteError(w, http.StatusBadRequest, errors.New("status must be APPROVED or DENIED"))
		return
	}
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, updated)
}
