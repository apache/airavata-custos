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
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"

	ldapdb "github.com/apache/airavata-custos/connectors/LDAP/Provisioner/db"
	"github.com/apache/airavata-custos/internal/db/dbtest"
)

func openAccounts(t *testing.T) (*Accounts, *sqlx.DB) {
	t.Helper()
	database := dbtest.Open(t, dbtest.CoreDSN(t), dbtest.Migrations{FS: ldapdb.MigrationFS(), Dir: "migrations", Name: "ldap"})
	if _, err := database.Exec("TRUNCATE ldap_posix_accounts"); err != nil {
		t.Fatal(err)
	}
	return NewAccounts(database), database
}

func account(id, username string) Account {
	return Account{
		ComputeClusterUserID: id,
		Username:             username,
		HomeDirectory:        "/home/" + username,
		LoginShell:           "/bin/bash",
		EntryDN:              "uid=" + username + ",ou=people,dc=test",
	}
}

func TestAccounts_CreateNumbersAboveEveryRow(t *testing.T) {
	s, _ := openAccounts(t)
	ctx := context.Background()

	first, err := s.Create(ctx, account("cu-1", "ada"), 50000)
	if err != nil {
		t.Fatal(err)
	}
	if first.UIDNumber != 50000 || first.GIDNumber != 50000 || first.Status != StatusActive {
		t.Fatalf("first row: %+v", first)
	}
	if err := s.SetStatus(ctx, "cu-1", StatusInactive); err != nil {
		t.Fatal(err)
	}
	// The inactive row still holds 50000.
	second, err := s.Create(ctx, account("cu-2", "bob"), 50000)
	if err != nil || second.UIDNumber != 50001 {
		t.Fatalf("second row: %+v, %v", second, err)
	}
	// A higher floor wins over the rows.
	third, err := s.Create(ctx, account("cu-3", "cy"), 60000)
	if err != nil || third.UIDNumber != 60000 {
		t.Fatalf("third row: %+v, %v", third, err)
	}
}

func TestAccounts_UsernameIsUnique(t *testing.T) {
	s, _ := openAccounts(t)
	ctx := context.Background()
	if _, err := s.Create(ctx, account("cu-1", "ada"), 50000); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create(ctx, account("cu-2", "ada"), 50000); !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
	got, err := s.GetByUsername(ctx, "ada")
	if err != nil || got == nil || got.ComputeClusterUserID != "cu-1" {
		t.Fatalf("GetByUsername: %+v, %v", got, err)
	}
}

func TestAccounts_ConcurrentCreatesGetDistinctNumbers(t *testing.T) {
	s, database := openAccounts(t)
	ctx := context.Background()
	const n = 20
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.Create(ctx, account(fmt.Sprintf("cu-%d", i), fmt.Sprintf("user%d", i)), 50000); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	var distinct, lo, hi int64
	if err := database.QueryRow("SELECT COUNT(DISTINCT uid_number), MIN(uid_number), MAX(uid_number) FROM ldap_posix_accounts").Scan(&distinct, &lo, &hi); err != nil {
		t.Fatal(err)
	}
	if distinct != n || lo != 50000 || hi != 50000+n-1 {
		t.Fatalf("got %d distinct numbers from %d to %d, want %d from 50000", distinct, lo, hi, n)
	}
}

func TestAccounts_RenumberMovesAboveEveryRow(t *testing.T) {
	s, _ := openAccounts(t)
	ctx := context.Background()
	for i, name := range []string{"ada", "bob"} {
		if _, err := s.Create(ctx, account(fmt.Sprintf("cu-%d", i), name), 50000); err != nil {
			t.Fatal(err)
		}
	}
	got, err := s.Renumber(ctx, "cu-0", 50000)
	if err != nil {
		t.Fatal(err)
	}
	if got.UIDNumber != 50002 || got.GIDNumber != 50002 {
		t.Fatalf("renumbered row: %+v", got)
	}
}

func TestAccounts_GetMissingReturnsNil(t *testing.T) {
	s, _ := openAccounts(t)
	got, err := s.Get(context.Background(), "nope")
	if err != nil || got != nil {
		t.Fatalf("got %+v, %v", got, err)
	}
}
