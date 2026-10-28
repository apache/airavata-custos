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

package store

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/db/dbtest"
)

// setupTestDB opens and migrates the test DB once, then truncates the event
// bus tables on every call.
func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	database := dbtest.Open(t, dbtest.CoreDSN(t))
	dbtest.Truncate(t, database,
		"event_deliveries",
		"events",
		"event_subscriptions",
	)
	return database
}

// inTx runs fn in a transaction and fails the test if it does not commit.
func inTx(t *testing.T, database *sqlx.DB, fn func(tx *sql.Tx) error) {
	t.Helper()
	if err := db.TxFn(context.Background(), database, fn); err != nil {
		t.Fatalf("tx: %v", err)
	}
}
