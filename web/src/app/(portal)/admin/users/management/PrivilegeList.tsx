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

import { Badge } from "@/shared/ui/badge";
import { PermissionRW } from "@/shared/users-admin/PermissionRW";

export type GroupedPrivilege =
  | { kind: "rw"; key: string; read: boolean; write: boolean }
  | { kind: "standalone"; key: string };

export function groupPrivileges(privileges: string[]): GroupedPrivilege[] {
  const unique = Array.from(new Set(privileges.filter(Boolean))).sort();
  const rw = new Map<string, { read: boolean; write: boolean }>();
  const standalone: string[] = [];

  for (const key of unique) {
    const match = key.match(/^(.*):(read|write)$/);
    if (!match) {
      standalone.push(key);
      continue;
    }
    const [, section, action] = match;
    if (!section || !action) continue;
    const current = rw.get(section) ?? { read: false, write: false };
    current[action as "read" | "write"] = true;
    rw.set(section, current);
  }

  return [
    ...Array.from(rw, ([key, value]) => ({ kind: "rw" as const, key, ...value })),
    ...standalone.map((key) => ({ kind: "standalone" as const, key })),
  ];
}

export function PrivilegeList({ privileges }: { privileges: string[] }) {
  const grouped = groupPrivileges(privileges);
  if (grouped.length === 0) {
    return <p className="text-sm text-muted-foreground">No privileges granted.</p>;
  }

  return (
    <ul className="space-y-2">
      {grouped.map((privilege) => (
        <li key={privilege.key} className="flex items-center justify-between gap-3 text-sm">
          <span className="break-all font-mono text-foreground">{privilege.key}</span>
          {privilege.kind === "rw" ? (
            <PermissionRW read={privilege.read} write={privilege.write} />
          ) : (
            <Badge variant="outline">granted</Badge>
          )}
        </li>
      ))}
    </ul>
  );
}
