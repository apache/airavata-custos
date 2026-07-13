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

import { useState } from "react";
import { Lock, Pencil } from "lucide-react";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { useUpdateMyName } from "../queries";
import type { UserProfile } from "../schemas";

function displayName(user: UserProfile): string {
  const full = [user.first_name, user.middle_name, user.last_name]
    .filter(Boolean)
    .join(" ")
    .trim();
  return full || user.email || "Unknown user";
}

function initials(user: UserProfile): string {
  const source = displayName(user);
  const parts = source.split(/\s+/).filter(Boolean);
  const first = parts.at(0) ?? source;
  const last = parts.at(-1) ?? "";
  const chars = parts.length >= 2 ? `${first.charAt(0)}${last.charAt(0)}` : source.slice(0, 2);
  return chars.toUpperCase();
}

function formatDate(iso?: string): string | null {
  if (!iso) return null;
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return null;
  return date.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <div className="mb-0.5 text-[11px] font-semibold tracking-wide text-muted-foreground uppercase">
        {label}
      </div>
      <div className="font-medium">{children}</div>
    </div>
  );
}

export function ProfileCard({ user }: { user: UserProfile }) {
  const [editing, setEditing] = useState(false);
  const [first, setFirst] = useState(user.first_name ?? "");
  const [middle, setMiddle] = useState(user.middle_name ?? "");
  const [last, setLast] = useState(user.last_name ?? "");
  const update = useUpdateMyName(user.id);

  function openEdit() {
    setFirst(user.first_name ?? "");
    setMiddle(user.middle_name ?? "");
    setLast(user.last_name ?? "");
    setEditing(true);
  }

  function onSave() {
    update.mutate(
      { first_name: first, middle_name: middle, last_name: last },
      { onSuccess: () => setEditing(false) },
    );
  }

  const memberSince = formatDate(undefined);

  return (
    <Card className="gap-0 divide-y divide-border py-0">
      <div className="flex items-center gap-4 px-6 py-5">
        <Avatar className="h-14 w-14 border-2 border-brand text-lg">
          <AvatarFallback>{initials(user)}</AvatarFallback>
        </Avatar>
        <div className="min-w-0">
          <div className="truncate text-base font-semibold">{displayName(user)}</div>
          <div className="truncate text-sm text-muted-foreground">{user.email}</div>
        </div>
        <div className="ml-auto flex items-center gap-3">
          {user.status === "ACTIVE" ? <StatusBadge variant="active" /> : null}
          <Button variant="outline" size="sm" onClick={openEdit} disabled={editing}>
            <Pencil className="h-3.5 w-3.5" />
            Edit name
          </Button>
        </div>
      </div>

      {editing ? (
        <div className="px-6 py-5">
          <div className="grid gap-3 sm:grid-cols-3">
            <div>
              <Label htmlFor="settings-first-name">First name</Label>
              <Input
                id="settings-first-name"
                className="mt-1"
                value={first}
                onChange={(e) => setFirst(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="settings-middle-name">Middle name</Label>
              <Input
                id="settings-middle-name"
                className="mt-1"
                placeholder="Optional"
                value={middle}
                onChange={(e) => setMiddle(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="settings-last-name">Last name</Label>
              <Input
                id="settings-last-name"
                className="mt-1"
                value={last}
                onChange={(e) => setLast(e.target.value)}
              />
            </div>
          </div>
          <div className="mt-4 flex items-center gap-2">
            <Button variant="brand" size="sm" onClick={onSave} disabled={update.isPending}>
              {update.isPending ? "Saving…" : "Save"}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setEditing(false)}
              disabled={update.isPending}
            >
              Cancel
            </Button>
            <span className="ml-auto flex items-center gap-1.5 text-xs text-muted-foreground">
              <Lock className="h-3 w-3" />
              Email and linked identities are managed by your identity provider.
            </span>
          </div>
        </div>
      ) : null}

      <div className="grid grid-cols-[repeat(auto-fit,minmax(180px,1fr))] gap-4 px-6 py-5">
        <Field label="Username">
          <span className="font-mono text-[13px]">{user.email?.split("@")[0] ?? "—"}</span>
        </Field>
        <Field label="Organization">{user.organization_id ?? "—"}</Field>
        <Field label="Member since">{memberSince ?? "—"}</Field>
      </div>
    </Card>
  );
}
