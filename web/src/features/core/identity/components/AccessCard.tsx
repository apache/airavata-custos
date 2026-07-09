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

import { useState } from "react";
import { cn } from "@/lib/utils";
import { Badge } from "@/shared/ui/badge";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { buildPrivilegeGroups } from "../privileges";
import type { PrivilegeAction } from "../privileges";
import type { MyAccess } from "../queries";
import type { RoleWithPrivileges, UserRole } from "../schemas";

function formatDate(iso?: string): string | null {
  if (!iso) return null;
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return null;
  return date.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

function grantedLine(grant: UserRole): string | null {
  const when = formatDate(grant.granted_at);
  const by = grant.granted_by;
  if (when && by) return `Granted ${when} · by ${by}`;
  if (when) return `Granted ${when}`;
  if (by) return `Granted by ${by}`;
  return null;
}

const columnHeading = "mb-3 text-xs font-semibold tracking-wide text-muted-foreground uppercase";

function ActionChip({ action }: { action: PrivilegeAction }) {
  const read = action.action === "read";
  return (
    <span
      className={cn(
        "rounded-[5px] px-2 py-0.5 text-[11px] font-semibold",
        read
          ? "bg-muted text-muted-foreground"
          : "bg-[color:var(--brand-tint)] text-[color:var(--accent-foreground)]",
      )}
    >
      {action.label}
    </span>
  );
}

export function AccessCard({ access }: { access: MyAccess }) {
  const [activeRole, setActiveRole] = useState<string | null>(null);
  const groups = buildPrivilegeGroups(access);

  return (
    <Card className="gap-0 py-0">
      <CardHeader className="px-6 pt-5 pb-4">
        <CardTitle>Roles &amp; effective privileges</CardTitle>
        <CardDescription>
          Roles group privileges. Hover a role to see exactly what it grants.
        </CardDescription>
      </CardHeader>

      <div className="grid grid-cols-1 border-t border-border md:grid-cols-2">
        <section className="px-6 py-5 md:border-r md:border-border">
          <h4 className={columnHeading}>Roles</h4>
          <RolesColumn
            roles={access.roles}
            activeRole={activeRole}
            onHover={setActiveRole}
          />
        </section>

        <section className="border-t border-border px-6 py-5 md:border-t-0">
          <h4 className={columnHeading}>Effective privileges</h4>
          {groups.length === 0 ? (
            <p className="text-sm text-muted-foreground">No privileges granted.</p>
          ) : (
            <div className="space-y-4">
              {groups.map((group) => (
                <div key={group.domain}>
                  <div className="mb-1.5 text-xs font-semibold text-foreground">{group.label}</div>
                  {group.rows.map((row) => {
                    const highlighted =
                      activeRole !== null &&
                      row.provenance === "role" &&
                      row.roleId === activeRole;
                    return (
                      <div
                        key={row.resource}
                        title={row.rawKeys.join(" · ")}
                        className={cn(
                          "flex items-center justify-between gap-3 rounded-md px-2 py-1 transition-colors",
                          highlighted && "bg-[color:var(--brand-tint)]",
                        )}
                      >
                        <span className="min-w-[110px] font-medium">{row.resourceLabel}</span>
                        <span className="flex flex-1 gap-1.5">
                          {row.actions.map((action) => (
                            <ActionChip key={action.action} action={action} />
                          ))}
                        </span>
                        <span
                          className={cn(
                            "text-[11px] whitespace-nowrap",
                            row.provenance === "direct"
                              ? "font-semibold text-[color:var(--accent-foreground)]"
                              : "text-muted-foreground",
                          )}
                        >
                          {row.provenanceLabel}
                        </span>
                      </div>
                    );
                  })}
                </div>
              ))}
            </div>
          )}
        </section>
      </div>
    </Card>
  );
}

function RolesColumn({
  roles,
  activeRole,
  onHover,
}: {
  roles: RoleWithPrivileges[];
  activeRole: string | null;
  onHover: (roleId: string | null) => void;
}) {
  if (roles.length === 0) {
    return <p className="text-sm text-muted-foreground">No roles assigned.</p>;
  }
  return (
    <div className="space-y-2.5">
      {roles.map((rwp) => {
        const roleId = rwp.role.id ?? "";
        const granted = grantedLine(rwp.grant);
        return (
          <div
            key={roleId}
            data-role-id={roleId}
            onMouseEnter={() => onHover(roleId)}
            onMouseLeave={() => onHover(null)}
            className={cn(
              "rounded-lg border border-border p-3 transition-colors",
              activeRole === roleId && "border-brand bg-[color:var(--brand-tint)]",
            )}
          >
            <div className="mb-0.5 flex items-center gap-2">
              <span className="font-semibold">{rwp.role.name}</span>
              {rwp.role.is_system ? (
                <Badge className="bg-[color:var(--brand-tint)] text-[color:var(--accent-foreground)]">
                  SYSTEM
                </Badge>
              ) : null}
            </div>
            {rwp.role.description ? (
              <div className="text-[13px] text-muted-foreground">{rwp.role.description}</div>
            ) : null}
            {granted ? <div className="mt-1.5 text-xs text-muted-foreground">{granted}</div> : null}
          </div>
        );
      })}
    </div>
  );
}
