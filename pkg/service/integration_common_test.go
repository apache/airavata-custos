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

package service

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
)

var (
	sharedDB     *sqlx.DB
	sharedDBOnce sync.Once
	sharedDBErr  error
)

// setupTestDB opens (once per process) and migrates the test DB pointed to by
// CORE_TEST_DATABASE_DSN or DATABASE_DSN. Subsequent calls truncate the core
// tables and return the shared handle.
func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("CORE_TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_DSN")
	}
	if dsn == "" {
		t.Skip("integration env not set: CORE_TEST_DATABASE_DSN or DATABASE_DSN required")
	}
	sharedDBOnce.Do(func() {
		database, err := db.Open(db.Config{
			DSN:          dsn,
			MaxOpenConns: 5,
			MaxIdleConns: 2,
		})
		if err != nil {
			sharedDBErr = err
			return
		}
		if err := db.MigrateEmbedded(database); err != nil {
			sharedDBErr = err
			return
		}
		sharedDB = database
	})
	if sharedDBErr != nil {
		t.Fatalf("setup db: %v", sharedDBErr)
	}
	truncateAll(t, sharedDB)
	return sharedDB
}

// truncateAll wipes the core tables that privilege tests touch.
func truncateAll(t *testing.T, database *sqlx.DB) {
	t.Helper()
	tables := []string{
		"user_roles",
		"role_privileges",
		"roles",
		"user_privileges",
		"audit_events",
		"user_identities",
		"users",
		"organizations",
	}
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		t.Fatalf("disable FK: %v", err)
	}
	for _, tbl := range tables {
		if _, err := database.Exec("TRUNCATE TABLE " + tbl); err != nil {
			t.Fatalf("truncate %s: %v", tbl, err)
		}
	}
	if _, err := database.Exec("SET FOREIGN_KEY_CHECKS = 1"); err != nil {
		t.Fatalf("re-enable FK: %v", err)
	}
}

func newTestService(database *sqlx.DB) *Service {
	return New(database, events.New())
}

func seedOrg(t *testing.T, database *sqlx.DB) string {
	t.Helper()
	orgID := uuid.NewString()
	if _, err := database.Exec(
		"INSERT INTO organizations (id, originated_id, name) VALUES (?, ?, ?)",
		orgID, "TEST-ORG-"+orgID[:8], "Test Org",
	); err != nil {
		t.Fatalf("seed org: %v", err)
	}
	return orgID
}

func seedUser(t *testing.T, database *sqlx.DB, email string) string {
	t.Helper()
	orgID := seedOrg(t, database)
	userID := uuid.NewString()
	if _, err := database.Exec(
		"INSERT INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		userID, orgID, "Test", "User", "", email, string(models.UserActive), string(models.UserTypeClusterLocal),
	); err != nil {
		t.Fatalf("seed user %s: %v", email, err)
	}
	return userID
}

// seedPrivilegeGrant directly inserts an active grant of privileges:grant for
// userID. Bypasses the service guards so tests can stand up a granter without
// a chicken-and-egg dependency.
func seedPrivilegeGrant(t *testing.T, database *sqlx.DB, userID string) {
	t.Helper()
	if _, err := database.Exec(
		`INSERT INTO user_privileges (id, user_id, privilege, granted_at, reason)
		 VALUES (?, ?, ?, NOW(6), 'seed')`,
		uuid.NewString(), userID, string(models.PrivilegeGrant),
	); err != nil {
		t.Fatalf("seed privileges:grant for %s: %v", userID, err)
	}
}

func countAuditEventsOfType(t *testing.T, database *sqlx.DB, eventType, entityID string) int {
	t.Helper()
	var n int
	if err := database.Get(&n,
		"SELECT COUNT(*) FROM audit_events WHERE event_type = ? AND entity_id = ?",
		eventType, entityID,
	); err != nil {
		t.Fatalf("count audit_events: %v", err)
	}
	return n
}

func ctx() context.Context { return context.Background() }
