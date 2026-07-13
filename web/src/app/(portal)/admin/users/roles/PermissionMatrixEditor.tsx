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
