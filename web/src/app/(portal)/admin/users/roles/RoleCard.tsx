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

import type { RoleRow } from "@/features/core/roles/queries";
import { Badge } from "@/shared/ui/badge";
import { Card, CardContent, CardHeader } from "@/shared/ui/card";
import { Pencil, ShieldCheck } from "lucide-react";
import { PermissionMatrixEditor } from "./PermissionMatrixEditor";

export function RoleCard({ role, onEdit }: { role: RoleRow; onEdit: () => void }) {
  return (
    <Card className="min-w-0">
      <CardHeader>
        <div className="flex flex-wrap items-start justify-between gap-2">
          <div className="min-w-0">
            <h2 className="flex items-center gap-1.5 [overflow-wrap:anywhere] font-heading text-base font-semibold text-foreground">
              {role.name}
              {role.is_system ? (
                <ShieldCheck className="size-3.5 text-muted-foreground" aria-label="System role" />
              ) : null}
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">{role.description}</p>
          </div>
          <Badge variant="secondary" className="shrink-0">
            {role.holderIds.length} {role.holderIds.length === 1 ? "member" : "members"}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="border-t border-border" />

        <PermissionMatrixEditor permissions={role.privileges} editable={false} />

        <div className="border-t border-border" />
        <button
          type="button"
          onClick={onEdit}
          aria-label={`Edit ${role.name}`}
          className="inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground hover:text-foreground"
        >
          <Pencil className="size-3.5" />
          Edit
        </button>
      </CardContent>
    </Card>
  );
}
