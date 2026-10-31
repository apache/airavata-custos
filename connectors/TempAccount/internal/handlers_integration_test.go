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

package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	"github.com/apache/airavata-custos/pkg/service"
)

func setup(t *testing.T) (*sqlx.DB, *Handlers, func(string, ...any), string) {
	dsn := os.Getenv("CORE_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("CORE_TEST_DATABASE_DSN not set")
	}
	database, err := db.Open(db.Config{DSN: dsn, MaxOpenConns: 5, MaxIdleConns: 2})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.MigrateEmbedded(database); err != nil {
		t.Fatal(err)
	}
	bus, err := events.New(context.Background(), database)
	if err != nil {
		t.Fatal(err)
	}
	exec := func(q string, args ...any) {
		if _, err := database.Exec(q, args...); err != nil {
			t.Fatal(err)
		}
	}
	orgID := uuid.NewString()
	exec(`INSERT INTO organizations (id, originated_id, name) VALUES ($1, $2, 'Org')`, orgID, orgID)
	return database, NewHandlers(service.New(database, bus)), exec, orgID
}

func TestUpdateAllocationKeepsMembershipOnVirtualUser(t *testing.T) {
	for _, userType := range []string{"CLUSTER_LOCAL", "VIRTUAL"} {
		t.Run(userType, func(t *testing.T) {
			database, h, exec, orgID := setup(t)
			userID, otherID, clusterID, projectID, allocID, memID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
			for id, typ := range map[string]string{userID: userType, otherID: "VIRTUAL"} {
				exec(`INSERT INTO users (id, organization_id, first_name, last_name, email, type) VALUES ($1, $2, 'T', 'U', $3, $4)`, id, orgID, id, typ)
			}
			exec(`INSERT INTO compute_clusters (id, name) VALUES ($1, $1)`, clusterID)
			exec(`INSERT INTO projects (id, originated_id, title, origination, project_pi_id) VALUES ($1, $1, 'P', 'TEST', $2)`, projectID, userID)
			exec(`INSERT INTO compute_allocations (id, project_id, name, status, compute_cluster_id, start_time, end_time)
			      VALUES ($1, $2, $1, 'ACTIVE', $3, NOW(), NOW() + INTERVAL '1 year')`, allocID, projectID, clusterID)
			exec(`INSERT INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status)
			      VALUES ($1, $2, $3, NOW(), NOW() + INTERVAL '1 year', 'ACTIVE')`, memID, allocID, userID)

			body := `{"id":"` + memID + `","user_id":"` + otherID + `","compute_allocation_id":"` + allocID + `","membership_status":"INACTIVE"}`
			rec := httptest.NewRecorder()
			h.updateAllocationToTempAccount(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)))
			var got struct {
				UserID string `db:"user_id"`
				Status string `db:"membership_status"`
			}
			if err := database.Get(&got, `SELECT user_id, membership_status FROM compute_allocation_memberships WHERE id = $1`, memID); err != nil {
				t.Fatal(err)
			}
			if got.UserID != userID || userType == "CLUSTER_LOCAL" && (rec.Code != http.StatusBadRequest || got.Status != "ACTIVE") {
				t.Errorf("got %d, user %s, status %s", rec.Code, got.UserID, got.Status)
			}
		})
	}
}

func TestCreateTempAccountIgnoresClientIDAndStatus(t *testing.T) {
	_, h, _, orgID := setup(t)
	chosenID := uuid.NewString()
	body := `{"id":"` + chosenID + `","status":"ACTIVE","type":"VIRTUAL","email":"` + chosenID + `@example.edu","organization_id":"` + orgID + `"}`
	rec := httptest.NewRecorder()
	h.createTempAccount(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)))

	var created models.User
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || rec.Code != http.StatusCreated {
		t.Fatalf("got %d %s, want 201", rec.Code, rec.Body)
	}
	if created.ID == chosenID || created.Status != models.UserPending {
		t.Errorf("created id %q status %q, want a server id and PENDING", created.ID, created.Status)
	}
}
