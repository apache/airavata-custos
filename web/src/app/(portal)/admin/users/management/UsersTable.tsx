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

import { type UserManagementRow, useUser, useUserRow } from "@/features/core/users/queries";
import type { Role, UserStatus } from "@/generated/core/types.gen";
import { enumLabel } from "@/shared/format";
import { useSearchDraft } from "@/shared/hooks/usePageClamp";
import {
  replaceShallowSearchParams,
  setSearchParam,
  useCanonicalParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { normalizeQueryId, queryUuid } from "@/shared/layout/nav";
import { DataTable, type DataTableColumn, NameCell, PastLastPage } from "@/shared/ui/DataTable";
import { QueryErrorState } from "@/shared/ui/ErrorState";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { StatusBadge, type StatusBadgeVariant } from "@/shared/ui/StatusBadge";
import { Badge } from "@/shared/ui/badge";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { displayName } from "@/shared/users-admin/displayName";
import { ChevronRight } from "lucide-react";
import * as React from "react";
import { BadgesCell } from "./BadgesCell";
import { IdentityLookupBox } from "./IdentityLookupBox";
import { PermissionsDrawer } from "./PermissionsDrawer";
import { IDENTITY_SOURCE_LABELS, identitySourceIcon, identitySourceLabel } from "./identities";

const STATUS_VARIANTS: Record<UserStatus, StatusBadgeVariant> = {
  ACTIVE: "active",
  PENDING: "pending",
  SUSPENDED: "warning",
  INACTIVE: "inactive",
  MERGED: "inactive",
};

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
  rolesReady,
  currentUserId,
  canManageRoles,
  canReadDirectPrivileges,
  page,
  pageSize,
  total,
}: {
  users: UserManagementRow[];
  rolesCatalog: Role[];
  // The catalog has loaded, so a ?role= outside it can be dropped.
  rolesReady: boolean;
  currentUserId: string | undefined;
  canManageRoles: boolean;
  canReadDirectPrivileges: boolean;
  page: number;
  pageSize: number;
  total: number;
}) {
  const searchParams = useShallowSearchParams();
  // ?user= holds the open user, so traces can link straight to them and a reload keeps the drawer.
  const selectedId = queryUuid(searchParams.get("user")) ?? null;
  const setSelectedId = (id: string | null) => setSearchParam(searchParams, "user", id);
  const expandedRow = useExpandableRow();
  const search = (searchParams.get("q") ?? "").trim();
  const [searchDraft, setSearchDraft] = useSearchDraft(search, (q) => updateFilterParam("q", q));
  const rawRole = normalizeQueryId(searchParams.get("role"));
  const roleFilter =
    (canManageRoles && rolesCatalog.find((role) => role.id === rawRole)?.id) || "all";
  const rawIdentity = searchParams.get("identity") ?? "";
  const identityFilter = Object.hasOwn(IDENTITY_SOURCE_LABELS, rawIdentity) ? rawIdentity : "all";
  const applied = (value: string) => (value === "all" ? undefined : value);
  useCanonicalParams({
    q: search || undefined,
    identity: applied(identityFilter),
    user: selectedId ?? undefined,
    ...(canManageRoles && !rolesReady ? {} : { role: applied(roleFilter) }),
  });
  const filtersActive =
    Boolean(search) || (canManageRoles && roleFilter !== "all") || identityFilter !== "all";

  function roleLabelFor(value: string): string {
    if (value === "all") return "All roles";
    return rolesCatalog.find((role) => role.id === value)?.name ?? value;
  }

  function identityLabelFor(value: string): string {
    return value === "all" ? "All external identities" : identitySourceLabel(value);
  }
  const pageUser = users.find((user) => user.id === selectedId);
  const lookupId = pageUser ? undefined : (selectedId ?? undefined);
  const lookedUpUser = useUserRow(lookupId, rolesCatalog, canManageRoles);
  const lookup = useUser(lookupId);
  const selectedUser = pageUser ?? lookedUpUser ?? null;

  // The search filters only this page; the other filters return to page 1 and close the drawer.
  function updateFilterParam(key: string, value: string | null) {
    const next = new URLSearchParams(searchParams);
    if (value && value !== "all") next.set(key, value);
    else next.delete(key);
    if (key !== "q") {
      next.delete("user");
      next.delete("page");
      expandedRow.clear();
    }
    replaceShallowSearchParams(next);
  }

  function goToPage(nextPage: number) {
    const next = new URLSearchParams(searchParams);
    next.delete("user");
    if (nextPage > 1) next.set("page", String(nextPage));
    else next.delete("page");
    expandedRow.clear();
    replaceShallowSearchParams(next);
  }

  function isCurrentUser(row: UserManagementRow): boolean {
    return row.id === currentUserId;
  }

  const filteredUsers = React.useMemo(() => {
    const needle = search.toLowerCase();
    return users.filter((user) => {
      if (needle) {
        const haystack = `${displayName(user)} ${user.email}`.toLowerCase();
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
      header: "Name",
      cell: (row) => (
        <span className="flex items-baseline font-medium text-foreground">
          <NameCell text={displayName(row)} />
          {isCurrentUser(row) ? (
            <span className="ml-1.5 shrink-0 font-normal text-muted-foreground">(You)</span>
          ) : null}
        </span>
      ),
    },
    {
      key: "email",
      header: "Email",
      cell: (row) => <NameCell text={row.email} className="text-muted-foreground" />,
    },
  ];

  if (canManageRoles) {
    columns.push({
      key: "roles",
      header: "Roles",
      width: "260px",
      interactive: true,
      cell: (row) => (
        <BadgesCell
          items={row.roles}
          label="Roles"
          className="w-[260px]"
          isLoading={row.rolesLoading}
          hasError={row.rolesError}
          expanded={row.id === expandedRow.expandedId}
          onToggleExpand={() => expandedRow.toggle(row.id)}
        >
          {(role) => (
            <Badge key={role.id} variant="outline">
              {role.name}
            </Badge>
          )}
        </BadgesCell>
      ),
    });
  }

  columns.push(
    {
      key: "status",
      header: "Status",
      width: "160px",
      cell: (row) => (
        <StatusBadge variant={STATUS_VARIANTS[row.status]} label={enumLabel(row.status)} />
      ),
    },
    {
      key: "identities",
      header: "External identities",
      width: "220px",
      interactive: true,
      cell: (row) => (
        <BadgesCell
          items={row.identities}
          label="Identities"
          className="w-[220px]"
          isLoading={row.identitiesLoading}
          hasError={row.identitiesError}
          expanded={row.id === expandedRow.expandedId}
          onToggleExpand={() => expandedRow.toggle(row.id)}
        >
          {(identity) => {
            const Icon = identitySourceIcon(identity.source);
            return (
              <Badge key={identity.id} variant="outline">
                <Icon data-icon="inline-start" />
                {identitySourceLabel(identity.source)}
              </Badge>
            );
          }}
        </BadgesCell>
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
      <div className="flex flex-col gap-3 rounded-md border bg-card p-4 sm:flex-row sm:flex-wrap sm:items-center">
        <Input
          type="search"
          placeholder="Search by name or email"
          value={searchDraft}
          onChange={(event) => setSearchDraft(event.target.value)}
          aria-label="Search users on this page"
          className="sm:w-72"
        />
        {canManageRoles ? (
          <Select value={roleFilter} onValueChange={(value) => updateFilterParam("role", value)}>
            <SelectTrigger aria-label="Filter this page by role" className="h-9 w-36 px-3">
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
          <SelectTrigger
            aria-label="Filter this page by external identity"
            className="h-9 w-56 px-3"
          >
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
        <div className="sm:ml-auto">
          <IdentityLookupBox onResolve={setSelectedId} />
        </div>
      </div>

      {filtersActive ? (
        <p className="text-sm text-muted-foreground">
          Showing {filteredUsers.length} match{filteredUsers.length === 1 ? "" : "es"} on this page;{" "}
          {total} users total.
        </p>
      ) : null}

      {users.length === 0 && page > 1 ? (
        <PastLastPage things="users" page={page} onPageChange={goToPage} />
      ) : (
        <DataTable
          caption="Users"
          columns={columns}
          rows={filteredUsers}
          rowKey={(row) => row.id}
          onRowClick={(row) => {
            const id = row.id;
            expandedRow.collapseUnless(id);
            setSelectedId(selectedId === id ? null : id);
          }}
          rowClassName={(row) =>
            isCurrentUser(row)
              ? "bg-[color:var(--brand-tint)]/40 hover:bg-[color:var(--brand-tint)]/60"
              : undefined
          }
          empty={
            <span className="text-sm text-muted-foreground">
              {filtersActive ? "No users match on this page." : "No users found."}
            </span>
          }
          pagination={{ page, pageSize, total, onPageChange: goToPage }}
        />
      )}
      <PermissionsDrawer
        key={selectedUser?.id ?? "closed"}
        user={selectedUser}
        rolesCatalog={rolesCatalog}
        canManageRoles={canManageRoles}
        canReadDirectPrivileges={canReadDirectPrivileges}
        currentUserId={currentUserId}
        onClose={() => setSelectedId(null)}
      />
      <SideDrawer
        open={Boolean(lookup.error)}
        recordKey={selectedId ?? undefined}
        onOpenChange={(open) => (open ? null : setSelectedId(null))}
        title="User"
      >
        {lookup.error ? (
          <QueryErrorState
            error={lookup.error}
            what="user"
            as="h3"
            onRetry={() => lookup.refetch()}
          />
        ) : null}
      </SideDrawer>
    </div>
  );
}
