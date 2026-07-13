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
import type { PermissionKey } from "@/shared/users-admin/permissions";
import { DEFAULT_PERMISSION_KEYS, permissionRowsFor } from "@/shared/users-admin/permissions";

export function PermissionMatrixEditor({
  permissions,
  catalog = DEFAULT_PERMISSION_KEYS,
  onTogglePermission,
  editable = true,
}: {
  permissions: PermissionKey[];
  catalog?: readonly PermissionKey[];
  onTogglePermission: (permission: PermissionKey) => void;
  editable?: boolean;
}) {
  const rows = permissionRowsFor(permissions, catalog);

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
                    ? "bg-[color:var(--custos-blue-50)] text-[color:var(--custos-blue-700)]"
                    : "bg-[color:var(--custos-gray-100)] text-[color:var(--custos-gray-400)]",
                  editable && "cursor-pointer transition-transform hover:scale-105",
                );
                if (!editable) {
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
