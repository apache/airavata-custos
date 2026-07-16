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

import type { UserIdentity } from "@/features/core/users/schemas";
import { Badge } from "@/shared/ui/badge";
import { identitySourceIcon, identitySourceLabel } from "./identities";

const VISIBLE_IDENTITY_COUNT = 2;

export function IdentitiesCell({
  identities,
  isLoading,
  hasError,
  expanded,
  onToggleExpand,
}: {
  identities: UserIdentity[];
  isLoading: boolean;
  hasError: boolean;
  expanded: boolean;
  onToggleExpand: () => void;
}) {
  const visible = expanded ? identities : identities.slice(0, VISIBLE_IDENTITY_COUNT);
  const hiddenCount = identities.length - visible.length;

  return (
    <div className="flex w-[220px] flex-wrap items-center gap-1.5">
      {isLoading ? (
        <span className="text-sm text-muted-foreground">Loading identities…</span>
      ) : hasError ? (
        <span className="text-sm text-muted-foreground">Identities unavailable</span>
      ) : identities.length === 0 ? (
        <span className="text-sm text-muted-foreground">No identities</span>
      ) : (
        visible.map((identity) => {
          const Icon = identitySourceIcon(identity.source);
          return (
            <Badge key={identity.id} variant="outline">
              <Icon data-icon="inline-start" />
              {identitySourceLabel(identity.source)}
            </Badge>
          );
        })
      )}
      {hiddenCount > 0 ? (
        <button
          type="button"
          onClick={onToggleExpand}
          className="text-xs font-medium text-muted-foreground underline-offset-2 hover:text-foreground hover:underline"
        >
          +{hiddenCount}
        </button>
      ) : expanded && identities.length > VISIBLE_IDENTITY_COUNT ? (
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
