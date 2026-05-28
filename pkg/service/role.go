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

	"github.com/apache/airavata-custos/pkg/models"
)

const (
	roleAuditCreated          = "ROLE_CREATED"
	roleAuditUpdated          = "ROLE_UPDATED"
	roleAuditDeleted          = "ROLE_DELETED"
	roleAuditPrivilegeAdded   = "ROLE_PRIVILEGE_ADDED"
	roleAuditPrivilegeRemoved = "ROLE_PRIVILEGE_REMOVED"
)

// CreateRole requires actorID to hold roles:manage. Role names are unique.
func (s *Service) CreateRole(ctx context.Context, name, description, actorID string) (*models.Role, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	if actorID == "" {
		return nil, fmt.Errorf("%w: actor_id is required", ErrInvalidInput)
	}
	role := &models.Role{
		ID:          newID(),
		Name:        name,
		Description: stringPtrOrNil(description),
		IsSystem:    false,
	}
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.PrivilegeRolesManage); err != nil {
			return err
		}
		if existing, err := s.roles.FindByName(ctx, name); err != nil {
			return fmt.Errorf("lookup role by name: %w", err)
		} else if existing != nil {
			return fmt.Errorf("%w: role %q already exists", ErrAlreadyExists, name)
		}
		if err := s.roles.Create(ctx, tx, role); err != nil {
			return fmt.Errorf("insert role: %w", err)
		}
		return s.writeRoleAuditTx(ctx, tx, roleAuditCreated, role.ID, map[string]any{
			"name":     role.Name,
			"actor_id": actorID,
		})
	}); err != nil {
		return nil, err
	}
	return role, nil
}

// UpdateRole requires actorID to hold roles:manage. System roles cannot be
// renamed.
func (s *Service) UpdateRole(ctx context.Context, roleID, name, description, actorID string) (*models.Role, error) {
	if roleID == "" {
		return nil, fmt.Errorf("%w: role_id is required", ErrInvalidInput)
	}
	if actorID == "" {
		return nil, fmt.Errorf("%w: actor_id is required", ErrInvalidInput)
	}
	var updated *models.Role
	if err := s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.PrivilegeRolesManage); err != nil {
			return err
		}
		existing, err := s.roles.FindByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("lookup role: %w", err)
		}
		if existing == nil {
			return fmt.Errorf("%w: role %q does not exist", ErrNotFound, roleID)
		}
		if existing.IsSystem && name != "" && name != existing.Name {
			return fmt.Errorf("%w: cannot rename system role %q", ErrInvalidInput, existing.Name)
		}
		if name != "" {
			existing.Name = name
		}
		if description != "" {
			existing.Description = stringPtrOrNil(description)
		}
		if err := s.roles.Update(ctx, tx, existing); err != nil {
			return fmt.Errorf("update role: %w", err)
		}
		updated = existing
		return s.writeRoleAuditTx(ctx, tx, roleAuditUpdated, existing.ID, map[string]any{
			"actor_id": actorID,
			"name":     existing.Name,
		})
	}); err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteRole requires actorID to hold roles:manage. System roles cannot be
// deleted. CASCADE drops all assignments + privilege rows for this role.
func (s *Service) DeleteRole(ctx context.Context, roleID, actorID string) error {
	if roleID == "" {
		return fmt.Errorf("%w: role_id is required", ErrInvalidInput)
	}
	if actorID == "" {
		return fmt.Errorf("%w: actor_id is required", ErrInvalidInput)
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.PrivilegeRolesManage); err != nil {
			return err
		}
		existing, err := s.roles.FindByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("lookup role: %w", err)
		}
		if existing == nil {
			return fmt.Errorf("%w: role %q does not exist", ErrNotFound, roleID)
		}
		if existing.IsSystem {
			return fmt.Errorf("%w: cannot delete system role %q", ErrInvalidInput, existing.Name)
		}
		if err := s.roles.Delete(ctx, tx, roleID); err != nil {
			return fmt.Errorf("delete role: %w", err)
		}
		return s.writeRoleAuditTx(ctx, tx, roleAuditDeleted, roleID, map[string]any{
			"actor_id": actorID,
			"name":     existing.Name,
		})
	})
}

func (s *Service) GetRole(ctx context.Context, roleID string) (*models.Role, error) {
	r, err := s.roles.FindByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}
	if r == nil {
		return nil, ErrNotFound
	}
	return r, nil
}

func (s *Service) ListRoles(ctx context.Context) ([]models.Role, error) {
	return s.roles.List(ctx)
}

func (s *Service) ListRolePrivileges(ctx context.Context, roleID string) ([]models.PrivilegeKey, error) {
	if roleID == "" {
		return nil, fmt.Errorf("%w: role_id is required", ErrInvalidInput)
	}
	return s.roles.ListPrivileges(ctx, roleID)
}

// AddPrivilegeToRole requires actorID to hold roles:manage. The change
// propagates to every holder of the role.
func (s *Service) AddPrivilegeToRole(ctx context.Context, roleID string, privilege models.PrivilegeKey, actorID string) error {
	if roleID == "" {
		return fmt.Errorf("%w: role_id is required", ErrInvalidInput)
	}
	if !models.IsKnownPrivilege(privilege) {
		return fmt.Errorf("%w: unknown privilege %q", ErrInvalidInput, privilege)
	}
	if actorID == "" {
		return fmt.Errorf("%w: actor_id is required", ErrInvalidInput)
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.PrivilegeRolesManage); err != nil {
			return err
		}
		existing, err := s.roles.FindByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("lookup role: %w", err)
		}
		if existing == nil {
			return fmt.Errorf("%w: role %q does not exist", ErrNotFound, roleID)
		}
		has, err := s.roles.HasPrivilege(ctx, tx, roleID, privilege)
		if err != nil {
			return fmt.Errorf("check role privilege: %w", err)
		}
		if has {
			return fmt.Errorf("%w: role %q already carries %q", ErrAlreadyExists, existing.Name, privilege)
		}
		if err := s.roles.AddPrivilege(ctx, tx, roleID, privilege); err != nil {
			return fmt.Errorf("add role privilege: %w", err)
		}
		return s.writeRoleAuditTx(ctx, tx, roleAuditPrivilegeAdded, roleID, map[string]any{
			"actor_id":  actorID,
			"privilege": privilege,
			"role_name": existing.Name,
		})
	})
}

// RemovePrivilegeFromRole detaches a privilege from a role. Caller must
// hold roles:manage. Refuses to remove privileges:grant or roles:manage if
// no other role carries it.
func (s *Service) RemovePrivilegeFromRole(ctx context.Context, roleID string, privilege models.PrivilegeKey, actorID string) error {
	if roleID == "" {
		return fmt.Errorf("%w: role_id is required", ErrInvalidInput)
	}
	if !models.IsKnownPrivilege(privilege) {
		return fmt.Errorf("%w: unknown privilege %q", ErrInvalidInput, privilege)
	}
	if actorID == "" {
		return fmt.Errorf("%w: actor_id is required", ErrInvalidInput)
	}
	return s.inTx(ctx, func(tx *sql.Tx) error {
		if err := s.assertHasPrivilegeTx(ctx, tx, actorID, models.PrivilegeRolesManage); err != nil {
			return err
		}
		existing, err := s.roles.FindByID(ctx, roleID)
		if err != nil {
			return fmt.Errorf("lookup role: %w", err)
		}
		if existing == nil {
			return fmt.Errorf("%w: role %q does not exist", ErrNotFound, roleID)
		}
		has, err := s.roles.HasPrivilege(ctx, tx, roleID, privilege)
		if err != nil {
			return fmt.Errorf("check role privilege: %w", err)
		}
		if !has {
			return fmt.Errorf("%w: role %q does not carry %q", ErrNotFound, existing.Name, privilege)
		}
		if privilege == models.PrivilegeGrant || privilege == models.PrivilegeRolesManage {
			count, err := s.roles.CountRolesGrantingPrivilege(ctx, tx, privilege)
			if err != nil {
				return fmt.Errorf("count roles granting privilege: %w", err)
			}
			if count <= 1 {
				return fmt.Errorf("%w: cannot remove the last source of %q from any role", ErrInvalidInput, privilege)
			}
		}
		if err := s.roles.RemovePrivilege(ctx, tx, roleID, privilege); err != nil {
			return fmt.Errorf("remove role privilege: %w", err)
		}
		return s.writeRoleAuditTx(ctx, tx, roleAuditPrivilegeRemoved, roleID, map[string]any{
			"actor_id":  actorID,
			"privilege": privilege,
			"role_name": existing.Name,
		})
	})
}

// assertHasPrivilegeTx fails with ErrInvalidInput when actorID does not
// hold the privilege directly or via any role.
func (s *Service) assertHasPrivilegeTx(ctx context.Context, tx *sql.Tx, actorID string, required models.PrivilegeKey) error {
	direct, err := s.privileges.FindForUpdate(ctx, tx, actorID, required)
	if err != nil {
		return fmt.Errorf("lookup actor direct privilege: %w", err)
	}
	if direct != nil {
		return nil
	}
	roleKeys, err := s.userRoles.PrivilegesForUser(ctx, actorID)
	if err != nil {
		return fmt.Errorf("lookup actor role privileges: %w", err)
	}
	for _, k := range roleKeys {
		if k == required {
			return nil
		}
	}
	return fmt.Errorf("%w: actor does not hold %s", ErrInvalidInput, required)
}

func (s *Service) writeRoleAuditTx(ctx context.Context, tx *sql.Tx, eventType string, entityID string, details map[string]any) error {
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
