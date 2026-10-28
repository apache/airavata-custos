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

// Package amietest opens the integration test database for AMIE connector tests, with the core and AMIE
// migrations applied. Tests skip unless the environment that scripts/run-amie-integration-tests.sh exports
// is set.
package amietest

import (
	"os"
	"testing"

	"github.com/jmoiron/sqlx"

	amiedb "github.com/apache/airavata-custos/connectors/ACCESS/AMIE-Processor/db"
	"github.com/apache/airavata-custos/internal/db/dbtest"
)

const ClusterID = "00000000-0000-0000-0000-000000000001"

// Tables lists every table an AMIE packet can write, children first.
var Tables = []string{
	"amie_audit_extras",
	"audit_events",
	"event_deliveries",
	"events",
	"event_subscriptions",
	"amie_processing_errors",
	"amie_processing_events",
	"amie_packets",
	"amie_user_dns",
	"compute_allocation_membership_resource_overrides",
	"compute_allocation_usages",
	"project_memberships",
	"compute_allocation_memberships",
	"compute_allocation_change_request_events",
	"compute_allocation_change_requests",
	"compute_allocation_diffs",
	"compute_allocation_resource_rates",
	"compute_allocation_resource_mappings",
	"compute_allocation_resources",
	"compute_allocations",
	"compute_cluster_users",
	"compute_clusters",
	"projects",
	"user_identities",
	"users",
	"organizations",
}

// OpenDB returns the shared test database with tables truncated.
func OpenDB(t *testing.T, tables ...string) *sqlx.DB {
	t.Helper()
	for _, key := range []string{"DATABASE_DSN", "AMIE_BASE_URL", "AMIE_SITE_CODE", "AMIE_API_KEY", "AMIE_CLUSTER_ID"} {
		if os.Getenv(key) == "" {
			t.Skip("integration env not set; run via scripts/run-amie-integration-tests.sh")
		}
	}
	database := dbtest.Open(t, os.Getenv("DATABASE_DSN"),
		dbtest.Migrations{FS: amiedb.MigrationFS(), Dir: "migrations", Name: "amie"})
	dbtest.Truncate(t, database, tables...)
	return database
}

// SeedCluster inserts the cluster ClusterID names; handlers refuse to run without it.
func SeedCluster(t *testing.T, database *sqlx.DB) {
	t.Helper()
	if _, err := database.Exec(
		"INSERT INTO compute_clusters (id, name) VALUES ($1, $2)",
		ClusterID, "default-cluster",
	); err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
}
