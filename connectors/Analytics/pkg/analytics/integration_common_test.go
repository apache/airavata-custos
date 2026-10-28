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
	"context"
	"net/http"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db/dbtest"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

// setupTestStack builds the analytics connector against a fresh test database
// and returns the router that serves its routes.
func setupTestStack(t *testing.T) (*sqlx.DB, *Service, http.Handler) {
	t.Helper()
	database := dbtest.Open(t, dbtest.CoreDSN(t))
	dbtest.Truncate(t, database,
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
	)

	bus, err := events.New(context.Background(), database)
	if err != nil {
		t.Fatalf("event bus: %v", err)
	}
	svc := NewService(coreservice.New(database, bus), database)
	router := identity.NewRouter(http.NewServeMux())
	NewHandlers(svc).RegisterRoutes(router)
	return database, svc, router
}

var seedUser = dbtest.SeedUser

// withTestCaller installs a caller plus a privilege set onto the request
// context for handler tests that run without the auth middleware.
func withTestCaller(r *http.Request, userID string, privs ...models.PrivilegeKey) *http.Request {
	ctx := identity.WithCaller(r.Context(), &identity.Caller{UserID: userID})
	ctx = identity.WithPrivilegesForTest(ctx, privs)
	return r.WithContext(ctx)
}
