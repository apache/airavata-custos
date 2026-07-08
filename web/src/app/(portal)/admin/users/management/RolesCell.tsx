"use client";

import { Badge } from "@/shared/ui/badge";
import type { RoleRow, UserRow } from "../types";
import { RoleAssignMenu } from "./RoleAssignMenu";

const VISIBLE_ROLE_COUNT = 2;

export function RolesCell({
  user,
  roles,
  expanded,
  onToggleExpand,
  onToggleRole,
}: {
  user: UserRow;
  roles: RoleRow[];
  expanded: boolean;
  onToggleExpand: () => void;
  onToggleRole: (roleId: string) => void;
}) {
  const userRoles = user.roles;
  const visibleRoles = expanded ? userRoles : userRoles.slice(0, VISIBLE_ROLE_COUNT);
  const hiddenCount = userRoles.length - visibleRoles.length;
  const heldRoleIds = new Set(userRoles.map((r) => r.id ?? ""));

  return (
    <div className="flex w-[260px] flex-wrap items-center gap-1.5">
      {userRoles.length === 0 ? (
        <span className="text-sm text-muted-foreground">No roles</span>
      ) : (
        visibleRoles.map((role) => (
          <Badge key={role.id} variant="outline">
            {role.name}
          </Badge>
        ))
      )}
      {hiddenCount > 0 ? (
        <button
          type="button"
          onClick={onToggleExpand}
          className="text-xs font-medium text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
        >
          +{hiddenCount}
        </button>
      ) : expanded && userRoles.length > VISIBLE_ROLE_COUNT ? (
        <button
          type="button"
          onClick={onToggleExpand}
          className="text-xs font-medium text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
        >
          Show less
        </button>
      ) : null}

      <RoleAssignMenu
        roles={roles}
        heldRoleIds={heldRoleIds}
        onToggleRole={onToggleRole}
        triggerLabel={`Assign a role to ${user.email ?? "this user"}`}
      />
    </div>
  );
}
