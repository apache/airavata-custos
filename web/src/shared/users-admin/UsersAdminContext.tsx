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

"use client";

import * as React from "react";
import type { PermissionKey } from "./permissions";
import { INITIAL_ROLES } from "./roles-catalog";
import { INITIAL_USERS } from "./seed-users";
import type { RoleRow, UserRow } from "./types";

export type RoleInput = {
  name: string;
  description: string;
  permissions: PermissionKey[];
};

type UsersAdminContextValue = {
  users: UserRow[];
  roles: RoleRow[];
  toggleUserRole: (userId: string, roleId: string) => void;
  addRole: (input: RoleInput, memberUserIds: string[]) => void;
  updateRole: (roleId: string, input: RoleInput, memberUserIds: string[]) => void;
};

const UsersAdminContext = React.createContext<UsersAdminContextValue | null>(null);

let nextRoleSuffix = 1;

export function UsersAdminProvider({ children }: { children: React.ReactNode }) {
  const [users, setUsers] = React.useState<UserRow[]>(INITIAL_USERS);
  const [roles, setRoles] = React.useState<RoleRow[]>(INITIAL_ROLES);

  function toggleUserRole(userId: string, roleId: string) {
    const role = roles.find((r) => r.id === roleId);
    if (!role) return;
    setUsers((prev) =>
      prev.map((user) => {
        if (user.id !== userId) return user;
        const hasRole = user.roles.some((r) => r.id === roleId);
        const nextRoles = hasRole
          ? user.roles.filter((r) => r.id !== roleId)
          : [...user.roles, role];
        return { ...user, roles: nextRoles };
      }),
    );
  }

  function addRole(input: RoleInput, memberUserIds: string[]) {
    const role: RoleRow = {
      id: `role-custom-${nextRoleSuffix++}`,
      name: input.name,
      description: input.description,
      is_system: false,
      permissions: input.permissions,
    };
    setRoles((prev) => [...prev, role]);
    if (memberUserIds.length > 0) {
      const memberSet = new Set(memberUserIds);
      setUsers((prev) =>
        prev.map((user) =>
          user.id && memberSet.has(user.id) ? { ...user, roles: [...user.roles, role] } : user,
        ),
      );
    }
  }

  // Role objects are embedded (by value) into each user's `roles` array
  // rather than referenced by id, mirroring how the table/drawer render
  // them — so updating a role here must also refresh the copy held by every
  // member, add it to newly-assigned members, and drop it from anyone
  // deselected.
  function updateRole(roleId: string, input: RoleInput, memberUserIds: string[]) {
    const existing = roles.find((r) => r.id === roleId);
    const updated: RoleRow = {
      id: roleId,
      name: input.name,
      description: input.description,
      is_system: existing?.is_system ?? false,
      permissions: input.permissions,
    };
    setRoles((prev) => prev.map((r) => (r.id === roleId ? updated : r)));
    const memberSet = new Set(memberUserIds);
    setUsers((prev) =>
      prev.map((user) => {
        const hasRole = user.roles.some((r) => r.id === roleId);
        const shouldHaveRole = Boolean(user.id) && memberSet.has(user.id as string);
        if (shouldHaveRole) {
          const nextRoles = hasRole
            ? user.roles.map((r) => (r.id === roleId ? updated : r))
            : [...user.roles, updated];
          return { ...user, roles: nextRoles };
        }
        if (hasRole) {
          return { ...user, roles: user.roles.filter((r) => r.id !== roleId) };
        }
        return user;
      }),
    );
  }

  return (
    <UsersAdminContext.Provider value={{ users, roles, toggleUserRole, addRole, updateRole }}>
      {children}
    </UsersAdminContext.Provider>
  );
}

export function useUsersAdmin(): UsersAdminContextValue {
  const ctx = React.useContext(UsersAdminContext);
  if (!ctx) throw new Error("useUsersAdmin must be used within UsersAdminProvider");
  return ctx;
}
