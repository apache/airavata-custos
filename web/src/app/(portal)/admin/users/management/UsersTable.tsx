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

import type { Role } from "@/features/core/users/schemas";
import type { UserManagementRow } from "@/features/core/users/types";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { Input } from "@/shared/ui/input";
import { ChevronRight } from "lucide-react";
import * as React from "react";
import { IdentitiesCell } from "./IdentitiesCell";
import { PermissionsDrawer } from "./PermissionsDrawer";
import { RolesCell } from "./RolesCell";
import { IDENTITY_SOURCE_LABELS } from "./identities";

function fullNameFor(user: UserManagementRow): string {
  const name = [user.first_name, user.last_name].filter(Boolean).join(" ");
  return name || user.email;
}

function useExpandableRow() {
  const [expandedId, setExpandedId] = React.useState<string | null>(null);
  return {
    expandedId,
    toggle: (id: string) => setExpandedId((previous) => (previous === id ? null : id)),
    collapseUnless: (id: string) =>
      setExpandedId((previous) => (previous === id ? previous : null)),
    clear: () => setExpandedId(null),
  };
}

export function UsersTable({
  users,
  rolesCatalog,
  currentUserEmail,
  canManageRoles,
  canReadDirectPrivileges,
  page,
  pageSize,
  total,
  onPageChange,
}: {
  users: UserManagementRow[];
  rolesCatalog: Role[];
  currentUserEmail: string | undefined;
  canManageRoles: boolean;
  canReadDirectPrivileges: boolean;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
}) {
  const [selectedId, setSelectedId] = React.useState<string | null>(null);
  const expandedRow = useExpandableRow();
  const searchParams = useShallowSearchParams();
  const search = searchParams.get("q") ?? "";
  const roleFilter = searchParams.get("role") ?? "all";
  const identityFilter = searchParams.get("identity") ?? "all";
  const filtersActive =
    Boolean(search.trim()) || (canManageRoles && roleFilter !== "all") || identityFilter !== "all";
  const selectedUser = users.find((user) => user.id === selectedId) ?? null;

  function resetSelection() {
    setSelectedId(null);
    expandedRow.clear();
  }

  function updateFilterParam(key: string, value: string | null) {
    const params = new URLSearchParams(searchParams.toString());
    if (!value || value === "all") params.delete(key);
    else params.set(key, value);
    replaceShallowSearchParams(params);
    if (key !== "q") {
      resetSelection();
      onPageChange(1);
    }
  }

  function isCurrentUser(row: UserManagementRow): boolean {
    return Boolean(currentUserEmail) && row.email === currentUserEmail;
  }

  const filteredUsers = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    return users.filter((user) => {
      if (needle) {
        const haystack = `${fullNameFor(user)} ${user.email}`.toLowerCase();
        if (!haystack.includes(needle)) return false;
      }
      if (canManageRoles && roleFilter !== "all") {
        if (
          !user.rolesLoading &&
          !user.rolesError &&
          !user.roles.some((role) => role.id === roleFilter)
        ) {
          return false;
        }
      }
      if (identityFilter !== "all") {
        if (
          !user.identitiesLoading &&
          !user.identitiesError &&
          !user.identities.some((identity) => identity.source === identityFilter)
        ) {
          return false;
        }
      }
      return true;
    });
  }, [users, search, roleFilter, identityFilter, canManageRoles]);

  const columns: Array<DataTableColumn<UserManagementRow>> = [
    {
      key: "username",
      header: "Username",
      cell: (row) => (
        <span className="font-medium text-foreground">
          {fullNameFor(row)}
          {isCurrentUser(row) ? (
            <span className="ml-1.5 font-normal text-muted-foreground">(You)</span>
          ) : null}
        </span>
      ),
    },
    {
      key: "email",
      header: "Email",
      cell: (row) => <span className="text-muted-foreground">{row.email}</span>,
    },
  ];

  if (canManageRoles) {
    columns.push({
      key: "roles",
      header: "Roles",
      width: "260px",
      interactive: true,
      cell: (row) => (
        <RolesCell
          roles={row.roles}
          isLoading={row.rolesLoading}
          hasError={row.rolesError}
          expanded={row.id === expandedRow.expandedId}
          onToggleExpand={() => expandedRow.toggle(row.id)}
        />
      ),
    });
  }

  columns.push(
    {
      key: "identities",
      header: "External Identities",
      width: "220px",
      interactive: true,
      cell: (row) => (
        <IdentitiesCell
          identities={row.identities}
          isLoading={row.identitiesLoading}
          hasError={row.identitiesError}
          expanded={row.id === expandedRow.expandedId}
          onToggleExpand={() => expandedRow.toggle(row.id)}
        />
      ),
    },
    {
      key: "actions",
      header: "",
      align: "right",
      cell: () => <ChevronRight className="ml-auto text-muted-foreground" strokeWidth={1.5} />,
    },
  );

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 rounded-md border bg-card p-4 sm:flex-row sm:items-center">
        <Input
          type="search"
          placeholder="Search this page by username or email"
          value={search}
          onChange={(event) => updateFilterParam("q", event.target.value)}
          aria-label="Search users on this page"
          className="sm:w-72"
        />
        {canManageRoles ? (
          <select
            value={roleFilter}
            onChange={(event) => updateFilterParam("role", event.target.value)}
            aria-label="Filter this page by role"
            className="h-9 rounded-md border bg-background px-3 text-sm"
          >
            <option value="all">All roles</option>
            {rolesCatalog.map((role) => (
              <option key={role.id} value={role.id}>
                {role.name}
              </option>
            ))}
          </select>
        ) : null}
        <select
          value={identityFilter}
          onChange={(event) => updateFilterParam("identity", event.target.value)}
          aria-label="Filter this page by external identity"
          className="h-9 rounded-md border bg-background px-3 text-sm"
        >
          <option value="all">All external identities</option>
          {Object.entries(IDENTITY_SOURCE_LABELS).map(([source, label]) => (
            <option key={source} value={source}>
              {label}
            </option>
          ))}
        </select>
      </div>

      {filtersActive ? (
        <p className="text-sm text-muted-foreground">
          Showing {filteredUsers.length} match{filteredUsers.length === 1 ? "" : "es"} on this page;{" "}
          {total} users total.
        </p>
      ) : null}

      <DataTable
        columns={columns}
        rows={filteredUsers}
        rowKey={(row) => row.id}
        onRowClick={(row) => {
          expandedRow.collapseUnless(row.id);
          setSelectedId((previous) => (previous === row.id ? null : row.id));
        }}
        rowClassName={(row) =>
          isCurrentUser(row)
            ? "bg-[color:var(--custos-blue-50)]/40 hover:bg-[color:var(--custos-blue-50)]/60"
            : undefined
        }
        empty={
          <span className="text-sm text-muted-foreground">
            {filtersActive ? "No users match on this page." : "No users found."}
          </span>
        }
        pagination={{
          page,
          pageSize,
          total,
          onPageChange: (nextPage) => {
            resetSelection();
            onPageChange(nextPage);
          },
        }}
      />
      <PermissionsDrawer
        key={selectedUser?.id ?? "closed"}
        user={selectedUser}
        rolesCatalog={rolesCatalog}
        canManageRoles={canManageRoles}
        canReadDirectPrivileges={canReadDirectPrivileges}
        currentUserEmail={currentUserEmail}
        onClose={() => setSelectedId(null)}
      />
    </div>
  );
}
