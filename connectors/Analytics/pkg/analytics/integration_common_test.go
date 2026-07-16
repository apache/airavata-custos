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

package analytics

import (
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

var (
	sharedDB     *sqlx.DB
	sharedDBOnce sync.Once
	sharedDBErr  error
)

// setupTestStack builds the analytics connector against a fresh test database
// and returns the router that serves its routes.
func setupTestStack(t *testing.T) (*sqlx.DB, *Service, http.Handler) {
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

	svc := NewService(coreservice.New(sharedDB, events.New()), sharedDB)
	router := identity.NewRouter(http.NewServeMux())
	NewHandlers(svc).RegisterRoutes(router)
	return sharedDB, svc, router
}

func truncateAll(t *testing.T, database *sqlx.DB) {
	t.Helper()
	tables := []string{
		"user_roles",
		"role_privileges",
		"roles",
		"user_privileges",
		"audit_events",
		"compute_allocation_usages",
		"compute_allocation_resource_mappings",
		"compute_allocation_resources",
		"compute_allocation_memberships",
		"compute_allocations",
		"project_memberships",
		"projects",
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

func seedUser(t *testing.T, database *sqlx.DB, email string) string {
	t.Helper()
	orgID := uuid.NewString()
	if _, err := database.Exec(
		"INSERT INTO organizations (id, originated_id, name) VALUES (?, ?, ?)",
		orgID, "TEST-ORG-"+orgID[:8], "Test Org",
	); err != nil {
		t.Fatalf("seed org: %v", err)
	}
	userID := uuid.NewString()
	if _, err := database.Exec(
		"INSERT INTO users (id, organization_id, first_name, last_name, middle_name, email, status) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, orgID, "Test", "User", "", email, string(models.UserActive),
	); err != nil {
		t.Fatalf("seed user %s: %v", email, err)
	}
	return userID
}

// withTestCaller installs a caller plus a privilege set onto the request
// context for handler tests that run without the auth middleware.
func withTestCaller(r *http.Request, userID string, privs ...models.PrivilegeKey) *http.Request {
	ctx := identity.WithCaller(r.Context(), &identity.Caller{UserID: userID})
	ctx = identity.WithPrivilegesForTest(ctx, privs)
	return r.WithContext(ctx)
}
