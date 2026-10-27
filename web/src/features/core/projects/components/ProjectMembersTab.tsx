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
  useAllocationsByProject,
  useMembershipsByAllocation,
  useToggleAllocationAccess,
} from "@/features/core/allocations/queries";
import type { AllocationMembership } from "@/features/core/allocations/schemas";
import { UserDetailsDrawer } from "@/features/core/users/components/UserDetailsDrawer";
import { useUserEmailSearch } from "@/features/core/users/queries";
import type { ComputeAllocation, User } from "@/generated/core/zod.gen";
import { cn } from "@/lib/utils";
import { requires, requiresAny } from "@/shared/casl/permissions";
import { CheckMark } from "@/shared/ui/CheckMark";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { GuardedButton } from "@/shared/ui/GuardedButton";
import { TableSkeleton } from "@/shared/ui/Loading";
import { SuggestInput } from "@/shared/ui/SuggestInput";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import { Button } from "@/shared/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableEmptyRow,
  TableHead,
  TableHeader,
  TableRow,
} from "@/shared/ui/table";
import { Pencil, Plus } from "lucide-react";
import Link from "next/link";
import * as React from "react";
import { toast } from "sonner";
import { useProjectMembers, useUpdateProjectMember } from "../queries";
import { type ProjectMemberRole, projectMemberRoleSchema } from "../schemas";

export type ProjectMembersTabProps = {
  projectId: string;
  canEditRoles: boolean;
  canEditAccess: boolean;
  canSearchUsers: boolean;
};

const ROLE_LABELS: Record<ProjectMemberRole, string> = {
  PI: "PI",
  CO_PI: "Co-PI",
  ALLOCATION_MANAGER: "Allocation Manager",
  MEMBER: "Member",
};

const ASSIGNABLE_ROLES: ProjectMemberRole[] = ["CO_PI", "ALLOCATION_MANAGER", "MEMBER"];

type GridRow = {
  userId: string;
  name: string;
  email: string;
  role: ProjectMemberRole;
};

type Change =
  | { kind: "role"; row: GridRow; to: ProjectMemberRole; isNew: boolean }
  | {
      kind: "access";
      row: GridRow;
      allocation: ComputeAllocation;
      membership: AllocationMembership | undefined;
    };

type Draft = {
  roles: Record<string, ProjectMemberRole>;
  access: Record<string, boolean>;
  added: GridRow[];
};

const EMPTY_DRAFT: Draft = { roles: {}, access: {}, added: [] };

const CHANGED = "bg-amber-50 dark:bg-amber-500/10";

function initialsFrom(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length >= 2) return `${parts[0]?.[0] ?? ""}${parts[1]?.[0] ?? ""}`.toUpperCase();
  return name.slice(0, 2).toUpperCase();
}

function displayName(u: User): string {
  return [u.first_name, u.last_name].filter(Boolean).join(" ") || u.email;
}

// The name truncates first so the email always stays visible at the right end.
function MemberCell({
  name,
  email,
  children,
}: {
  name: string;
  email: string;
  children?: React.ReactNode;
}) {
  return (
    <div className="flex w-96 min-w-0 items-center gap-3">
      <Avatar className="size-8 shrink-0">
        <AvatarFallback>{initialsFrom(name)}</AvatarFallback>
      </Avatar>
      <div className="min-w-0 flex-1 leading-tight">
        <div className="truncate font-medium text-foreground">{name}</div>
        <div className="truncate text-muted-foreground text-xs">{email}</div>
      </div>
      {children}
    </div>
  );
}

function RoleSelect({
  value,
  onChange,
  label,
  disabled,
}: {
  value: ProjectMemberRole | "";
  onChange: (role: ProjectMemberRole) => void;
  label: string;
  disabled: boolean;
}) {
  return (
    <select
      aria-label={label}
      value={value}
      disabled={disabled}
      title={disabled ? requires("projectsWrite") : undefined}
      onChange={(e) => onChange(projectMemberRoleSchema.parse(e.target.value))}
      className="h-8 w-full rounded-md border bg-background px-2 text-sm disabled:cursor-not-allowed disabled:opacity-60"
    >
      {value === "" ? (
        <option value="" disabled>
          Select role
        </option>
      ) : null}
      {ASSIGNABLE_ROLES.map((r) => (
        <option key={r} value={r}>
          {ROLE_LABELS[r]}
        </option>
      ))}
    </select>
  );
}

function describe(change: Change, changes: Change[]): string {
  if (change.kind === "role" && change.isNew) {
    const ticked = changes.some((c) => c.kind === "access" && c.row.userId === change.row.userId);
    return `${change.row.name}: add as ${ROLE_LABELS[change.to]}${
      change.to === "MEMBER" && !ticked ? " (no allocation ticked, so nothing is saved)" : ""
    }`;
  }
  if (change.kind === "role") return `${change.row.name}: role → ${ROLE_LABELS[change.to]}`;
  return `${change.row.name}: ${change.membership ? "remove from" : "add to"} ${change.allocation.name}`;
}

export function ProjectMembersTab({
  projectId,
  canEditRoles,
  canEditAccess,
  canSearchUsers,
}: ProjectMembersTabProps) {
  const membersQuery = useProjectMembers(projectId);
  const allocationsQuery = useAllocationsByProject(projectId);
  const allocations = allocationsQuery.data?.items ?? [];
  const memberships = useMembershipsByAllocation(allocations.map((a) => a.id));
  const toggleAccess = useToggleAllocationAccess();
  const roleMutation = useUpdateProjectMember(projectId);
  const [editing, setEditing] = React.useState(false);
  const [draft, setDraft] = React.useState<Draft>(EMPTY_DRAFT);
  const [reviewing, setReviewing] = React.useState(false);
  const [applying, setApplying] = React.useState(false);
  const [adding, setAdding] = React.useState(false);
  const [viewing, setViewing] = React.useState<string | null>(null);
  const [email, setEmail] = React.useState("");
  const [newRole, setNewRole] = React.useState<ProjectMemberRole | "">("");
  const search = useUserEmailSearch(email, editing && canSearchUsers);

  const error = membersQuery.error ?? allocationsQuery.error ?? memberships.error;
  if (membersQuery.isLoading || allocationsQuery.isLoading || memberships.isLoading) {
    return <TableSkeleton rows={4} columns={4} />;
  }
  if (error) {
    return (
      <ErrorState
        message={error.message}
        onRetry={() => {
          membersQuery.refetch();
          allocationsQuery.refetch();
          memberships.refetch();
        }}
      />
    );
  }
  if (allocations.length === 0) {
    return (
      <EmptyState
        heading="No allocations on this project"
        description="Members get access per allocation, so add an allocation first."
      />
    );
  }

  const members: GridRow[] = (membersQuery.data ?? []).map((m) => ({
    userId: m.user_id,
    name: m.display_name,
    email: m.email,
    role: m.role,
  }));
  const rows = [
    ...members,
    ...draft.added.filter((a) => !members.some((m) => m.userId === a.userId)),
  ];
  const membershipOf = (allocationId: string, userId: string) =>
    memberships.byAllocation[allocationId]?.find((m) => m.user_id === userId);
  const accessKey = (allocationId: string, userId: string) => `${allocationId}:${userId}`;
  const roleOf = (row: GridRow) => draft.roles[row.userId] ?? row.role;
  const hasAccess = (allocationId: string, userId: string) =>
    draft.access[accessKey(allocationId, userId)] ?? Boolean(membershipOf(allocationId, userId));

  const isNew = (row: GridRow) => draft.added.some((a) => a.userId === row.userId);
  const changes: Change[] = rows.flatMap((row) => [
    ...(isNew(row) || roleOf(row) !== row.role
      ? [{ kind: "role" as const, row, to: roleOf(row), isNew: isNew(row) }]
      : []),
    ...allocations.flatMap((allocation) => {
      const membership = membershipOf(allocation.id, row.userId);
      return hasAccess(allocation.id, row.userId) !== Boolean(membership)
        ? [{ kind: "access" as const, row, allocation, membership }]
        : [];
    }),
  ]);

  const resetAdding = () => {
    setAdding(false);
    setEmail("");
    setNewRole("");
  };
  const exitEditing = () => {
    setEditing(false);
    setDraft(EMPTY_DRAFT);
    resetAdding();
  };
  const confirmAdd = () => {
    const user = search.match;
    if (!user || !newRole) return;
    if (!rows.some((r) => r.userId === user.id)) {
      setDraft((d) => ({
        ...d,
        added: [
          ...d.added,
          { userId: user.id, name: displayName(user), email: user.email, role: "MEMBER" },
        ],
        roles: { ...d.roles, [user.id]: newRole },
      }));
    }
    resetAdding();
  };
  // Applied one at a time so a failure leaves a clear point to resume from after the refresh.
  const apply = async () => {
    setApplying(true);
    try {
      for (const change of changes) {
        if (change.kind === "role") {
          if (change.isNew && change.to === "MEMBER") continue;
          await roleMutation.mutateAsync({
            userId: change.row.userId,
            payload: { role: change.to },
          });
        } else {
          await toggleAccess.mutateAsync({
            allocation: change.allocation,
            userId: change.row.userId,
            membership: change.membership,
          });
        }
      }
      toast.success("Member changes applied.");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : String(err));
    } finally {
      setApplying(false);
      setReviewing(false);
      exitEditing();
    }
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center gap-3">
        <h2 className="font-display text-lg font-semibold text-foreground">Memberships</h2>
        <div className="flex gap-2">
          {editing ? (
            <>
              <Button variant="ghost" size="sm" onClick={exitEditing}>
                Cancel
              </Button>
              <Button size="sm" disabled={changes.length === 0} onClick={() => setReviewing(true)}>
                Save
              </Button>
            </>
          ) : (
            <GuardedButton
              size="sm"
              allowed={canEditRoles || canEditAccess}
              reason={requiresAny("projectsWrite", "allocationsWrite")}
              onClick={() => setEditing(true)}
            >
              <Pencil />
              Edit
            </GuardedButton>
          )}
        </div>
      </div>
      <div className="overflow-x-auto rounded-lg border border-foreground/20">
        <Table>
          <TableHeader className="bg-muted/60">
            <TableRow>
              <TableHead>Member</TableHead>
              <TableHead className="w-48">Role</TableHead>
              {allocations.map((a) => (
                <TableHead key={a.id} className="text-center">
                  <Link href={`/allocations/${a.id}`} className="hover:underline">
                    {a.name}
                  </Link>
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.length === 0 ? (
              <TableEmptyRow colSpan={allocations.length + 2}>No members yet.</TableEmptyRow>
            ) : null}
            {rows.map((row) => {
              const role = roleOf(row);
              return (
                <TableRow
                  key={row.userId}
                  className={cn(
                    isNew(row) && CHANGED,
                    !editing && "cursor-pointer hover:bg-muted/60",
                  )}
                  onClick={editing ? undefined : () => setViewing(row.userId)}
                >
                  <TableCell>
                    <MemberCell name={row.name} email={row.email} />
                  </TableCell>
                  <TableCell className={role !== row.role ? CHANGED : undefined}>
                    {editing && row.role !== "PI" ? (
                      <RoleSelect
                        label={`Role of ${row.name}`}
                        disabled={!canEditRoles}
                        value={role}
                        onChange={(next) =>
                          setDraft((d) => ({ ...d, roles: { ...d.roles, [row.userId]: next } }))
                        }
                      />
                    ) : (
                      <span className="text-sm">{ROLE_LABELS[role]}</span>
                    )}
                  </TableCell>
                  {allocations.map((a) => {
                    const checked = hasAccess(a.id, row.userId);
                    const changed = checked !== Boolean(membershipOf(a.id, row.userId));
                    const label = `${row.name} in ${a.name}`;
                    return (
                      <TableCell
                        key={a.id}
                        className={changed ? `text-center ${CHANGED}` : "text-center"}
                      >
                        {editing ? (
                          <input
                            type="checkbox"
                            aria-label={label}
                            className="size-4 accent-[color:var(--custos-blue-600)] disabled:cursor-not-allowed"
                            disabled={!canEditAccess}
                            title={canEditAccess ? undefined : requires("allocationsWrite")}
                            checked={checked}
                            onChange={() =>
                              setDraft((d) => ({
                                ...d,
                                access: { ...d.access, [accessKey(a.id, row.userId)]: !checked },
                              }))
                            }
                          />
                        ) : (
                          <CheckMark checked={checked} label={label} />
                        )}
                      </TableCell>
                    );
                  })}
                </TableRow>
              );
            })}
            {editing && !adding ? (
              <TableRow className="hover:bg-transparent">
                <TableCell colSpan={allocations.length + 2}>
                  <GuardedButton
                    variant="ghost"
                    size="sm"
                    allowed={canSearchUsers}
                    reason={requires("usersRead")}
                    onClick={() => {
                      setNewRole(canEditRoles ? "" : "MEMBER");
                      setAdding(true);
                    }}
                  >
                    <Plus />
                    Add member
                  </GuardedButton>
                </TableCell>
              </TableRow>
            ) : null}
            {editing && adding ? (
              <TableRow className="hover:bg-transparent">
                <TableCell>
                  {search.match ? (
                    <MemberCell name={displayName(search.match)} email={search.match.email}>
                      <Button
                        variant="ghost"
                        size="icon-xs"
                        aria-label="Change member"
                        onClick={() => setEmail("")}
                      >
                        <Pencil />
                      </Button>
                    </MemberCell>
                  ) : (
                    <div className="w-96">
                      <SuggestInput
                        aria-label="New member email"
                        type="email"
                        autoFocus
                        value={email}
                        onValueChange={setEmail}
                        suggestions={search.suggestions}
                        emptyMessage={search.emptyMessage}
                        placeholder="Search by email"
                      />
                    </div>
                  )}
                </TableCell>
                <TableCell>
                  <RoleSelect
                    label="Role of new member"
                    disabled={!canEditRoles}
                    value={newRole}
                    onChange={setNewRole}
                  />
                </TableCell>
                <TableCell colSpan={allocations.length}>
                  <div className="flex justify-end gap-2">
                    <Button variant="ghost" size="sm" onClick={resetAdding}>
                      Cancel
                    </Button>
                    <Button size="sm" onClick={confirmAdd} disabled={!search.match || !newRole}>
                      Confirm
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ) : null}
          </TableBody>
        </Table>
      </div>
      <UserDetailsDrawer userId={viewing} onClose={() => setViewing(null)} />
      <Dialog open={reviewing} onOpenChange={(open) => !applying && setReviewing(open)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Apply member changes</DialogTitle>
            <DialogDescription>
              {changes.length} change{changes.length === 1 ? "" : "s"} will be applied to this
              project.
            </DialogDescription>
          </DialogHeader>
          <ul className="max-h-72 list-disc space-y-1 overflow-y-auto pl-5 text-sm">
            {changes.map((c) => (
              <li
                key={
                  c.kind === "role"
                    ? `role:${c.row.userId}`
                    : `access:${c.allocation.id}:${c.row.userId}`
                }
              >
                {describe(c, changes)}
              </li>
            ))}
          </ul>
          <DialogFooter>
            <Button variant="outline" disabled={applying} onClick={() => setReviewing(false)}>
              Back
            </Button>
            <Button disabled={applying} onClick={apply}>
              {applying ? "Applying…" : "Apply changes"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
