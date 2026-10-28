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
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// LoginDateLayout is the calendar-day layout shared by login facts. Dates are
// user-local days, never instants.
const LoginDateLayout = "2006-01-02"

type pgLoginActivityStore struct {
	db *sqlx.DB
}

// NewLoginActivityStore returns a PostgreSQL-backed LoginActivityStore.
func NewLoginActivityStore(db *sqlx.DB) LoginActivityStore {
	return &pgLoginActivityStore{db: db}
}

// loginSummary is the denormalized activity state carried on the user row.
type loginSummary struct {
	lastLogin     *time.Time
	lastLocalDate *string
	loginCount    uint64
	loginDayCount uint64
	loginStreak   uint64
}

// Record stores one login event and folds it into the daily facts and the
// user's summary columns. It reports false when event_key was already
// recorded, which makes repeated portal callbacks for the same OIDC token
// idempotent rather than inflating the counts.
func (s *pgLoginActivityStore) Record(ctx context.Context, tx *sql.Tx, event LoginEvent) (bool, error) {
	summary, err := lockLoginSummary(ctx, tx, event.UserID)
	if err != nil {
		return false, err
	}
	inserted, err := insertLoginEvent(ctx, tx, event)
	if err != nil {
		return false, err
	}
	if !inserted {
		return false, nil
	}
	newDay, err := upsertLoginDay(ctx, tx, event)
	if err != nil {
		return false, err
	}
	// An event older than the recorded last login cannot be folded in
	// incrementally without corrupting the streak, so replay the facts.
	if outOfOrder(summary, event) {
		return true, rebuildLoginSummary(ctx, tx, event.UserID)
	}
	return true, applyLoginSummary(ctx, tx, event, summary, newDay)
}

func lockLoginSummary(ctx context.Context, tx *sql.Tx, userID string) (loginSummary, error) {
	var summary loginSummary
	row := tx.QueryRowContext(ctx,
		`SELECT last_login, TO_CHAR(last_login_local_date, 'YYYY-MM-DD'),
		        login_count, login_day_count, login_streak
		 FROM users WHERE id = $1 FOR UPDATE`, userID)
	if err := row.Scan(&summary.lastLogin, &summary.lastLocalDate,
		&summary.loginCount, &summary.loginDayCount, &summary.loginStreak); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return summary, fmt.Errorf("login activity: unknown user %s", userID)
		}
		return summary, err
	}
	return summary, nil
}

// insertLoginEvent returns false when this event_key is already stored for the
// same user. A key already owned by a different user is a collision, not a
// retry, and must not silently pass.
func insertLoginEvent(ctx context.Context, tx *sql.Tx, event LoginEvent) (bool, error) {
	result, err := tx.ExecContext(ctx,
		`INSERT INTO user_login_events
		     (id, event_key, user_id, occurred_at, local_date, timezone, provider, session_id)
		 VALUES ($1, $2, $3, $4, $5::date, $6, NULLIF($7, ''), NULLIF($8, ''))
		 ON CONFLICT (event_key) DO NOTHING`,
		event.ID, event.EventKey, event.UserID, event.OccurredAt,
		event.LocalDate, event.Timezone, event.Provider, event.SessionID)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if affected > 0 {
		return true, nil
	}
	var owner string
	if err := tx.QueryRowContext(ctx,
		`SELECT user_id FROM user_login_events WHERE event_key = $1`, event.EventKey).Scan(&owner); err != nil {
		return false, err
	}
	if owner != event.UserID {
		return false, fmt.Errorf("login activity: event key already owned by user %s", owner)
	}
	return false, nil
}

// upsertLoginDay reports whether this local date is newly active for the user.
func upsertLoginDay(ctx context.Context, tx *sql.Tx, event LoginEvent) (bool, error) {
	var existed bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM user_login_daily WHERE user_id = $1 AND local_date = $2::date)`,
		event.UserID, event.LocalDate).Scan(&existed); err != nil {
		return false, err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO user_login_daily
		     (user_id, local_date, timezone, login_count, first_login_at, last_login_at)
		 VALUES ($1, $2::date, $3, 1, $4, $4)
		 ON CONFLICT (user_id, local_date) DO UPDATE SET
		     login_count = user_login_daily.login_count + 1,
		     timezone = EXCLUDED.timezone,
		     first_login_at = LEAST(user_login_daily.first_login_at, EXCLUDED.first_login_at),
		     last_login_at = GREATEST(user_login_daily.last_login_at, EXCLUDED.last_login_at)`,
		event.UserID, event.LocalDate, event.Timezone, event.OccurredAt); err != nil {
		return false, err
	}
	return !existed, nil
}

func outOfOrder(summary loginSummary, event LoginEvent) bool {
	if summary.lastLogin != nil && event.OccurredAt.Before(*summary.lastLogin) {
		return true
	}
	if summary.lastLocalDate == nil {
		return false
	}
	return event.LocalDate < *summary.lastLocalDate
}

// applyLoginSummary advances the user row for an event that is at least as
// recent as everything already folded in.
func applyLoginSummary(ctx context.Context, tx *sql.Tx, event LoginEvent, summary loginSummary, newDay bool) error {
	streak := summary.loginStreak
	dayCount := summary.loginDayCount
	if newDay {
		dayCount++
		streak = nextStreak(summary, event.LocalDate)
	}
	_, err := tx.ExecContext(ctx,
		`UPDATE users SET last_login = $1, last_login_local_date = $2::date,
		        login_count = $3, login_day_count = $4, login_streak = $5
		 WHERE id = $6`,
		event.OccurredAt, event.LocalDate, summary.loginCount+1, dayCount, streak, event.UserID)
	return err
}

// nextStreak extends the streak only for the calendar day right after the
// last recorded one; any gap restarts it.
func nextStreak(summary loginSummary, localDate string) uint64 {
	if summary.lastLocalDate == nil {
		return 1
	}
	previous, err := time.Parse(LoginDateLayout, *summary.lastLocalDate)
	if err != nil {
		return 1
	}
	current, err := time.Parse(LoginDateLayout, localDate)
	if err != nil {
		return 1
	}
	if previous.AddDate(0, 0, 1).Equal(current) {
		return summary.loginStreak + 1
	}
	return 1
}

// rebuildLoginSummary recomputes the user row from the stored facts. Used for
// late or out-of-order events, where incremental arithmetic cannot recover the
// correct streak.
func rebuildLoginSummary(ctx context.Context, tx *sql.Tx, userID string) error {
	var loginCount uint64
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_login_events WHERE user_id = $1`, userID).Scan(&loginCount); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx,
		`SELECT TO_CHAR(local_date, 'YYYY-MM-DD') FROM user_login_daily
		 WHERE user_id = $1 ORDER BY local_date DESC`, userID)
	if err != nil {
		return err
	}
	var dates []string
	for rows.Next() {
		var date string
		if err := rows.Scan(&date); err != nil {
			rows.Close()
			return err
		}
		dates = append(dates, date)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	var lastLogin *time.Time
	var lastLocalDate *string
	err = tx.QueryRowContext(ctx,
		`SELECT occurred_at, TO_CHAR(local_date, 'YYYY-MM-DD') FROM user_login_events
		 WHERE user_id = $1 ORDER BY occurred_at DESC, id DESC LIMIT 1`, userID).
		Scan(&lastLogin, &lastLocalDate)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	_, err = tx.ExecContext(ctx,
		`UPDATE users SET last_login = $1, last_login_local_date = $2::date,
		        login_count = $3, login_day_count = $4, login_streak = $5
		 WHERE id = $6`,
		lastLogin, lastLocalDate, loginCount, len(dates), streakFromDescendingDates(dates), userID)
	return err
}

// streakFromDescendingDates counts consecutive calendar days ending at the
// most recent active day.
func streakFromDescendingDates(dates []string) uint64 {
	if len(dates) == 0 {
		return 0
	}
	previous, err := time.Parse(LoginDateLayout, dates[0])
	if err != nil {
		return 0
	}
	streak := uint64(1)
	for _, value := range dates[1:] {
		current, err := time.Parse(LoginDateLayout, value)
		if err != nil {
			break
		}
		if !previous.AddDate(0, 0, -1).Equal(current) {
			break
		}
		streak++
		previous = current
	}
	return streak
}
