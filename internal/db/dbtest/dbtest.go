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

// Package dbtest gives integration tests one migrated database handle per test binary.
// Tests share that handle and reset state by truncating the tables they touch, so they
// must run with -p 1 against a database reserved for tests.
package dbtest

import (
	"io/fs"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/pkg/models"
)

type Migrations struct {
	FS   fs.FS
	Dir  string
	Name string
}

var (
	once      sync.Once
	shared    *sqlx.DB
	sharedErr error
)

func CoreDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv("CORE_TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_DSN")
	}
	if dsn == "" {
		t.Skip("integration env not set: CORE_TEST_DATABASE_DSN or DATABASE_DSN required")
	}
	return dsn
}

// Open opens dsn on the first call and applies the core migrations, then each connector's.
// Later calls return the same handle, whatever their arguments.
func Open(t *testing.T, dsn string, connectors ...Migrations) *sqlx.DB {
	t.Helper()
	once.Do(func() {
		database, err := db.Open(db.Config{DSN: dsn, MaxOpenConns: 5, MaxIdleConns: 2})
		if err == nil {
			err = db.MigrateEmbedded(database)
		}
		for _, m := range connectors {
			if err != nil {
				break
			}
			err = db.MigrateConnectorFS(database, m.FS, m.Dir, m.Name)
		}
		shared, sharedErr = database, err
	})
	if sharedErr != nil {
		t.Fatalf("setup db: %v", sharedErr)
	}
	return shared
}

func Truncate(t *testing.T, database *sqlx.DB, tables ...string) {
	t.Helper()
	if _, err := database.Exec("TRUNCATE TABLE " + strings.Join(tables, ", ") + " CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func CountRows(t *testing.T, database *sqlx.DB, table string) int {
	t.Helper()
	var n int
	if err := database.Get(&n, "SELECT COUNT(*) FROM "+table); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func SeedOrg(t *testing.T, database *sqlx.DB) string {
	t.Helper()
	orgID := uuid.NewString()
	if _, err := database.Exec(
		"INSERT INTO organizations (id, originated_id, name) VALUES ($1, $2, $3)",
		orgID, "TEST-ORG-"+orgID[:8], "Test Org",
	); err != nil {
		t.Fatalf("seed org: %v", err)
	}
	return orgID
}

// SeedUser inserts an active user in a new organization.
func SeedUser(t *testing.T, database *sqlx.DB, email string) string {
	t.Helper()
	orgID := SeedOrg(t, database)
	userID := uuid.NewString()
	if _, err := database.Exec(
		"INSERT INTO users (id, organization_id, first_name, last_name, middle_name, email, status, type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		userID, orgID, "Test", "User", "", email, string(models.UserActive), string(models.UserTypeClusterLocal),
	); err != nil {
		t.Fatalf("seed user %s: %v", email, err)
	}
	return userID
}

// SeedPrivilege grants p to userID directly, bypassing the service guards, so a test can stand up a
// granter without one.
func SeedPrivilege(t *testing.T, database *sqlx.DB, userID string, p models.PrivilegeKey) {
	t.Helper()
	if _, err := database.Exec(
		`INSERT INTO user_privileges (id, user_id, privilege, granted_at, reason)
		 VALUES ($1, $2, $3, NOW(), 'seed')`,
		uuid.NewString(), userID, string(p),
	); err != nil {
		t.Fatalf("seed privilege %s for %s: %v", p, userID, err)
	}
}
