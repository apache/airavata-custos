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

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

type actorView struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type userAgreementListItem struct {
	models.UserAgreement
	AcceptedCount int `json:"accepted_count"`
	PendingCount  int `json:"pending_count"`
}

type userAgreementDetail struct {
	models.UserAgreement
	AcceptedCount int        `json:"accepted_count"`
	PendingCount  int        `json:"pending_count"`
	CreatedBy     *actorView `json:"created_by"`
	PublishedBy   *actorView `json:"published_by"`
	RemovedBy     *actorView `json:"removed_by"`
}

type acceptanceRow struct {
	User       actorView `json:"user"`
	AcceptedAt time.Time `json:"accepted_at"`
}

type pendingRow struct {
	User          actorView  `json:"user"`
	FirstSignInAt *time.Time `json:"first_sign_in_at"`
}

type acceptanceReport struct {
	Status   string          `json:"status"`
	Accepted []acceptanceRow `json:"accepted"`
	Pending  []pendingRow    `json:"pending"`
}

type userAgreementRequest struct {
	Title        string `json:"title"`
	VersionLabel string `json:"version_label"`
	Summary      string `json:"summary"`
	BodyMD       string `json:"body_md"`
	ChangeNote   string `json:"change_note"`
}

func (r userAgreementRequest) input() service.UserAgreementInput {
	return service.UserAgreementInput{Title: r.Title, VersionLabel: r.VersionLabel, Summary: r.Summary, BodyMD: r.BodyMD, ChangeNote: r.ChangeNote}
}

func userView(u *models.User) actorView {
	return actorView{ID: u.ID, Name: strings.TrimSpace(u.FirstName + " " + u.LastName), Email: u.Email}
}

// actor returns nil for an empty id or a user that no longer exists.
func (s *Server) actor(ctx context.Context, id string) (*actorView, error) {
	if id == "" {
		return nil, nil
	}
	u, err := s.svc.GetUser(ctx, id)
	if errors.Is(err, service.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	v := userView(u)
	return &v, nil
}

func (s *Server) agreementDetail(ctx context.Context, a *models.UserAgreement) (*userAgreementDetail, error) {
	d := &userAgreementDetail{UserAgreement: *a}
	var err error
	if d.AcceptedCount, d.PendingCount, err = s.svc.UserAgreementCounts(ctx, a.ID); err != nil {
		return nil, err
	}
	if d.CreatedBy, err = s.actor(ctx, a.CreatedBy); err != nil {
		return nil, err
	}
	if a.PublishedBy != nil {
		if d.PublishedBy, err = s.actor(ctx, *a.PublishedBy); err != nil {
			return nil, err
		}
	}
	if a.RemovedBy != nil {
		if d.RemovedBy, err = s.actor(ctx, *a.RemovedBy); err != nil {
			return nil, err
		}
	}
	return d, nil
}

// @Summary	List user agreements with acceptance counts
// @Tags	Agreements
// @Security	BearerAuth
// @Produce	json
// @Success	200	{array}	userAgreementListItem
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks agreements:read"
// @Router	/agreements [get]
func (s *Server) listUserAgreements(w http.ResponseWriter, r *http.Request) {
	rows, err := s.svc.ListUserAgreements(r.Context())
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	items := make([]userAgreementListItem, 0, len(rows))
	for _, a := range rows {
		accepted, pending, err := s.svc.UserAgreementCounts(r.Context(), a.ID)
		if err != nil {
			common.WriteServiceError(w, err)
			return
		}
		items = append(items, userAgreementListItem{UserAgreement: a, AcceptedCount: accepted, PendingCount: pending})
	}
	common.WriteJSON(w, http.StatusOK, items)
}

// @Summary	Get a user agreement with counts and actor names
// @Tags	Agreements
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Agreement ID"
// @Success	200	{object}	userAgreementDetail
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks agreements:read"
// @Failure	404	{object}	object{error=string}
// @Router	/agreements/{id} [get]
func (s *Server) getUserAgreement(w http.ResponseWriter, r *http.Request) {
	a, err := s.svc.GetUserAgreement(r.Context(), r.PathValue("id"))
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.writeAgreementDetail(w, r, http.StatusOK, a)
}

func (s *Server) writeAgreementDetail(w http.ResponseWriter, r *http.Request, status int, a *models.UserAgreement) {
	d, err := s.agreementDetail(r.Context(), a)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, status, d)
}

// @Summary	Create a draft user agreement
// @Tags	Agreements
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	request	body	userAgreementRequest	true	"Agreement draft"
// @Success	201	{object}	userAgreementDetail
// @Failure	400	{object}	object{error=string}
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks agreements:manage"
// @Failure	409	{object}	object{error=string}	"That version already exists"
// @Router	/agreements [post]
func (s *Server) createUserAgreement(w http.ResponseWriter, r *http.Request) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	var req userAgreementRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	a, err := s.svc.CreateUserAgreementDraft(r.Context(), req.input(), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.writeAgreementDetail(w, r, http.StatusCreated, a)
}

// @Summary	Update a draft user agreement
// @Description	Only drafts can be edited.
// @Tags	Agreements
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Param	id	path	string	true	"Agreement ID"
// @Param	request	body	userAgreementRequest	true	"Agreement draft"
// @Success	200	{object}	userAgreementDetail
// @Failure	400	{object}	object{error=string}	"Not a draft or invalid content"
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks agreements:manage"
// @Failure	404	{object}	object{error=string}
// @Failure	409	{object}	object{error=string}	"That version already exists"
// @Router	/agreements/{id} [put]
func (s *Server) updateUserAgreement(w http.ResponseWriter, r *http.Request) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	var req userAgreementRequest
	if err := common.DecodeJSON(r, &req); err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	a, err := s.svc.UpdateUserAgreementDraft(r.Context(), r.PathValue("id"), req.input(), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.writeAgreementDetail(w, r, http.StatusOK, a)
}

// @Summary	Delete a draft user agreement
// @Tags	Agreements
// @Security	BearerAuth
// @Param	id	path	string	true	"Agreement ID"
// @Success	204	"No Content"
// @Failure	400	{object}	object{error=string}	"Not a draft"
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks agreements:manage"
// @Failure	404	{object}	object{error=string}
// @Router	/agreements/{id} [delete]
func (s *Server) deleteUserAgreement(w http.ResponseWriter, r *http.Request) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	if err := s.svc.DeleteUserAgreementDraft(r.Context(), r.PathValue("id"), caller.UserID); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// @Summary	Publish a draft user agreement
// @Description	It becomes the current version. Every user must accept it at their next request.
// @Tags	Agreements
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Agreement ID"
// @Success	200	{object}	userAgreementDetail
// @Failure	400	{object}	object{error=string}	"Not a draft"
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks agreements:manage"
// @Failure	404	{object}	object{error=string}
// @Router	/agreements/{id}/publish [post]
func (s *Server) publishUserAgreement(w http.ResponseWriter, r *http.Request) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	a, err := s.svc.PublishUserAgreement(r.Context(), r.PathValue("id"), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.svc.InvalidateAllIdentities()
	s.writeAgreementDetail(w, r, http.StatusOK, a)
}

// @Summary	Remove a published user agreement
// @Description	If it was the current version, the previous published version becomes current again.
// @Tags	Agreements
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Agreement ID"
// @Success	200	{object}	userAgreementDetail
// @Failure	400	{object}	object{error=string}	"Not published"
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks agreements:manage"
// @Failure	404	{object}	object{error=string}
// @Router	/agreements/{id}/remove [post]
func (s *Server) removeUserAgreement(w http.ResponseWriter, r *http.Request) {
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	a, err := s.svc.RemoveUserAgreement(r.Context(), r.PathValue("id"), caller.UserID)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	s.svc.InvalidateAllIdentities()
	s.writeAgreementDetail(w, r, http.StatusOK, a)
}

// @Summary	Report who accepted or has not yet accepted an agreement
// @Tags	Agreements
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"Agreement ID"
// @Param	status	query	string	false	"accepted (default) or pending"
// @Param	limit	query	int	false	"Page size (default 50, max 200)"
// @Param	offset	query	int	false	"Rows to skip"
// @Success	200	{object}	acceptanceReport
// @Failure	400	{object}	object{error=string}	"Unknown status"
// @Failure	401	{object}	object{error=string}
// @Failure	403	{object}	object{error=string}	"Caller lacks agreements:read"
// @Failure	404	{object}	object{error=string}
// @Router	/agreements/{id}/acceptances [get]
func (s *Server) listUserAgreementAcceptances(w http.ResponseWriter, r *http.Request) {
	ctx, id, q := r.Context(), r.PathValue("id"), r.URL.Query()
	status := q.Get("status")
	if status == "" {
		status = "accepted"
	}
	if status != "accepted" && status != "pending" {
		common.WriteError(w, http.StatusBadRequest, errors.New("status must be accepted or pending"))
		return
	}
	if _, err := s.svc.GetUserAgreement(ctx, id); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	limit, offset := common.Paged(q)
	rep := acceptanceReport{Status: status, Accepted: []acceptanceRow{}, Pending: []pendingRow{}}
	if status == "accepted" {
		rows, err := s.svc.ListUserAgreementAcceptances(ctx, id, limit, offset)
		if err != nil {
			common.WriteServiceError(w, err)
			return
		}
		for _, a := range rows {
			who, err := s.actor(ctx, a.UserID)
			if err != nil {
				common.WriteServiceError(w, err)
				return
			}
			if who == nil {
				who = &actorView{ID: a.UserID}
			}
			rep.Accepted = append(rep.Accepted, acceptanceRow{User: *who, AcceptedAt: a.AcceptedAt})
		}
	} else {
		users, err := s.svc.ListUserAgreementPendingUsers(ctx, id, limit, offset)
		if err != nil {
			common.WriteServiceError(w, err)
			return
		}
		for i := range users {
			ident, err := s.svc.GetOIDCIdentity(ctx, users[i].ID)
			if err != nil {
				common.WriteServiceError(w, err)
				return
			}
			row := pendingRow{User: userView(&users[i])}
			if ident != nil {
				row.FirstSignInAt = &ident.CreatedAt
			}
			rep.Pending = append(rep.Pending, row)
		}
	}
	common.WriteJSON(w, http.StatusOK, rep)
}
