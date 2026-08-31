// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied. See the License for the
// specific language governing permissions and limitations
// under the License.

//go:build integration
// +build integration

// Integration tests for the persistent UID sequence store. Requires a
// live MariaDB / MySQL — set LDAP_TEST_DSN, e.g.:
//
//   export LDAP_TEST_DSN='admin:admin@tcp(localhost:3306)/custos?parseTime=true'
//   go test -tags integration ./connectors/LDAP/Provisioner/internal/store/...
//
// The tests use the ldap_uid_sequence table created by
// connectors/LDAP/Provisioner/db/migrations. Run the connector once
// against the target DB (or the same migration set) before executing
// these tests.
package store

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func openTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("LDAP_TEST_DSN")
	if dsn == "" {
		t.Skip("LDAP_TEST_DSN not set; skipping integration test")
	}
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		t.Fatalf("connect DB: %v", err)
	}
	// Each test uses a fresh cluster_id so tests don't stomp on each
	// other's counters.
	return db
}

func freshClusterID() string {
	return "test-" + time.Now().Format("20060102150405.000000")
}

func TestUIDSequence_SeedThenAllocateMonotonically(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	seq := NewUIDSequence(db)
	ctx := context.Background()
	clusterID := freshClusterID()

	if err := seq.Seed(ctx, clusterID, 60000); err != nil {
		t.Fatalf("Seed: %v", err)
	}

	// Ten sequential allocations return 60000, 60001, ..., 60009.
	for want := int64(60000); want < 60010; want++ {
		got, err := seq.Allocate(ctx, clusterID)
		if err != nil {
			t.Fatalf("Allocate: %v", err)
		}
		if got != want {
			t.Fatalf("uid: got %d, want %d", got, want)
		}
	}
}

func TestUIDSequence_SeedIsIdempotent_NeverRegresses(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	seq := NewUIDSequence(db)
	ctx := context.Background()
	clusterID := freshClusterID()

	if err := seq.Seed(ctx, clusterID, 60000); err != nil {
		t.Fatalf("Seed 60000: %v", err)
	}
	// Advance the counter.
	for i := 0; i < 5; i++ {
		if _, err := seq.Allocate(ctx, clusterID); err != nil {
			t.Fatalf("Allocate: %v", err)
		}
	}
	// Peek should now show 60005.
	next, err := seq.Peek(ctx, clusterID)
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if next != 60005 {
		t.Fatalf("Peek: got %d, want 60005", next)
	}

	// Re-seed with a LOWER value (as would happen if a fresh LDAP scan
	// missed some entries). Counter must NOT regress.
	if err := seq.Seed(ctx, clusterID, 60002); err != nil {
		t.Fatalf("re-Seed lower: %v", err)
	}
	next, _ = seq.Peek(ctx, clusterID)
	if next != 60005 {
		t.Fatalf("Seed regressed counter: Peek got %d, want 60005", next)
	}

	// Re-seed with a HIGHER value should advance.
	if err := seq.Seed(ctx, clusterID, 70000); err != nil {
		t.Fatalf("re-Seed higher: %v", err)
	}
	next, _ = seq.Peek(ctx, clusterID)
	if next != 70000 {
		t.Fatalf("Seed did not advance: Peek got %d, want 70000", next)
	}
}

func TestUIDSequence_ConcurrentAllocatesReturnDistinctValues(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	seq := NewUIDSequence(db)
	ctx := context.Background()
	clusterID := freshClusterID()

	if err := seq.Seed(ctx, clusterID, 80000); err != nil {
		t.Fatalf("Seed: %v", err)
	}

	const n = 50
	got := make([]int64, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			uid, err := seq.Allocate(ctx, clusterID)
			if err != nil {
				t.Errorf("Allocate: %v", err)
				return
			}
			got[idx] = uid
		}(i)
	}
	wg.Wait()

	// Every allocation must be a distinct value in [80000, 80050).
	seen := make(map[int64]bool, n)
	for _, uid := range got {
		if uid < 80000 || uid >= 80000+n {
			t.Errorf("uid out of expected range [80000,%d): %d", 80000+n, uid)
		}
		if seen[uid] {
			t.Errorf("duplicate uid: %d", uid)
		}
		seen[uid] = true
	}
	if len(seen) != n {
		t.Errorf("expected %d distinct uids, got %d", n, len(seen))
	}
}

func TestUIDSequence_ClusterIDsAreIsolated(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	seq := NewUIDSequence(db)
	ctx := context.Background()
	clusterA := freshClusterID() + "-A"
	clusterB := freshClusterID() + "-B"

	if err := seq.Seed(ctx, clusterA, 60000); err != nil {
		t.Fatalf("Seed A: %v", err)
	}
	if err := seq.Seed(ctx, clusterB, 70000); err != nil {
		t.Fatalf("Seed B: %v", err)
	}

	uidA, _ := seq.Allocate(ctx, clusterA)
	uidB, _ := seq.Allocate(ctx, clusterB)
	if uidA != 60000 {
		t.Errorf("cluster A first uid: got %d, want 60000", uidA)
	}
	if uidB != 70000 {
		t.Errorf("cluster B first uid: got %d, want 70000", uidB)
	}

	// Advancing A must not affect B.
	for i := 0; i < 10; i++ {
		_, _ = seq.Allocate(ctx, clusterA)
	}
	nextB, _ := seq.Peek(ctx, clusterB)
	if nextB != 70001 {
		t.Errorf("cluster B counter drifted: Peek got %d, want 70001", nextB)
	}
}

func TestUIDSequence_AllocateWithoutSeedErrors(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	seq := NewUIDSequence(db)

	_, err := seq.Allocate(context.Background(), freshClusterID())
	if err == nil {
		t.Fatal("expected error when allocating without prior Seed")
	}
}
