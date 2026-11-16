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

package store

import (
	"context"
	"slices"
	"testing"
)

// seedDirectory adds one organization, three users with cluster accounts, a role, an identity and two projects.
func seedDirectory(t *testing.T) {
	t.Helper()
	database := setupTestDB(t)
	if _, err := database.Exec(`TRUNCATE TABLE compute_cluster_users, compute_clusters, projects, user_roles, roles, user_identities, users, organizations CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	for _, stmt := range []string{
		`INSERT INTO organizations (id, originated_id, name) VALUES ('org-gt', 'GT-01', 'Georgia Institute of Technology')`,
		`INSERT INTO users (id, organization_id, first_name, last_name, email) VALUES
			('u-jane', 'org-gt', 'Jane', 'Doe', 'jdoe@example.edu'),
			('u-john', 'org-gt', 'John', 'Roe', 'jroe@example.edu'),
			('u-mary', 'org-gt', 'Mary', 'Major', 'mmajor@example.edu')`,
		`INSERT INTO roles (id, name) VALUES ('role-ops', 'Operator')`,
		`INSERT INTO user_roles (user_id, role_id) VALUES ('u-john', 'role-ops')`,
		`INSERT INTO user_identities (id, user_id, source, external_id) VALUES ('ui-1', 'u-mary', 'access', 'mmajor')`,
		`INSERT INTO projects (id, originated_id, title, origination, project_pi_id) VALUES
			('p-access', 'BIO-130000', 'Genomics', 'ACCESS', 'u-jane'),
			('4f2c9e1a-0000-0000-0000-000000000000', '', 'Sandbox', 'internal', 'u-jane')`,
		`INSERT INTO compute_clusters (id, name) VALUES ('cl-a', 'Cluster A')`,
		`INSERT INTO compute_cluster_users (id, compute_cluster_id, user_id, local_username) VALUES
			('cu-jane', 'cl-a', 'u-jane', 'jdoe'),
			('cu-john', 'cl-a', 'u-john', 'jroe'),
			('cu-mary', 'cl-a', 'u-mary', 'mmajor')`,
	} {
		if _, err := database.Exec(stmt); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func userEmails(t *testing.T, filter UserListFilter) []string {
	t.Helper()
	rows, total, err := NewUserStore(setupTestDB(t)).List(context.Background(), filter, 50, 0)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if total != len(rows) {
		t.Fatalf("total %d, rows %d", total, len(rows))
	}
	emails := make([]string, len(rows))
	for i, u := range rows {
		emails[i] = u.Email
	}
	return emails
}

// The user list matches a full name and narrows by role and identity source across every page.
func TestUserListFilters(t *testing.T) {
	seedDirectory(t)
	cases := []struct {
		name   string
		filter UserListFilter
		want   []string
	}{
		{"full name", UserListFilter{Query: "jane doe"}, []string{"jdoe@example.edu"}},
		{"email", UserListFilter{Query: "JROE"}, []string{"jroe@example.edu"}},
		{"role", UserListFilter{RoleID: "role-ops"}, []string{"jroe@example.edu"}},
		{"identity source", UserListFilter{IdentitySource: "access"}, []string{"mmajor@example.edu"}},
		{"role and query", UserListFilter{RoleID: "role-ops", Query: "jane"}, []string{}},
		{"none", UserListFilter{}, []string{"jdoe@example.edu", "jroe@example.edu", "mmajor@example.edu"}},
	}
	for _, c := range cases {
		got := userEmails(t, c.filter)
		if !slices.Equal(got, c.want) {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

// A project search matches the ref the portal shows, for originated and internal projects.
func TestProjectListMatchesRef(t *testing.T) {
	seedDirectory(t)
	projects := NewProjectStore(setupTestDB(t))
	for query, want := range map[string]string{
		"access-bio13":      "p-access",
		"ACCESS-BIO130000":  "p-access",
		"internal-4f2c9e1a": "4f2c9e1a-0000-0000-0000-000000000000",
	} {
		rows, _, err := projects.ListWithPI(context.Background(), ProjectListFilter{Query: query, Limit: 50})
		if err != nil {
			t.Fatalf("list projects: %v", err)
		}
		if len(rows) != 1 || rows[0].ID != want {
			t.Errorf("%q: got %d rows, want %s", query, len(rows), want)
		}
	}
}

// The project list narrows by origination regardless of case.
func TestProjectListOrigination(t *testing.T) {
	seedDirectory(t)
	rows, total, err := NewProjectStore(setupTestDB(t)).ListWithPI(context.Background(), ProjectListFilter{Origination: "access", Limit: 50})
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if total != 1 || len(rows) != 1 || rows[0].ID != "p-access" {
		t.Errorf("total %d, rows %d, want p-access only", total, len(rows))
	}
}

// The cluster account list narrows by role and identity source like the user list.
func TestClusterAccountListFilters(t *testing.T) {
	seedDirectory(t)
	accounts := NewComputeClusterUserStore(setupTestDB(t))
	for _, c := range []struct {
		filter UserListFilter
		want   string
	}{
		{UserListFilter{RoleID: "role-ops"}, "cu-john"},
		{UserListFilter{IdentitySource: "access"}, "cu-mary"},
		{UserListFilter{Query: "jdoe"}, "cu-jane"},
	} {
		rows, total, err := accounts.ListByApprovalStatus(context.Background(), "", c.filter, 50, 0)
		if err != nil {
			t.Fatalf("list cluster accounts: %v", err)
		}
		if total != 1 || len(rows) != 1 || rows[0].ID != c.want {
			t.Errorf("%+v: total %d, rows %v, want %s", c.filter, total, rows, c.want)
		}
	}
}

// The delivery list narrows by connector.
func TestDeliveryListFilters(t *testing.T) {
	database := setupTestDB(t)
	for _, stmt := range []string{
		`TRUNCATE TABLE event_deliveries, events CASCADE`,
		`INSERT INTO events (id, event_type, payload, source, trace_id, span_id) VALUES
			('ev-1', 'user.created', '{}', 'core', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa', 'aaaaaaaaaaaaaaaa'),
			('ev-2', 'project.created', '{}', 'amie', 'bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb', 'bbbbbbbbbbbbbbbb')`,
		`INSERT INTO event_deliveries (id, event_id, subscriber, status, next_run_at, last_error) VALUES
			('d-1', 'ev-1', 'slurm', 'FAILED', now(), 'ssh: connection refused'),
			('d-2', 'ev-2', 'slurm', 'SUCCEEDED', now(), NULL),
			('d-3', 'ev-2', 'comanage', 'PENDING', now(), NULL)`,
	} {
		if _, err := database.Exec(stmt); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
	deliveries := NewEventDeliveryStore(database)
	for subscriber, want := range map[string]int{"slurm": 2, "comanage": 1, "": 3} {
		rows, err := deliveries.ListDeliveries(context.Background(), "", subscriber, 50)
		if err != nil {
			t.Fatalf("list deliveries: %v", err)
		}
		if len(rows) != want {
			t.Errorf("subscriber %q: got %d rows, want %d", subscriber, len(rows), want)
		}
	}
}
