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
	"database/sql"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

const testServiceAccountPrivilege models.PrivilegeKey = "test:items:create"

// testServiceAccountPrivileges registers the test key as allowed for service accounts and returns it as an account's privileges.
func testServiceAccountPrivileges() []models.PrivilegeKey {
	models.AllowForServiceAccounts(testServiceAccountPrivilege)
	return []models.PrivilegeKey{testServiceAccountPrivilege}
}

func seedServiceAccountAdmin(t *testing.T, database *sqlx.DB) string {
	t.Helper()
	userID := seedUser(t, database, "service-account-admin@example.edu")
	if _, err := database.Exec(
		`INSERT INTO user_privileges (id, user_id, privilege, granted_at, reason) VALUES ($1, $2, $3, NOW(), 'seed')`,
		uuid.NewString(), userID, string(models.ServiceAccountsManage),
	); err != nil {
		t.Fatalf("seed service-accounts:manage: %v", err)
	}
	return userID
}

func newTestServiceAccount(t *testing.T, svc *Service, adminID string) (*models.ServiceAccount, string) {
	t.Helper()
	account, token, err := svc.CreateServiceAccount(ctx(), "test-gateway", "Test Gateway", "", testServiceAccountPrivileges(), adminID)
	if err != nil {
		t.Fatalf("create service account: %v", err)
	}
	return account, token
}

// Make sure the token from creation finds the service account with its privileges, and the secret isn't stored.
func TestServiceAccount_TokenAuthenticatesAndSecretIsNotStored(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	account, token := newTestServiceAccount(t, svc, seedServiceAccountAdmin(t, database))

	resolved, err := svc.ResolveServiceAccount(ctx(), token)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resolved.ID != account.ID || !slices.Equal(resolved.Privileges, []models.PrivilegeKey{testServiceAccountPrivilege}) {
		t.Fatalf("service account = %+v", resolved)
	}
	_, secret, _ := strings.Cut(token, ".")
	var stored int
	if err := database.Get(&stored, `SELECT COUNT(*) FROM service_accounts WHERE secret_hash = $1 OR secret_hint = $1`, secret); err != nil {
		t.Fatal(err)
	}
	if stored != 0 {
		t.Fatal("the secret was stored")
	}
}

// Make sure a service account no longer gets a stored privilege once that key is not allowed for service accounts.
func TestServiceAccount_DropsPrivilegesNoLongerAllowed(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	account, token := newTestServiceAccount(t, svc, seedServiceAccountAdmin(t, database))
	// Stands in for a key that a later release took off AllowForServiceAccounts.
	if _, err := database.Exec(`INSERT INTO service_account_privileges (service_account_id, privilege) VALUES ($1, $2)`, account.ID, models.UsersWrite); err != nil {
		t.Fatal(err)
	}

	resolved, err := svc.ResolveServiceAccount(ctx(), token)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(resolved.Privileges, []models.PrivilegeKey{testServiceAccountPrivilege}) {
		t.Fatalf("privileges = %v", resolved.Privileges)
	}
}

// Make sure a token with a wrong or missing secret, or a malformed key, is refused.
func TestServiceAccount_BadTokenIsRefused(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	newTestServiceAccount(t, svc, seedServiceAccountAdmin(t, database))

	prefix := identity.ServiceAccountTokenPrefix
	for _, token := range []string{prefix + "test-gateway.wrong", prefix + "test-gateway", prefix + "\xff.secret"} {
		if _, err := svc.ResolveServiceAccount(ctx(), token); !errors.Is(err, identity.ErrInvalidServiceAccountToken) {
			t.Errorf("%s: got %v", token, err)
		}
	}
}

// Make sure a key or name that is already taken is reported as a conflict.
func TestServiceAccount_CreateRefusesDuplicateKeyAndName(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedServiceAccountAdmin(t, database)
	newTestServiceAccount(t, svc, adminID)

	for _, account := range []struct{ key, name string }{
		{"test-gateway", "Another Gateway"},
		{"another-gateway", "Test Gateway"},
	} {
		if _, _, err := svc.CreateServiceAccount(ctx(), account.key, account.name, "", testServiceAccountPrivileges(), adminID); !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("%s: got %v", account.key, err)
		}
	}
}

// Make sure a key that would break the token is refused.
func TestServiceAccount_CreateRefusesBadKey(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedServiceAccountAdmin(t, database)

	for _, key := range []string{"has.dot", "has:colon", "Upper", ""} {
		if _, _, err := svc.CreateServiceAccount(ctx(), key, "Test Gateway", "", testServiceAccountPrivileges(), adminID); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%q: %v", key, err)
		}
	}
}

// Make sure the old token stops working as soon as a new secret is issued.
func TestServiceAccount_NewSecretStopsOldToken(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedServiceAccountAdmin(t, database)
	account, oldToken := newTestServiceAccount(t, svc, adminID)

	_, newToken, err := svc.IssueServiceAccountSecret(ctx(), account.ID, adminID)
	if err != nil {
		t.Fatalf("issue secret: %v", err)
	}
	if _, err := svc.ResolveServiceAccount(ctx(), oldToken); !errors.Is(err, identity.ErrInvalidServiceAccountToken) {
		t.Errorf("old token: %v", err)
	}
	if _, err := svc.ResolveServiceAccount(ctx(), newToken); err != nil {
		t.Errorf("new token: %v", err)
	}
}

// Make sure a disabled service account is refused and works again once it is enabled.
func TestServiceAccount_DisabledAccountIsRefused(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedServiceAccountAdmin(t, database)
	account, token := newTestServiceAccount(t, svc, adminID)

	if _, err := svc.DisableServiceAccount(ctx(), account.ID, adminID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResolveServiceAccount(ctx(), token); !errors.Is(err, identity.ErrServiceAccountDisabled) {
		t.Errorf("disabled: %v", err)
	}
	if _, err := svc.EnableServiceAccount(ctx(), account.ID, adminID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResolveServiceAccount(ctx(), token); err != nil {
		t.Errorf("enabled: %v", err)
	}
}

// createTestServiceAccount tries to create a service account with the given privileges.
func createTestServiceAccount(t *testing.T, privileges ...models.PrivilegeKey) error {
	t.Helper()
	database := setupTestDB(t)
	svc := newTestService(database)
	_, _, err := svc.CreateServiceAccount(ctx(), "test-gateway", "Test Gateway", "", privileges, seedServiceAccountAdmin(t, database))
	return err
}

// Make sure a service account can't be given a privilege that isn't allowed for service accounts, such as managing users.
func TestServiceAccount_CreateRefusesPrivilegeNotAllowed(t *testing.T) {
	if err := createTestServiceAccount(t, models.UsersWrite); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

// Make sure a service account needs at least one privilege, since without one it can't call anything.
func TestServiceAccount_CreateRefusesAccountWithoutPrivileges(t *testing.T) {
	if err := createTestServiceAccount(t); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

// Make sure only a holder of core:service-accounts:manage can create a service account.
func TestServiceAccount_CreateNeedsServiceAccountsManage(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	userID := seedUser(t, database, "plain-user@example.edu")

	if _, _, err := svc.CreateServiceAccount(ctx(), "test-gateway", "Test Gateway", "", testServiceAccountPrivileges(), userID); err == nil {
		t.Fatal("created a service account without core:service-accounts:manage")
	}
}

// Make sure the admin role is not given core:service-accounts:manage, so only super admins run service accounts.
func TestServiceAccount_AdminRoleCannotManageServiceAccounts(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)

	var role *models.Role
	if err := svc.inTx(ctx(), func(tx *sql.Tx) (err error) {
		role, err = svc.ensureAdminRoleTx(ctx(), tx)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	keys, err := svc.roles.ListPrivileges(ctx(), role.ID)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(keys, models.ServiceAccountsManage) {
		t.Fatal("the admin role holds core:service-accounts:manage")
	}
}

// Make sure a change made through a service account records the account in the audit trail.
func TestServiceAccount_AuditRecordsTheAccount(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedServiceAccountAdmin(t, database)
	account, _ := newTestServiceAccount(t, svc, adminID)

	viaAccount := identity.WithCaller(ctx(), &identity.Caller{ServiceAccountKey: "gateway"})
	if _, err := svc.DisableServiceAccount(viaAccount, account.ID, adminID); err != nil {
		t.Fatal(err)
	}

	var details string
	if err := database.Get(&details, `SELECT details FROM audit_events WHERE entity_id = $1 AND event_type = $2`, account.ID, serviceAccountAuditDisabled); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(details, `"service_account_key":"gateway"`) || !strings.Contains(details, `"actor_id":"`+adminID+`"`) {
		t.Fatalf("details = %s", details)
	}
}

// Make sure an active service account can't be deleted, so it has to be disabled first.
func TestServiceAccount_DeleteNeedsTheAccountDisabled(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedServiceAccountAdmin(t, database)
	account, _ := newTestServiceAccount(t, svc, adminID)

	if err := svc.DeleteServiceAccount(ctx(), account.ID, adminID); !errors.Is(err, ErrInUse) {
		t.Fatalf("got %v", err)
	}
}

// Make sure a deleted service account can never be used or changed again.
func TestServiceAccount_DeletedAccountCannotBeUsedAgain(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedServiceAccountAdmin(t, database)
	account, token := newTestServiceAccount(t, svc, adminID)
	if _, err := svc.DisableServiceAccount(ctx(), account.ID, adminID); err != nil {
		t.Fatal(err)
	}

	if err := svc.DeleteServiceAccount(ctx(), account.ID, adminID); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.ResolveServiceAccount(ctx(), token); !errors.Is(err, identity.ErrInvalidServiceAccountToken) {
		t.Errorf("token: %v", err)
	}
	if _, err := svc.EnableServiceAccount(ctx(), account.ID, adminID); !errors.Is(err, ErrNotFound) {
		t.Errorf("enable: %v", err)
	}
	if _, _, err := svc.IssueServiceAccountSecret(ctx(), account.ID, adminID); !errors.Is(err, ErrNotFound) {
		t.Errorf("new secret: %v", err)
	}
	if accounts, err := svc.ListServiceAccounts(ctx()); err != nil || len(accounts) != 0 {
		t.Errorf("still listed: %v %v", accounts, err)
	}
}

// Make sure a deleted service account's key is never given to another one, so older records that store the key,
// such as audit rows, keep pointing at the right account. Its name is free to use again.
func TestServiceAccount_DeletedAccountKeepsItsKey(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedServiceAccountAdmin(t, database)
	account, _ := newTestServiceAccount(t, svc, adminID)
	if _, err := svc.DisableServiceAccount(ctx(), account.ID, adminID); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteServiceAccount(ctx(), account.ID, adminID); err != nil {
		t.Fatal(err)
	}

	if _, _, err := svc.CreateServiceAccount(ctx(), account.Key, "Another Gateway", "", testServiceAccountPrivileges(), adminID); !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("same key: %v", err)
	}
	if _, _, err := svc.CreateServiceAccount(ctx(), "new-gateway", account.Name, "", testServiceAccountPrivileges(), adminID); err != nil {
		t.Errorf("same name: %v", err)
	}
}
