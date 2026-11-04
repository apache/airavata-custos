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
	"net/url"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/db/dbtest"
)

// Make sure 000009 readies only the memberships the scheduler sweep grants, 000010 keeps existing work done, and down drops both.
func TestMigration000009_000010_BackfillAndDown(t *testing.T) {
	admin := dbtest.Open(t, dbtest.CoreDSN(t))
	name := "custos_upgrade_" + uuid.NewString()[:8]
	if _, err := admin.Exec("CREATE DATABASE " + name); err != nil {
		t.Fatalf("create database: %v", err)
	}
	u, err := url.Parse(dbtest.CoreDSN(t))
	if err != nil {
		t.Fatalf("parse dsn: %v", err)
	}
	u.Path = "/" + name
	dsn := u.String()
	if err := db.MigrateEmbeddedTo(dsn, 8); err != nil {
		t.Fatalf("migrate to 8: %v", err)
	}
	database, err := db.Open(db.Config{DSN: dsn, MaxOpenConns: 2, MaxIdleConns: 1})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() {
		database.Close()
		admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)")
	})
	if _, err := database.Exec(`
		INSERT INTO organizations (id, originated_id, name) VALUES ('o1', 'o1', 'Org');
		INSERT INTO users (id, organization_id, first_name, last_name, email) VALUES
			('u1', 'o1', 'A', 'B', 'u1@example.edu'), ('u2', 'o1', 'C', 'D', 'u2@example.edu');
		INSERT INTO compute_clusters (id, name) VALUES ('c1', 'cluster');
		INSERT INTO compute_cluster_users (id, compute_cluster_id, user_id, local_username, provisioned_at) VALUES
			('cu1', 'c1', 'u1', 'ua', now()), ('cu2', 'c1', 'u2', 'ub', NULL);
		INSERT INTO projects (id, originated_id, title, origination, project_pi_id) VALUES ('p1', 'p1', 'P', 'test', 'u1');
		INSERT INTO compute_allocations (id, project_id, name, status, compute_cluster_id, start_time, end_time) VALUES
			('a1', 'p1', 'one', 'ACTIVE', 'c1', now() - interval '1 day', now() + interval '1 year'),
			('a2', 'p1', 'two', 'ACTIVE', 'c1', now() - interval '1 day', now() - interval '1 hour'),
			('a3', 'p1', 'three', 'ACTIVE', 'c1', now() - interval '1 day', now() + interval '1 year');
		INSERT INTO compute_allocation_memberships (id, compute_allocation_id, user_id, start_time, end_time, membership_status) VALUES
			('m1', 'a1', 'u1', now(), now() + interval '1 year', 'ACTIVE'),
			('m2', 'a1', 'u2', now(), now() + interval '1 year', 'ACTIVE'),
			('m3', 'a2', 'u1', now(), now() + interval '1 year', 'ACTIVE'),
			('m4', 'a3', 'u1', now(), now() + interval '1 year', 'INACTIVE');`); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := db.MigrateEmbeddedTo(dsn, 9); err != nil {
		t.Fatalf("up: %v", err)
	}
	var ready []string
	if err := database.Select(&ready, `SELECT id FROM compute_allocation_memberships WHERE ready_at IS NOT NULL ORDER BY id`); err != nil ||
		!slices.Equal(ready, []string{"m1", "m3"}) {
		t.Fatalf("ready %v, %v, want [m1 m3]", ready, err)
	}
	if err := db.MigrateEmbeddedTo(dsn, 10); err != nil {
		t.Fatalf("up 10: %v", err)
	}
	var pending, done []string
	if err := database.Select(&pending, `SELECT id FROM compute_allocations WHERE scheduler_provisioned_at IS NULL`); err != nil || len(pending) != 0 {
		t.Fatalf("allocations without a scheduler mark %v, %v, want none", pending, err)
	}
	if err := database.Select(&done, `SELECT id FROM compute_cluster_users WHERE storage_provisioned_at IS NOT NULL AND ready_at IS NOT NULL`); err != nil ||
		!slices.Equal(done, []string{"cu1"}) {
		t.Fatalf("ready accounts %v, %v, want [cu1]", done, err)
	}

	if err := db.MigrateEmbeddedTo(dsn, 8); err != nil {
		t.Fatalf("down: %v", err)
	}
	var n int
	if err := database.Get(&n, `SELECT count(*) FROM information_schema.columns
		WHERE column_name IN ('ready_at', 'storage_provisioned_at', 'scheduler_provisioned_at')`); err != nil || n != 0 {
		t.Fatalf("added columns after down = %d, %v, want 0", n, err)
	}
}
