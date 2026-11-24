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

package events

import (
	"context"
	"strings"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db/dbtest"
)

// setupTestDB opens and migrates the test DB once, then truncates the event bus tables on every call.
func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	sharedDB := dbtest.Open(t, dbtest.CoreDSN(t))
	tables := []string{
		"audit_events",
		"event_deliveries",
		"events",
		"event_subscriptions",
	}
	if _, err := sharedDB.Exec("TRUNCATE TABLE " + strings.Join(tables, ", ") + " CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return sharedDB
}

// newBus creates a bus on the test database and fails the test if that fails.
func newBus(t *testing.T, database *sqlx.DB) *Bus {
	t.Helper()
	bus, err := New(context.Background(), database)
	if err != nil {
		t.Fatalf("new bus: %v", err)
	}
	return bus
}
