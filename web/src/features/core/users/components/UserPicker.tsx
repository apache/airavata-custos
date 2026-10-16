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

import { useUser, useUserPages } from "@/features/core/users/queries";
import type { User } from "@/generated/core/types.gen";
import { listOf } from "@/shared/api/listOf";
import { isGone } from "@/shared/api/useOnGone";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { displayName, emailBeside } from "@/shared/users-admin/displayName";
import * as React from "react";

// Users a merge or a new cluster account may target.
export const isAssignableUser = (u: User) => u.status === "ACTIVE" || u.status === "PENDING";

// A user's name for callers who may read users; nobody sees the raw id.
export function UserName({ id }: { id: string | undefined }) {
  const canRead = useAbility().can("read", "User");
  const query = useUser(canRead ? id : undefined);
  if (query.data) return <>{displayName(query.data)}</>;
  return <>{isGone(query.error) ? "Unknown user" : "—"}</>;
}

export function UserPicker({
  id,
  label,
  enabled,
  selected,
  onToggle,
  single = false,
  eligible = () => true,
}: {
  id: string;
  label: string;
  enabled: boolean;
  selected: ReadonlySet<string>;
  onToggle: (userId: string) => void;
  single?: boolean;
  eligible?: (user: User) => boolean;
}) {
  const [search, setSearch] = React.useState("");
  const query = useUserPages(enabled);
  const users = (query.data?.pages.flatMap((p) => listOf(p.items)) ?? []).filter(eligible);
  const needle = search.trim().toLowerCase();
  const matching = (
    needle
      ? users.filter((u) => `${displayName(u)} ${u.email}`.toLowerCase().includes(needle))
      : users
  ).sort((a, b) => displayName(a).localeCompare(displayName(b)));

  return (
    <div className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <Input
        id={id}
        type="search"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Search loaded users by name or email"
      />
      <ul className="max-h-40 space-y-1 overflow-y-auto rounded-md border p-2">
        {query.isError ? (
          <li className="px-1 py-1 text-sm text-muted-foreground">
            Couldn't load users. Try again.
          </li>
        ) : query.isLoading ? (
          <li className="px-1 py-1 text-sm text-muted-foreground">Loading users…</li>
        ) : matching.length === 0 ? (
          <li className="px-1 py-1 text-sm text-muted-foreground">No users match.</li>
        ) : (
          matching.map((u) => (
            <li key={u.id}>
              <label className="flex cursor-pointer items-center gap-2 rounded-sm px-1 py-1 text-sm hover:bg-muted">
                <input
                  type={single ? "radio" : "checkbox"}
                  name={id}
                  checked={selected.has(u.id)}
                  onChange={() => onToggle(u.id)}
                  className="size-4 rounded border-input"
                />
                <span className="font-medium text-foreground">{displayName(u)}</span>
                <span className="text-xs text-muted-foreground">
                  {emailBeside(displayName(u), u.email)}
                </span>
              </label>
            </li>
          ))
        )}
      </ul>
      <div className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
        <span>
          {!single && selected.size > 0
            ? `${selected.size} user${selected.size === 1 ? "" : "s"} selected`
            : null}
        </span>
        {query.hasNextPage ? (
          <Button
            variant="ghost"
            size="sm"
            type="button"
            pending={query.isFetchingNextPage}
            onClick={() => query.fetchNextPage()}
          >
            Load more users
          </Button>
        ) : null}
      </div>
    </div>
  );
}
