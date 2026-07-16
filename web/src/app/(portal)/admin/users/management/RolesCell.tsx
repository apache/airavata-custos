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

import { Badge } from "@/shared/ui/badge";
import type { Role } from "@/features/core/users/schemas";

const VISIBLE_ROLE_COUNT = 2;

export function RolesCell({
  roles,
  isLoading,
  hasError,
  expanded,
  onToggleExpand,
}: {
  roles: Role[];
  isLoading: boolean;
  hasError: boolean;
  expanded: boolean;
  onToggleExpand: () => void;
}) {
  const visibleRoles = expanded ? roles : roles.slice(0, VISIBLE_ROLE_COUNT);
  const hiddenCount = roles.length - visibleRoles.length;

  return (
    <div className="flex w-[260px] flex-wrap items-center gap-1.5">
      {isLoading ? (
        <span className="text-sm text-muted-foreground">Loading roles…</span>
      ) : hasError ? (
        <span className="text-sm text-muted-foreground">Roles unavailable</span>
      ) : roles.length === 0 ? (
        <span className="text-sm text-muted-foreground">No roles</span>
      ) : (
        visibleRoles.map((role) => (
          <Badge key={role.id} variant="outline">
            {role.name}
          </Badge>
        ))
      )}
      {hiddenCount > 0 ? (
        <button
          type="button"
          onClick={onToggleExpand}
          className="text-xs font-medium text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
        >
          +{hiddenCount}
        </button>
      ) : expanded && roles.length > VISIBLE_ROLE_COUNT ? (
        <button
          type="button"
          onClick={onToggleExpand}
          className="text-xs font-medium text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
        >
          Show less
        </button>
      ) : null}
    </div>
  );
}
