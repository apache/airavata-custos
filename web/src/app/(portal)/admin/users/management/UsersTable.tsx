"use client";

import { ChevronRight } from "lucide-react";
import * as React from "react";
import { useCurrentUser } from "@/features/core/identity/queries";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { Input } from "@/shared/ui/input";
import { useUsersAdmin } from "../UsersAdminContext";
import type { UserRow } from "../types";
import { IDENTITY_SOURCE_LABELS } from "./identities";
import { IdentitiesCell } from "./IdentitiesCell";
import { PermissionsDrawer } from "./PermissionsDrawer";
import { RolesCell } from "./RolesCell";

function fullNameFor(user: UserRow): string {
  const name = [user.first_name, user.last_name].filter(Boolean).join(" ");
  return name || (user.email ?? "Unknown user");
}

// Roles and identities collapse to a couple of chips with a "+N" toggle.
// Both columns share one row id so expanding either from a row expands both
// — the row is expanded either way — and only one row is expanded at a
// time: switching rows (expanding a different one, or opening its drawer)
// collapses whichever was open before.
function useExpandableRow() {
  const [expandedId, setExpandedId] = React.useState<string | null>(null);
  function toggle(id: string) {
    setExpandedId((prev) => (prev === id ? null : id));
  }
  function collapseUnless(id: string) {
    setExpandedId((prev) => (prev === id ? prev : null));
  }
  return { expandedId, toggle, collapseUnless };
}

export function UsersTable() {
  const { user: currentUser } = useCurrentUser();
  const { users, roles, toggleUserRole } = useUsersAdmin();
  const [selectedId, setSelectedId] = React.useState<string | null>(null);
  const selectedUser = users.find((u) => u.id === selectedId) ?? null;
  const expandedRow = useExpandableRow();

  const searchParams = useShallowSearchParams();
  const search = searchParams.get("q") ?? "";
  const roleFilter = searchParams.get("role") ?? "all";
  const identityFilter = searchParams.get("identity") ?? "all";

  function updateFilterParam(key: string, value: string | null) {
    const params = new URLSearchParams(searchParams.toString());
    if (!value || value === "all") params.delete(key);
    else params.set(key, value);
    replaceShallowSearchParams(params);
  }

  function isCurrentUser(row: UserRow): boolean {
    return Boolean(currentUser?.email) && row.email === currentUser?.email;
  }

  const filteredUsers = React.useMemo(() => {
    const needle = search.trim().toLowerCase();
    return users.filter((user) => {
      if (needle) {
        const hay = `${fullNameFor(user)} ${user.email ?? ""}`.toLowerCase();
        if (!hay.includes(needle)) return false;
      }
      if (roleFilter !== "all" && !user.roles.some((r) => r.id === roleFilter)) return false;
      if (identityFilter !== "all" && !user.identities.some((i) => i.source === identityFilter)) {
        return false;
      }
      return true;
    });
  }, [users, search, roleFilter, identityFilter]);

  const columns: Array<DataTableColumn<UserRow>> = [
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
    {
      key: "roles",
      header: "Roles",
      width: "260px",
      interactive: true,
      cell: (row) => (
        <RolesCell
          user={row}
          roles={roles}
          expanded={row.id !== undefined && row.id === expandedRow.expandedId}
          onToggleExpand={() => row.id && expandedRow.toggle(row.id)}
          onToggleRole={(roleId) => row.id && toggleUserRole(row.id, roleId)}
        />
      ),
    },
    {
      key: "identities",
      header: "External Identities",
      width: "220px",
      interactive: true,
      cell: (row) => (
        <IdentitiesCell
          identities={row.identities}
          expanded={row.id !== undefined && row.id === expandedRow.expandedId}
          onToggleExpand={() => row.id && expandedRow.toggle(row.id)}
        />
      ),
    },
    {
      key: "actions",
      header: "",
      align: "right",
      cell: () => <ChevronRight className="ml-auto text-muted-foreground" strokeWidth={1.5} />,
    },
  ];

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-3 rounded-md border bg-card p-4 sm:flex-row sm:items-center">
        <Input
          type="search"
          placeholder="Search by username or email"
          value={search}
          onChange={(e) => updateFilterParam("q", e.target.value)}
          aria-label="Search users"
          className="sm:w-72"
        />
        <select
          value={roleFilter}
          onChange={(e) => updateFilterParam("role", e.target.value)}
          aria-label="Filter by role"
          className="h-9 rounded-md border bg-background px-3 text-sm"
        >
          <option value="all">All roles</option>
          {roles.map((role) => (
            <option key={role.id} value={role.id}>
              {role.name}
            </option>
          ))}
        </select>
        <select
          value={identityFilter}
          onChange={(e) => updateFilterParam("identity", e.target.value)}
          aria-label="Filter by external identity"
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

      <DataTable
        columns={columns}
        rows={filteredUsers}
        rowKey={(row) => row.id ?? row.email ?? ""}
        onRowClick={(row) => {
          if (row.id) expandedRow.collapseUnless(row.id);
          setSelectedId((prev) => (prev === row.id ? null : (row.id ?? null)));
        }}
        rowClassName={(row) =>
          isCurrentUser(row)
            ? "bg-[color:var(--custos-blue-50)]/40 hover:bg-[color:var(--custos-blue-50)]/60"
            : undefined
        }
        empty={
          <span className="text-sm text-muted-foreground">
            No users match the current filters.
          </span>
        }
      />
      <PermissionsDrawer user={selectedUser} onClose={() => setSelectedId(null)} />
    </div>
  );
}
