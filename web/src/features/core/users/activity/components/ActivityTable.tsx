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
import { Button } from "@/shared/ui/button";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/shared/ui/table";
import { ArrowDown, ArrowUp } from "lucide-react";
import { activityActionLabel, formatLastLogin } from "../lib";
import type { ActivitySortKey, SortDirection, UserActivityRow } from "../schemas";
import { ActivityStatus } from "./ActivityStatus";

export function ActivityTable({
  rows,
  windowDays,
  page,
  pageSize,
  total,
  sort,
  direction,
  busy = false,
  onSort,
  onPage,
  onPageSize,
  onSelect,
}: {
  rows: UserActivityRow[];
  windowDays: number;
  page: number;
  pageSize: number;
  total: number;
  sort: ActivitySortKey;
  direction: SortDirection;
  busy?: boolean;
  onSort: (key: ActivitySortKey) => void;
  onPage: (page: number) => void;
  onPageSize: (size: number) => void;
  onSelect: (user: UserActivityRow) => void;
}) {
  function sortHeader(key: ActivitySortKey, label: string) {
    return (
      <TableHead
        className={key === "login_count" ? "text-right" : undefined}
        aria-sort={sort === key ? (direction === "asc" ? "ascending" : "descending") : "none"}
      >
        <button
          type="button"
          disabled={busy}
          className="inline-flex items-center gap-1 rounded uppercase tracking-wide focus-visible:outline focus-visible:outline-2"
          onClick={() => onSort(key)}
        >
          {label}
          {sort === key ? (
            direction === "asc" ? (
              <ArrowUp size={12} aria-hidden />
            ) : (
              <ArrowDown size={12} aria-hidden />
            )
          ) : null}
        </button>
      </TableHead>
    );
  }
  return (
    <>
      <Table className="[&_td]:px-4 [&_td]:py-3 [&_th]:px-4">
        <caption className="sr-only">User login activity</caption>
        <TableHeader className="bg-muted [&_th]:font-mono [&_th]:text-xs [&_th]:uppercase [&_th]:tracking-wide [&_th]:text-muted-foreground">
          <TableRow>
            {sortHeader("name", "User")}
            <TableHead>Status</TableHead>
            {sortHeader("last_login", "Last sign-in")}
            <TableHead className="text-right">In window</TableHead>
            {sortHeader("login_count", "All time")}
            <TableHead className="text-right">Action</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {rows.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="py-10 text-center text-muted-foreground">
                No users match this view.
              </TableCell>
            </TableRow>
          ) : (
            rows.map((user) => {
              const reviewAccess = activityActionLabel(user, windowDays) === "Review access";
              return (
                <TableRow key={user.user_id}>
                  <TableCell>
                    <p className="font-medium text-brand">{user.name}</p>
                    <p className="text-xs text-muted-foreground">
                      {user.email}
                      {user.role_names?.length ? ` · ${user.role_names.join(", ")}` : ""}
                    </p>
                  </TableCell>
                  <TableCell>
                    <ActivityStatus user={user} windowDays={windowDays} />
                  </TableCell>
                  <TableCell>
                    <span className="block">{formatLastLogin(user)}</span>
                    <span className="block font-mono text-xs text-muted-foreground tabular-nums">
                      {user.last_login
                        ? new Intl.DateTimeFormat("en-US", {
                            month: "short",
                            day: "numeric",
                            year: "numeric",
                            timeZone: "UTC",
                          }).format(new Date(user.last_login))
                        : "No recorded sign-in"}
                    </span>
                  </TableCell>
                  <TableCell className="text-right tabular-nums">
                    {user.window_login_count}
                  </TableCell>
                  <TableCell className="text-right tabular-nums text-muted-foreground">
                    {user.login_count}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      size="sm"
                      variant="outline"
                      disabled={busy}
                      onClick={() => onSelect(user)}
                      aria-label={`${reviewAccess ? "Review access" : "View activity"} for ${user.name}`}
                    >
                      {reviewAccess ? "Review access" : "View"}
                    </Button>
                  </TableCell>
                </TableRow>
              );
            })
          )}
        </TableBody>
      </Table>
      <div className="flex flex-wrap items-center justify-between gap-3 border-t px-4 py-3 text-sm text-muted-foreground">
        <span>
          {total === 0
            ? "No users"
            : `${(page - 1) * pageSize + 1}–${Math.min(page * pageSize, total)} of ${total}`}
        </span>
        <div className="flex items-center gap-2">
          <label>
            Rows per page{" "}
            <select
              className="rounded border bg-background p-1"
              aria-label="Rows per page"
              disabled={busy}
              value={pageSize}
              onChange={(event) => onPageSize(Number(event.target.value))}
            >
              <option value={10}>10</option>
              <option value={15}>15</option>
            </select>
          </label>
          <Button
            variant="outline"
            size="sm"
            disabled={busy || page <= 1}
            onClick={() => onPage(page - 1)}
          >
            Previous
          </Button>
          <span className="px-1 tabular-nums" aria-live="polite" aria-label="Current page">
            {busy ? "Loading page " : null}
            <span className="font-medium text-foreground">{page}</span> of{" "}
            {Math.max(1, Math.ceil(total / pageSize))}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={busy || page * pageSize >= total}
            onClick={() => onPage(page + 1)}
          >
            Next
          </Button>
        </div>
      </div>
    </>
  );
}
