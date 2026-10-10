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

import { formatDate } from "@/shared/format";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { Badge } from "@/shared/ui/badge";
import { actionChipClass } from "@/shared/users-admin/permissions";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import { buildPrivilegeRows } from "../privileges";
import type { HeldRole, MyAccess } from "../queries";

const grantedLine = ({ granted_at, granted_by }: HeldRole) =>
  `Granted ${formatDate(granted_at)}${granted_by ? ` · by ${granted_by}` : ""}`;

const columnHeading = "mb-3 text-xs font-semibold tracking-wide text-muted-foreground uppercase";

export function AccessCard({ access }: { access: MyAccess }) {
  const [activeRole, setActiveRole] = useState<string | null>(null);
  const rows = buildPrivilegeRows(access);

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
          {rows.length === 0 ? (
            <p className="text-sm text-muted-foreground">No privileges granted.</p>
          ) : (
            <div className="space-y-0.5">
              {rows.map((row) => {
                const highlighted =
                  activeRole !== null &&
                  row.provenance === "role" &&
                  row.roleId === activeRole;
                return (
                  <div
                    key={row.prefix}
                    className={cn(
                      "grid grid-cols-[minmax(185px,auto)_1fr_auto] items-center gap-3 rounded-md px-2 py-1 transition-colors",
                      highlighted && "bg-[color:var(--brand-tint)]",
                    )}
                  >
                    <span className="font-mono text-[13px] whitespace-nowrap">{row.prefix}</span>
                    <span className="flex gap-1.5">
                      {row.actions.map((action) => (
                        <span
                          key={action}
                          className={cn(
                            "inline-flex h-6 items-center justify-center rounded px-2 text-xs font-medium",
                            actionChipClass(action),
                          )}
                        >
                          {action}
                        </span>
                      ))}
                    </span>
                    {row.provenanceLabel ? (
                      <span
                        className={cn(
                          "rounded-full px-2.5 py-0.5 text-[11px] whitespace-nowrap",
                          row.provenance === "direct"
                            ? "border border-[color:var(--brand)] bg-[color:var(--brand-tint)] font-semibold text-[color:var(--brand)]"
                            : "bg-[color:var(--brand-tint)] text-[color:var(--brand)]",
                        )}
                      >
                        {row.provenanceLabel}
                      </span>
                    ) : (
                      <span />
                    )}
                  </div>
                );
              })}
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
  roles: HeldRole[];
  activeRole: string | null;
  onHover: (roleId: string | null) => void;
}) {
  if (roles.length === 0) {
    return <p className="text-sm text-muted-foreground">No roles assigned.</p>;
  }
  return (
    <div className="space-y-2.5">
      {roles.map((held) => {
        const roleId = held.role.id;
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
              <span className="font-semibold">{held.role.name}</span>
              {held.role.is_system ? (
                <Badge className="bg-[color:var(--brand-tint)] text-[color:var(--brand)]">
                  SYSTEM
                </Badge>
              ) : null}
            </div>
            {held.role.description ? (
              <div className="text-[13px] text-muted-foreground">{held.role.description}</div>
            ) : null}
            <div className="mt-1.5 text-xs text-muted-foreground">{grantedLine(held)}</div>
          </div>
        );
      })}
    </div>
  );
}
