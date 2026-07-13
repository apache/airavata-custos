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

import { Badge } from "@/shared/ui/badge";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { PermissionRW } from "@/shared/users-admin/PermissionRW";
import { permissionsFromRoles, rwStateFor } from "@/shared/users-admin/permissions";
import type { UserRow } from "@/shared/users-admin/types";
import { useUsersAdmin } from "@/shared/users-admin/UsersAdminContext";
import { identitySourceIcon, identitySourceLabel } from "./identities";
import { RoleAssignMenu } from "./RoleAssignMenu";

export function PermissionsDrawer({
  user,
  onClose,
}: {
  user: UserRow | null;
  onClose: () => void;
}) {
  const { roles, toggleUserRole } = useUsersAdmin();
  const rwPermissions = rwStateFor(user ? permissionsFromRoles(user.roles) : []).filter(
    (p) => p.read || p.write,
  );
  const heldRoleIds = new Set(user?.roles.map((r) => r.id ?? "") ?? []);

  return (
    <SideDrawer
      open={user !== null}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      title={user ? [user.first_name, user.last_name].filter(Boolean).join(" ") : undefined}
      description={user?.email}
      width="sm"
      modal={false}
      disablePointerDismissal
    >
      {user && (
        <div className="space-y-5">
          <section>
            <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Roles
            </h3>
            <div className="flex flex-wrap items-center gap-1.5">
              {user.roles.length === 0 ? (
                <span className="text-sm text-muted-foreground">No roles</span>
              ) : (
                user.roles.map((role) => (
                  <Badge key={role.id} variant="outline">
                    {role.name}
                  </Badge>
                ))
              )}
              <RoleAssignMenu
                roles={roles}
                heldRoleIds={heldRoleIds}
                onToggleRole={(roleId) => user.id && toggleUserRole(user.id, roleId)}
                triggerLabel={`Assign a role to ${user.email ?? "this user"}`}
              />
            </div>
          </section>

          <div className="border-t border-border" />

          <section>
            <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              External Identities
            </h3>
            <div className="flex flex-wrap items-center gap-1.5">
              {user.identities.length === 0 ? (
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

          <div className="border-t border-border" />

          <section>
            <h3 className="mb-3 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
              Effective Privileges
            </h3>
            {rwPermissions.length === 0 ? (
              <p className="text-sm text-muted-foreground">No privileges granted.</p>
            ) : (
              <ul className="space-y-2">
                {rwPermissions.map((p) => (
                  <li key={p.section} className="flex items-center justify-between text-sm">
                    <span className="font-mono text-foreground">{p.section}</span>
                    <PermissionRW read={p.read} write={p.write} />
                  </li>
                ))}
              </ul>
            )}
          </section>
        </div>
      )}
    </SideDrawer>
  );
}
