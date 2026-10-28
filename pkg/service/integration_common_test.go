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
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db/dbtest"
	"github.com/apache/airavata-custos/internal/tracing/tracingtest"
	"github.com/apache/airavata-custos/pkg/events"
)

// setupTestDB opens (once per process) and migrates the test DB pointed to by
// CORE_TEST_DATABASE_DSN or DATABASE_DSN. Subsequent calls truncate the core
// tables and return the shared handle.
func setupTestDB(t *testing.T) *sqlx.DB {
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
		"user_identities",
		"users",
		"organizations",
	)
	return database
}

func newTestService(database *sqlx.DB) *Service {
	return New(database, newTestBus(database))
}

// newTestBus creates a bus on the test database and panics if that fails.
func newTestBus(database *sqlx.DB) *events.Bus {
	bus, err := events.New(context.Background(), database)
	if err != nil {
		panic("event bus: " + err.Error())
	}
	return bus
}

var (
	seedOrg       = dbtest.SeedOrg
	seedUser      = dbtest.SeedUser
	seedPrivilege = dbtest.SeedPrivilege
)

func countAuditEventsOfType(t *testing.T, database *sqlx.DB, eventType, entityID string) int {
	t.Helper()
	var n int
	if err := database.Get(&n,
		"SELECT COUNT(*) FROM audit_events WHERE event_type = $1 AND entity_id = $2",
		eventType, entityID,
	); err != nil {
		t.Fatalf("count audit_events: %v", err)
	}
	return n
}

func ctx() context.Context { return tracingtest.Context() }
