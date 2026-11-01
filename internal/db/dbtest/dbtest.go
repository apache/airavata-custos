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
// They truncate what they touch, so they must run with -p 1 against a database reserved for tests.
package dbtest

import (
	"io/fs"
	"os"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/internal/db"
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
