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
  useDirectPrivileges,
  useRoleDetails,
  useUpdateUserRoles,
} from "@/features/core/users/queries";
import type { UserManagementRow } from "@/features/core/users/queries";
import type { Role } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { Badge } from "@/shared/ui/badge";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { DrawerSection } from "./DrawerSection";
import { DirectPrivilegesSection } from "./DirectPrivilegesSection";
import { IdentitiesSection } from "./IdentitiesSection";
import { PrivilegeList } from "./PrivilegeList";
import { RoleAssignMenu } from "./RoleAssignMenu";
import { UserActivitySections } from "./UserActivitySections";
import { UserStatusSection } from "./UserStatusSection";

export function PermissionsDrawer({
  user,
  rolesCatalog,
  canManageRoles,
  canReadDirectPrivileges,
  currentUserId,
  onClose,
}: {
  user: UserManagementRow | null;
  rolesCatalog: Role[];
  canManageRoles: boolean;
  canReadDirectPrivileges: boolean;
  currentUserId: string | undefined;
  onClose: () => void;
}) {
  const roleIds = user?.roles.map((role) => role.id) ?? [];
  const roleDetails = useRoleDetails(roleIds, Boolean(user) && canManageRoles);
  const directPrivileges = useDirectPrivileges(
    user?.id,
    Boolean(user) && canManageRoles && canReadDirectPrivileges,
  );
  const updateRoles = useUpdateUserRoles();
  const ability = useAbility();
  const canWriteUsers = ability.can("write", "User");
  const isCurrentUser = Boolean(currentUserId) && user?.id === currentUserId;

  // Direct grants stay unfetched without canReadDirectPrivileges, leaving the role-derived set.
  const displayedPrivileges = [
    ...new Set([
      ...roleDetails.details.flatMap((detail) => detail.privileges ?? []),
      ...(directPrivileges.data ?? []).map((grant) => grant.privilege),
    ]),
  ];

  function handleSaveRoles(
    desiredRoleIds: string[],
    reason: string | undefined,
    onSaved: () => void,
  ) {
    if (!user) return;
    updateRoles.mutate(
      { userId: user.id, currentRoleIds: roleIds, desiredRoleIds, reason },
      toastOnSuccess("User roles updated", onSaved),
    );
  }

  return (
    <SideDrawer
      open={user !== null}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      title={user ? [user.first_name, user.last_name].filter(Boolean).join(" ") : undefined}
      description={user?.email}
      width="md"
      modal={false}
      disablePointerDismissal
    >
      {user ? (
        <div className="space-y-5">
          <UserStatusSection user={user} canWrite={canWriteUsers} onMerged={onClose} />

          {canManageRoles ? (
            <DrawerSection title="Roles" isLoading={user.rolesLoading} isError={user.rolesError}>
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
                  roles={rolesCatalog}
                  heldRoleIds={new Set(roleIds)}
                  onSave={handleSaveRoles}
                  triggerLabel={`Assign a role to ${user.email}`}
                  isCurrentUser={isCurrentUser}
                  isPending={updateRoles.isPending}
                />
              </div>
            </DrawerSection>
          ) : null}

          <IdentitiesSection user={user} canWrite={canWriteUsers} />

          {canManageRoles ? (
            <DrawerSection
              title={canReadDirectPrivileges ? "Effective Privileges" : "Role-derived Privileges"}
              isLoading={roleDetails.isLoading || directPrivileges.isLoading}
              isError={roleDetails.isError || directPrivileges.isError}
            >
              <PrivilegeList privileges={displayedPrivileges} />
            </DrawerSection>
          ) : null}

          {canReadDirectPrivileges ? (
            <DirectPrivilegesSection
              userId={user.id}
              email={user.email}
              isCurrentUser={isCurrentUser}
            />
          ) : null}

          <UserActivitySections
            userId={user.id}
            canReadClusters={ability.can("read", "Cluster")}
            canReadAllocations={ability.can("read", "Allocation")}
          />
        </div>
      ) : null}
    </SideDrawer>
  );
}
