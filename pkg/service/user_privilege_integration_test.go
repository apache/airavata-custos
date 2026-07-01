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
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/apache/airavata-custos/pkg/models"
)

func TestGrantPrivilege_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)

	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegesGrant(t, database, granter)

	grant, err := svc.GrantPrivilege(ctx(), target, models.ClustersRead, granter, "ops view")
	if err != nil {
		t.Fatalf("GrantPrivilege: %v", err)
	}
	if grant.UserID != target || grant.Privilege != models.ClustersRead {
		t.Errorf("grant payload mismatch: %+v", grant)
	}

	has, err := svc.HasPrivilege(ctx(), target, models.ClustersRead)
	if err != nil || !has {
		t.Errorf("HasPrivilege after grant: has=%v err=%v", has, err)
	}
	if got := countAuditEventsOfType(t, database, "PRIVILEGE_GRANTED", target); got != 1 {
		t.Errorf("audit PRIVILEGE_GRANTED: got %d, want 1", got)
	}
}

func TestGrantPrivilege_RejectsGranterWithoutMeta(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	// granter has NO privileges:grant.
	_, err := svc.GrantPrivilege(ctx(), target, models.ClustersRead, granter, "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("granter w/o meta: got err=%v, want ErrInvalidInput", err)
	}
	if !strings.Contains(err.Error(), string(models.PrivilegesGrant)) {
		t.Errorf("error should mention the missing privilege, got: %v", err)
	}
}

func TestGrantPrivilege_RejectsUnknownPrivilege(t *testing.T) {
	svc := newTestService(setupTestDB(t))
	_, err := svc.GrantPrivilege(ctx(), "u", "bogus:thing", "g", "")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("unknown privilege: got err=%v, want ErrInvalidInput", err)
	}
}

func TestGrantPrivilege_DuplicateActiveReturnsAlreadyExists(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegesGrant(t, database, granter)

	if _, err := svc.GrantPrivilege(ctx(), target, models.ClustersRead, granter, ""); err != nil {
		t.Fatalf("first grant: %v", err)
	}
	_, err := svc.GrantPrivilege(ctx(), target, models.ClustersRead, granter, "")
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("duplicate grant: got err=%v, want ErrAlreadyExists", err)
	}
}

func TestRevokePrivilege_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegesGrant(t, database, granter)
	if _, err := svc.GrantPrivilege(ctx(), target, models.ClustersRead, granter, ""); err != nil {
		t.Fatalf("grant: %v", err)
	}
	if err := svc.RevokePrivilege(ctx(), target, models.ClustersRead, granter, "no longer needed"); err != nil {
		t.Fatalf("RevokePrivilege: %v", err)
	}
	if has, err := svc.HasPrivilege(ctx(), target, models.ClustersRead); err != nil || has {
		t.Errorf("HasPrivilege after revoke: has=%v err=%v", has, err)
	}
	if got := countAuditEventsOfType(t, database, "PRIVILEGE_REVOKED", target); got != 1 {
		t.Errorf("audit PRIVILEGE_REVOKED: got %d, want 1", got)
	}
}

func TestRevokePrivilege_RejectsSelfRevokeOfMeta(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	caller := seedUser(t, database, "caller@example.edu")
	seedPrivilegesGrant(t, database, caller)

	err := svc.RevokePrivilege(ctx(), caller, models.PrivilegesGrant, caller, "self")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("self-revoke meta: got err=%v, want ErrInvalidInput", err)
	}
}

func TestRevokePrivilege_NoActiveGrant_ReturnsNotFound(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegesGrant(t, database, granter)
	err := svc.RevokePrivilege(ctx(), target, models.ClustersRead, granter, "")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("revoke with no active grant: got err=%v, want ErrNotFound", err)
	}
}

func TestHasPrivilege_ReturnsFalseForUngranted(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	user := seedUser(t, database, "u@example.edu")
	has, err := svc.HasPrivilege(ctx(), user, models.ClustersRead)
	if err != nil {
		t.Fatalf("HasPrivilege: %v", err)
	}
	if has {
		t.Errorf("HasPrivilege for ungranted user: got true, want false")
	}
}

func TestListUserPrivileges_ReturnsActiveOnly(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilegesGrant(t, database, granter)
	if _, err := svc.GrantPrivilege(ctx(), target, models.ClustersWrite, granter, ""); err != nil {
		t.Fatalf("grant: %v", err)
	}
	if _, err := svc.GrantPrivilege(ctx(), target, models.ClustersRead, granter, ""); err != nil {
		t.Fatalf("grant 2: %v", err)
	}
	if err := svc.RevokePrivilege(ctx(), target, models.ClustersRead, granter, ""); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	rows, err := svc.ListUserPrivileges(ctx(), target)
	if err != nil {
		t.Fatalf("ListUserPrivileges: %v", err)
	}
	if len(rows) != 1 || rows[0].Privilege != models.ClustersWrite {
		t.Errorf("active set: got %v, want [core:clusters:write]", rows)
	}
}

func TestListPrivilegeHolders(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	granter := seedUser(t, database, "granter@example.edu")
	a := seedUser(t, database, "a@example.edu")
	b := seedUser(t, database, "b@example.edu")
	seedPrivilegesGrant(t, database, granter)
	if _, err := svc.GrantPrivilege(ctx(), a, models.ClustersRead, granter, ""); err != nil {
		t.Fatalf("grant a: %v", err)
	}
	if _, err := svc.GrantPrivilege(ctx(), b, models.ClustersRead, granter, ""); err != nil {
		t.Fatalf("grant b: %v", err)
	}
	rows, err := svc.ListPrivilegeHolders(ctx(), models.ClustersRead)
	if err != nil {
		t.Fatalf("ListPrivilegeHolders: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("holders: got %d, want 2", len(rows))
	}
}

func TestPrivilegeCatalog(t *testing.T) {
	svc := newTestService(setupTestDB(t))
	cat := svc.PrivilegeCatalog()
	seen := map[models.PrivilegeKey]bool{}
	for _, k := range cat {
		seen[k] = true
	}
	if !seen[models.PrivilegesGrant] || !seen[models.ClustersRead] {
		t.Errorf("catalog missing expected keys, got: %v", cat)
	}
}

func TestBootstrapSuperAdmin_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	user := seedUser(t, database, "boot@example.edu")
	if err := svc.BootstrapSuperAdmin(ctx(), "boot@example.edu", "env:TEST"); err != nil {
		t.Fatalf("BootstrapSuperAdmin: %v", err)
	}
	if has, err := svc.HasPrivilege(ctx(), user, models.PrivilegesGrant); err != nil || !has {
		t.Errorf("HasPrivilege grant after bootstrap: has=%v err=%v", has, err)
	}
	if has, err := svc.HasPrivilege(ctx(), user, models.RolesManage); err != nil || !has {
		t.Errorf("HasPrivilege roles:manage after bootstrap: has=%v err=%v", has, err)
	}
	if got := countAuditEventsOfType(t, database, "ROLE_BOOTSTRAPPED", user); got != 1 {
		t.Errorf("audit ROLE_BOOTSTRAPPED: got %d, want 1", got)
	}
}

func TestBootstrapSuperAdmin_NoOpWhenRoleHasHolder(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	first := seedUser(t, database, "first@example.edu")
	another := seedUser(t, database, "another@example.edu")
	_ = another
	if err := svc.BootstrapSuperAdmin(ctx(), "first@example.edu", "env:TEST"); err != nil {
		t.Fatalf("first bootstrap: %v", err)
	}
	if err := svc.BootstrapSuperAdmin(ctx(), "another@example.edu", "env:TEST"); err != nil {
		t.Fatalf("second bootstrap: %v", err)
	}
	// only `first` should hold super_admin
	if has, err := svc.HasPrivilege(ctx(), first, models.PrivilegesGrant); err != nil || !has {
		t.Errorf("first should still hold grant: has=%v err=%v", has, err)
	}
	if has, err := svc.HasPrivilege(ctx(), another, models.PrivilegesGrant); err != nil || has {
		t.Errorf("another should NOT hold grant: has=%v err=%v", has, err)
	}
}

func TestBootstrapSuperAdmin_NoOpWhenEmailNotFound(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	if err := svc.BootstrapSuperAdmin(ctx(), "missing@example.edu", "env:TEST"); err != nil {
		t.Fatalf("BootstrapSuperAdmin: %v", err)
	}
	rows, err := svc.ListPrivilegeHolders(ctx(), models.PrivilegesGrant)
	if err != nil {
		t.Fatalf("ListPrivilegeHolders: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("holders after missing-email bootstrap: got %d, want 0", len(rows))
	}
}

func TestSchema_RejectsTwoActivePrivilegesOfSameKey(t *testing.T) {
	database := setupTestDB(t)
	user := seedUser(t, database, "u@example.edu")
	if _, err := database.Exec(
		`INSERT INTO user_privileges (id, user_id, privilege, granted_at) VALUES (?, ?, ?, NOW(6))`,
		uuid.NewString(), user, string(models.ClustersRead),
	); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	_, err := database.Exec(
		`INSERT INTO user_privileges (id, user_id, privilege, granted_at) VALUES (?, ?, ?, NOW(6))`,
		uuid.NewString(), user, string(models.ClustersRead),
	)
	if err == nil {
		t.Fatal("expected UNIQUE violation, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "duplicate") &&
		!strings.Contains(strings.ToLower(err.Error()), "unique") {
		t.Errorf("expected duplicate/unique error, got: %v", err)
	}
}
