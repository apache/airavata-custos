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
  type UserManagementRow,
  useDirectPrivileges,
  useRoleDetails,
  useUpdateUserRoles,
} from "@/features/core/users/queries";
import type { Role } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { Badge } from "@/shared/ui/badge";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { displayName, emailBeside } from "@/shared/users-admin/displayName";
import { DirectPrivilegesSection } from "./DirectPrivilegesSection";
import { DrawerSection } from "./DrawerSection";
import { EditUserNameDialog } from "./EditUserNameDialog";
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
  const directPrivileges = useDirectPrivileges(user?.id);
  const displayedPrivileges = [
    ...new Set([
      ...roleDetails.details.flatMap((detail) => detail.privileges),
      ...(directPrivileges.data ?? []).map((grant) => grant.privilege),
    ]),
  ];
  const updateRoles = useUpdateUserRoles();
  const ability = useAbility();
  const canWriteUsers = ability.can("write", "User");
  const isCurrentUser = Boolean(currentUserId) && user?.id === currentUserId;

  function handleSaveRoles(
    desiredRoleIds: string[],
    reason: string | undefined,
    onSaved: () => void,
  ) {
    if (!user) return;
    toastOnSuccess(
      updateRoles.mutateAsync({ userId: user.id, currentRoleIds: roleIds, desiredRoleIds, reason }),
      "User roles updated",
      onSaved,
    );
  }

  return (
    <SideDrawer
      open={user !== null}
      recordKey={user?.id}
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      title={user ? displayName(user) : undefined}
      description={user ? emailBeside(displayName(user), user.email) : undefined}
      width="md"
    >
      {user ? (
        <div className="space-y-5">
          {canWriteUsers ? <EditUserNameDialog user={user} /> : null}
          <UserStatusSection
            user={user}
            canWrite={canWriteUsers}
            isCurrentUser={isCurrentUser}
            onMerged={onClose}
          />

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
                  triggerLabel={`Assign a role to ${displayName(user)}`}
                  isCurrentUser={isCurrentUser}
                  isPending={updateRoles.isPending}
                />
              </div>
            </DrawerSection>
          ) : null}

          <IdentitiesSection user={user} canWrite={canWriteUsers} />

          {canManageRoles ? (
            <DrawerSection
              title={canReadDirectPrivileges ? "Effective privileges" : "Role-derived privileges"}
              isLoading={roleDetails.isLoading || directPrivileges.isLoading}
              isError={roleDetails.isError || directPrivileges.isError}
            >
              <PrivilegeList privileges={displayedPrivileges} />
            </DrawerSection>
          ) : null}

          {canReadDirectPrivileges ? (
            <DirectPrivilegesSection
              userId={user.id}
              name={displayName(user)}
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
