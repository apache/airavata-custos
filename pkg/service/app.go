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
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/apache/airavata-custos/pkg/identity"
	"github.com/apache/airavata-custos/pkg/models"
)

const (
	appAuditCreated      = "APP_CREATED"
	appAuditUpdated      = "APP_UPDATED"
	appAuditDisabled     = "APP_DISABLED"
	appAuditEnabled      = "APP_ENABLED"
	appAuditSecretIssued = "APP_SECRET_ISSUED"

	appNameMaxLength    = 100
	appSecretHintLength = 4
)

// The token splits on "." and the certificate key ID splits on ":", so an app key can have neither.
var appKeyRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func hashAppSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// The returned token is never stored.
func setNewAppSecret(a *models.App) string {
	secret := rand.Text()
	a.SecretHash, a.SecretHint = hashAppSecret(secret), secret[:appSecretHintLength]
	return identity.AppTokenPrefix + a.Key + "." + secret
}

func validateAppKey(key string) error {
	if len(key) > 50 || !appKeyRegex.MatchString(key) {
		return fmt.Errorf("%w: app key must be lowercase letters, digits and hyphens, at most 50 characters", ErrInvalidInput)
	}
	return nil
}

func validateAppName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > appNameMaxLength {
		return "", fmt.Errorf("%w: name is required, at most %d characters", ErrInvalidInput, appNameMaxLength)
	}
	return name, nil
}

func appConnector(actions []string) (string, error) {
	if len(actions) == 0 {
		return "", fmt.Errorf("%w: at least one action is required", ErrInvalidInput)
	}
	name, _, _ := strings.Cut(actions[0], ":")
	connector, ok := models.FindAppConnector(name)
	if !ok {
		return "", fmt.Errorf("%w: unknown action %s", ErrInvalidInput, actions[0])
	}
	for _, action := range actions {
		if !slices.Contains(connector.Actions, action) {
			return "", fmt.Errorf("%w: %s is not an action of the %s connector", ErrInvalidInput, action, name)
		}
	}
	return name, nil
}

func appWriteErr(op string, err error) error {
	if strings.Contains(err.Error(), "uq_apps_key") {
		return fmt.Errorf("%w: an app with this key already exists", ErrAlreadyExists)
	}
	if strings.Contains(err.Error(), "uq_apps_name") {
		return fmt.Errorf("%w: an app with this name already exists", ErrAlreadyExists)
	}
	return fmt.Errorf("%s: %w", op, err)
}

func (s *Service) ListApps(ctx context.Context) ([]models.App, error) {
	return s.apps.List(ctx)
}

func (s *Service) GetApp(ctx context.Context, id string) (*models.App, error) {
	a, err := s.apps.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("lookup app: %w", err)
	}
	if a == nil {
		return nil, ErrNotFound
	}
	return a, nil
}

// CreateApp The returned token cannot be read again.
func (s *Service) CreateApp(ctx context.Context, key, name, description string, actions []string, actorID string) (*models.App, string, error) {
	if err := validateAppKey(key); err != nil {
		return nil, "", err
	}
	name, err := validateAppName(name)
	if err != nil {
		return nil, "", err
	}
	slices.Sort(actions)
	actions = slices.Compact(actions)
	connector, err := appConnector(actions)
	if err != nil {
		return nil, "", err
	}
	a := &models.App{
		ID:          newID(),
		Key:         key,
		Name:        name,
		Description: description,
		Connector:   connector,
		Status:      models.AppActive,
		CreatedBy:   actorID,
		Actions:     actions,
	}
	token := setNewAppSecret(a)
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.AppsManage); err != nil {
			return err
		}
		if err := s.apps.Create(ctx, tx, a); err != nil {
			return appWriteErr("insert app", err)
		}
		return s.writeAuditTx(ctx, tx, "app", appAuditCreated, a.ID, map[string]any{"actor_id": actorID, "name": a.Name, "actions": a.Actions})
	}); err != nil {
		return nil, "", err
	}
	saved, err := s.GetApp(ctx, a.ID)
	return saved, token, err
}

func (s *Service) changeApp(ctx context.Context, id, actorID, auditType string, change func(*models.App)) (*models.App, error) {
	var a *models.App
	if err := s.inTx(ctx, func(tx *sql.Tx) (err error) {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.AppsManage); err != nil {
			return err
		}
		if a, err = s.apps.FindByIDTx(ctx, tx, id); err != nil {
			return fmt.Errorf("lookup app: %w", err)
		}
		if a == nil {
			return ErrNotFound
		}
		change(a)
		if err := s.apps.Update(ctx, tx, a); err != nil {
			return appWriteErr("update app", err)
		}
		return s.writeAuditTx(ctx, tx, "app", auditType, id, map[string]any{"actor_id": actorID})
	}); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) UpdateApp(ctx context.Context, id, name, description, actorID string) (*models.App, error) {
	name, err := validateAppName(name)
	if err != nil {
		return nil, err
	}
	return s.changeApp(ctx, id, actorID, appAuditUpdated, func(a *models.App) {
		a.Name, a.Description = name, description
	})
}

func (s *Service) DisableApp(ctx context.Context, id, actorID string) (*models.App, error) {
	return s.changeApp(ctx, id, actorID, appAuditDisabled, func(a *models.App) { a.Status = models.AppDisabled })
}

func (s *Service) EnableApp(ctx context.Context, id, actorID string) (*models.App, error) {
	return s.changeApp(ctx, id, actorID, appAuditEnabled, func(a *models.App) { a.Status = models.AppActive })
}

func (s *Service) IssueAppSecret(ctx context.Context, id, actorID string) (*models.App, string, error) {
	var token string
	a, err := s.changeApp(ctx, id, actorID, appAuditSecretIssued, func(a *models.App) { token = setNewAppSecret(a) })
	return a, token, err
}

func (s *Service) ResolveApp(ctx context.Context, token string) (*identity.Caller, error) {
	key, secret, ok := strings.Cut(strings.TrimPrefix(token, identity.AppTokenPrefix), ".")
	if !ok {
		return nil, identity.ErrInvalidAppToken
	}
	a, err := s.apps.FindByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("lookup app by key: %w", err)
	}
	if a == nil || subtle.ConstantTimeCompare([]byte(a.SecretHash), []byte(hashAppSecret(secret))) != 1 {
		return nil, identity.ErrInvalidAppToken
	}
	if a.Status != models.AppActive {
		return nil, identity.ErrAppDisabled
	}
	return &identity.Caller{AppID: a.ID, AppKey: a.Key, AppActions: a.Actions}, nil
}
