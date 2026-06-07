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

	"github.com/apache/airavata-custos/pkg/models"
)

func TestGrantRoleToUser_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilege(t, database, granter, models.PrivilegeRolesManage)
	role, _ := svc.CreateRole(ctx(), "amie-viewer", "", granter)
	_ = svc.AddPrivilegeToRole(ctx(), role.ID, models.PrivilegeAMIERead, granter)

	if _, err := svc.GrantRoleToUser(ctx(), target, role.ID, granter, "ops"); err != nil {
		t.Fatalf("GrantRoleToUser: %v", err)
	}
	// target should now have amie:read via the role
	if has, err := svc.HasPrivilege(ctx(), target, models.PrivilegeAMIERead); err != nil || !has {
		t.Errorf("HasPrivilege amie:read via role: has=%v err=%v", has, err)
	}
}

func TestGrantRoleToUser_RejectsDuplicate(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilege(t, database, granter, models.PrivilegeRolesManage)
	role, _ := svc.CreateRole(ctx(), "viewer", "", granter)

	if _, err := svc.GrantRoleToUser(ctx(), target, role.ID, granter, ""); err != nil {
		t.Fatalf("first: %v", err)
	}
	_, err := svc.GrantRoleToUser(ctx(), target, role.ID, granter, "")
	if !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestRevokeRoleFromUser_HappyPath(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilege(t, database, granter, models.PrivilegeRolesManage)
	role, _ := svc.CreateRole(ctx(), "viewer", "", granter)
	_ = svc.AddPrivilegeToRole(ctx(), role.ID, models.PrivilegeAMIERead, granter)
	if _, err := svc.GrantRoleToUser(ctx(), target, role.ID, granter, ""); err != nil {
		t.Fatalf("grant: %v", err)
	}
	if err := svc.RevokeRoleFromUser(ctx(), target, role.ID, granter, "rotated"); err != nil {
		t.Fatalf("RevokeRoleFromUser: %v", err)
	}
	if has, err := svc.HasPrivilege(ctx(), target, models.PrivilegeAMIERead); err != nil || has {
		t.Errorf("HasPrivilege after revoke: has=%v err=%v", has, err)
	}
}

func TestRevokeRoleFromUser_RejectsLastMetaHolder(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	boot := seedUser(t, database, "boot@example.edu")
	if err := svc.BootstrapSuperAdmin(ctx(), "boot@example.edu", "test"); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	roles, _ := svc.ListRoles(ctx())
	var superAdmin *models.Role
	for i := range roles {
		if roles[i].Name == models.SystemRoleSuperAdmin {
			superAdmin = &roles[i]
		}
	}
	// boot is the only holder of super_admin. To call revoke we need an actor
	// with roles:manage. Use boot themselves — but revoking boot's own role
	// would leave 0 holders of meta-privileges, so the guard should refuse.
	err := svc.RevokeRoleFromUser(ctx(), boot, superAdmin.ID, boot, "self")
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput (last meta holder), got %v", err)
	}
}

func TestHasPrivilege_UnionsDirectAndRole(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilege(t, database, granter, models.PrivilegeRolesManage)
	seedPrivilege(t, database, granter, models.PrivilegeGrant)

	// Grant direct amie:read
	if _, err := svc.GrantPrivilege(ctx(), target, models.PrivilegeAMIERead, granter, ""); err != nil {
		t.Fatalf("direct grant: %v", err)
	}
	// Grant role with hpc:read
	role, _ := svc.CreateRole(ctx(), "hpc-viewer", "", granter)
	_ = svc.AddPrivilegeToRole(ctx(), role.ID, models.PrivilegeHPCRead, granter)
	if _, err := svc.GrantRoleToUser(ctx(), target, role.ID, granter, ""); err != nil {
		t.Fatalf("role grant: %v", err)
	}
	// Target should have BOTH amie:read (direct) and hpc:read (via role)
	for _, key := range []models.PrivilegeKey{models.PrivilegeAMIERead, models.PrivilegeHPCRead} {
		has, err := svc.HasPrivilege(ctx(), target, key)
		if err != nil || !has {
			t.Errorf("HasPrivilege %s: has=%v err=%v", key, has, err)
		}
	}
	// Effective should include both
	keys, err := svc.EffectivePrivileges(ctx(), target)
	if err != nil {
		t.Fatalf("EffectivePrivileges: %v", err)
	}
	seen := map[models.PrivilegeKey]bool{}
	for _, k := range keys {
		seen[k] = true
	}
	if !seen[models.PrivilegeAMIERead] || !seen[models.PrivilegeHPCRead] {
		t.Errorf("effective set missing keys: %v", keys)
	}
}

func TestUpdateRolePrivileges_PropagatesToHolders(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	granter := seedUser(t, database, "granter@example.edu")
	target := seedUser(t, database, "target@example.edu")
	seedPrivilege(t, database, granter, models.PrivilegeRolesManage)
	role, _ := svc.CreateRole(ctx(), "evolving", "", granter)
	if _, err := svc.GrantRoleToUser(ctx(), target, role.ID, granter, ""); err != nil {
		t.Fatalf("grant role: %v", err)
	}
	// Target holds the role but no privilege yet.
	if has, err := svc.HasPrivilege(ctx(), target, models.PrivilegeAMIERead); err != nil || has {
		t.Errorf("pre-add HasPrivilege: has=%v err=%v", has, err)
	}
	// Add a privilege to the role — target should gain it transparently.
	if err := svc.AddPrivilegeToRole(ctx(), role.ID, models.PrivilegeAMIERead, granter); err != nil {
		t.Fatalf("add: %v", err)
	}
	if has, err := svc.HasPrivilege(ctx(), target, models.PrivilegeAMIERead); err != nil || !has {
		t.Errorf("post-add HasPrivilege: has=%v err=%v", has, err)
	}
	// Remove it — target loses it transparently.
	if err := svc.RemovePrivilegeFromRole(ctx(), role.ID, models.PrivilegeAMIERead, granter); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if has, err := svc.HasPrivilege(ctx(), target, models.PrivilegeAMIERead); err != nil || has {
		t.Errorf("post-remove HasPrivilege: has=%v err=%v", has, err)
	}
}

func TestRolesManageRequiredForRoleOps(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	// Actor has privileges:grant but NOT roles:manage — role ops must refuse.
	actor := seedUser(t, database, "actor@example.edu")
	seedPrivilege(t, database, actor, models.PrivilegeGrant)

	_, err := svc.CreateRole(ctx(), "x", "", actor)
	if !errors.Is(err, ErrInvalidInput) {
		t.Errorf("CreateRole without roles:manage: got %v, want ErrInvalidInput", err)
	}
}
