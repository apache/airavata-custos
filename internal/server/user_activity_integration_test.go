//go:build integration

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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/store"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// activityPopulation is the seeded fixture the activity assertions read.
type activityPopulation struct {
	alice string // signed in today and yesterday
	bob   string // dormant at 40 days
	carol string // dormant beyond 90 days
	dave  string // never signed in
	erin  string // never linked an identity provider
}

func seedActivityPopulation(t *testing.T, database *sqlx.DB, svc *service.Service) activityPopulation {
	t.Helper()
	pop := activityPopulation{
		alice: seedOIDCUser(t, database, "alice@example.org", "America/New_York"),
		bob:   seedOIDCUser(t, database, "bob@example.org", ""),
		carol: seedOIDCUser(t, database, "carol@example.org", "Asia/Shanghai"),
		dave:  seedOIDCUser(t, database, "dave@example.org", ""),
		erin:  seedUser(t, database, "erin@example.org"),
	}
	seedRole(t, database, pop.alice, "Staff")
	now := time.Now().UTC()
	// Ascending order: the incremental summary path is what production uses.
	recordLogin(t, svc, pop.alice, now.AddDate(0, 0, -1))
	recordLogin(t, svc, pop.alice, now)
	recordLogin(t, svc, pop.bob, now.AddDate(0, 0, -40))
	recordLogin(t, svc, pop.carol, now.AddDate(0, 0, -120))
	return pop
}

func seedOIDCUser(t *testing.T, database *sqlx.DB, email, timezone string) string {
	t.Helper()
	userID := seedUser(t, database, email)
	if timezone != "" {
		if _, err := database.Exec(`UPDATE users SET timezone = $1 WHERE id = $2`, timezone, userID); err != nil {
			t.Fatalf("set timezone for %s: %v", email, err)
		}
	}
	if _, err := database.Exec(
		`INSERT INTO user_identities (id, user_id, source, external_id, email, oidc_sub)
		 VALUES ($1, $2, 'oidc', $3, $4, $3)`,
		uuid.NewString(), userID, "sub-"+userID, email,
	); err != nil {
		t.Fatalf("link oidc identity for %s: %v", email, err)
	}
	return userID
}

func seedRole(t *testing.T, database *sqlx.DB, userID, name string) {
	t.Helper()
	roleID := uuid.NewString()
	if _, err := database.Exec(`INSERT INTO roles (id, name) VALUES ($1, $2)`, roleID, name); err != nil {
		t.Fatalf("seed role %s: %v", name, err)
	}
	if _, err := database.Exec(
		`INSERT INTO user_roles (user_id, role_id, granted_at) VALUES ($1, $2, NOW())`,
		userID, roleID,
	); err != nil {
		t.Fatalf("grant role %s: %v", name, err)
	}
}

func recordLogin(t *testing.T, svc *service.Service, userID string, occurredAt time.Time) {
	t.Helper()
	result, err := svc.RecordLoginEvent(t.Context(), userID, service.RecordLoginEventInput{
		EventKey:   fmt.Sprintf("seed:%s:%d", userID, occurredAt.UnixNano()),
		OccurredAt: occurredAt,
		Provider:   "oidc",
		SessionID:  "seed-session",
	})
	if err != nil {
		t.Fatalf("record login for %s: %v", userID, err)
	}
	if !result.Recorded {
		t.Fatalf("seed login for %s was treated as a duplicate", userID)
	}
}

func activityList(t *testing.T, srv *Server, query string) UserActivityListResponse {
	t.Helper()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/users/activity?"+query, nil)
	req = withTestCaller(req, "admin-1", models.UsersActivityRead)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET /users/activity?%s: got %d, want 200 (%s)", query, rr.Code, rr.Body.String())
	}
	var body UserActivityListResponse
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode activity list: %v", err)
	}
	return body
}

func activityAnalytics(t *testing.T, srv *Server, path string) store.UserActivityAnalytics {
	t.Helper()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req = withTestCaller(req, "admin-1", models.UsersActivityRead)
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("GET %s: got %d, want 200 (%s)", path, rr.Code, rr.Body.String())
	}
	var body store.UserActivityAnalytics
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode analytics: %v", err)
	}
	return body
}

func TestUserActivity_RequiresActivityPrivilege(t *testing.T) {
	_, _, srv := setupTestStack(t)
	for _, path := range []string{
		"/users/activity?window=30&status=all",
		"/users/activity/analytics?window=30",
		"/users/u-1/activity/analytics?window=30",
	} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		// Reading the user directory must not imply reading sign-in history.
		req = withTestCaller(req, "admin-1", models.UsersRead)
		srv.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Errorf("GET %s with users:read only: got %d, want 403", path, rr.Code)
		}
	}
}

func TestUserActivity_RejectsInvalidFilters(t *testing.T) {
	_, _, srv := setupTestStack(t)
	cases := []string{
		"/users/activity?status=all",
		"/users/activity?window=0&status=all",
		"/users/activity?window=366&status=all",
		"/users/activity?window=abc&status=all",
		"/users/activity?window=30",
		"/users/activity?window=30&status=stale",
		"/users/activity?window=30&status=all&sort=email",
		"/users/activity?window=30&status=all&direction=sideways",
		"/users/activity/analytics?window=0",
		"/users/activity/analytics?window=366",
	}
	for _, path := range cases {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req = withTestCaller(req, "admin-1", models.UsersActivityRead)
		srv.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("GET %s: got %d, want 400", path, rr.Code)
		}
	}
}

func TestUserActivity_ListsDisjointStatusesInUserLocalDays(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	pop := seedActivityPopulation(t, database, svc)

	all := activityList(t, srv, "window=30&status=all&limit=50")
	if all.Total != 4 {
		t.Fatalf("total for the OIDC-linked population: got %d, want 4", all.Total)
	}
	if all.WindowDays != 30 || all.Status != "all" || all.Limit != 50 || all.Offset != 0 {
		t.Errorf("response must echo the normalized filter, got %+v", all)
	}
	for _, row := range all.Items {
		if row.UserID == pop.erin {
			t.Error("a user with no OIDC identity must stay out of the activity population")
		}
	}

	byStatus := map[string][]string{}
	for _, status := range []string{"active", "dormant", "never"} {
		page := activityList(t, srv, "window=30&status="+status+"&limit=50")
		if page.Status != status {
			t.Errorf("status echo: got %q, want %q", page.Status, status)
		}
		if page.Total != len(page.Items) {
			t.Errorf("%s: total %d disagrees with %d returned rows", status, page.Total, len(page.Items))
		}
		for _, row := range page.Items {
			byStatus[status] = append(byStatus[status], row.UserID)
		}
	}
	assertSameUsers(t, "active", byStatus["active"], []string{pop.alice})
	assertSameUsers(t, "dormant", byStatus["dormant"], []string{pop.bob, pop.carol})
	assertSameUsers(t, "never", byStatus["never"], []string{pop.dave})

	// A 90-day window moves the 40-day user from dormant to active without
	// disturbing the never group.
	wide := activityList(t, srv, "window=90&status=active&limit=50")
	assertSameUsers(t, "active at 90 days", userIDs(wide.Items), []string{pop.alice, pop.bob})
}

func TestUserActivity_RowCarriesWindowAndLifetimeCounts(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	pop := seedActivityPopulation(t, database, svc)

	page := activityList(t, srv, "window=30&status=all&query=alice&limit=50")
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("search for alice: total=%d items=%d", page.Total, len(page.Items))
	}
	row := page.Items[0]
	if row.UserID != pop.alice {
		t.Fatalf("unexpected row %s", row.UserID)
	}
	if row.InactiveDays == nil || *row.InactiveDays != 0 {
		t.Errorf("inactive_days: got %v, want 0", row.InactiveDays)
	}
	if row.LastLogin == nil {
		t.Error("last_login must be set for a user who signed in")
	}
	if row.LoginCount != 2 || row.WindowLoginCount != 2 || row.LoginDayCount != 2 {
		t.Errorf("counts: login=%d window=%d days=%d, want 2/2/2",
			row.LoginCount, row.WindowLoginCount, row.LoginDayCount)
	}
	if row.CurrentStreak != 2 {
		t.Errorf("current_streak: got %d, want 2", row.CurrentStreak)
	}
	if len(row.RoleNames) != 1 || row.RoleNames[0] != "Staff" {
		t.Errorf("role_names: got %v, want [Staff]", row.RoleNames)
	}

	// A one-day window still reports lifetime totals but only today's count.
	narrow := activityList(t, srv, "window=1&status=all&query=alice&limit=50")
	if narrow.Items[0].WindowLoginCount != 1 || narrow.Items[0].LoginCount != 2 {
		t.Errorf("one-day window: window=%d lifetime=%d, want 1/2",
			narrow.Items[0].WindowLoginCount, narrow.Items[0].LoginCount)
	}

	never := activityList(t, srv, "window=30&status=never&limit=50")
	if never.Items[0].LastLogin != nil || never.Items[0].InactiveDays != nil {
		t.Error("a never-signed-in row must report null last_login and inactive_days")
	}
	if never.Items[0].LoginCount != 0 || never.Items[0].WindowLoginCount != 0 {
		t.Error("a never-signed-in row must report zero counts")
	}
	if never.Items[0].RoleNames == nil {
		t.Error("role_names must be an empty list, not null, when no roles are held")
	}
}

func TestUserActivity_SortsAndPagesStably(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	seedActivityPopulation(t, database, svc)

	desc := activityList(t, srv, "window=30&status=all&sort=last_login&direction=desc&limit=50")
	asc := activityList(t, srv, "window=30&status=all&sort=last_login&direction=asc&limit=50")
	// Never-signed-in rows carry no date and stay last in both directions.
	if last := desc.Items[len(desc.Items)-1]; last.LastLogin != nil {
		t.Error("descending order must leave the never-signed-in row last")
	}
	if last := asc.Items[len(asc.Items)-1]; last.LastLogin != nil {
		t.Error("ascending order must leave the never-signed-in row last")
	}

	byName := activityList(t, srv, "window=30&status=all&sort=name&direction=asc&limit=50")
	names := make([]string, 0, len(byName.Items))
	for _, row := range byName.Items {
		names = append(names, row.Name)
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("name sort is not ascending: %v", names)
			break
		}
	}

	first := activityList(t, srv, "window=30&status=all&limit=2&offset=0")
	second := activityList(t, srv, "window=30&status=all&limit=2&offset=2")
	if first.Total != 4 || second.Total != 4 {
		t.Errorf("total must describe the filtered set, not the page: %d/%d", first.Total, second.Total)
	}
	if second.Offset != 2 || len(second.Items) != 2 {
		t.Errorf("second page: offset=%d items=%d", second.Offset, len(second.Items))
	}
	seen := map[string]bool{}
	for _, row := range append(first.Items, second.Items...) {
		if seen[row.UserID] {
			t.Errorf("user %s appears on both pages", row.UserID)
		}
		seen[row.UserID] = true
	}
}

func TestUserActivity_AnalyticsMatchTheListPopulation(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	seedActivityPopulation(t, database, svc)

	analytics := activityAnalytics(t, srv, "/users/activity/analytics?window=30")
	if analytics.WindowDays != 30 {
		t.Errorf("window_days: got %d, want 30", analytics.WindowDays)
	}
	if analytics.TotalUsers != 4 || analytics.UsersEverLoggedIn != 3 || analytics.ActiveUsers != 1 {
		t.Errorf("population: total=%d ever=%d active=%d, want 4/3/1",
			analytics.TotalUsers, analytics.UsersEverLoggedIn, analytics.ActiveUsers)
	}
	if analytics.ActiveUsers > analytics.UsersEverLoggedIn || analytics.UsersEverLoggedIn > analytics.TotalUsers {
		t.Error("analytics must describe mutually exclusive populations")
	}
	if analytics.LifetimeLoginCount != 4 || analytics.LifetimeActiveDays != 4 {
		t.Errorf("lifetime: logins=%d days=%d, want 4/4",
			analytics.LifetimeLoginCount, analytics.LifetimeActiveDays)
	}
	if analytics.WindowLoginCount != 2 || analytics.WindowActiveDays != 2 {
		t.Errorf("window: logins=%d days=%d, want 2/2",
			analytics.WindowLoginCount, analytics.WindowActiveDays)
	}
	if len(analytics.Trend) != 2 {
		t.Errorf("trend buckets: got %d, want 2 active days", len(analytics.Trend))
	}
	for _, point := range analytics.Trend {
		if point.ActiveUsers != 1 || point.LoginCount != 1 {
			t.Errorf("trend point %s: active=%d logins=%d, want 1/1",
				point.Date, point.ActiveUsers, point.LoginCount)
		}
	}
	// The 40-day user sits in the window immediately before this one.
	if analytics.PriorActiveUsers == nil || *analytics.PriorActiveUsers != 1 {
		t.Errorf("prior_active_users: got %v, want 1", analytics.PriorActiveUsers)
	}
	// Only the 120-day user is dormant beyond 90 days.
	if analytics.DormantOver90Days == nil || *analytics.DormantOver90Days != 1 {
		t.Errorf("dormant_over_90_days: got %v, want 1", analytics.DormantOver90Days)
	}
	if analytics.OldestNeverCreatedAt == nil {
		t.Error("oldest_never_created_at must be set while a never-signed-in user exists")
	}

	wide := activityAnalytics(t, srv, "/users/activity/analytics?window=90")
	if wide.ActiveUsers != 2 {
		t.Errorf("active users at 90 days: got %d, want 2", wide.ActiveUsers)
	}
	if wide.DormantOver90Days == nil || *wide.DormantOver90Days != 1 {
		t.Errorf("dormant beyond 90 days at a 90-day window: got %v, want 1", wide.DormantOver90Days)
	}
}

func TestUserActivity_PerUserAnalyticsIsolateTheUser(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	pop := seedActivityPopulation(t, database, svc)

	alice := activityAnalytics(t, srv, "/users/"+pop.alice+"/activity/analytics?window=30")
	if alice.TotalUsers != 1 || alice.UsersEverLoggedIn != 1 || alice.ActiveUsers != 1 {
		t.Errorf("alice: total=%d ever=%d active=%d, want 1/1/1",
			alice.TotalUsers, alice.UsersEverLoggedIn, alice.ActiveUsers)
	}
	if alice.LifetimeLoginCount != 2 || alice.WindowLoginCount != 2 || len(alice.Trend) != 2 {
		t.Errorf("alice: lifetime=%d window=%d trend=%d, want 2/2/2",
			alice.LifetimeLoginCount, alice.WindowLoginCount, len(alice.Trend))
	}

	dave := activityAnalytics(t, srv, "/users/"+pop.dave+"/activity/analytics?window=30")
	if dave.UsersEverLoggedIn != 0 || dave.LifetimeLoginCount != 0 || len(dave.Trend) != 0 {
		t.Errorf("never-signed-in user: ever=%d lifetime=%d trend=%d, want 0/0/0",
			dave.UsersEverLoggedIn, dave.LifetimeLoginCount, len(dave.Trend))
	}

	for _, id := range []string{uuid.NewString(), pop.erin} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/users/"+id+"/activity/analytics?window=30", nil)
		req = withTestCaller(req, "admin-1", models.UsersActivityRead)
		srv.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Errorf("analytics for %s: got %d, want 404", id, rr.Code)
		}
	}
}

func TestRecordLoginEvent_IsIdempotentPerToken(t *testing.T) {
	database, _, srv := setupTestStack(t)
	userID := seedOIDCUser(t, database, "grace@example.org", "")

	issuedAt := time.Now().UTC().Add(-time.Minute).Unix()
	caller := &identity.Caller{
		UserID:   userID,
		Issuer:   "https://issuer.example.org",
		Subject:  "sub-" + userID,
		TokenID:  "jti-1",
		IssuedAt: issuedAt,
	}
	if code, recorded := postLoginEvent(t, srv, caller); code != http.StatusCreated || !recorded {
		t.Fatalf("first capture: got %d recorded=%v, want 201 true", code, recorded)
	}
	// The portal retries the same callback; the counts must not move.
	if code, recorded := postLoginEvent(t, srv, caller); code != http.StatusOK || recorded {
		t.Fatalf("repeat capture: got %d recorded=%v, want 200 false", code, recorded)
	}

	page := activityList(t, srv, "window=30&status=active&query=grace&limit=10")
	if page.Total != 1 {
		t.Fatalf("captured user should be active: total=%d", page.Total)
	}
	if page.Items[0].LoginCount != 1 || page.Items[0].LoginDayCount != 1 || page.Items[0].CurrentStreak != 1 {
		t.Errorf("after one recorded login: count=%d days=%d streak=%d, want 1/1/1",
			page.Items[0].LoginCount, page.Items[0].LoginDayCount, page.Items[0].CurrentStreak)
	}

	// A second sign-in with a fresh token on the same day adds a login but
	// not another active day.
	next := *caller
	next.TokenID = "jti-2"
	next.IssuedAt = issuedAt + 60
	if code, recorded := postLoginEvent(t, srv, &next); code != http.StatusCreated || !recorded {
		t.Fatalf("second token: got %d recorded=%v, want 201 true", code, recorded)
	}
	page = activityList(t, srv, "window=30&status=active&query=grace&limit=10")
	if page.Items[0].LoginCount != 2 || page.Items[0].LoginDayCount != 1 {
		t.Errorf("same-day second sign-in: count=%d days=%d, want 2/1",
			page.Items[0].LoginCount, page.Items[0].LoginDayCount)
	}
}

func TestRecordLoginEvent_RebuildsSummaryForLateEvents(t *testing.T) {
	database, svc, srv := setupTestStack(t)
	userID := seedOIDCUser(t, database, "ivan@example.org", "")
	now := time.Now().UTC()

	// Three consecutive days arriving newest-first: the middle day lands
	// after the streak was already advanced, so the summary is replayed.
	recordLogin(t, svc, userID, now)
	recordLogin(t, svc, userID, now.AddDate(0, 0, -2))
	recordLogin(t, svc, userID, now.AddDate(0, 0, -1))

	page := activityList(t, srv, "window=30&status=active&query=ivan&limit=10")
	if page.Total != 1 {
		t.Fatalf("expected one active row, got %d", page.Total)
	}
	row := page.Items[0]
	if row.LoginCount != 3 || row.LoginDayCount != 3 {
		t.Errorf("counts after replay: logins=%d days=%d, want 3/3", row.LoginCount, row.LoginDayCount)
	}
	if row.CurrentStreak != 3 {
		t.Errorf("streak after replay: got %d, want 3 consecutive days", row.CurrentStreak)
	}
	if row.InactiveDays == nil || *row.InactiveDays != 0 {
		t.Errorf("inactive_days must follow the newest event, got %v", row.InactiveDays)
	}
}

func TestRecordLoginEvent_RejectsTokenWithoutSessionEvidence(t *testing.T) {
	database, _, srv := setupTestStack(t)
	userID := seedOIDCUser(t, database, "heidi@example.org", "")
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/me/login-events", nil)
	req = req.WithContext(identity.WithCaller(req.Context(), &identity.Caller{UserID: userID}))
	srv.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("capture without jti/iat or sid/auth_time: got %d, want 400 (%s)", rr.Code, rr.Body.String())
	}
}

func postLoginEvent(t *testing.T, srv *Server, caller *identity.Caller) (int, bool) {
	t.Helper()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/me/login-events", nil)
	req = req.WithContext(identity.WithCaller(req.Context(), caller))
	srv.ServeHTTP(rr, req)
	var body service.RecordLoginEventResult
	if rr.Code == http.StatusCreated || rr.Code == http.StatusOK {
		if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
			t.Fatalf("decode capture result: %v", err)
		}
	}
	return rr.Code, body.Recorded
}

func userIDs(rows []store.UserActivityRow) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.UserID)
	}
	return ids
}

func assertSameUsers(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("%s: got %d users %v, want %d %v", label, len(got), got, len(want), want)
		return
	}
	expected := map[string]bool{}
	for _, id := range want {
		expected[id] = true
	}
	for _, id := range got {
		if !expected[id] {
			t.Errorf("%s: unexpected user %s in %v", label, id, got)
		}
	}
}
