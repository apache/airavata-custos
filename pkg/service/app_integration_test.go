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

const testAppPrivilege models.PrivilegeKey = "test:items:create"

func registerTestAppPrivilege() {
	models.AllowForApps(testAppPrivilege)
}

// newServiceApp creates a service app named temp-tool.
func newServiceApp(t *testing.T, svc *Service, adminID string) *models.App {
	t.Helper()
	registerTestAppPrivilege()
	app, _, err := svc.CreateApp(ctx(), "temp-tool", "Temp Tool", "", models.AppService, []models.PrivilegeKey{testAppPrivilege}, adminID)
	if err != nil {
		t.Fatalf("create service app: %v", err)
	}
	return app
}

func seedAppAdmin(t *testing.T, database *sqlx.DB) string {
	t.Helper()
	userID := seedUser(t, database, "app-admin@example.edu")
	if _, err := database.Exec(
		`INSERT INTO user_privileges (id, user_id, privilege, granted_at, reason) VALUES ($1, $2, $3, NOW(), 'seed')`,
		uuid.NewString(), userID, string(models.AppsManage),
	); err != nil {
		t.Fatalf("seed apps:manage: %v", err)
	}
	return userID
}

func newTestApp(t *testing.T, svc *Service, adminID string) (*models.App, string) {
	t.Helper()
	registerTestAppPrivilege()
	app, token, err := svc.CreateApp(ctx(), "test-gateway", "Test Gateway", "", models.AppDelegated, []models.PrivilegeKey{testAppPrivilege}, adminID)
	if err != nil {
		t.Fatalf("create app: %v", err)
	}
	return app, token
}

// Make sure the token from creation finds the app with its mode and privileges, and the secret isn't stored.
func TestApp_TokenAuthenticatesAndSecretIsNotStored(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	app, token := newTestApp(t, svc, seedAppAdmin(t, database))

	resolved, err := svc.ResolveApp(ctx(), token)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if resolved.ID != app.ID || resolved.Mode != models.AppDelegated || !slices.Equal(resolved.Privileges, []models.PrivilegeKey{testAppPrivilege}) {
		t.Fatalf("app = %+v", resolved)
	}
	_, secret, _ := strings.Cut(token, ".")
	var stored int
	if err := database.Get(&stored, `SELECT COUNT(*) FROM apps WHERE secret_hash = $1 OR secret_hint = $1`, secret); err != nil {
		t.Fatal(err)
	}
	if stored != 0 {
		t.Fatal("the secret was stored")
	}
}

// Make sure an app no longer gets a stored privilege once that key is not allowed for apps.
func TestApp_DropsPrivilegesNoLongerAllowedForApps(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	app, token := newTestApp(t, svc, seedAppAdmin(t, database))
	// Stands in for a key that a later release took off AllowForApps.
	if _, err := database.Exec(`INSERT INTO app_privileges (app_id, privilege) VALUES ($1, $2)`, app.ID, models.UsersWrite); err != nil {
		t.Fatal(err)
	}

	resolved, err := svc.ResolveApp(ctx(), token)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(resolved.Privileges, []models.PrivilegeKey{testAppPrivilege}) {
		t.Fatalf("privileges = %v", resolved.Privileges)
	}
}

// Make sure a token with a wrong or missing secret, or a malformed app key, is refused.
func TestApp_BadTokenIsRefused(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	newTestApp(t, svc, seedAppAdmin(t, database))

	for _, token := range []string{identity.AppTokenPrefix + "test-gateway.wrong", identity.AppTokenPrefix + "test-gateway", identity.AppTokenPrefix + "\xff.secret"} {
		if _, err := svc.ResolveApp(ctx(), token); !errors.Is(err, identity.ErrInvalidAppToken) {
			t.Errorf("%s: got %v", token, err)
		}
	}
}

// Make sure an app key or name that is already taken is reported as a conflict.
func TestApp_CreateRefusesDuplicateKeyAndName(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)
	newTestApp(t, svc, adminID)

	for _, app := range []struct{ key, name string }{
		{"test-gateway", "Another Gateway"},
		{"another-gateway", "Test Gateway"},
	} {
		if _, _, err := svc.CreateApp(ctx(), app.key, app.name, "", models.AppDelegated, nil, adminID); !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("%s: got %v", app.key, err)
		}
	}
}

// Make sure an app key that would break the token or the certificate key ID is refused.
func TestApp_CreateRefusesBadAppKey(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)

	for _, key := range []string{"has.dot", "has:colon", "Upper", ""} {
		if _, _, err := svc.CreateApp(ctx(), key, "Test Gateway", "", models.AppDelegated, nil, adminID); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%q: %v", key, err)
		}
	}
}

// Make sure the old token stops working as soon as a new secret is issued.
func TestApp_NewSecretStopsOldToken(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)
	app, oldToken := newTestApp(t, svc, adminID)

	_, newToken, err := svc.IssueAppSecret(ctx(), app.ID, adminID)
	if err != nil {
		t.Fatalf("issue secret: %v", err)
	}
	if _, err := svc.ResolveApp(ctx(), oldToken); !errors.Is(err, identity.ErrInvalidAppToken) {
		t.Errorf("old token: %v", err)
	}
	if _, err := svc.ResolveApp(ctx(), newToken); err != nil {
		t.Errorf("new token: %v", err)
	}
}

// Make sure a disabled app is refused and works again once it is enabled.
func TestApp_DisabledAppIsRefused(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)
	app, token := newTestApp(t, svc, adminID)

	if _, err := svc.DisableApp(ctx(), app.ID, adminID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResolveApp(ctx(), token); !errors.Is(err, identity.ErrAppDisabled) {
		t.Errorf("disabled: %v", err)
	}
	if _, err := svc.EnableApp(ctx(), app.ID, adminID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResolveApp(ctx(), token); err != nil {
		t.Errorf("enabled: %v", err)
	}
}

// createTestApp tries to create an app with the given mode and privileges.
func createTestApp(t *testing.T, mode models.AppMode, privileges ...models.PrivilegeKey) error {
	t.Helper()
	database := setupTestDB(t)
	svc := newTestService(database)
	registerTestAppPrivilege()
	_, _, err := svc.CreateApp(ctx(), "test-gateway", "Test Gateway", "", mode, privileges, seedAppAdmin(t, database))
	return err
}

// Make sure an app can't be given a privilege that isn't allowed for apps, such as managing users.
func TestApp_CreateRefusesPrivilegeNotAllowedForApps(t *testing.T) {
	if err := createTestApp(t, models.AppDelegated, models.UsersWrite); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

// Make sure a service app needs at least one privilege, since without one it can't call anything.
func TestApp_CreateRefusesServiceAppWithoutPrivileges(t *testing.T) {
	if err := createTestApp(t, models.AppService); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

// Make sure an app can't be created with a mode other than service or delegated.
func TestApp_CreateRefusesUnknownMode(t *testing.T) {
	if err := createTestApp(t, "OTHER", testAppPrivilege); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v", err)
	}
}

// Make sure a user can allow an app and remove it again.
func TestApp_UserCanAuthorizeAndRemove(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	app, _ := newTestApp(t, svc, seedAppAdmin(t, database))
	userID := seedUser(t, database, "member@example.edu")

	if err := svc.AuthorizeApp(ctx(), app.ID, userID); err != nil {
		t.Fatal(err)
	}
	if authorized, err := svc.IsAppAuthorized(ctx(), app.ID, userID); err != nil || !authorized {
		t.Fatalf("after allow: %v %v", authorized, err)
	}
	if err := svc.RemoveAppAuthorization(ctx(), app.ID, userID); err != nil {
		t.Fatal(err)
	}
	if authorized, err := svc.IsAppAuthorized(ctx(), app.ID, userID); err != nil || authorized {
		t.Fatalf("after remove: %v %v", authorized, err)
	}
}

// Make sure the member list only shows active delegated apps.
func TestApp_AuthorizationListShowsOnlyActiveDelegatedApps(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)
	app, _ := newTestApp(t, svc, adminID)
	newServiceApp(t, svc, adminID)
	userID := seedUser(t, database, "member@example.edu")
	if err := svc.AuthorizeApp(ctx(), app.ID, userID); err != nil {
		t.Fatal(err)
	}

	listed, err := svc.ListAppAuthorizations(ctx(), userID)
	if err != nil || len(listed) != 1 || listed[0].App.ID != app.ID || !listed[0].Authorized {
		t.Fatalf("listed %+v, err %v", listed, err)
	}
	if _, err := svc.DisableApp(ctx(), app.ID, adminID); err != nil {
		t.Fatal(err)
	}
	if listed, err := svc.ListAppAuthorizations(ctx(), userID); err != nil || len(listed) != 0 {
		t.Fatalf("disabled app listed %+v, err %v", listed, err)
	}
	if authorized, err := svc.IsAppAuthorized(ctx(), app.ID, userID); err != nil || !authorized {
		t.Fatalf("allow lost when disabled: %v %v", authorized, err)
	}
}

// Make sure a user's allowed apps move to the surviving user when two users are merged.
func TestApp_AuthorizationMovesOnUserMerge(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	app, _ := newTestApp(t, svc, seedAppAdmin(t, database))
	retiring := seedUser(t, database, "retiring@example.edu")
	survivor := seedUser(t, database, "survivor@example.edu")
	if err := svc.AuthorizeApp(ctx(), app.ID, retiring); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.MergeUsers(ctx(), survivor, retiring); err != nil {
		t.Fatal(err)
	}

	if authorized, err := svc.IsAppAuthorized(ctx(), app.ID, survivor); err != nil || !authorized {
		t.Fatalf("survivor: %v %v", authorized, err)
	}
}

// Make sure a service app or a disabled app can't be allowed.
func TestApp_CannotAuthorizeAppOutsideTheList(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)
	disabled, _ := newTestApp(t, svc, adminID)
	if _, err := svc.DisableApp(ctx(), disabled.ID, adminID); err != nil {
		t.Fatal(err)
	}
	service := newServiceApp(t, svc, adminID)
	userID := seedUser(t, database, "member@example.edu")

	for _, appID := range []string{disabled.ID, service.ID} {
		if err := svc.AuthorizeApp(ctx(), appID, userID); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%s: got %v", appID, err)
		}
	}
}

// Make sure only a holder of core:apps:manage can create an app.
func TestApp_CreateNeedsAppsManage(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	userID := seedUser(t, database, "plain-user@example.edu")

	if _, _, err := svc.CreateApp(ctx(), "test-gateway", "Test Gateway", "", models.AppDelegated, nil, userID); err == nil {
		t.Fatal("created an app without core:apps:manage")
	}
}

// Make sure the admin role is not given core:apps:manage, so only super admins run apps.
func TestApp_AdminRoleCannotManageApps(t *testing.T) {
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
	if slices.Contains(keys, models.AppsManage) {
		t.Fatal("the admin role holds core:apps:manage")
	}
}

// Make sure a change made through an app records the app in the audit trail, next to the user.
func TestApp_AuditRecordsTheApp(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)
	app, _ := newTestApp(t, svc, adminID)

	viaApp := identity.WithCaller(ctx(), &identity.Caller{UserID: adminID, AppID: app.ID, AppKey: "gateway"})
	if _, err := svc.DisableApp(viaApp, app.ID, adminID); err != nil {
		t.Fatal(err)
	}

	var details string
	if err := database.Get(&details, `SELECT details FROM audit_events WHERE entity_id = $1 AND event_type = $2`, app.ID, appAuditDisabled); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(details, `"app_key":"gateway"`) || !strings.Contains(details, `"actor_id":"`+adminID+`"`) {
		t.Fatalf("details = %s", details)
	}
}

// Make sure an active app can't be deleted, so it has to be disabled first.
func TestApp_DeleteNeedsTheAppDisabled(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)
	app, _ := newTestApp(t, svc, adminID)

	if err := svc.DeleteApp(ctx(), app.ID, adminID); !errors.Is(err, ErrInUse) {
		t.Fatalf("got %v", err)
	}
}

// Make sure a deleted app can never be used or changed again and loses every user's authorizations.
func TestApp_DeletedAppCannotBeUsedAgain(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)
	app, token := newTestApp(t, svc, adminID)
	userID := seedUser(t, database, "member@example.edu")
	if err := svc.AuthorizeApp(ctx(), app.ID, userID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.DisableApp(ctx(), app.ID, adminID); err != nil {
		t.Fatal(err)
	}

	if err := svc.DeleteApp(ctx(), app.ID, adminID); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.ResolveApp(ctx(), token); !errors.Is(err, identity.ErrInvalidAppToken) {
		t.Errorf("token: %v", err)
	}
	if _, err := svc.EnableApp(ctx(), app.ID, adminID); !errors.Is(err, ErrNotFound) {
		t.Errorf("enable: %v", err)
	}
	if _, _, err := svc.IssueAppSecret(ctx(), app.ID, adminID); !errors.Is(err, ErrNotFound) {
		t.Errorf("new secret: %v", err)
	}
	if authorized, err := svc.IsAppAuthorized(ctx(), app.ID, userID); err != nil || authorized {
		t.Errorf("allow kept: %v %v", authorized, err)
	}
	if apps, err := svc.ListApps(ctx()); err != nil || len(apps) != 0 {
		t.Errorf("still listed: %v %v", apps, err)
	}
}

// Make sure a deleted app's key is never given to another app, so older records that store the key,
// such as certificates, keep pointing at the right app. Its name is free to use again.
func TestApp_DeletedAppKeepsItsKey(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)
	app, _ := newTestApp(t, svc, adminID)
	if _, err := svc.DisableApp(ctx(), app.ID, adminID); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeleteApp(ctx(), app.ID, adminID); err != nil {
		t.Fatal(err)
	}

	if _, _, err := svc.CreateApp(ctx(), app.Key, "Another Gateway", "", models.AppDelegated, nil, adminID); !errors.Is(err, ErrAlreadyExists) {
		t.Errorf("same key: %v", err)
	}
	if _, _, err := svc.CreateApp(ctx(), "new-gateway", app.Name, "", models.AppDelegated, nil, adminID); err != nil {
		t.Errorf("same name: %v", err)
	}
	names, err := svc.ListAppNames(ctx())
	if err != nil || !slices.ContainsFunc(names, func(a models.App) bool { return a.Key == app.Key && a.Name == app.Name }) {
		t.Errorf("name of the deleted app missing: %v %v", names, err)
	}
}
