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
	serviceAccountAuditCreated      = "SERVICE_ACCOUNT_CREATED"
	serviceAccountAuditUpdated      = "SERVICE_ACCOUNT_UPDATED"
	serviceAccountAuditDisabled     = "SERVICE_ACCOUNT_DISABLED"
	serviceAccountAuditEnabled      = "SERVICE_ACCOUNT_ENABLED"
	serviceAccountAuditSecretIssued = "SERVICE_ACCOUNT_SECRET_ISSUED"
	serviceAccountAuditDeleted      = "SERVICE_ACCOUNT_DELETED"

	serviceAccountEntity           = "service_account"
	serviceAccountNameMaxLength    = 100
	serviceAccountSecretHintLength = 4
)

// The token splits on ".", so a service account key can't have one.
var serviceAccountKeyRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func hashServiceAccountSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// The returned token is never stored.
func setNewServiceAccountSecret(a *models.ServiceAccount) string {
	secret := rand.Text()
	a.SecretHash, a.SecretHint = hashServiceAccountSecret(secret), secret[:serviceAccountSecretHintLength]
	return identity.ServiceAccountTokenPrefix + a.Key + "." + secret
}

func validateServiceAccountKey(key string) error {
	if len(key) > 50 || !serviceAccountKeyRegex.MatchString(key) {
		return fmt.Errorf("%w: service account key must be lowercase letters, digits and hyphens, at most 50 characters", ErrInvalidInput)
	}
	return nil
}

func validateServiceAccountName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > serviceAccountNameMaxLength {
		return "", fmt.Errorf("%w: name is required, at most %d characters", ErrInvalidInput, serviceAccountNameMaxLength)
	}
	return name, nil
}

// A service account with no privileges could not call anything.
func validateServiceAccountPrivileges(privileges []models.PrivilegeKey) error {
	if len(privileges) == 0 {
		return fmt.Errorf("%w: a service account needs at least one privilege", ErrInvalidInput)
	}
	for _, privilege := range privileges {
		if !models.IsServiceAccountPrivilege(privilege) {
			return fmt.Errorf("%w: service accounts cannot hold %s", ErrInvalidInput, privilege)
		}
	}
	return nil
}

func serviceAccountWriteErr(op string, err error) error {
	if strings.Contains(err.Error(), "uq_service_accounts_key") {
		return fmt.Errorf("%w: this service account key is taken, deleted service accounts keep theirs", ErrAlreadyExists)
	}
	if strings.Contains(err.Error(), "uq_service_accounts_name") {
		return fmt.Errorf("%w: a service account with this name already exists", ErrAlreadyExists)
	}
	return fmt.Errorf("%s: %w", op, err)
}

func (s *Service) ListServiceAccounts(ctx context.Context) ([]models.ServiceAccount, error) {
	accounts, err := s.serviceAccounts.List(ctx)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(accounts, func(a models.ServiceAccount) bool { return a.Status == models.ServiceAccountDeleted }), nil
}

func (s *Service) GetServiceAccount(ctx context.Context, id string) (*models.ServiceAccount, error) {
	a, err := s.serviceAccounts.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("lookup service account: %w", err)
	}
	if a == nil || a.Status == models.ServiceAccountDeleted {
		return nil, ErrNotFound
	}
	return a, nil
}

// CreateServiceAccount The returned token cannot be read again.
func (s *Service) CreateServiceAccount(ctx context.Context, key, name, description string, privileges []models.PrivilegeKey, actorID string) (*models.ServiceAccount, string, error) {
	if err := validateServiceAccountKey(key); err != nil {
		return nil, "", err
	}
	name, err := validateServiceAccountName(name)
	if err != nil {
		return nil, "", err
	}
	slices.Sort(privileges)
	privileges = slices.Compact(privileges)
	if err := validateServiceAccountPrivileges(privileges); err != nil {
		return nil, "", err
	}
	a := &models.ServiceAccount{
		ID:          newID(),
		Key:         key,
		Name:        name,
		Description: description,
		Status:      models.ServiceAccountActive,
		CreatedBy:   actorID,
		Privileges:  privileges,
	}
	token := setNewServiceAccountSecret(a)
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.ServiceAccountsManage); err != nil {
			return err
		}
		if err := s.serviceAccounts.Create(ctx, tx, a); err != nil {
			return serviceAccountWriteErr("insert service account", err)
		}
		return s.writeAuditTx(ctx, tx, serviceAccountEntity, serviceAccountAuditCreated, a.ID, map[string]any{"actor_id": actorID, "name": a.Name, "privileges": a.Privileges})
	}); err != nil {
		return nil, "", err
	}
	saved, err := s.GetServiceAccount(ctx, a.ID)
	return saved, token, err
}

func (s *Service) changeServiceAccount(ctx context.Context, id, actorID, auditType string, change func(*models.ServiceAccount)) (*models.ServiceAccount, error) {
	var a *models.ServiceAccount
	if err := s.inTx(ctx, func(tx *sql.Tx) (err error) {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.ServiceAccountsManage); err != nil {
			return err
		}
		if a, err = s.serviceAccounts.FindByIDTx(ctx, tx, id); err != nil {
			return fmt.Errorf("lookup service account: %w", err)
		}
		// A deleted service account can't be edited, enabled, or given a new secret.
		if a == nil || a.Status == models.ServiceAccountDeleted {
			return ErrNotFound
		}
		change(a)
		if err := s.serviceAccounts.Update(ctx, tx, a); err != nil {
			return serviceAccountWriteErr("update service account", err)
		}
		return s.writeAuditTx(ctx, tx, serviceAccountEntity, auditType, id, map[string]any{"actor_id": actorID})
	}); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) UpdateServiceAccount(ctx context.Context, id, name, description, actorID string) (*models.ServiceAccount, error) {
	name, err := validateServiceAccountName(name)
	if err != nil {
		return nil, err
	}
	return s.changeServiceAccount(ctx, id, actorID, serviceAccountAuditUpdated, func(a *models.ServiceAccount) {
		a.Name, a.Description = name, description
	})
}

func (s *Service) DisableServiceAccount(ctx context.Context, id, actorID string) (*models.ServiceAccount, error) {
	return s.changeServiceAccount(ctx, id, actorID, serviceAccountAuditDisabled, func(a *models.ServiceAccount) { a.Status = models.ServiceAccountDisabled })
}

func (s *Service) EnableServiceAccount(ctx context.Context, id, actorID string) (*models.ServiceAccount, error) {
	return s.changeServiceAccount(ctx, id, actorID, serviceAccountAuditEnabled, func(a *models.ServiceAccount) { a.Status = models.ServiceAccountActive })
}

// DeleteServiceAccount soft deletes a disabled service account. The row is kept so the key is never reused.
func (s *Service) DeleteServiceAccount(ctx context.Context, id, actorID string) error {
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.ServiceAccountsManage); err != nil {
			return err
		}
		a, err := s.serviceAccounts.FindByIDTx(ctx, tx, id)
		if err != nil {
			return fmt.Errorf("lookup service account: %w", err)
		}
		if a == nil || a.Status == models.ServiceAccountDeleted {
			return ErrNotFound
		}
		if a.Status != models.ServiceAccountDisabled {
			return fmt.Errorf("%w: disable the service account before deleting it", ErrInUse)
		}
		// Clear the secret so the old token can never work again even if the status is changed back.
		a.Status, a.SecretHash = models.ServiceAccountDeleted, ""
		if err := s.serviceAccounts.Update(ctx, tx, a); err != nil {
			return fmt.Errorf("delete service account: %w", err)
		}
		return s.writeAuditTx(ctx, tx, serviceAccountEntity, serviceAccountAuditDeleted, id, map[string]any{"actor_id": actorID, "key": a.Key, "name": a.Name})
	})
}

func (s *Service) IssueServiceAccountSecret(ctx context.Context, id, actorID string) (*models.ServiceAccount, string, error) {
	var token string
	a, err := s.changeServiceAccount(ctx, id, actorID, serviceAccountAuditSecretIssued, func(a *models.ServiceAccount) { token = setNewServiceAccountSecret(a) })
	return a, token, err
}

func (s *Service) ResolveServiceAccount(ctx context.Context, token string) (*models.ServiceAccount, error) {
	key, secret, ok := strings.Cut(strings.TrimPrefix(token, identity.ServiceAccountTokenPrefix), ".")
	if !ok || validateServiceAccountKey(key) != nil {
		return nil, identity.ErrInvalidServiceAccountToken
	}
	a, err := s.serviceAccounts.FindByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("lookup service account by key: %w", err)
	}
	if a == nil || subtle.ConstantTimeCompare([]byte(a.SecretHash), []byte(hashServiceAccountSecret(secret))) != 1 {
		return nil, identity.ErrInvalidServiceAccountToken
	}
	if a.Status != models.ServiceAccountActive {
		return nil, identity.ErrServiceAccountDisabled
	}
	// A key taken off AllowForServiceAccounts in a later release stops working for accounts that were given it before.
	a.Privileges = slices.DeleteFunc(a.Privileges, func(p models.PrivilegeKey) bool { return !models.IsServiceAccountPrivilege(p) })
	return a, nil
}
