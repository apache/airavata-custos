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
	"context"
	"net/http"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db/dbtest"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/service"
)

func setupTestStack(t *testing.T) (*sqlx.DB, *service.Service, *Server) {
	t.Helper()
	database := dbtest.Open(t, dbtest.CoreDSN(t))
	dbtest.Truncate(t, database,
		"user_roles",
		"role_privileges",
		"roles",
		"user_privileges",
		"audit_events",
		"event_deliveries",
		"events",
		"event_subscriptions",
		"compute_allocation_usages",
		"compute_allocation_memberships",
		"compute_allocation_resource_mappings",
		"compute_allocation_resources",
		"compute_allocations",
		"project_memberships",
		"projects",
		"compute_cluster_users",
		"compute_clusters",
		"user_identities",
		"users",
		"organizations",
	)
	bus, err := events.New(context.Background(), database)
	if err != nil {
		t.Fatalf("event bus: %v", err)
	}
	svc := service.New(database, bus)
	router := identity.NewRouter(http.NewServeMux())
	return database, svc, New(svc, router)
}

var (
	seedUser      = dbtest.SeedUser
	seedPrivilege = dbtest.SeedPrivilege
)
