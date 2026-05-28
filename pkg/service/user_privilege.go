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
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

const (
	privilegeAuditGrant     = "PRIVILEGE_GRANTED"
	privilegeAuditRevoke    = "PRIVILEGE_REVOKED"
	privilegeAuditBootstrap = "PRIVILEGE_BOOTSTRAPPED"
)

// GrantPrivilege attaches privilege to userID. Caller (granterID) must hold
// an active privileges:grant.
func (s *Service) GrantPrivilege(
	ctx context.Context,
	userID string,
	privilege models.PrivilegeKey,
	granterID string,
	reason string,
) (*models.UserPrivilege, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if granterID == "" {
		return nil, fmt.Errorf("%w: granter_id is required", ErrInvalidInput)
	}
	if !models.IsKnownPrivilege(privilege) {
		return nil, fmt.Errorf("%w: unknown privilege %q", ErrInvalidInput, privilege)
	}

	grant := &models.UserPrivilege{
		ID:        newID(),
		UserID:    userID,
		Privilege: privilege,
		GrantedBy: stringPtrOrNil(granterID),
		GrantedAt: nowUTC(),
		Reason:    stringPtrOrNil(reason),
	}

	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertGranterTx(ctx, tx, granterID); err != nil {
			return err
		}
		existing, err := s.privileges.FindForUpdate(ctx, tx, userID, privilege)
		if err != nil {
			return fmt.Errorf("lookup existing grant: %w", err)
		}
		if existing != nil {
			return fmt.Errorf("%w: privilege %q is already active for user", ErrAlreadyExists, privilege)
		}
		if err := s.privileges.Create(ctx, tx, grant); err != nil {
			return fmt.Errorf("insert privilege grant: %w", err)
		}
		return s.writePrivilegeAuditTx(ctx, tx, privilegeAuditGrant, userID, map[string]any{
			"privilege": privilege,
			"actor_id":  granterID,
			"reason":    reason,
		})
	}); err != nil {
		return nil, err
	}
	return grant, nil
}

// RevokePrivilege removes the user's grant for privilege via DELETE. The
// full revoke history (who, when, why) is captured in audit_events. The
// meta-privilege (privileges:grant) cannot be self-revoked and cannot be
// removed from the last holder.
func (s *Service) RevokePrivilege(
	ctx context.Context,
	userID string,
	privilege models.PrivilegeKey,
	revokerID string,
	reason string,
) error {
	if userID == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if revokerID == "" {
		return fmt.Errorf("%w: revoker_id is required", ErrInvalidInput)
	}
	if !models.IsKnownPrivilege(privilege) {
		return fmt.Errorf("%w: unknown privilege %q", ErrInvalidInput, privilege)
	}
	if privilege == models.PrivilegeGrant && revokerID == userID {
		return fmt.Errorf("%w: cannot self-revoke %s", ErrInvalidInput, models.PrivilegeGrant)
	}

	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertGranterTx(ctx, tx, revokerID); err != nil {
			return err
		}
		existing, err := s.privileges.FindForUpdate(ctx, tx, userID, privilege)
		if err != nil {
			return fmt.Errorf("lookup active grant: %w", err)
		}
		if existing == nil {
			return fmt.Errorf("%w: no active grant for privilege %q", ErrNotFound, privilege)
		}
		if privilege == models.PrivilegeGrant {
			count, err := s.privileges.CountByPrivilege(ctx, tx, models.PrivilegeGrant)
			if err != nil {
				return fmt.Errorf("count meta holders: %w", err)
			}
			if count <= 1 {
				return fmt.Errorf("%w: cannot revoke the last active %s", ErrInvalidInput, models.PrivilegeGrant)
			}
		}
		if err := s.privileges.Delete(ctx, tx, userID, privilege); err != nil {
			return fmt.Errorf("delete grant: %w", err)
		}
		return s.writePrivilegeAuditTx(ctx, tx, privilegeAuditRevoke, userID, map[string]any{
			"privilege": privilege,
			"actor_id":  revokerID,
			"reason":    reason,
		})
	})
}

// HasPrivilege returns true iff an active grant of the named privilege
// exists for userID.
func (s *Service) HasPrivilege(ctx context.Context, userID string, privilege models.PrivilegeKey) (bool, error) {
	if userID == "" {
		return false, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if !models.IsKnownPrivilege(privilege) {
		return false, fmt.Errorf("%w: unknown privilege %q", ErrInvalidInput, privilege)
	}
	row, err := s.privileges.Find(ctx, userID, privilege)
	if err != nil {
		return false, fmt.Errorf("lookup privilege: %w", err)
	}
	return row != nil, nil
}

// ListUserPrivileges returns the user's active privileges.
func (s *Service) ListUserPrivileges(ctx context.Context, userID string) ([]models.UserPrivilege, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	rows, err := s.privileges.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list privileges: %w", err)
	}
	return rows, nil
}

// ListPrivilegeHolders returns the active holders of privilege.
func (s *Service) ListPrivilegeHolders(ctx context.Context, privilege models.PrivilegeKey) ([]models.UserPrivilege, error) {
	if !models.IsKnownPrivilege(privilege) {
		return nil, fmt.Errorf("%w: unknown privilege %q", ErrInvalidInput, privilege)
	}
	rows, err := s.privileges.ListByPrivilege(ctx, privilege)
	if err != nil {
		return nil, fmt.Errorf("list privilege holders: %w", err)
	}
	return rows, nil
}

// PrivilegeCatalog returns the declared catalog of privilege keys.
func (s *Service) PrivilegeCatalog() []models.PrivilegeKey {
	return models.KnownPrivileges()
}

// BootstrapPrivilegeGrant is called once at server startup if
// CUSTOS_BOOTSTRAP_ADMIN_EMAIL is set. Looks up the user by email and grants
// PrivilegeGrant if no active holder exists. Returns nil on every no-op
// case (no env user, user not found, holder already present); startup must
// not fail because of this.
func (s *Service) BootstrapPrivilegeGrant(ctx context.Context, email, source string) error {
	if email == "" {
		return nil
	}
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("lookup bootstrap user: %w", err)
	}
	if user == nil {
		slog.Warn("bootstrap: user not found, skipping", "email", email)
		return nil
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		existing, err := s.privileges.CountByPrivilege(ctx, tx, models.PrivilegeGrant)
		if err != nil {
			return fmt.Errorf("count grant-holders: %w", err)
		}
		if existing > 0 {
			slog.Info("bootstrap: privileges:grant already held by another user, skipping", "email", email)
			return nil
		}
		grant := &models.UserPrivilege{
			ID:        newID(),
			UserID:    user.ID,
			Privilege: models.PrivilegeGrant,
			GrantedBy: nil,
			GrantedAt: nowUTC(),
			Reason:    stringPtrOrNil("bootstrap"),
		}
		if err := s.privileges.Create(ctx, tx, grant); err != nil {
			return fmt.Errorf("insert bootstrap grant: %w", err)
		}
		if err := s.writePrivilegeAuditTx(ctx, tx, privilegeAuditBootstrap, user.ID, map[string]any{
			"privilege": models.PrivilegeGrant,
			"source":    source,
		}); err != nil {
			return fmt.Errorf("audit bootstrap grant: %w", err)
		}
		slog.Info("bootstrap: privileges:grant granted", "user_id", user.ID, "email", email, "source", source)
		return nil
	})
}

// assertGranterTx fails with ErrInvalidInput when actorID does not hold an
// active privileges:grant. The check runs inside the supplied tx with
// SELECT FOR UPDATE so concurrent grant + revoke serialize.
func (s *Service) assertGranterTx(ctx context.Context, tx *sql.Tx, actorID string) error {
	grant, err := s.privileges.FindForUpdate(ctx, tx, actorID, models.PrivilegeGrant)
	if err != nil {
		return fmt.Errorf("lookup actor meta privilege: %w", err)
	}
	if grant == nil {
		return fmt.Errorf("%w: actor does not hold %s", ErrInvalidInput, models.PrivilegeGrant)
	}
	return nil
}

// writePrivilegeAuditTx records a privilege lifecycle event in audit_events.
func (s *Service) writePrivilegeAuditTx(
	ctx context.Context,
	tx *sql.Tx,
	eventType string,
	entityID string,
	details map[string]any,
) error {
	payload, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("marshal audit details: %w", err)
	}
	return s.auditEvents.Create(ctx, tx, &models.AuditEvent{
		ID:        newID(),
		EventType: eventType,
		EventTime: nowUTC(),
		EntityID:  entityID,
		Details:   string(payload),
	})
}

// stringPtrOrNil returns a pointer to s, or nil when s is empty. Used for
// optional VARCHAR / TEXT columns whose absence we want to encode as NULL
// rather than the empty string.
func stringPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
