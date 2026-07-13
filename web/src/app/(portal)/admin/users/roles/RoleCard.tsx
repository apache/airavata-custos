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

import { Pencil, ShieldCheck } from "lucide-react";
import type { RoleRow } from "@/features/core/roles/schemas";
import { Badge } from "@/shared/ui/badge";
import { Card, CardContent, CardHeader } from "@/shared/ui/card";
import type { PermissionKey } from "@/shared/users-admin/permissions";
import { PermissionMatrixEditor } from "./PermissionMatrixEditor";
import { RoleFormDialog } from "./RoleFormDialog";

export function RoleCard({ role, catalog }: { role: RoleRow; catalog: readonly PermissionKey[] }) {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <div>
            <div className="flex items-center gap-1.5 font-heading text-base font-semibold text-foreground">
              {role.name}
              {role.is_system ? (
                <ShieldCheck
                  className="size-3.5 text-muted-foreground"
                  aria-label="System role"
                />
              ) : null}
            </div>
            <p className="mt-1 text-sm text-muted-foreground">{role.description}</p>
          </div>
          <Badge variant="secondary" className="shrink-0">
            {role.memberCount} {role.memberCount === 1 ? "member" : "members"}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="border-t border-border" />

        <PermissionMatrixEditor
          permissions={role.privileges}
          catalog={catalog}
          onTogglePermission={() => undefined}
          editable={false}
        />

        <div className="border-t border-border" />

        <RoleFormDialog
          role={role}
          triggerRender={
            <button
              type="button"
              className="inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground hover:text-foreground"
            />
          }
          triggerContent={
            <>
              <Pencil className="size-3.5" />
              Edit
            </>
          }
        />
      </CardContent>
    </Card>
  );
}
