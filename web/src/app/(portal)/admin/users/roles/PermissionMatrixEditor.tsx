"use client";

import { PermissionRW } from "@/shared/users-admin/PermissionRW";
import type { PermissionKey } from "@/shared/users-admin/permissions";
import { rwStateFor } from "@/shared/users-admin/permissions";

export function PermissionMatrixEditor({
  permissions,
  onTogglePermission,
  editable = true,
}: {
  permissions: PermissionKey[];
  onTogglePermission: (permission: PermissionKey) => void;
  editable?: boolean;
}) {
  const rwPermissions = rwStateFor(permissions);

  return (
    <div>
      <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
        Effective Privileges
      </h4>
      <ul className="space-y-2">
        {rwPermissions.map((p) => (
          <li key={p.section} className="flex items-center justify-between text-sm">
            <span className="font-mono text-foreground">{p.section}</span>
            <PermissionRW
              read={p.read}
              write={p.write}
              onToggleRead={editable ? () => onTogglePermission(`${p.section}:read`) : undefined}
              onToggleWrite={
                editable ? () => onTogglePermission(`${p.section}:write`) : undefined
              }
            />
          </li>
        ))}
      </ul>
    </div>
  );
}
