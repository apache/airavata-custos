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

import { useRoleRows } from "@/features/core/roles/queries";
import { Button } from "@/shared/ui/button";
import { RoleCard } from "./RoleCard";

export function RolesGrid() {
  const rolesQuery = useRoleRows();

  if (rolesQuery.isPending) {
    return (
      <div className="rounded-md border bg-card p-6 text-sm text-muted-foreground">
        Loading roles...
      </div>
    );
  }

  if (rolesQuery.isError) {
    const error = rolesQuery.error;
    return (
      <div className="space-y-3 rounded-md border bg-card p-6">
        <p className="text-sm text-destructive">
          {error instanceof Error ? error.message : "Could not load roles."}
        </p>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => {
            rolesQuery.refetch();
          }}
        >
          Retry
        </Button>
      </div>
    );
  }

  const roles = rolesQuery.data;

  if (roles.length === 0) {
    return (
      <div className="rounded-md border bg-card p-6 text-sm text-muted-foreground">
        No roles have been created yet.
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {roles.map((role) => (
        <RoleCard key={role.id} role={role} />
      ))}
    </div>
  );
}
