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

func registerTestAppConnectors() {
	models.RegisterAppConnector(models.AppConnector{Name: "signer", NeedsUserGrant: true, Actions: []string{"signer:app:generate", "signer:app:revoke"}})
	models.RegisterAppConnector(models.AppConnector{Name: "temp-account", Actions: []string{"temp-account:app:create"}})
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
	registerTestAppConnectors()
	app, token, err := svc.CreateApp(ctx(), "test-gateway", "Test Gateway", "", []string{"signer:app:revoke", "signer:app:generate"}, adminID)
	if err != nil {
		t.Fatalf("create app: %v", err)
	}
	return app, token
}

// Make sure the token returned at creation authenticates the app with its actions, and the secret itself is not stored.
func TestApp_TokenAuthenticatesAndSecretIsNotStored(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	app, token := newTestApp(t, svc, seedAppAdmin(t, database))

	caller, err := svc.ResolveApp(ctx(), token)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if caller.AppID != app.ID || caller.AppKey != "test-gateway" || caller.UserID != "" || !slices.Equal(caller.AppActions, []string{"signer:app:generate", "signer:app:revoke"}) {
		t.Fatalf("caller = %+v", caller)
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
		if _, _, err := svc.CreateApp(ctx(), app.key, app.name, "", []string{"signer:app:generate"}, adminID); !errors.Is(err, ErrAlreadyExists) {
			t.Errorf("%s: got %v", app.key, err)
		}
	}
}

// Make sure an app key that would break the token or the certificate key ID is refused.
func TestApp_CreateRefusesBadAppKey(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)
	registerTestAppConnectors()

	for _, key := range []string{"has.dot", "has:colon", "Upper", ""} {
		if _, _, err := svc.CreateApp(ctx(), key, "Test Gateway", "", []string{"signer:app:generate"}, adminID); !errors.Is(err, ErrInvalidInput) {
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

// Make sure an app cannot be created with actions from two connectors or with an unknown action.
func TestApp_CreateRefusesBadActions(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	adminID := seedAppAdmin(t, database)
	registerTestAppConnectors()

	for _, actions := range [][]string{
		{"signer:app:generate", "temp-account:app:create"},
		{"signer:app:unknown"},
		{},
	} {
		if _, _, err := svc.CreateApp(ctx(), "test-gateway", "Test Gateway", "", actions, adminID); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("%v: %v", actions, err)
		}
	}
}

// Make sure only a holder of core:apps:manage can create an app.
func TestApp_CreateNeedsAppsManage(t *testing.T) {
	database := setupTestDB(t)
	svc := newTestService(database)
	registerTestAppConnectors()
	userID := seedUser(t, database, "plain-user@example.edu")

	if _, _, err := svc.CreateApp(ctx(), "test-gateway", "Test Gateway", "", []string{"signer:app:generate"}, userID); err == nil {
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
