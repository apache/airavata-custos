// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements. See the NOTICE file distributed with this work
// for additional information regarding copyright ownership. The ASF licenses
// this file to you under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package useractivity

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/common"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

type ActivityRow struct {
	UserID           string     `json:"user_id" db:"user_id"`
	Name             string     `json:"name" db:"name"`
	Email            string     `json:"email" db:"email"`
	RoleNames        []string   `json:"role_names" db:"-"`
	LastLogin        *time.Time `json:"last_login" db:"last_login"`
	InactiveDays     *int       `json:"inactive_days" db:"inactive_days"`
	LoginCount       int64      `json:"login_count" db:"login_count"`
	WindowLoginCount int64      `json:"window_login_count" db:"window_login_count"`
	LoginDayCount    int64      `json:"login_day_count" db:"login_day_count"`
	CurrentStreak    int64      `json:"current_streak" db:"-"`
}

type TrendPoint struct {
	Date        string `json:"date" db:"date"`
	ActiveUsers int64  `json:"active_users" db:"active_users"`
	LoginCount  int64  `json:"login_count" db:"login_count"`
}

type Analytics struct {
	GeneratedAt          time.Time    `json:"generated_at"`
	WindowDays           int          `json:"window_days"`
	TotalUsers           int64        `json:"total_users" db:"total_users"`
	UsersEverLoggedIn    int64        `json:"users_ever_logged_in" db:"users_ever_logged_in"`
	ActiveUsers          int64        `json:"active_users" db:"active_users"`
	PriorActiveUsers     int64        `json:"prior_active_users" db:"prior_active_users"`
	DormantOver90Days    int64        `json:"dormant_over_90_days" db:"dormant_over_90_days"`
	OldestNeverCreatedAt *time.Time   `json:"oldest_never_created_at" db:"oldest_never_created_at"`
	LifetimeLoginCount   int64        `json:"lifetime_login_count" db:"lifetime_login_count"`
	LifetimeActiveDays   int64        `json:"lifetime_active_days" db:"lifetime_active_days"`
	WindowLoginCount     int64        `json:"window_login_count" db:"window_login_count"`
	WindowActiveDays     int64        `json:"window_active_days" db:"window_active_days"`
	Trend                []TrendPoint `json:"trend" db:"-"`
}

const population = `SELECT u.id, u.created_at FROM users u WHERE EXISTS (
	SELECT 1 FROM user_identities i WHERE i.user_id = u.id AND i.oidc_sub IS NOT NULL)`

func (c *Connector) registerRoutes(router *identity.Router) {
	router.RequirePrivilege("GET /users/activity", models.UsersActivityRead, c.list)
	router.RequirePrivilege("GET /users/activity/analytics", models.UsersActivityRead, c.analytics)
	router.RequirePrivilege("GET /users/{id}/activity/analytics", models.UsersActivityRead, c.userAnalytics)
}

func parseWindow(raw string, fallback int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > 365 {
		return 0, errors.New("window must be between 1 and 365 days")
	}
	return n, nil
}

func (c *Connector) analytics(w http.ResponseWriter, r *http.Request) {
	window, err := parseWindow(r.URL.Query().Get("window"), 30)
	if err != nil {
		common.WriteError(w, 400, err)
		return
	}
	result, err := c.readAnalytics(r.Context(), "", window)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, 200, result)
}

func (c *Connector) userAnalytics(w http.ResponseWriter, r *http.Request) {
	window, err := parseWindow(r.URL.Query().Get("window"), 30)
	if err != nil {
		common.WriteError(w, 400, err)
		return
	}
	result, err := c.readAnalytics(r.Context(), r.PathValue("id"), window)
	if err != nil {
		common.WriteServiceError(w, err)
		return
	}
	if result.TotalUsers == 0 {
		common.WriteError(w, 404, errors.New("user not found"))
		return
	}
	common.WriteJSON(w, 200, result)
}

func (c *Connector) readAnalytics(ctx context.Context, userID string, window int) (*Analytics, error) {
	now := time.Now().UTC()
	today := now.Format("2006-01-02")
	pop := population
	args := []any{}
	if userID != "" {
		pop += " AND u.id = ?"
		args = append(args, userID)
	}
	query := `WITH pop AS (` + pop + `), facts AS (
		SELECT a.entity_id AS user_id, (a.event_time AT TIME ZONE 'UTC')::date AS day
		FROM audit_events a JOIN pop p ON p.id = a.entity_id
		WHERE a.event_type = 'PORTAL_LOGIN' AND a.entity_type = 'user'
	), last AS (SELECT user_id, MAX(day) AS day FROM facts GROUP BY user_id)
	SELECT (SELECT COUNT(*) FROM pop) AS total_users,
		(SELECT COUNT(*) FROM last) AS users_ever_logged_in,
		(SELECT COUNT(*) FROM last WHERE day BETWEEN ?::date - (?::int - 1) AND ?::date) AS active_users,
		(SELECT COUNT(DISTINCT user_id) FROM facts WHERE day BETWEEN ?::date - (2 * ?::int - 1) AND ?::date - ?::int) AS prior_active_users,
		(SELECT COUNT(*) FROM last WHERE day < ?::date - GREATEST(?::int, 91) + 1) AS dormant_over_90_days,
		(SELECT MIN(p.created_at) FROM pop p LEFT JOIN last l ON l.user_id = p.id WHERE l.user_id IS NULL) AS oldest_never_created_at,
		(SELECT COUNT(*) FROM facts) AS lifetime_login_count,
		(SELECT COUNT(*) FROM (SELECT DISTINCT user_id, day FROM facts) d) AS lifetime_active_days,
		(SELECT COUNT(*) FROM facts WHERE day BETWEEN ?::date - (?::int - 1) AND ?::date) AS window_login_count,
		(SELECT COUNT(*) FROM (SELECT DISTINCT user_id, day FROM facts WHERE day BETWEEN ?::date - (?::int - 1) AND ?::date) d) AS window_active_days`
	qArgs := append(append([]any{}, args...), today, window, today, today, window, today, window, today, window, today, window, today, today, window, today)
	var result Analytics
	if err := c.db.GetContext(ctx, &result, c.db.Rebind(query), qArgs...); err != nil {
		return nil, err
	}
	result.GeneratedAt, result.WindowDays = now, window
	result.Trend = make([]TrendPoint, 0, window)
	trendQuery := `WITH pop AS (` + pop + `)
	SELECT TO_CHAR(a.event_time AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS date,
		COUNT(DISTINCT a.entity_id) AS active_users, COUNT(*) AS login_count
	FROM audit_events a JOIN pop p ON p.id = a.entity_id
	WHERE a.event_type = 'PORTAL_LOGIN' AND a.entity_type = 'user'
		AND a.event_time >= ((?::date - (?::int - 1))::timestamp AT TIME ZONE 'UTC')
		AND a.event_time < ((?::date + 1)::timestamp AT TIME ZONE 'UTC')
	GROUP BY date ORDER BY date`
	var found []TrendPoint
	tArgs := append(append([]any{}, args...), today, window, today)
	if err := c.db.SelectContext(ctx, &found, c.db.Rebind(trendQuery), tArgs...); err != nil {
		return nil, err
	}
	byDay := make(map[string]TrendPoint, len(found))
	for _, point := range found {
		byDay[point.Date] = point
	}
	start := now.AddDate(0, 0, 1-window)
	for i := 0; i < window; i++ {
		date := start.AddDate(0, 0, i).Format("2006-01-02")
		point, ok := byDay[date]
		if !ok {
			point.Date = date
		}
		result.Trend = append(result.Trend, point)
	}
	return &result, nil
}

func parsePageInt(raw string, fallback, max int) (int, error) {
	if raw == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 || n > max {
		return 0, fmt.Errorf("invalid pagination value")
	}
	return n, nil
}

func (c *Connector) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	window, err := parseWindow(q.Get("window"), 30)
	if err != nil {
		common.WriteError(w, 400, err)
		return
	}
	limit, err := parsePageInt(q.Get("limit"), 10, 200)
	if err != nil || limit == 0 {
		common.WriteError(w, 400, errors.New("limit must be 1..200"))
		return
	}
	offset, err := parsePageInt(q.Get("offset"), 0, 1000000)
	if err != nil {
		common.WriteError(w, 400, err)
		return
	}
	status := q.Get("status")
	if status == "" {
		status = "all"
	}
	if status != "all" && status != "active" && status != "dormant" && status != "never" {
		common.WriteError(w, 400, errors.New("invalid status"))
		return
	}
	sortKey := q.Get("sort")
	if sortKey == "" {
		sortKey = "last_login"
	}
	sortColumns := map[string]string{"name": "name", "last_login": "last_login", "login_count": "login_count"}
	order, ok := sortColumns[sortKey]
	if !ok {
		common.WriteError(w, 400, errors.New("invalid sort"))
		return
	}
	direction := strings.ToLower(q.Get("direction"))
	if direction == "" {
		direction = "desc"
	}
	if direction != "asc" && direction != "desc" {
		common.WriteError(w, 400, errors.New("invalid direction"))
		return
	}
	today := time.Now().UTC().Format("2006-01-02")
	base := `WITH pop AS (` + population + `), counts AS (
		SELECT a.entity_id AS user_id, MAX(a.event_time) AS last_login,
			COUNT(*) AS login_count, COUNT(DISTINCT (a.event_time AT TIME ZONE 'UTC')::date) AS login_day_count,
			COUNT(*) FILTER (WHERE (a.event_time AT TIME ZONE 'UTC')::date BETWEEN ?::date - (?::int - 1) AND ?::date) AS window_login_count
	FROM audit_events a JOIN pop p ON p.id = a.entity_id
	WHERE a.event_type = 'PORTAL_LOGIN' AND a.entity_type = 'user'
	GROUP BY a.entity_id), rows AS (
		SELECT u.id AS user_id, COALESCE(NULLIF(TRIM(CONCAT_WS(' ', u.first_name, u.last_name)), ''), u.email::text) AS name, u.email,
			c.last_login, CASE WHEN c.last_login IS NULL THEN NULL ELSE (?::date - (c.last_login AT TIME ZONE 'UTC')::date) END AS inactive_days,
		COALESCE(c.login_count, 0) AS login_count, COALESCE(c.window_login_count, 0) AS window_login_count,
		COALESCE(c.login_day_count, 0) AS login_day_count
	FROM pop p JOIN users u ON u.id = p.id LEFT JOIN counts c ON c.user_id = p.id)
	SELECT * FROM rows WHERE (? = 'all' OR (? = 'active' AND inactive_days < ?)
		OR (? = 'dormant' AND inactive_days >= ?) OR (? = 'never' AND last_login IS NULL))
	AND (LOWER(name) LIKE ? OR LOWER(email) LIKE ?)`
	pattern := "%" + strings.ToLower(strings.TrimSpace(q.Get("query"))) + "%"
	args := []any{today, window, today, today, status, status, window, status, window, status, pattern, pattern}
	var total int
	countSQL := `SELECT COUNT(*) FROM (` + base + `) t`
	if err := c.db.GetContext(r.Context(), &total, c.db.Rebind(countSQL), args...); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	rows := []ActivityRow{}
	listSQL := base + ` ORDER BY ` + order + ` ` + direction + ` NULLS LAST, user_id ASC LIMIT ? OFFSET ?`
	if err := c.db.SelectContext(r.Context(), &rows, c.db.Rebind(listSQL), append(args, limit, offset)...); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	if err := c.decorateRows(r.Context(), rows, today); err != nil {
		common.WriteServiceError(w, err)
		return
	}
	common.WriteJSON(w, 200, map[string]any{"items": rows, "total": total, "limit": limit, "offset": offset, "window_days": window, "status": status})
}

func (c *Connector) decorateRows(ctx context.Context, rows []ActivityRow, today string) error {
	if len(rows) == 0 {
		return nil
	}
	ids := make([]string, len(rows))
	index := make(map[string]int, len(rows))
	for i := range rows {
		ids[i] = rows[i].UserID
		index[ids[i]] = i
		rows[i].RoleNames = []string{}
	}
	type roleRow struct {
		UserID string `db:"user_id"`
		Name   string `db:"name"`
	}
	roleSQL, roleArgs, err := sqlx.In(`SELECT ur.user_id, r.name FROM user_roles ur JOIN roles r ON r.id = ur.role_id WHERE ur.user_id IN (?) ORDER BY r.name`, ids)
	if err != nil {
		return err
	}
	var roles []roleRow
	if err := c.db.SelectContext(ctx, &roles, c.db.Rebind(roleSQL), roleArgs...); err != nil {
		return err
	}
	for _, role := range roles {
		rows[index[role.UserID]].RoleNames = append(rows[index[role.UserID]].RoleNames, role.Name)
	}
	type dayRow struct {
		UserID string    `db:"user_id"`
		Day    time.Time `db:"day"`
	}
	daySQL, dayArgs, err := sqlx.In(`SELECT DISTINCT entity_id AS user_id, (event_time AT TIME ZONE 'UTC')::date AS day
		FROM audit_events WHERE entity_id IN (?) AND event_type = 'PORTAL_LOGIN'
		ORDER BY user_id, day DESC`, ids)
	if err != nil {
		return err
	}
	var days []dayRow
	if err := c.db.SelectContext(ctx, &days, c.db.Rebind(daySQL), dayArgs...); err != nil {
		return err
	}
	anchors := make(map[string]time.Time, len(rows))
	base, _ := time.Parse("2006-01-02", today)
	for _, row := range rows {
		anchors[row.UserID] = base
	}
	stopped := make(map[string]bool, len(rows))
	for _, entry := range days {
		if stopped[entry.UserID] {
			continue
		}
		i := index[entry.UserID]
		anchor := anchors[entry.UserID]
		diff := int(anchor.Sub(entry.Day).Hours() / 24)
		if rows[i].CurrentStreak == 0 && diff == 1 {
			anchor = anchor.AddDate(0, 0, -1)
			diff = 0
		}
		if diff != int(rows[i].CurrentStreak) {
			stopped[entry.UserID] = true
			continue
		}
		rows[i].CurrentStreak++
		anchors[entry.UserID] = anchor
	}
	return nil
}
