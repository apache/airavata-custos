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

import * as React from "react";
import { useUser, useUserPages } from "@/features/core/users/queries";
import { useAbility } from "@/shared/casl/AbilityProvider";
import type { User } from "@/generated/core/types.gen";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";

export function fullNameFor(user: User): string {
  const name = [user.first_name, user.last_name].filter(Boolean).join(" ");
  return name || user.email;
}

// A user's name for callers who may read users; everyone else sees the id.
export function UserName({ id }: { id: string | undefined }) {
  const canRead = useAbility().can("read", "User");
  const user = useUser(canRead ? id : undefined).data;
  return <>{user ? fullNameFor(user) : id}</>;
}

export function UserPicker({
  id,
  label,
  enabled,
  selected,
  onToggle,
  single = false,
  excludeId,
}: {
  id: string;
  label: string;
  enabled: boolean;
  selected: ReadonlySet<string>;
  onToggle: (userId: string) => void;
  single?: boolean;
  excludeId?: string;
}) {
  const [search, setSearch] = React.useState("");
  const query = useUserPages(enabled);
  const users = (query.data?.pages.flatMap((p) => p.items) ?? []).filter((u) => u.id !== excludeId);
  const needle = search.trim().toLowerCase();
  const matching = needle
    ? users.filter((u) => `${fullNameFor(u)} ${u.email}`.toLowerCase().includes(needle))
    : users;

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
            Could not load users: {query.error.message}
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
                <span className="font-medium text-foreground">{fullNameFor(u)}</span>
                <span className="text-xs text-muted-foreground">{u.email}</span>
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
            disabled={query.isFetchingNextPage}
            onClick={() => query.fetchNextPage()}
          >
            Load more users
          </Button>
        ) : null}
      </div>
    </div>
  );
}
