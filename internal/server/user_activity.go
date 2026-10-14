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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
)

// UserActivityListResponse is the activity audit envelope. It echoes the
// normalized window and status so the portal can reject a response from a
// backend that ignored those filters.
type UserActivityListResponse struct {
	Items      []store.UserActivityRow `json:"items"`
	Total      int                     `json:"total"`
	Limit      int                     `json:"limit"`
	Offset     int                     `json:"offset"`
	WindowDays int                     `json:"window_days"`
	Status     string                  `json:"status"`
}

// @Summary	List sign-in activity for users with an OIDC-linked identity
// @Description	One row per user with an OIDC-linked identity. `window` and `status` are applied before pagination, and the response echoes both so a caller can detect an unfiltered answer. Sign-in statuses are disjoint: active has signed in within the window, dormant has signed in earlier, never has no recorded sign-in.
// @Tags	Users
// @Security	BearerAuth
// @Produce	json
// @Param	window	query	integer	true	"Activity window in user-local calendar days (1-365)"
// @Param	status	query	string	true	"all | active | dormant | never"
// @Param	query	query	string	false	"Substring match on name or email"
// @Param	limit	query	integer	false	"Page size (default 10, max 200)"
// @Param	offset	query	integer	false	"Page offset (default 0)"
// @Param	sort	query	string	false	"name | last_login | login_count (default last_login)"
// @Param	direction	query	string	false	"asc | desc (default desc)"
// @Success	200	{object}	UserActivityListResponse
// @Failure	400	{object}	object{error=string}	"Invalid window, status, sort or direction"
// @Failure	403	{object}	object{code=string,message=string}	"Caller lacks core:users:activity:read"
// @Router	/users/activity [get]
func (s *Server) listUserActivity(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	window, err := activityWindowParam(q.Get("window"), 0)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	page, err := s.svc.ListUserActivity(r.Context(), store.UserActivityFilter{
		WindowDays: window,
		Status:     q.Get("status"),
		Query:      q.Get("query"),
		Limit:      atoiOr(q.Get("limit"), 0),
		Offset:     atoiOr(q.Get("offset"), 0),
		Sort:       q.Get("sort"),
		Direction:  q.Get("direction"),
	})
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, UserActivityListResponse{
		Items:      page.Items,
		Total:      page.Total,
		Limit:      page.Limit,
		Offset:     page.Offset,
		WindowDays: page.WindowDays,
		Status:     page.Status,
	})
}

// @Summary	Aggregate sign-in analytics for all users with an OIDC-linked identity
// @Description	Population-wide sign-in activity over the requested window. Supplemental fields (`prior_active_users`, `dormant_over_90_days`, `oldest_never_created_at`) may be null when the population cannot produce them.
// @Tags	Users
// @Security	BearerAuth
// @Produce	json
// @Param	window	query	integer	false	"Activity window in user-local calendar days (1-365, default 30)"
// @Success	200	{object}	store.UserActivityAnalytics
// @Failure	400	{object}	object{error=string}	"Invalid window"
// @Failure	403	{object}	object{code=string,message=string}	"Caller lacks core:users:activity:read"
// @Router	/users/activity/analytics [get]
func (s *Server) getActivityAnalytics(w http.ResponseWriter, r *http.Request) {
	window, err := activityWindowParam(r.URL.Query().Get("window"), defaultActivityWindow)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	analytics, err := s.svc.GetActivityAnalytics(r.Context(), window)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, analytics)
}

// @Summary	Aggregate sign-in analytics for one user with an OIDC-linked identity
// @Tags	Users
// @Security	BearerAuth
// @Produce	json
// @Param	id	path	string	true	"User ID"
// @Param	window	query	integer	false	"Activity window in user-local calendar days (1-365, default 30)"
// @Success	200	{object}	store.UserActivityAnalytics
// @Failure	400	{object}	object{error=string}	"Invalid window"
// @Failure	403	{object}	object{code=string,message=string}	"Caller lacks core:users:activity:read"
// @Failure	404	{object}	object{error=string}	"No such OIDC-linked user"
// @Router	/users/{id}/activity/analytics [get]
func (s *Server) getSelectedUserActivityAnalytics(w http.ResponseWriter, r *http.Request) {
	window, err := activityWindowParam(r.URL.Query().Get("window"), defaultActivityWindow)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	analytics, err := s.svc.GetUserActivityAnalytics(r.Context(), r.PathValue("id"), window)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, http.StatusOK, analytics)
}

const defaultActivityWindow = 30

// activityWindowParam rejects a malformed window instead of falling back to a
// default, so the portal never renders one window's data under another's label.
func activityWindowParam(raw string, fallback int) (int, error) {
	if raw == "" {
		if fallback == 0 {
			return 0, errors.New("window is required")
		}
		return fallback, nil
	}
	window, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("window must be a whole number of days: %w", err)
	}
	return window, nil
}

// recordLoginEventRequest is intentionally empty: the portal posts `{}` and
// the server derives every recorded fact from the verified bearer token.
type recordLoginEventRequest struct{}

// @Summary	Record the caller's sign-in
// @Description	Records one sign-in for the authenticated caller. Identity and timing come from the verified OIDC token, never the request body, and the body must be empty or `{}`. Repeating the call for the same token is idempotent and answers 200 without changing any counts.
// @Tags	Users
// @Security	BearerAuth
// @Accept	json
// @Produce	json
// @Success	201	{object}	service.RecordLoginEventResult	"Sign-in recorded"
// @Success	200	{object}	service.RecordLoginEventResult	"Already recorded for this token"
// @Failure	400	{object}	object{error=string}	"Body is not empty, or the token carries no usable session evidence"
// @Failure	401	{object}	object{error=string}	"No authenticated caller"
// @Router	/me/login-events [post]
func (s *Server) recordLoginEvent(w http.ResponseWriter, r *http.Request) {
	var req recordLoginEventRequest
	if err := common.DecodeJSON(r, &req); err != nil && !errors.Is(err, io.EOF) {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	caller := requireCaller(w, r)
	if caller == nil {
		return
	}
	input, err := trustedLoginEvent(caller)
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, err)
		return
	}
	result, err := s.svc.RecordLoginEvent(r.Context(), caller.UserID, input)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	status := http.StatusOK
	if result.Recorded {
		status = http.StatusCreated
	} else {
		slog.DebugContext(r.Context(), "login event already recorded", "user_id", caller.UserID)
	}
	common.WriteJSON(w, status, result)
}

// trustedLoginEvent derives the dedup key and timestamp from verified claims.
// A token identifier plus its issue time is preferred: two portal sign-ins
// that share one identity-provider session still produce distinct tokens,
// while the session claims alone would collapse them into one event.
func trustedLoginEvent(caller *identity.Caller) (service.RecordLoginEventInput, error) {
	if caller.TokenID != "" && caller.IssuedAt > 0 {
		return service.RecordLoginEventInput{
			EventKey:   hashedEventKey("token", caller.Issuer, caller.Subject, caller.TokenID, caller.IssuedAt),
			OccurredAt: time.Unix(caller.IssuedAt, 0).UTC(),
			Provider:   "oidc",
			SessionID:  caller.SessionID,
		}, nil
	}
	if caller.SessionID != "" && caller.AuthTime > 0 {
		return service.RecordLoginEventInput{
			EventKey:   hashedEventKey("session", caller.Issuer, caller.Subject, caller.SessionID, caller.AuthTime),
			OccurredAt: time.Unix(caller.AuthTime, 0).UTC(),
			Provider:   "oidc",
			SessionID:  caller.SessionID,
		}, nil
	}
	return service.RecordLoginEventInput{}, errors.New(
		"token carries no session evidence: jti with iat, or sid with auth_time, is required")
}

// hashedEventKey builds a collision-resistant key over NUL-separated claim
// values so a value containing the separator cannot forge another key.
func hashedEventKey(kind, issuer, subject, sessionID string, timestamp int64) string {
	canonical := fmt.Sprintf("v1\x00%s\x00%s\x00%s\x00%s\x00%d", kind, issuer, subject, sessionID, timestamp)
	sum := sha256.Sum256([]byte(canonical))
	return "v1:" + hex.EncodeToString(sum[:])
}
