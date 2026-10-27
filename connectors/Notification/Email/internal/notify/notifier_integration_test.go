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

package notify

import (
	"context"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/tracing/tracingtest"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

// syncRelay is a mockRelay the event worker can call from its own goroutine.
type syncRelay struct {
	mu   sync.Mutex
	sent []string
}

func (r *syncRelay) Send(_ context.Context, m Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sent = append(r.sent, m.To+" | "+m.Subject)
	return nil
}

func (r *syncRelay) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.sent)
}

// Make sure for each account event: pending, ready, and denied an email is sent only once.
func TestNotifier_SendsOneEmailPerAccountEvent(t *testing.T) {
	dsn := os.Getenv("CORE_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("integration env not set: CORE_TEST_DATABASE_DSN required")
	}
	database, err := db.Open(db.Config{DSN: dsn, MaxOpenConns: 5, MaxIdleConns: 2})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.MigrateEmbedded(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec("TRUNCATE TABLE audit_events, event_deliveries, events, event_subscriptions, compute_cluster_users, compute_clusters, users, organizations CASCADE"); err != nil {
		t.Fatal(err)
	}
	ctx := tracingtest.Context()
	bus, err := events.New(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	svc := service.New(database, bus)
	relay := &syncRelay{}
	NewNotifier(svc, relay, Site{SiteName: "Example HPC", PortalURL: "https://portal.example.edu", ClusterHost: "login.example.edu", SupportEmail: "help@example.edu"}).RegisterSubscribers(bus, "email-notifier")

	org, err := svc.CreateOrganization(ctx, &models.Organization{OriginatedID: uuid.NewString(), Name: "Example University"})
	if err != nil {
		t.Fatal(err)
	}
	cluster, err := svc.CreateComputeCluster(ctx, &models.ComputeCluster{Name: "hpc-" + uuid.NewString()[:8]})
	if err != nil {
		t.Fatal(err)
	}
	newUser := func(first, username string) string {
		user, err := svc.CreateUser(ctx, &models.User{OrganizationID: org.ID, FirstName: first, LastName: "Doe", Email: username + "@example.edu"})
		if err != nil {
			t.Fatal(err)
		}
		return user.ID
	}
	admin := newUser("Ada", "aadmin")
	account := func(first, username string, status models.ClusterAccountApproval) *models.ComputeClusterUser {
		cu := &models.ComputeClusterUser{ComputeClusterID: cluster.ID, UserID: newUser(first, username), LocalUsername: username, ApprovalStatus: status}
		if status == models.ClusterAccountApproved {
			cu.ReviewedBy = &admin
		}
		cu, err := svc.CreateComputeClusterUser(ctx, cu)
		if err != nil {
			t.Fatal(err)
		}
		return cu
	}
	account("Jane", "jdoe", models.ClusterAccountPending)
	john := account("John", "johndoe", models.ClusterAccountPending)
	if _, err := svc.DenyComputeClusterUser(ctx, john.ID, admin, "Not on the allocation."); err != nil {
		t.Fatal(err)
	}
	amy := account("Amy", "amydoe", models.ClusterAccountApproved)
	if err := svc.MarkComputeClusterUserProvisioned(ctx, amy.ID); err != nil {
		t.Fatal(err)
	}

	runCtx, stop := context.WithCancel(ctx)
	go bus.Run(runCtx)
	deadline := time.Now().Add(5 * time.Second)
	for relay.count() < 4 && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	stop()

	want := []string{
		"amydoe@example.edu | Your Example HPC account is ready",
		"jdoe@example.edu | We received a request for your Example HPC account",
		"johndoe@example.edu | We received a request for your Example HPC account",
		"johndoe@example.edu | Your Example HPC account request was not approved",
	}
	slices.Sort(relay.sent)
	if !slices.Equal(relay.sent, want) {
		t.Fatalf("emails sent:\n%v\nwant:\n%v", relay.sent, want)
	}
	var audited int
	if err := database.Get(&audited, "SELECT COUNT(*) FROM audit_events WHERE event_type = $1 AND source = 'email-notifier'", AuditNotificationSent); err != nil {
		t.Fatal(err)
	}
	if audited != len(want) {
		t.Fatalf("expected %d sent-email audit rows, got %d", len(want), audited)
	}
}
