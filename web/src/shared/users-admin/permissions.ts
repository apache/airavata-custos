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

import type { PrivilegeKey } from "@/generated/core/types.gen";

const ACTION_ORDER = ["read", "write", "grant", "manage"];

// Read and write chips get their own tones; other verbs (grant, manage) the accent tone.
export const actionChipClass = (action: string) =>
  ({
    read: "bg-[color:var(--tone-info-bg)] text-[color:var(--tone-info-fg)]",
    write: "bg-[color:var(--tone-ok-bg)] text-[color:var(--tone-ok-fg)]",
  })[action] ?? "bg-[color:var(--tone-accent-bg)] text-[color:var(--tone-accent-fg)]";

function splitPermission(key: PrivilegeKey) {
  const parts = key.split(":");
  const action = parts.pop() ?? key;
  return { section: parts.join(":") || key, action };
}

export function permissionRowsFor(
  permissions: PrivilegeKey[],
  catalog: readonly PrivilegeKey[] = permissions,
) {
  const held = new Set(permissions);
  const keys = Array.from(new Set([...catalog, ...permissions])).sort();
  const rows = new Map<
    string,
    {
      section: string;
      actions: Array<{ action: string; key: PrivilegeKey; active: boolean }>;
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
      const ai = ACTION_ORDER.indexOf(a.action);
      const bi = ACTION_ORDER.indexOf(b.action);
      if (ai !== bi) {
        return (ai === -1 ? ACTION_ORDER.length : ai) - (bi === -1 ? ACTION_ORDER.length : bi);
      }
      return a.action.localeCompare(b.action);
    }),
  }));
}

export function toggleId(ids: ReadonlySet<string>, id: string): Set<string> {
  const next = new Set(ids);
  if (!next.delete(id)) next.add(id);
  return next;
}

// Write implies read (you can't write what you can't read), so granting
// write also grants its catalog read key, and revoking read also revokes write.
export function togglePermission(
  permissions: PrivilegeKey[],
  key: PrivilegeKey,
  catalog: readonly PrivilegeKey[],
): PrivilegeKey[] {
  const { action } = splitPermission(key);
  const paired =
    action === "read" ? key.replace(/:read$/, ":write") : key.replace(/:write$/, ":read");
  if (permissions.includes(key)) {
    return permissions.filter((k) => k !== key && (action !== "read" || k !== paired));
  }
  const read = action === "write" ? catalog.filter((k) => k === paired) : [];
  return Array.from(new Set([...permissions, key, ...read]));
}
