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

import { Link2 } from "lucide-react";
import { Badge } from "@/shared/ui/badge";
import { Card, CardDescription, CardHeader, CardTitle } from "@/shared/ui/card";
import type { UserIdentity } from "../schemas";

function formatDate(iso?: string): string | null {
  if (!iso) return null;
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return null;
  return date.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

export function IdentitiesCard({ identities }: { identities: UserIdentity[] }) {
  return (
    <Card className="gap-0 py-0">
      <CardHeader className="px-6 pt-5 pb-4">
        <CardTitle>External identities</CardTitle>
        <CardDescription>
          Accounts linked to your Custos user. Managed by your identity provider.
        </CardDescription>
      </CardHeader>
      {identities.length === 0 ? (
        <p className="border-t border-border px-6 py-6 text-sm text-muted-foreground">
          No identities linked to this account.
        </p>
      ) : (
        <ul className="divide-y divide-border border-t border-border">
          {identities.map((identity) => {
            const linked = formatDate(identity.created_at);
            return (
              <li
                key={identity.id ?? `${identity.source}-${identity.external_id}`}
                className="flex flex-wrap items-center gap-3 px-6 py-3.5"
              >
                <Link2 className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
                <span className="min-w-[88px] font-semibold">{identity.source ?? "—"}</span>
                <span className="flex-1 font-mono text-[13px]">
                  {identity.external_id || identity.email || "—"}
                </span>
                {identity.oidc_sub ? (
                  <Badge variant="secondary">OIDC</Badge>
                ) : (
                  <Badge variant="secondary">Registry</Badge>
                )}
                {linked ? (
                  <span className="ml-auto text-xs text-muted-foreground">Linked {linked}</span>
                ) : null}
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}
