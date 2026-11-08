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

package internal

import (
	"os"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/internal/db"
	"github.com/apache/airavata-custos/internal/tracing/tracingtest"
	"github.com/apache/airavata-custos/pkg/events"
	"github.com/apache/airavata-custos/pkg/models"
	coreservice "github.com/apache/airavata-custos/pkg/service"
)

// Make sure the super admin holds the signer privileges and can grant read to another user, since the signer reads these keys from that user's /me.
func TestSignerPrivileges_GrantableAndHeldBySuperAdmin(t *testing.T) {
	dsn := os.Getenv("CORE_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("integration env not set: CORE_TEST_DATABASE_DSN required")
	}
	database, err := db.Open(db.Config{DSN: dsn, MaxOpenConns: 5, MaxIdleConns: 2})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.MigrateEmbedded(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec("TRUNCATE TABLE user_roles, role_privileges, roles, user_privileges, audit_events, user_identities, users, organizations CASCADE"); err != nil {
		t.Fatal(err)
	}
	ctx := tracingtest.Context()
	bus, err := events.New(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	svc := coreservice.New(database, bus)

	if err := svc.BootstrapSuperAdmin(ctx, "admin@example.edu", "oidc"); err != nil {
		t.Fatal(err)
	}
	var adminID string
	if err := database.Get(&adminID, "SELECT id FROM users WHERE email = $1", "admin@example.edu"); err != nil {
		t.Fatal(err)
	}
	adminPrivs, err := svc.EffectivePrivileges(ctx, adminID)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(adminPrivs, CertificatesRead) || !slices.Contains(adminPrivs, CertificatesWrite) {
		t.Fatalf("super admin privileges %v lack the signer keys", adminPrivs)
	}

	operatorID := uuid.NewString()
	if _, err := database.Exec("INSERT INTO users (id, organization_id, first_name, last_name, middle_name, email, status) SELECT $1, organization_id, 'Jane', 'Doe', '', 'jdoe@example.edu', $2 FROM users WHERE id = $3", operatorID, string(models.UserActive), adminID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.GrantPrivilege(ctx, operatorID, CertificatesRead, adminID, "certificate audits"); err != nil {
		t.Fatal(err)
	}
	operatorPrivs, err := svc.EffectivePrivileges(ctx, operatorID)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(operatorPrivs, []models.PrivilegeKey{CertificatesRead}) {
		t.Fatalf("operator privileges %v, want only %s", operatorPrivs, CertificatesRead)
	}
}
