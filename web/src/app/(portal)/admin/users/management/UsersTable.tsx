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

import {
  clusterAccountLabel,
  clusterAccountStatus,
} from "@/features/core/cluster-accounts/components/ClusterAccountBadge";
import { PermissionsDrawer } from "@/features/core/users/components/PermissionsDrawer";
import { userStatusLabel } from "@/features/core/users/components/UserStatusBadge";
import {
  IDENTITY_SOURCE_LABELS,
  identitySourceLabel,
} from "@/features/core/users/components/identities";
import type {
  ClusterAccountsView,
  UserAccountRow,
  UserManagementRow,
} from "@/features/core/users/types";
import type { ClusterAccountApproval, Role } from "@/generated/core/zod.gen";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { CheckMark } from "@/shared/ui/CheckMark";
import { DataTable, type DataTableColumn } from "@/shared/ui/DataTable";
import { StatusIcon } from "@/shared/ui/StatusIcon";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { ChevronRight, UserRound } from "lucide-react";
import * as React from "react";
import { IdentitiesCell } from "./IdentitiesCell";
import { RolesCell } from "./RolesCell";

export type AccountStatusFilter = ClusterAccountApproval | "ALL";

const ACCOUNT_STATUS_LABELS: Record<AccountStatusFilter, string> = {
  ALL: "All statuses",
  PENDING: "Pending",
  APPROVED: "Approved",
  DENIED: "Denied",
};

export const toAccountStatus = (value: string | null): AccountStatusFilter =>
  Object.keys(ACCOUNT_STATUS_LABELS).find((s): s is AccountStatusFilter => s === value) ?? "ALL";

type ClusterAccountActions = ClusterAccountsView & {
  status: AccountStatusFilter;
  onStatusChange: (status: AccountStatusFilter) => void;
};

const AGO = new Intl.RelativeTimeFormat("en");

function createdAgo(iso: string): string {
  const minutes = Math.floor((Date.now() - Date.parse(iso)) / 60_000);
  if (minutes < 1) return "a few seconds ago";
  if (minutes < 60) return AGO.format(-minutes, "minute");
  if (minutes < 1440) return AGO.format(-Math.floor(minutes / 60), "hour");
  if (minutes < 10080) return AGO.format(-Math.floor(minutes / 1440), "day");
  return new Date(iso).toLocaleDateString(undefined, { dateStyle: "medium" });
}

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
  canReadAllocations,
  search,
  onSearchChange,
  page,
  pageSize,
  total,
  onPageChange,
  toolbarAction,
  clusterAccounts,
  placeholder,
  sort,
}: {
  users: UserAccountRow[];
  rolesCatalog: Role[];
  currentUserEmail: string | undefined;
  canManageRoles: boolean;
  canReadDirectPrivileges: boolean;
  canReadAllocations: boolean;
  search: string;
  onSearchChange: (text: string) => void;
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  toolbarAction?: React.ReactNode;
  // Present when the viewer can read cluster accounts.
  clusterAccounts: ClusterAccountActions | undefined;
  // Shown in place of the table, under the toolbar, while rows load or fail.
  placeholder?: React.ReactNode;
  sort?: string | undefined;
}) {
  const [selectedId, setSelectedId] = React.useState<string | null>(null);
  const expandedRow = useExpandableRow();
  const searchParams = useShallowSearchParams();
  const roleFilter = searchParams.get("role") ?? "all";
  const identityFilter = searchParams.get("identity") ?? "all";
  const filtersActive =
    Boolean(search.trim()) || (canManageRoles && roleFilter !== "all") || identityFilter !== "all";

  function roleLabelFor(value: string): string {
    if (value === "all") return "All roles";
    return rolesCatalog.find((role) => role.id === value)?.name ?? value;
  }

  function identityLabelFor(value: string): string {
    return value === "all" ? "All external identities" : identitySourceLabel(value);
  }
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
    resetSelection();
    onPageChange(1);
  }

  function isCurrentUser(row: UserManagementRow): boolean {
    return Boolean(currentUserEmail) && row.email === currentUserEmail;
  }

  const columns: Array<DataTableColumn<UserAccountRow>> = [
    {
      key: "name",
      header: "Name",
      sortable: true,
      cell: (row) => (
        <span className="inline-flex items-center gap-2">
          <StatusIcon
            icon={UserRound}
            active={row.status === "ACTIVE"}
            label={userStatusLabel(row.status)}
          />
          <span>
            <span className="block font-medium text-foreground">
              {fullNameFor(row)}
              {isCurrentUser(row) ? (
                <span className="ml-1.5 font-normal text-muted-foreground">(You)</span>
              ) : null}
            </span>
            <span className="block text-xs text-muted-foreground">{row.email}</span>
          </span>
        </span>
      ),
    },
  ];

  for (const cluster of clusterAccounts?.clusters ?? []) {
    columns.push({
      key: `cluster:${cluster.id}`,
      header: cluster.name,
      align: "center",
      cell: (row) => {
        const account = row.accounts.find((a) => a.compute_cluster_id === cluster.id);
        const active = account ? clusterAccountStatus(account).variant === "active" : false;
        return <CheckMark checked={active} label={clusterAccountLabel(account)} />;
      },
    });
  }

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
      key: "created_at",
      width: "150px",
      header: "Created at",
      sortable: true,
      cell: (row) => (
        <span className="whitespace-nowrap text-muted-foreground">
          {createdAgo(row.created_at)}
        </span>
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
          placeholder="Search by name or email"
          value={search}
          onChange={(event) => onSearchChange(event.target.value)}
          aria-label="Search users"
          className="sm:w-72"
        />
        {canManageRoles ? (
          <Select value={roleFilter} onValueChange={(value) => updateFilterParam("role", value)}>
            <SelectTrigger aria-label="Filter by role" className="h-9 w-36 px-3">
              <SelectValue>{(value: string) => roleLabelFor(value)}</SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All roles</SelectItem>
              {rolesCatalog.map((role) => (
                <SelectItem key={role.id} value={role.id}>
                  {role.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : null}
        <Select
          value={identityFilter}
          onValueChange={(value) => updateFilterParam("identity", value)}
        >
          <SelectTrigger aria-label="Filter by external identity" className="h-9 w-56 px-3">
            <SelectValue>{(value: string) => identityLabelFor(value)}</SelectValue>
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All external identities</SelectItem>
            {Object.entries(IDENTITY_SOURCE_LABELS).map(([source, label]) => (
              <SelectItem key={source} value={source}>
                {label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {clusterAccounts ? (
          <Select
            value={clusterAccounts.status}
            onValueChange={(value) => clusterAccounts.onStatusChange(toAccountStatus(value))}
          >
            <SelectTrigger aria-label="Filter by cluster account status" className="h-9 w-36 px-3">
              <SelectValue>
                {(value: AccountStatusFilter) => ACCOUNT_STATUS_LABELS[value]}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {Object.entries(ACCOUNT_STATUS_LABELS).map(([value, label]) => (
                <SelectItem key={value} value={value}>
                  {label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : null}
        {toolbarAction}
      </div>

      {placeholder ?? (
        <DataTable
          columns={columns}
          rows={users}
          rowKey={(row) => row.id}
          onRowClick={(row) => {
            expandedRow.collapseUnless(row.id);
            setSelectedId((previous) => (previous === row.id ? null : row.id));
          }}
          rowClassName={(row) =>
            isCurrentUser(row)
              ? "bg-[color:var(--brand-tint)]/40 hover:bg-[color:var(--brand-tint)]/60"
              : undefined
          }
          sort={
            sort === undefined
              ? undefined
              : { key: sort.replace(/^-/, ""), direction: sort.startsWith("-") ? "desc" : "asc" }
          }
          onSortChange={
            sort === undefined
              ? undefined
              : ({ key, direction }) =>
                  updateFilterParam("sort", `${direction === "desc" ? "-" : ""}${key}`)
          }
          empty={filtersActive ? "No users match these filters." : "No users found."}
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
      )}
      <PermissionsDrawer
        key={selectedUser?.id ?? "closed"}
        user={selectedUser}
        clusterAccounts={clusterAccounts}
        rolesCatalog={rolesCatalog}
        canManageRoles={canManageRoles}
        canReadDirectPrivileges={canReadDirectPrivileges}
        canReadAllocations={canReadAllocations}
        currentUserEmail={currentUserEmail}
        onClose={() => setSelectedId(null)}
      />
    </div>
  );
}
