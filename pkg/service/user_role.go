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
	"fmt"
	"log/slog"

	"github.com/apache/airavata-custos/pkg/models"
)

const (
	userRoleAuditGranted   = "ROLE_GRANTED"
	userRoleAuditRevoked   = "ROLE_REVOKED"
	userRoleAuditBootstrap = "ROLE_BOOTSTRAPPED"
)

// GrantRoleToUser requires granterID to hold roles:manage.
func (s *Service) GrantRoleToUser(ctx context.Context, userID, roleID, granterID, reason string) (*models.UserRole, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if roleID == "" {
		return nil, fmt.Errorf("%w: role_id is required", ErrInvalidInput)
	}
	if granterID == "" {
		return nil, fmt.Errorf("%w: granter_id is required", ErrInvalidInput)
	}
	assignment := &models.UserRole{
		UserID:    userID,
		RoleID:    roleID,
		GrantedBy: stringPtrOrNil(granterID),
		GrantedAt: nowUTC(),
		Reason:    stringPtrOrNil(reason),
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, granterID, models.PrivilegeRolesManage); err != nil {
			return err
		}
		role, err := s.roles.FindByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("lookup role: %w", err)
		}
		if role == nil {
			return fmt.Errorf("%w: role %q does not exist", ErrNotFound, roleID)
		}
		existing, err := s.userRoles.FindForUpdate(ctx, tx, userID, roleID)
		if err != nil {
			return fmt.Errorf("lookup existing assignment: %w", err)
		}
		if existing != nil {
			return fmt.Errorf("%w: user already holds role %q", ErrAlreadyExists, role.Name)
		}
		if err := s.userRoles.Create(ctx, tx, assignment); err != nil {
			return fmt.Errorf("insert role assignment: %w", err)
		}
		return s.writeRoleAuditTx(ctx, tx, userRoleAuditGranted, userID, map[string]any{
			"actor_id":  granterID,
			"role_id":   roleID,
			"role_name": role.Name,
			"reason":    reason,
		})
	}); err != nil {
		return nil, err
	}
	return assignment, nil
}

// RevokeRoleFromUser requires revokerID to hold roles:manage. Refuses if
// the revoke would leave no holder of privileges:grant or roles:manage anywhere.
func (s *Service) RevokeRoleFromUser(ctx context.Context, userID, roleID, revokerID, reason string) error {
	if userID == "" {
		return fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	if roleID == "" {
		return fmt.Errorf("%w: role_id is required", ErrInvalidInput)
	}
	if revokerID == "" {
		return fmt.Errorf("%w: revoker_id is required", ErrInvalidInput)
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, revokerID, models.PrivilegeRolesManage); err != nil {
			return err
		}
		role, err := s.roles.FindByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("lookup role: %w", err)
		}
		if role == nil {
			return fmt.Errorf("%w: role %q does not exist", ErrNotFound, roleID)
		}
		existing, err := s.userRoles.FindForUpdate(ctx, tx, userID, roleID)
		if err != nil {
			return fmt.Errorf("lookup assignment: %w", err)
		}
		if existing == nil {
			return fmt.Errorf("%w: user does not hold role %q", ErrNotFound, role.Name)
		}
		if err := s.assertNotLastMetaHolderTx(ctx, tx, userID, roleID); err != nil {
			return err
		}
		if err := s.userRoles.Delete(ctx, tx, userID, roleID); err != nil {
			return fmt.Errorf("delete role assignment: %w", err)
		}
		return s.writeRoleAuditTx(ctx, tx, userRoleAuditRevoked, userID, map[string]any{
			"actor_id":  revokerID,
			"role_id":   roleID,
			"role_name": role.Name,
			"reason":    reason,
		})
	})
}

func (s *Service) ListUserRoles(ctx context.Context, userID string) ([]models.UserRole, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user_id is required", ErrInvalidInput)
	}
	return s.userRoles.ListByUser(ctx, userID)
}

func (s *Service) ListRoleHolders(ctx context.Context, roleID string) ([]models.UserRole, error) {
	if roleID == "" {
		return nil, fmt.Errorf("%w: role_id is required", ErrInvalidInput)
	}
	return s.userRoles.ListByRole(ctx, roleID)
}

// BootstrapSuperAdmin ensures the super_admin role exists and grants it to
// the user with the named email. Idempotent — returns nil on every no-op
// case so a failed bootstrap never blocks server start.
func (s *Service) BootstrapSuperAdmin(ctx context.Context, email, source string) error {
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
		role, err := s.ensureSuperAdminRoleTx(ctx, tx)
		if err != nil {
			return fmt.Errorf("ensure super_admin role: %w", err)
		}
		existing, err := s.userRoles.FindForUpdate(ctx, tx, user.ID, role.ID)
		if err != nil {
			return fmt.Errorf("lookup existing assignment: %w", err)
		}
		if existing != nil {
			slog.Info("bootstrap: super_admin already granted to user, skipping", "email", email)
			return nil
		}
		// If another user already holds super_admin, install is past bootstrap.
		holders, err := s.userRoles.ListByRole(ctx, role.ID)
		if err != nil {
			return fmt.Errorf("count super_admin holders: %w", err)
		}
		if len(holders) > 0 {
			slog.Info("bootstrap: super_admin already held by another user, skipping", "email", email)
			return nil
		}
		assignment := &models.UserRole{
			UserID:    user.ID,
			RoleID:    role.ID,
			GrantedBy: nil,
			GrantedAt: nowUTC(),
			Reason:    stringPtrOrNil("bootstrap"),
		}
		if err := s.userRoles.Create(ctx, tx, assignment); err != nil {
			return fmt.Errorf("insert bootstrap assignment: %w", err)
		}
		if err := s.writeRoleAuditTx(ctx, tx, userRoleAuditBootstrap, user.ID, map[string]any{
			"role_id":   role.ID,
			"role_name": role.Name,
			"source":    source,
		}); err != nil {
			return fmt.Errorf("audit bootstrap assignment: %w", err)
		}
		slog.Info("bootstrap: super_admin granted", "user_id", user.ID, "email", email, "source", source)
		return nil
	})
}

// ensureSuperAdminRoleTx creates the super_admin role with the meta
// privileges if it doesn't exist, otherwise returns the existing row.
func (s *Service) ensureSuperAdminRoleTx(ctx context.Context, tx *sql.Tx) (*models.Role, error) {
	role, err := s.roles.FindByName(ctx, models.SystemRoleSuperAdmin)
	if err != nil {
		return nil, fmt.Errorf("lookup super_admin role: %w", err)
	}
	if role == nil {
		role = &models.Role{
			ID:          newID(),
			Name:        models.SystemRoleSuperAdmin,
			Description: stringPtrOrNil("bootstrap role carrying privileges:grant and roles:manage"),
			IsSystem:    true,
		}
		if err := s.roles.Create(ctx, tx, role); err != nil {
			return nil, fmt.Errorf("create super_admin role: %w", err)
		}
	}
	for _, key := range []models.PrivilegeKey{models.PrivilegeGrant, models.PrivilegeRolesManage} {
		has, err := s.roles.HasPrivilege(ctx, tx, role.ID, key)
		if err != nil {
			return nil, fmt.Errorf("check super_admin privilege %s: %w", key, err)
		}
		if !has {
			if err := s.roles.AddPrivilege(ctx, tx, role.ID, key); err != nil {
				return nil, fmt.Errorf("attach %s to super_admin: %w", key, err)
			}
		}
	}
	return role, nil
}

// assertNotLastMetaHolderTx refuses revoke if it would leave no holder of
// privileges:grant or roles:manage anywhere in the system.
func (s *Service) assertNotLastMetaHolderTx(ctx context.Context, tx *sql.Tx, userID, roleID string) error {
	for _, key := range []models.PrivilegeKey{models.PrivilegeGrant, models.PrivilegeRolesManage} {
		rolePrivs, err := s.roles.ListPrivileges(ctx, roleID)
		if err != nil {
			return fmt.Errorf("list role privileges: %w", err)
		}
		grants := false
		for _, k := range rolePrivs {
			if k == key {
				grants = true
				break
			}
		}
		if !grants {
			continue
		}
		count, err := s.countUsersHoldingPrivilegeTx(ctx, tx, key)
		if err != nil {
			return fmt.Errorf("count holders of %s: %w", key, err)
		}
		// Would the revoke drop count to 0? Only if this user is the lone
		// holder AND has no other source for this key.
		if count <= 1 {
			other, err := s.userHasPrivilegeOutsideTx(ctx, tx, userID, key, roleID)
			if err != nil {
				return fmt.Errorf("check alternative source: %w", err)
			}
			if !other {
				return fmt.Errorf("%w: cannot revoke; would leave no holder of %s", ErrInvalidInput, key)
			}
		}
	}
	return nil
}

func (s *Service) countUsersHoldingPrivilegeTx(ctx context.Context, tx *sql.Tx, key models.PrivilegeKey) (int, error) {
	var n int
	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT user_id) FROM (
		   SELECT user_id FROM user_privileges WHERE privilege = ?
		   UNION
		   SELECT ur.user_id FROM user_roles ur
		     JOIN role_privileges rp ON rp.role_id = ur.role_id
		     WHERE rp.privilege = ?
		 ) AS holders`, key, key).Scan(&n)
	return n, err
}

// userHasPrivilegeOutsideTx checks if userID gets the privilege from any
// source other than excludeRoleID.
func (s *Service) userHasPrivilegeOutsideTx(ctx context.Context, tx *sql.Tx, userID string, key models.PrivilegeKey, excludeRoleID string) (bool, error) {
	var n int
	err := tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM (
		   SELECT 1 FROM user_privileges WHERE user_id = ? AND privilege = ?
		   UNION ALL
		   SELECT 1 FROM user_roles ur
		     JOIN role_privileges rp ON rp.role_id = ur.role_id
		     WHERE ur.user_id = ? AND ur.role_id <> ? AND rp.privilege = ?
		 ) AS sources`, userID, key, userID, excludeRoleID, key).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
