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

// Title-case a slug segment: "unmapped" → "Unmapped", "amie" → "Amie" is
// then upper-cased for known acronym domains below.
function titleCase(slug: string): string {
  return slug
    .split(/[-_]/)
    .filter(Boolean)
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
}

const DOMAIN_LABELS: Record<string, string> = {
  core: "Core",
  amie: "AMIE",
};

export function domainLabel(domain: string): string {
  return DOMAIN_LABELS[domain] ?? titleCase(domain);
}

export function resourceLabel(resource: string): string {
  return titleCase(resource);
}

export type ProvenanceKind = "role" | "direct";

export type PrivilegeAction = {
  action: string;
  label: string;
  rawKey: string;
};

export type PrivilegeRow = {
  resource: string;
  resourceLabel: string;
  actions: PrivilegeAction[];
  rawKeys: string[];
  // Where this row's grant comes from: a role (with its id + name) or a direct grant.
  provenance: ProvenanceKind;
  roleId?: string;
  provenanceLabel: string;
};

export type PrivilegeGroup = {
  domain: string;
  label: string;
  rows: PrivilegeRow[];
};

type Source = { roleId?: string; kind: ProvenanceKind; label: string };

// Build the domain → resource → actions view from the caller's access payload.
// Every effective key is placed; its provenance is the first role that grants
// it, else the direct grant, else (defensively) "Direct grant".
export function buildPrivilegeGroups(access: MyAccess): PrivilegeGroup[] {
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

  const groups = new Map<string, Map<string, PrivilegeRow>>();
  for (const key of access.privileges) {
    const [domain, resource, action] = key.split(":");
    if (!domain || !resource || !action) continue;
    const source = sourceByKey.get(key) ?? { kind: "direct" as const, label: "Direct grant" };

    let byResource = groups.get(domain);
    if (!byResource) {
      byResource = new Map();
      groups.set(domain, byResource);
    }
    let row = byResource.get(resource);
    if (!row) {
      row = {
        resource,
        resourceLabel: resourceLabel(resource),
        actions: [],
        rawKeys: [],
        provenance: source.kind,
        roleId: source.roleId,
        provenanceLabel: source.label,
      };
      byResource.set(resource, row);
    }
    if (!row.actions.some((a) => a.action === action)) {
      row.actions.push({ action, label: titleCase(action), rawKey: key });
    }
    row.rawKeys.push(key);
  }

  return [...groups.entries()].map(([domain, byResource]) => ({
    domain,
    label: domainLabel(domain),
    rows: [...byResource.values()],
  }));
}
