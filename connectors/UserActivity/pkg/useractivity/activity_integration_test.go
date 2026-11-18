//go:build integration

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
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/db/dbtest"
	"github.com/apache/airavata-custos/internal/tracing/tracingtest"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
	"github.com/google/uuid"
)

func TestAuditLoginIsIdempotentAndFeedsDashboard(t *testing.T) {
	database := dbtest.Open(t, dbtest.CoreDSN(t))
	_, err := database.Exec(`TRUNCATE audit_events, user_identities, users, organizations CASCADE`)
	if err != nil {
		t.Fatal(err)
	}
	orgID, userID, identityID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	if _, err = database.Exec(`INSERT INTO organizations (id, originated_id, name) VALUES ($1,$2,'Test')`, orgID, orgID); err != nil {
		t.Fatal(err)
	}
	if _, err = database.Exec(`INSERT INTO users (id, organization_id, first_name, last_name, email, status) VALUES ($1,$2,'Ada','Lovelace','ada@example.org','ACTIVE')`, userID, orgID); err != nil {
		t.Fatal(err)
	}
	if _, err = database.Exec(`INSERT INTO user_identities (id, user_id, source, external_id, oidc_sub) VALUES ($1,$2,'oidc','ada-sub','ada-sub')`, identityID, userID); err != nil {
		t.Fatal(err)
	}
	c := &Connector{db: database}
	ctx := tracingtest.Context()
	login := service.PortalLogin{ID: "portal-login:test-key", UserID: userID, OccurredAt: time.Now().UTC()}
	for i := 0; i < 2; i++ {
		if err := c.recordLogin(ctx, events.Event{}, login); err != nil {
			t.Fatal(err)
		}
	}
	var count int
	if err := database.Get(&count, `SELECT COUNT(*) FROM audit_events WHERE event_type='PORTAL_LOGIN' AND entity_id=$1`, userID); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("duplicate login stored %d times", count)
	}
	// A historical login and a never-signed-in identity exercise the disjoint
	// dashboard statuses without creating any derived counters.
	dormantID, neverID := uuid.NewString(), uuid.NewString()
	for _, user := range []struct{ id, email, sub string }{
		{dormantID, "dormant@example.org", "dormant-sub"},
		{neverID, "never@example.org", "never-sub"},
	} {
		if _, err := database.Exec(`INSERT INTO users (id, organization_id, first_name, last_name, email, status) VALUES ($1,$2,'Test','User',$3,'ACTIVE')`, user.id, orgID, user.email); err != nil {
			t.Fatal(err)
		}
		if _, err := database.Exec(`INSERT INTO user_identities (id, user_id, source, external_id, oidc_sub) VALUES ($1,$2,'oidc',$3,$3)`, uuid.NewString(), user.id, user.sub); err != nil {
			t.Fatal(err)
		}
	}
	if err := c.recordLogin(ctx, events.Event{}, service.PortalLogin{ID: "portal-login:old-key", UserID: dormantID, OccurredAt: time.Now().AddDate(0, 0, -10).UTC()}); err != nil {
		t.Fatal(err)
	}
	analytics, err := c.readAnalytics(ctx, "", 7)
	if err != nil {
		t.Fatal(err)
	}
	if analytics.TotalUsers != 3 || analytics.UsersEverLoggedIn != 2 || analytics.ActiveUsers != 1 || analytics.PriorActiveUsers != 1 || analytics.WindowLoginCount != 1 || len(analytics.Trend) != 7 {
		t.Fatalf("unexpected analytics: %+v", analytics)
	}
	r := httptest.NewRequest("GET", "/users/activity?window=7&status=active&limit=10", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	c.list(w, r)
	if w.Code != 200 {
		t.Fatalf("list status %d: %s", w.Code, w.Body.String())
	}
	var page struct {
		Total int           `json:"total"`
		Items []ActivityRow `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].LoginCount != 1 || page.Items[0].CurrentStreak != 1 {
		t.Fatalf("unexpected activity page: %+v", page)
	}
	for _, status := range []string{"dormant", "never"} {
		r := httptest.NewRequest("GET", "/users/activity?window=7&status="+status+"&limit=10", nil).WithContext(ctx)
		w := httptest.NewRecorder()
		c.list(w, r)
		if w.Code != 200 {
			t.Fatalf("%s list status %d: %s", status, w.Code, w.Body.String())
		}
		if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
			t.Fatal(err)
		}
		if page.Total != 1 || len(page.Items) != 1 {
			t.Fatalf("%s list: %+v", status, page)
		}
	}
}

func TestVerifiedPortalLoginReachesAuditThroughDurableBus(t *testing.T) {
	database := dbtest.Open(t, dbtest.CoreDSN(t))
	if _, err := database.Exec(`TRUNCATE event_deliveries, events, event_subscriptions, audit_events, user_identities, users, organizations CASCADE`); err != nil {
		t.Fatal(err)
	}
	orgID, userID := uuid.NewString(), uuid.NewString()
	if _, err := database.Exec(`INSERT INTO organizations (id, originated_id, name) VALUES ($1,$2,'Test')`, orgID, orgID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO users (id, organization_id, first_name, last_name, email, status) VALUES ($1,$2,'Ada','Lovelace','ada2@example.org','ACTIVE')`, userID, orgID); err != nil {
		t.Fatal(err)
	}
	bus, err := events.New(context.Background(), database)
	if err != nil {
		t.Fatal(err)
	}
	c := &Connector{db: database}
	bus.Subscribe(Type, events.PortalLoginEvent, c.recordLogin)
	svc := service.New(database, bus)
	claims := &identity.Claims{Sub: "ada-sub", Issuer: "https://issuer.example.org", TokenID: "token-1", IssuedAt: time.Now().Add(-time.Minute).Unix()}
	for i := 0; i < 2; i++ {
		if err := svc.PublishPortalLogin(tracingtest.Context(), userID, claims); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go bus.Run(ctx)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var count, completed int
		if err := database.Get(&count, `SELECT COUNT(*) FROM audit_events WHERE event_type='PORTAL_LOGIN'`); err != nil {
			t.Fatal(err)
		}
		if err := database.Get(&completed, `SELECT COUNT(*) FROM event_deliveries WHERE status='SUCCEEDED'`); err != nil {
			t.Fatal(err)
		}
		if count == 1 && completed == 2 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("both deliveries should complete while only one portal login is recorded")
}

func TestTrendUsesUTCDayWithNonUTCSession(t *testing.T) {
	dsn := dbtest.CoreDSN(t)
	dbtest.Open(t, dsn) // apply migrations on the reserved test database
	database, err := db.Open(db.Config{DSN: dsn, MaxOpenConns: 1, MaxIdleConns: 1})
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.Exec(`TRUNCATE event_deliveries, events, event_subscriptions, audit_events, user_identities, users, organizations CASCADE`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`SET TIME ZONE 'America/New_York'`); err != nil {
		t.Fatal(err)
	}
	orgID, userID := uuid.NewString(), uuid.NewString()
	if _, err := database.Exec(`INSERT INTO organizations (id, originated_id, name) VALUES ($1,$2,'Test')`, orgID, orgID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO users (id, organization_id, first_name, last_name, email, status) VALUES ($1,$2,'UTC','User','utc@example.org','ACTIVE')`, userID, orgID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO user_identities (id, user_id, source, external_id, oidc_sub) VALUES ($1,$2,'oidc','utc-sub','utc-sub')`, uuid.NewString(), userID); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	loginTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 30, 0, 0, time.UTC)
	c := &Connector{db: database}
	if err := c.recordLogin(tracingtest.Context(), events.Event{}, service.PortalLogin{ID: "portal-login:utc-boundary", UserID: userID, OccurredAt: loginTime}); err != nil {
		t.Fatal(err)
	}
	result, err := c.readAnalytics(tracingtest.Context(), "", 1)
	if err != nil {
		t.Fatal(err)
	}
	if result.WindowLoginCount != 1 || len(result.Trend) != 1 || result.Trend[0].LoginCount != 1 {
		t.Fatalf("UTC day mismatch: %+v", result)
	}
}
