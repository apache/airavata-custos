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
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/models"
)

func TestCreateRole_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	actor := seedUser(t, database, "actor@example.edu")
	seedPrivilege(t, database, actor, models.RolesManage)

	role, err := svc.CreateRole(ctx(), "operator", "view AMIE + HPC", actor)
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if role.Name != "operator" || role.IsSystem {
		t.Errorf("role: %+v", role)
	}
	if got := countAuditEventsOfType(t, database, "ROLE_CREATED", role.ID); got != 1 {
		t.Errorf("audit ROLE_CREATED: got %d, want 1", got)
	}
}

func TestCreateRole_RejectsWithoutRolesManage(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	actor := seedUser(t, database, "actor@example.edu")
	// actor has nothing
	_, err := svc.CreateRole(ctx(), "operator", "", actor)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateRole_RejectsDuplicateName(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	actor := seedUser(t, database, "actor@example.edu")
	seedPrivilege(t, database, actor, models.RolesManage)
	if _, err := svc.CreateRole(ctx(), "operator", "", actor); err != nil {
		t.Fatalf("first: %v", err)
	}
	_, err := svc.CreateRole(ctx(), "operator", "", actor)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestAddPrivilegeToRole_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	actor := seedUser(t, database, "actor@example.edu")
	seedPrivilege(t, database, actor, models.RolesManage)
	role, err := svc.CreateRole(ctx(), "amie-viewer", "", actor)
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}

	if err := svc.AddPrivilegeToRole(ctx(), role.ID, models.ClustersRead, actor); err != nil {
		t.Fatalf("AddPrivilegeToRole: %v", err)
	}
	keys, err := svc.ListRolePrivileges(ctx(), role.ID)
	if err != nil {
		t.Fatalf("ListRolePrivileges: %v", err)
	}
	if len(keys) != 1 || keys[0] != models.ClustersRead {
		t.Errorf("role privileges: got %v, update [core:clusters:read]", keys)
	}
	if got := countAuditEventsOfType(t, database, "ROLE_PRIVILEGE_ADDED", role.ID); got != 1 {
		t.Errorf("audit ROLE_PRIVILEGE_ADDED: got %d, want 1", got)
	}
}

func TestAddPrivilegeToRole_RejectsDuplicate(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	actor := seedUser(t, database, "actor@example.edu")
	seedPrivilege(t, database, actor, models.RolesManage)
	role, _ := svc.CreateRole(ctx(), "amie-viewer", "", actor)
	if err := svc.AddPrivilegeToRole(ctx(), role.ID, models.ClustersRead, actor); err != nil {
		t.Fatalf("first add: %v", err)
	}
	err := svc.AddPrivilegeToRole(ctx(), role.ID, models.ClustersRead, actor)
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestRemovePrivilegeFromRole_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	actor := seedUser(t, database, "actor@example.edu")
	seedPrivilege(t, database, actor, models.RolesManage)
	role, _ := svc.CreateRole(ctx(), "amie-viewer", "", actor)
	_ = svc.AddPrivilegeToRole(ctx(), role.ID, models.ClustersRead, actor)

	if err := svc.RemovePrivilegeFromRole(ctx(), role.ID, models.ClustersRead, actor); err != nil {
		t.Fatalf("RemovePrivilegeFromRole: %v", err)
	}
	keys, _ := svc.ListRolePrivileges(ctx(), role.ID)
	if len(keys) != 0 {
		t.Errorf("role privileges after remove: %v", keys)
	}
}

func TestRemovePrivilegeFromRole_RejectsRemovingLastMetaSource(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	actor := seedUser(t, database, "actor@example.edu")
	seedPrivilege(t, database, actor, models.RolesManage)
	role, _ := svc.CreateRole(ctx(), "controllers", "", actor)
	_ = svc.AddPrivilegeToRole(ctx(), role.ID, models.PrivilegesGrant, actor)
	// `actor`'s privileges:grant is via direct seed; the role we just added it
	// to is the ONLY role carrying it. Removing from the role would leave 0
	// roles carrying privileges:grant — guard refuses.
	err := svc.RemovePrivilegeFromRole(ctx(), role.ID, models.PrivilegesGrant, actor)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput (last source), got %v", err)
	}
}

func TestDeleteRole_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	actor := seedUser(t, database, "actor@example.edu")
	seedPrivilege(t, database, actor, models.RolesManage)
	role, _ := svc.CreateRole(ctx(), "tmp", "", actor)
	if err := svc.DeleteRole(ctx(), role.ID, actor); err != nil {
		t.Fatalf("DeleteRole: %v", err)
	}
	if _, err := svc.GetRole(ctx(), role.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteRole_RefusesSystemRole(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	bootstrap := seedUser(t, database, "boot@example.edu")
	if err := svc.BootstrapSuperAdmin(ctx(), "boot@example.edu", "test"); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	roles, _ := svc.ListRoles(ctx())
	var superAdmin *models.Role
	for i := range roles {
		if roles[i].Name == models.SystemRoleSuperAdmin {
			superAdmin = &roles[i]
			break
		}
	}
	if superAdmin == nil {
		t.Fatal("super_admin role not found")
	}
	err := svc.DeleteRole(ctx(), superAdmin.ID, bootstrap)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput (system role), got %v", err)
	}
}

// seedPrivilege directly inserts an active grant for any privilege key.
func seedPrivilege(t *testing.T, database *sqlx.DB, userID string, p models.PrivilegeKey) {
	t.Helper()
	if _, err := database.Exec(
		`INSERT INTO user_privileges (id, user_id, privilege, granted_at, reason)
		 VALUES (?, ?, ?, NOW(6), 'seed')`,
		uuid.NewString(), userID, string(p),
	); err != nil {
		t.Fatalf("seed privilege %s for %s: %v", p, userID, err)
	}
}
