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

package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Activity reads describe users with an OIDC-linked identity only; provisioned users who
// never linked an identity provider cannot have sign-in history.
const oidcLinkedUser = `EXISTS (
	SELECT 1 FROM user_identities ui
	WHERE ui.user_id = u.id AND ui.oidc_sub IS NOT NULL AND ui.oidc_sub <> ''
)`

const activityDisplayName = `COALESCE(NULLIF(TRIM(CONCAT_WS(' ', u.first_name, u.last_name)), ''), u.email)`

// activityRoleNames aggregates role names as JSON so the driver never has to
// decode a Postgres array. Absent roles come back as an empty list.
const activityRoleNames = `COALESCE((
	SELECT JSON_AGG(r.name ORDER BY r.name)
	FROM user_roles ur JOIN roles r ON r.id = ur.role_id
	WHERE ur.user_id = u.id
), '[]')::text`

// activityCalendar resolves "today" for every timezone in use so a single
// request compares all users against their own calendar day.
type activityCalendar struct {
	byZone map[string]string
	utc    string
}

func newActivityCalendar(ctx context.Context, q interface {
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}, now time.Time) (activityCalendar, error) {
	calendar := activityCalendar{
		byZone: map[string]string{},
		utc:    now.UTC().Format(LoginDateLayout),
	}
	var zones []string
	if err := q.SelectContext(ctx, &zones,
		`SELECT DISTINCT timezone FROM users WHERE timezone IS NOT NULL AND timezone <> ''`); err != nil {
		return calendar, err
	}
	for _, zone := range zones {
		location, err := time.LoadLocation(zone)
		if err != nil {
			// An unusable zone falls back to UTC rather than dropping the user.
			continue
		}
		calendar.byZone[zone] = now.In(location).Format(LoginDateLayout)
	}
	return calendar, nil
}

// todaySQL renders the local current date per user as a CASE over the stored
// timezone. NULL and unknown zones land on the UTC branch.
func (c activityCalendar) todaySQL(column string) (string, []any) {
	if len(c.byZone) == 0 {
		return `?::date`, []any{c.utc}
	}
	zones := make([]string, 0, len(c.byZone))
	for zone := range c.byZone {
		zones = append(zones, zone)
	}
	sort.Strings(zones)
	var sb strings.Builder
	args := make([]any, 0, len(zones)*2+1)
	sb.WriteString(`CASE ` + column)
	for _, zone := range zones {
		sb.WriteString(` WHEN ? THEN ?::date`)
		args = append(args, zone, c.byZone[zone])
	}
	sb.WriteString(` ELSE ?::date END`)
	args = append(args, c.utc)
	return sb.String(), args
}

type activityRowScan struct {
	UserID           string     `db:"user_id"`
	Name             string     `db:"name"`
	Email            string     `db:"email"`
	RoleNames        string     `db:"role_names"`
	LastLogin        *time.Time `db:"last_login"`
	InactiveDays     *int       `db:"inactive_days"`
	LoginCount       uint64     `db:"login_count"`
	WindowLoginCount uint64     `db:"window_login_count"`
	LoginDayCount    uint64     `db:"login_day_count"`
	CurrentStreak    uint64     `db:"current_streak"`
}

// ListActivity returns one page of the filtered activity population plus the
// filtered total. Status and window are applied in SQL before pagination so
// the total always describes the same rows the page came from.
func (s *pgUserStore) ListActivity(ctx context.Context, f UserActivityFilter) ([]UserActivityRow, int, error) {
	calendar, err := newActivityCalendar(ctx, s.db, f.Now)
	if err != nil {
		return nil, 0, err
	}
	today, todayArgs := calendar.todaySQL("u.timezone")
	inactiveDays := `(` + today + ` - u.last_login_local_date)`

	where := []string{oidcLinkedUser}
	var args []any
	if f.Query != "" {
		pattern := "%" + strings.ToLower(escapeLike(f.Query)) + "%"
		where = append(where, `(LOWER(u.email) LIKE ? ESCAPE '\'
			OR LOWER(TRIM(CONCAT_WS(' ', u.first_name, u.middle_name, u.last_name))) LIKE ? ESCAPE '\')`)
		args = append(args, pattern, pattern)
	}
	switch f.Status {
	case UserActivityStatusActive:
		where = append(where, `u.last_login_local_date IS NOT NULL AND `+inactiveDays+` < ?`)
		args = append(args, todayArgs...)
		args = append(args, f.WindowDays)
	case UserActivityStatusDormant:
		where = append(where, `u.last_login_local_date IS NOT NULL AND `+inactiveDays+` >= ?`)
		args = append(args, todayArgs...)
		args = append(args, f.WindowDays)
	case UserActivityStatusNever:
		where = append(where, `u.last_login_local_date IS NULL`)
	}
	clause := ` WHERE ` + strings.Join(where, " AND ")

	var total int
	if err := s.db.GetContext(ctx, &total,
		s.db.Rebind(`SELECT COUNT(*) FROM users u`+clause), args...); err != nil {
		return nil, 0, err
	}

	// Select args are laid out in statement order: projection, lateral window
	// bound, filters, then pagination.
	selectArgs := make([]any, 0, len(args)+len(todayArgs)*4+3)
	selectArgs = append(selectArgs, todayArgs...) // inactive_days
	selectArgs = append(selectArgs, todayArgs...) // current_streak
	selectArgs = append(selectArgs, todayArgs...) // window upper bound
	selectArgs = append(selectArgs, todayArgs...) // window lower bound
	selectArgs = append(selectArgs, f.WindowDays-1)
	selectArgs = append(selectArgs, args...)
	selectArgs = append(selectArgs, f.Limit, f.Offset)

	query := `SELECT u.id AS user_id,
		       ` + activityDisplayName + ` AS name,
		       u.email,
		       ` + activityRoleNames + ` AS role_names,
		       u.last_login,
		       CASE WHEN u.last_login_local_date IS NULL THEN NULL
		            ELSE (` + today + ` - u.last_login_local_date) END AS inactive_days,
		       u.login_count,
		       w.window_login_count,
		       u.login_day_count,
		       CASE WHEN u.last_login_local_date IS NOT NULL
		                 AND (` + today + ` - u.last_login_local_date) BETWEEN 0 AND 1
		            THEN u.login_streak ELSE 0 END AS current_streak
		FROM users u
		LEFT JOIN LATERAL (
		    SELECT COALESCE(SUM(d.login_count), 0) AS window_login_count
		    FROM user_login_daily d
		    WHERE d.user_id = u.id
		      AND d.local_date <= ` + today + `
		      AND d.local_date >= ` + today + ` - ?::int
		) w ON TRUE` + clause + `
		ORDER BY ` + activityOrderSQL(f.Sort, f.Direction) + `
		LIMIT ? OFFSET ?`

	var scanned []activityRowScan
	if err := s.db.SelectContext(ctx, &scanned, s.db.Rebind(query), selectArgs...); err != nil {
		return nil, 0, err
	}
	rows := make([]UserActivityRow, 0, len(scanned))
	for _, row := range scanned {
		var roles []string
		if err := json.Unmarshal([]byte(row.RoleNames), &roles); err != nil {
			return nil, 0, fmt.Errorf("decode role names for %s: %w", row.UserID, err)
		}
		rows = append(rows, UserActivityRow{
			UserID:           row.UserID,
			Name:             row.Name,
			Email:            row.Email,
			RoleNames:        roles,
			LastLogin:        row.LastLogin,
			InactiveDays:     row.InactiveDays,
			LoginCount:       row.LoginCount,
			WindowLoginCount: row.WindowLoginCount,
			LoginDayCount:    row.LoginDayCount,
			CurrentStreak:    row.CurrentStreak,
		})
	}
	return rows, total, nil
}

// activityOrderSQL keeps never-signed-in rows last in both directions and
// breaks ties on the user ID so paging is stable.
func activityOrderSQL(sortKey, direction string) string {
	dir := "DESC"
	if direction == "asc" {
		dir = "ASC"
	}
	switch sortKey {
	case "name":
		return activityDisplayName + ` ` + dir + `, u.id ASC`
	case "login_count":
		return `u.login_count ` + dir + `, ` + activityDisplayName + ` ASC, u.id ASC`
	default:
		return `u.last_login IS NULL, u.last_login ` + dir + ` NULLS LAST, u.id ASC`
	}
}

func escapeLike(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "%", `\%`)
	return strings.ReplaceAll(value, "_", `\_`)
}

type analyticsScan struct {
	TotalUsers           uint64     `db:"total_users"`
	UsersEverLoggedIn    uint64     `db:"users_ever_logged_in"`
	ActiveUsers          uint64     `db:"active_users"`
	PriorActiveUsers     *uint64    `db:"prior_active_users"`
	DormantOver90Days    *uint64    `db:"dormant_over_90_days"`
	OldestNeverCreatedAt *time.Time `db:"oldest_never_created_at"`
	LifetimeLoginCount   uint64     `db:"lifetime_login_count"`
	LifetimeActiveDays   uint64     `db:"lifetime_active_days"`
	WindowLoginCount     uint64     `db:"window_login_count"`
	WindowActiveDays     uint64     `db:"window_active_days"`
}

// GetActivityAnalytics aggregates the whole OIDC-linked population.
func (s *pgUserStore) GetActivityAnalytics(ctx context.Context, now time.Time, windowDays int) (*UserActivityAnalytics, error) {
	return s.activityAnalytics(ctx, "", now, windowDays)
}

// GetUserActivityAnalytics aggregates one OIDC-linked user. A user outside
// that population yields nil so the caller can answer 404.
func (s *pgUserStore) GetUserActivityAnalytics(ctx context.Context, userID string, now time.Time, windowDays int) (*UserActivityAnalytics, error) {
	if userID == "" {
		return nil, nil
	}
	return s.activityAnalytics(ctx, userID, now, windowDays)
}

func (s *pgUserStore) activityAnalytics(ctx context.Context, userID string, now time.Time, windowDays int) (*UserActivityAnalytics, error) {
	calendar, err := newActivityCalendar(ctx, s.db, now)
	if err != nil {
		return nil, err
	}
	today, todayArgs := calendar.todaySQL("u.timezone")

	population := `SELECT u.id, u.created_at, u.last_login_local_date, u.login_count,
		       u.login_day_count, ` + today + ` AS local_today
		FROM users u
		WHERE ` + oidcLinkedUser
	popArgs := append([]any{}, todayArgs...)
	if userID != "" {
		population += ` AND u.id = ?`
		popArgs = append(popArgs, userID)
	}

	// Window bounds are per user: the selected window is that user's local
	// today plus the previous window-1 local days.
	query := `WITH pop AS (` + population + `),
		facts AS (
		    SELECT p.id AS user_id, p.local_today, d.local_date, d.login_count
		    FROM pop p JOIN user_login_daily d ON d.user_id = p.id
		)
		SELECT
		    (SELECT COUNT(*) FROM pop) AS total_users,
		    (SELECT COUNT(*) FROM pop WHERE last_login_local_date IS NOT NULL) AS users_ever_logged_in,
		    (SELECT COUNT(*) FROM pop
		      WHERE last_login_local_date IS NOT NULL
		        AND (local_today - last_login_local_date) < ?::int) AS active_users,
		    (SELECT COUNT(DISTINCT user_id) FROM facts
		      WHERE local_date <= local_today - ?::int
		        AND local_date >= local_today - (2 * ?::int - 1)) AS prior_active_users,
		    (SELECT COUNT(*) FROM pop
		      WHERE last_login_local_date IS NOT NULL
		        AND (local_today - last_login_local_date) >= ?::int
		        AND (local_today - last_login_local_date) > 90) AS dormant_over_90_days,
		    (SELECT MIN(created_at) FROM pop WHERE last_login_local_date IS NULL) AS oldest_never_created_at,
		    (SELECT COALESCE(SUM(login_count), 0) FROM pop) AS lifetime_login_count,
		    (SELECT COALESCE(SUM(login_day_count), 0) FROM pop) AS lifetime_active_days,
		    (SELECT COALESCE(SUM(login_count), 0) FROM facts
		      WHERE local_date <= local_today
		        AND local_date >= local_today - (?::int - 1)) AS window_login_count,
		    (SELECT COUNT(*) FROM facts
		      WHERE local_date <= local_today
		        AND local_date >= local_today - (?::int - 1)) AS window_active_days`

	args := append([]any{}, popArgs...)
	args = append(args, windowDays, windowDays, windowDays, windowDays, windowDays, windowDays)

	var scanned analyticsScan
	if err := s.db.GetContext(ctx, &scanned, s.db.Rebind(query), args...); err != nil {
		return nil, err
	}
	if userID != "" && scanned.TotalUsers == 0 {
		return nil, nil
	}

	trendQuery := `WITH pop AS (` + population + `)
		SELECT TO_CHAR(d.local_date, 'YYYY-MM-DD') AS date,
		       COUNT(DISTINCT d.user_id) AS active_users,
		       COALESCE(SUM(d.login_count), 0) AS login_count
		FROM pop p JOIN user_login_daily d ON d.user_id = p.id
		WHERE d.local_date <= p.local_today
		  AND d.local_date >= p.local_today - (?::int - 1)
		GROUP BY d.local_date
		ORDER BY d.local_date`
	trendArgs := append([]any{}, popArgs...)
	trendArgs = append(trendArgs, windowDays)

	trend := []UserActivityTrendPoint{}
	if err := s.db.SelectContext(ctx, &trend, s.db.Rebind(trendQuery), trendArgs...); err != nil {
		return nil, err
	}

	return &UserActivityAnalytics{
		GeneratedAt:          now.UTC(),
		WindowDays:           windowDays,
		TotalUsers:           scanned.TotalUsers,
		UsersEverLoggedIn:    scanned.UsersEverLoggedIn,
		ActiveUsers:          scanned.ActiveUsers,
		PriorActiveUsers:     scanned.PriorActiveUsers,
		DormantOver90Days:    scanned.DormantOver90Days,
		OldestNeverCreatedAt: scanned.OldestNeverCreatedAt,
		LifetimeLoginCount:   scanned.LifetimeLoginCount,
		LifetimeActiveDays:   scanned.LifetimeActiveDays,
		WindowLoginCount:     scanned.WindowLoginCount,
		WindowActiveDays:     scanned.WindowActiveDays,
		Trend:                trend,
	}, nil
}
