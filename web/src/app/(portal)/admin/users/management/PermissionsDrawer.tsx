"use client";

import {
  useDirectPrivileges,
  useRoleDetails,
  useUpdateUserRoles,
} from "@/features/core/users/queries";
import type { Role } from "@/features/core/users/schemas";
import type { UserManagementRow } from "@/features/core/users/types";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { Badge } from "@/shared/ui/badge";
import * as React from "react";
import { toast } from "sonner";
import { PrivilegeList } from "./PrivilegeList";
import { RoleAssignMenu } from "./RoleAssignMenu";
import { identitySourceIcon, identitySourceLabel } from "./identities";

export function PermissionsDrawer({
  user,
  rolesCatalog,
  canManageRoles,
  canReadDirectPrivileges,
  currentUserEmail,
  onClose,
}: {
  user: UserManagementRow | null;
  rolesCatalog: Role[];
  canManageRoles: boolean;
  canReadDirectPrivileges: boolean;
  currentUserEmail: string | undefined;
  onClose: () => void;
}) {
  const roleIds = user?.roles.flatMap((role) => (role.id ? [role.id] : [])) ?? [];
  const roleDetails = useRoleDetails(roleIds, Boolean(user) && canManageRoles);
  const directPrivileges = useDirectPrivileges(
    user?.id,
    Boolean(user) && canManageRoles && canReadDirectPrivileges,
  );
  const updateRoles = useUpdateUserRoles();
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

  async function handleSaveRoles(desiredRoleIds: string[]): Promise<boolean> {
    if (!user?.id) return false;
    setAssignmentError(null);
    try {
      await updateRoles.mutateAsync({
        userId: user.id,
        currentRoleIds: roleIds,
        desiredRoleIds,
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
      width="sm"
      modal={false}
      disablePointerDismissal
    >
      {user ? (
        <div className="space-y-5">
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
                      triggerLabel={`Assign a role to ${user.email ?? "this user"}`}
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
