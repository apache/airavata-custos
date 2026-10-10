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

import { cn } from "@/lib/utils";
import type { PrivilegeKey } from "@/generated/core/types.gen";
import { actionChipClass, permissionRowsFor } from "@/shared/users-admin/permissions";

const INACTIVE_CHIP_CLASS = "border border-border text-muted-foreground";

export function PermissionMatrixEditor({
  permissions,
  catalog,
  onTogglePermission,
  editable = true,
}: {
  permissions: PrivilegeKey[];
  catalog?: readonly PrivilegeKey[];
  onTogglePermission?: (permission: PrivilegeKey) => void;
  editable?: boolean;
}) {
  const rows = permissionRowsFor(permissions, catalog);
  const canEdit = editable && onTogglePermission;

  return (
    <div>
      <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">
        Effective Privileges
      </h4>
      <ul className="space-y-2">
        {rows.map((row) => (
          <li key={row.section} className="flex items-center justify-between gap-3 text-sm">
            <span className="font-mono text-foreground">{row.section}</span>
            <div className="flex flex-wrap justify-end gap-1">
              {row.actions.map((privilege) => {
                const className = cn(
                  "inline-flex h-6 items-center justify-center rounded px-2 text-xs font-medium",
                  privilege.active
                    ? actionChipClass(privilege.action)
                    : INACTIVE_CHIP_CLASS,
                  canEdit && "cursor-pointer transition-transform hover:scale-105",
                );
                if (!canEdit) {
                  return (
                    <span key={privilege.key} title={privilege.key} className={className}>
                      {privilege.action}
                    </span>
                  );
                }
                return (
                  <button
                    key={privilege.key}
                    type="button"
                    title={privilege.key}
                    aria-pressed={privilege.active}
                    onClick={() => onTogglePermission(privilege.key)}
                    className={className}
                  >
                    {privilege.action}
                  </button>
                );
              })}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
