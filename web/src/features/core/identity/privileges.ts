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

import type { MyAccess } from "./queries";

export type ProvenanceKind = "role" | "direct" | "unknown";

export type PrivilegeRow = {
  prefix: string;
  actions: string[];
  provenance: ProvenanceKind;
  roleId?: string;
  provenanceLabel: string;
};

type Source = { roleId?: string; kind: ProvenanceKind; label: string };

const UNKNOWN_SOURCE: Source = { kind: "unknown", label: "" };

export function buildPrivilegeRows(access: MyAccess): PrivilegeRow[] {
  const sourceByKey = new Map<string, Source>();
  for (const rwp of access.roles) {
    const roleId = rwp.role.id ?? "";
    const label = rwp.role.name ?? "Role";
    for (const key of rwp.privileges) {
      if (!sourceByKey.has(key)) sourceByKey.set(key, { roleId, kind: "role", label });
    }
  }
  for (const grant of access.direct) {
    const key = grant.privilege;
    if (key && !sourceByKey.has(key)) {
      sourceByKey.set(key, { kind: "direct", label: "Direct grant" });
    }
  }

  // Without the admin-gated provenance reads, never claim "Direct grant".
  const fallback = access.provenance
    ? { kind: "direct" as const, label: "Direct grant" }
    : UNKNOWN_SOURCE;

  const rows = new Map<string, PrivilegeRow>();
  for (const key of [...access.privileges].sort()) {
    const lastColon = key.lastIndexOf(":");
    if (lastColon <= 0) continue;
    const prefix = key.slice(0, lastColon);
    const action = key.slice(lastColon + 1);
    const source = sourceByKey.get(key) ?? fallback;
    let row = rows.get(prefix);
    if (!row) {
      row = {
        prefix,
        actions: [],
        provenance: source.kind,
        roleId: source.roleId,
        provenanceLabel: source.label,
      };
      rows.set(prefix, row);
    }
    if (!row.actions.includes(action)) row.actions.push(action);
  }
  return [...rows.values()];
}
