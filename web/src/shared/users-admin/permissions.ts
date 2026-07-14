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

export type PermissionKey = string;

type PermissionParts = {
  section: string;
  action: string;
};

function splitPermission(key: PermissionKey): PermissionParts {
  const parts = key.split(":");
  const action = parts.pop() ?? key;
  return { section: parts.join(":") || key, action };
}

export function permissionRowsFor(
  permissions: PermissionKey[],
  catalog: readonly PermissionKey[] = permissions,
) {
  const held = new Set(permissions);
  const keys = Array.from(new Set([...catalog, ...permissions])).sort();
  const rows = new Map<
    string,
    {
      section: string;
      actions: Array<{ action: string; key: PermissionKey; active: boolean }>;
    }
  >();

  for (const key of keys) {
    const { section, action } = splitPermission(key);
    const row = rows.get(section) ?? { section, actions: [] };
    row.actions.push({ action, key, active: held.has(key) });
    rows.set(section, row);
  }

  return Array.from(rows.values()).map((row) => ({
    ...row,
    actions: row.actions.sort((a, b) => {
      const order = ["read", "write", "grant", "manage"];
      const ai = order.indexOf(a.action);
      const bi = order.indexOf(b.action);
      if (ai !== bi) {
        return (ai === -1 ? order.length : ai) - (bi === -1 ? order.length : bi);
      }
      return a.action.localeCompare(b.action);
    }),
  }));
}

export function rwStateFor(permissions: PermissionKey[], catalog?: readonly PermissionKey[]) {
  return permissionRowsFor(permissions, catalog).map((row) => ({
    section: row.section,
    read: row.actions.some((a) => a.action === "read" && a.active),
    write: row.actions.some((a) => a.action === "write" && a.active),
  }));
}

// Write implies read (you can't write what you can't read), so granting
// write also grants read, and revoking read also revokes write.
export function togglePermission(permissions: PermissionKey[], key: PermissionKey): PermissionKey[] {
  const held = new Set(permissions);
  const { action } = splitPermission(key);
  const isRead = action === "read";
  const isWrite = action === "write";
  const pairedKey = (
    isRead ? key.replace(/:read$/, ":write") : key.replace(/:write$/, ":read")
  ) as PermissionKey;

  if (held.has(key)) {
    held.delete(key);
    if (isRead) held.delete(pairedKey);
  } else {
    held.add(key);
    if (isWrite) held.add(pairedKey);
  }
  return Array.from(held);
}

// A user's effective permissions are the union of everything granted by
// each role they hold.
export function permissionsFromRoles(
  roles: Array<{ permissions: PermissionKey[] }>,
): PermissionKey[] {
  return Array.from(new Set(roles.flatMap((r) => r.permissions)));
}
