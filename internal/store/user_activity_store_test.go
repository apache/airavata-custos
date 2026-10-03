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
	"reflect"
	"strings"
	"testing"
	"time"
)

// zoneSelector stands in for the database when only the timezone lookup
// matters.
type zoneSelector struct {
	zones []string
	err   error
}

func (z zoneSelector) SelectContext(_ context.Context, dest any, _ string, _ ...any) error {
	if z.err != nil {
		return z.err
	}
	out, ok := dest.(*[]string)
	if !ok {
		return nil
	}
	*out = z.zones
	return nil
}

func TestActivityCalendarUsesEachZonesOwnDate(t *testing.T) {
	// 01:30 UTC is already the next day in Auckland and still the previous
	// day in Los Angeles.
	now := time.Date(2026, 9, 20, 1, 30, 0, 0, time.UTC)
	calendar, err := newActivityCalendar(t.Context(),
		zoneSelector{zones: []string{"Pacific/Auckland", "America/Los_Angeles"}}, now)
	if err != nil {
		t.Fatalf("build calendar: %v", err)
	}
	if calendar.utc != "2026-09-20" {
		t.Errorf("utc date: got %q, want 2026-09-20", calendar.utc)
	}
	if got := calendar.byZone["Pacific/Auckland"]; got != "2026-09-20" {
		t.Errorf("Auckland date: got %q, want 2026-09-20", got)
	}
	if got := calendar.byZone["America/Los_Angeles"]; got != "2026-09-19" {
		t.Errorf("Los Angeles date: got %q, want 2026-09-19", got)
	}
}

func TestActivityCalendarDropsUnusableZones(t *testing.T) {
	now := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	calendar, err := newActivityCalendar(t.Context(),
		zoneSelector{zones: []string{"Mars/Olympus_Mons", "Europe/Berlin"}}, now)
	if err != nil {
		t.Fatalf("build calendar: %v", err)
	}
	if _, ok := calendar.byZone["Mars/Olympus_Mons"]; ok {
		t.Error("an unusable zone must fall through to the UTC branch")
	}
	if _, ok := calendar.byZone["Europe/Berlin"]; !ok {
		t.Error("a usable zone must stay in the calendar")
	}
}

func TestActivityCalendarTodaySQLIsOrderedAndParameterized(t *testing.T) {
	calendar := activityCalendar{
		byZone: map[string]string{
			"Europe/Berlin":    "2026-09-20",
			"America/New_York": "2026-09-19",
		},
		utc: "2026-09-20",
	}
	sql, args := calendar.todaySQL("u.timezone")
	if strings.Count(sql, "WHEN ?") != 2 {
		t.Errorf("one branch per zone expected, got %q", sql)
	}
	if !strings.HasSuffix(sql, "ELSE ?::date END") {
		t.Errorf("unknown and null zones need a UTC fallback branch, got %q", sql)
	}
	// Sorted zones keep the statement text stable so the driver can reuse
	// its prepared-statement cache.
	want := []any{"America/New_York", "2026-09-19", "Europe/Berlin", "2026-09-20", "2026-09-20"}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("args: got %v, want %v", args, want)
	}
}

func TestActivityCalendarTodaySQLWithoutZones(t *testing.T) {
	calendar := activityCalendar{byZone: map[string]string{}, utc: "2026-09-20"}
	sql, args := calendar.todaySQL("u.timezone")
	if sql != "?::date" || !reflect.DeepEqual(args, []any{"2026-09-20"}) {
		t.Errorf("no stored zones should bind the UTC date alone, got %q %v", sql, args)
	}
}

func TestEscapeLikeNeutralizesWildcards(t *testing.T) {
	cases := map[string]string{
		"plain":    "plain",
		"50%":      `50\%`,
		"a_b":      `a\_b`,
		`back\ate`: `back\\ate`,
	}
	for input, want := range cases {
		if got := escapeLike(input); got != want {
			t.Errorf("escapeLike(%q): got %q, want %q", input, got, want)
		}
	}
}

func TestActivityOrderSQLKeepsNeverRowsLast(t *testing.T) {
	for _, direction := range []string{"asc", "desc"} {
		order := activityOrderSQL("last_login", direction)
		if !strings.HasPrefix(order, "u.last_login IS NULL") {
			t.Errorf("%s: never-signed-in rows must sort last, got %q", direction, order)
		}
		if !strings.HasSuffix(order, "u.id ASC") {
			t.Errorf("%s: paging needs a stable tie-break, got %q", direction, order)
		}
	}
	if order := activityOrderSQL("name", "asc"); !strings.Contains(order, "ASC") {
		t.Errorf("name sort should honour the direction, got %q", order)
	}
	if order := activityOrderSQL("login_count", "desc"); !strings.Contains(order, "u.login_count DESC") {
		t.Errorf("login_count sort should order by the count, got %q", order)
	}
	// An unvalidated key must not reach the statement.
	if order := activityOrderSQL("email; DROP TABLE users", "asc"); strings.Contains(order, "DROP") {
		t.Errorf("unknown sort keys must fall back to the default order, got %q", order)
	}
}

func TestStreakFromDescendingDates(t *testing.T) {
	cases := []struct {
		name  string
		dates []string
		want  uint64
	}{
		{"no activity", nil, 0},
		{"single day", []string{"2026-09-20"}, 1},
		{"three consecutive", []string{"2026-09-20", "2026-09-19", "2026-09-18"}, 3},
		{"gap ends the streak", []string{"2026-09-20", "2026-09-18", "2026-09-17"}, 1},
		{"across a month boundary", []string{"2026-10-01", "2026-09-30"}, 2},
		{"across a leap day", []string{"2024-03-01", "2024-02-29", "2024-02-28"}, 3},
	}
	for _, tc := range cases {
		if got := streakFromDescendingDates(tc.dates); got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, got, tc.want)
		}
	}
}

func TestNextStreakOnlyExtendsOnTheFollowingDay(t *testing.T) {
	previous := "2026-09-19"
	summary := loginSummary{lastLocalDate: &previous, loginStreak: 4}
	if got := nextStreak(summary, "2026-09-20"); got != 5 {
		t.Errorf("consecutive day: got %d, want 5", got)
	}
	if got := nextStreak(summary, "2026-09-21"); got != 1 {
		t.Errorf("skipped day should restart: got %d, want 1", got)
	}
	if got := nextStreak(loginSummary{}, "2026-09-20"); got != 1 {
		t.Errorf("first ever day: got %d, want 1", got)
	}
}

func TestOutOfOrderDetectsLateEvents(t *testing.T) {
	lastLogin := time.Date(2026, 9, 20, 12, 0, 0, 0, time.UTC)
	lastDate := "2026-09-20"
	summary := loginSummary{lastLogin: &lastLogin, lastLocalDate: &lastDate}

	if outOfOrder(summary, LoginEvent{OccurredAt: lastLogin.Add(time.Hour), LocalDate: "2026-09-20"}) {
		t.Error("a newer event on the same day is in order")
	}
	if !outOfOrder(summary, LoginEvent{OccurredAt: lastLogin.Add(-time.Hour), LocalDate: "2026-09-20"}) {
		t.Error("an event older than the recorded last login is late")
	}
	if !outOfOrder(summary, LoginEvent{OccurredAt: lastLogin.Add(time.Hour), LocalDate: "2026-09-19"}) {
		t.Error("an earlier calendar day is late even with a newer timestamp")
	}
	if outOfOrder(loginSummary{}, LoginEvent{OccurredAt: lastLogin, LocalDate: "2026-09-20"}) {
		t.Error("the first event for a user is never late")
	}
}
