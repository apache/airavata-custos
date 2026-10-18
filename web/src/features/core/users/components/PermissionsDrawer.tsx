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

import { Server } from "lucide-react";
import Link from "next/link";
import { useAllocation, useUserMemberships } from "@/features/core/allocations/queries";
import {
  ClusterAccountBadge,
  ClusterAccountDecisions,
} from "@/features/core/cluster-accounts/components/ClusterAccountBadge";
import {
  useDirectPrivileges,
  useRoleDetails,
  useUpdateUserRoles,
} from "@/features/core/users/queries";
import type { ComputeAllocationMembership, Role } from "@/generated/core/zod.gen";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { StatusIcon } from "@/shared/ui/StatusIcon";
import { TableEmptyRow } from "@/shared/ui/table";
import { Badge } from "@/shared/ui/badge";
import * as React from "react";
import { toast } from "sonner";
import { PrivilegeList } from "./PrivilegeList";
import { RoleAssignMenu } from "./RoleAssignMenu";
import { identitySourceIcon, identitySourceLabel } from "./identities";
import type { ClusterAccountsView, UserAccountRow } from "../types";
import { UserStatusBadge } from "./UserStatusBadge";

const HEADING = "mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground";

function formatDate(iso: string | null | undefined): string {
  return iso ? new Date(iso).toLocaleDateString(undefined, { dateStyle: "medium" }) : "";
}

const CELL = "py-2 pr-3";

// Headers stay visible when empty, so the reader always sees what the columns hold.
function DrawerTable({
  headers,
  empty,
  children,
}: {
  headers: string[];
  empty: string | null;
  children: React.ReactNode;
}) {
  return (
    <table className="w-full text-left text-sm">
      <thead className="text-xs text-muted-foreground">
        <tr>
          {headers.map((h, i) => (
            <th key={`${i}-${h}`} className="pb-2 font-medium">
              {h}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {empty ? <TableEmptyRow colSpan={headers.length}>{empty}</TableEmptyRow> : children}
      </tbody>
    </table>
  );
}

function MembershipRow({ membership }: { membership: ComputeAllocationMembership }) {
  const allocation = useAllocation(membership.compute_allocation_id).data;
  const active = membership.membership_status === "ACTIVE";
  return (
    <tr className="border-t border-border/60">
      <td className={CELL}>
        <span className="inline-flex items-center gap-2">
          <StatusIcon
            icon={Server}
            active={active}
            label={membership.membership_status.toLowerCase()}
          />
          <Link
            href={`/allocations/${membership.compute_allocation_id}`}
            className="hover:underline"
          >
            {allocation?.name ?? "…"}
          </Link>
        </span>
      </td>
      <td className={CELL}>{formatDate(membership.start_time)}</td>
      <td className={CELL}>{formatDate(membership.end_time)}</td>
    </tr>
  );
}

export function PermissionsDrawer({
  user,
  clusterAccounts,
  rolesCatalog,
  canManageRoles,
  canReadDirectPrivileges,
  canReadAllocations,
  currentUserEmail,
  onClose,
}: {
  user: UserAccountRow | null;
  // Present when the viewer can read cluster accounts.
  clusterAccounts: ClusterAccountsView | undefined;
  rolesCatalog: Role[];
  canManageRoles: boolean;
  canReadDirectPrivileges: boolean;
  canReadAllocations: boolean;
  currentUserEmail: string | undefined;
  onClose: () => void;
}) {
  const roleIds = user?.roles.map((role) => role.id) ?? [];
  const roleDetails = useRoleDetails(roleIds, Boolean(user) && canManageRoles);
  const directPrivileges = useDirectPrivileges(
    user?.id,
    Boolean(user) && canManageRoles && canReadDirectPrivileges,
  );
  const updateRoles = useUpdateUserRoles();
  const memberships = useUserMemberships(user?.id, Boolean(user) && canReadAllocations);
  const accounts = user?.accounts ?? [];
  const [assignmentError, setAssignmentError] = React.useState<{
    userId: string;
    message: string;
  } | null>(null);
  const currentAssignmentError =
    assignmentError && assignmentError.userId === user?.id ? assignmentError.message : null;

  const roleDerivedPrivileges = Array.from(
    new Set(roleDetails.roles.flatMap((role) => role.privileges)),
  );
  const displayedPrivileges = canReadDirectPrivileges
    ? Array.from(
        new Set([
          ...roleDerivedPrivileges,
          ...(directPrivileges.data ?? []).map((grant) => grant.privilege),
        ]),
      )
    : roleDerivedPrivileges;

  async function handleSaveRoles(desiredRoleIds: string[], reason?: string): Promise<boolean> {
    if (!user) return false;
    setAssignmentError(null);
    try {
      await updateRoles.mutateAsync({
        userId: user.id,
        currentRoleIds: roleIds,
        desiredRoleIds,
        reason,
      });
      toast.success("User roles updated");
      return true;
    } catch (error) {
      setAssignmentError({
        userId: user.id,
        message: error instanceof Error ? error.message : "Failed to update user roles",
      });
      return false;
    }
  }

  return (
    <SideDrawer
      open={user !== null}
      onOpenChange={(open) => {
        if (!open) {
          setAssignmentError(null);
          onClose();
        }
      }}
      title={user ? [user.first_name, user.last_name].filter(Boolean).join(" ") : undefined}
      description={user?.email}
      width="lg"
      modal={false}
      disablePointerDismissal
    >
      {user ? (
        <div className="space-y-5">
          <section>
            <h3 className={HEADING}>Details</h3>
            <dl className="grid grid-cols-[max-content_1fr] items-center gap-x-4 gap-y-2 text-sm">
              <dt className="text-muted-foreground">Status</dt>
              <dd>
                <UserStatusBadge status={user.status} />
              </dd>
              <dt className="text-muted-foreground">Type</dt>
              <dd className="capitalize">{user.type.toLowerCase()}</dd>
              <dt className="text-muted-foreground">User ID</dt>
              <dd className="truncate font-mono text-xs">{user.id}</dd>
            </dl>
          </section>

          {clusterAccounts ? (
            <section>
              <h3 className={HEADING}>Cluster accounts</h3>
              <DrawerTable
                headers={["Cluster", "Username", "Created on", ""]}
                empty={clusterAccounts.clusters.length === 0 ? "No clusters" : null}
              >
                {clusterAccounts.clusters.map((cluster) => {
                  const account = accounts.find((a) => a.compute_cluster_id === cluster.id);
                  return (
                    <tr key={cluster.id} className="border-t border-border/60">
                      <td className={CELL}>
                        <ClusterAccountBadge name={cluster.name} account={account} />
                      </td>
                      <td className={`${CELL} font-mono text-xs`}>
                        {account?.local_username ?? "—"}
                      </td>
                      <td className={CELL}>{formatDate(account?.provisioned_at) || "—"}</td>
                      <td className="py-2 text-right">
                        {account ? (
                          <ClusterAccountDecisions
                            account={account}
                            onApprove={clusterAccounts.onApprove}
                            onDeny={clusterAccounts.onDeny}
                          />
                        ) : null}
                      </td>
                    </tr>
                  );
                })}
              </DrawerTable>
            </section>
          ) : null}

          {canReadAllocations ? (
            <section>
              <h3 className={HEADING}>Allocations</h3>
              <DrawerTable
                headers={["Allocation", "Created on", "Valid until"]}
                empty={
                  memberships.isLoading
                    ? "Loading allocations…"
                    : (memberships.data ?? []).length === 0
                      ? "No allocations"
                      : null
                }
              >
                {(memberships.data ?? []).map((m) => (
                  <MembershipRow key={m.id} membership={m} />
                ))}
              </DrawerTable>
            </section>
          ) : null}
          <div className="border-t border-border" />
          {canManageRoles ? (
            <>
              <section>
                <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  Roles
                </h3>
                <div className="flex flex-wrap items-center gap-1.5">
                  {user.rolesLoading ? (
                    <span className="text-sm text-muted-foreground">Loading roles…</span>
                  ) : user.rolesError ? (
                    <span className="text-sm text-muted-foreground">Roles unavailable</span>
                  ) : user.roles.length === 0 ? (
                    <span className="text-sm text-muted-foreground">No roles</span>
                  ) : (
                    user.roles.map((role) => (
                      <Badge key={role.id} variant="outline">
                        {role.name}
                      </Badge>
                    ))
                  )}
                  {!user.rolesLoading && !user.rolesError ? (
                    <RoleAssignMenu
                      roles={rolesCatalog}
                      heldRoleIds={new Set(roleIds)}
                      onSave={handleSaveRoles}
                      triggerLabel={`Assign a role to ${user.email}`}
                      isCurrentUser={Boolean(currentUserEmail) && user.email === currentUserEmail}
                      isPending={updateRoles.isPending}
                      error={currentAssignmentError}
                    />
                  ) : null}
                </div>
              </section>
              <div className="border-t border-border" />
            </>
          ) : null}

          <section>
            <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              External Identities
            </h3>
            <div className="flex flex-wrap items-center gap-1.5">
              {user.identitiesLoading ? (
                <span className="text-sm text-muted-foreground">Loading identities…</span>
              ) : user.identitiesError ? (
                <span className="text-sm text-muted-foreground">Identities unavailable</span>
              ) : user.identities.length === 0 ? (
                <span className="text-sm text-muted-foreground">No identities</span>
              ) : (
                user.identities.map((identity) => {
                  const Icon = identitySourceIcon(identity.source);
                  return (
                    <Badge key={identity.id} variant="outline">
                      <Icon data-icon="inline-start" />
                      {identitySourceLabel(identity.source)}
                    </Badge>
                  );
                })
              )}
            </div>
          </section>

          {canManageRoles ? (
            <>
              <div className="border-t border-border" />
              <section>
                <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                  {canReadDirectPrivileges ? "Effective Privileges" : "Role-derived Privileges"}
                </h3>
                {roleDetails.isLoading ||
                (canReadDirectPrivileges && directPrivileges.isLoading) ? (
                  <p className="text-sm text-muted-foreground">Loading privileges…</p>
                ) : roleDetails.isError || (canReadDirectPrivileges && directPrivileges.isError) ? (
                  <p className="text-sm text-muted-foreground">Privileges unavailable.</p>
                ) : (
                  <PrivilegeList privileges={displayedPrivileges} />
                )}
              </section>
            </>
          ) : null}
        </div>
      ) : null}
    </SideDrawer>
  );
}
