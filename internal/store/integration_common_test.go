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
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db"
)

var (
	sharedDBOnce sync.Once
	sharedDB     *sqlx.DB
	sharedDBErr  error
)

// setupTestDB opens and migrates the test DB once, then truncates the event
// bus tables on every call.
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
	tables := []string{
		"event_deliveries",
		"events",
		"event_subscriptions",
	}
	if _, err := sharedDB.Exec("TRUNCATE TABLE " + strings.Join(tables, ", ") + " CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return sharedDB
}

// inTx runs fn in a transaction and fails the test if it does not commit.
func inTx(t *testing.T, database *sqlx.DB, fn func(tx *sql.Tx) error) {
	t.Helper()
	if err := db.TxFn(context.Background(), database, fn); err != nil {
		t.Fatalf("tx: %v", err)
	}
}
