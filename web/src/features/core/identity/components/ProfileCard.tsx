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

import type { User } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { ErrorState } from "@/shared/ui/ErrorState";
import { edited, useFreshOnOpen } from "@/shared/ui/FormDialog";
import { CardSkeleton } from "@/shared/ui/Loading";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { Avatar, AvatarFallback } from "@/shared/ui/avatar";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { displayName, emailBeside, initials } from "@/shared/users-admin/displayName";
import { Lock, Pencil } from "lucide-react";
import { useEffect, useState } from "react";
import { useMe, useUpdateMyName } from "../queries";

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

export function ProfileCard({
  user,
  organizationName,
  showsIdentities = false,
}: {
  user: User;
  // Absent when the caller may not read organizations.
  organizationName?: string;
  // The page shows the identities card, so the lock note may mention identities.
  showsIdentities?: boolean;
}) {
  const [editing, setEditing] = useState(false);
  const [first, setFirst] = useState(user.first_name);
  const [middle, setMiddle] = useState(user.middle_name ?? "");
  const [last, setLast] = useState(user.last_name);
  const update = useUpdateMyName();
  const canEdit = useAbility().can("write", "User");
  const me = useMe();
  const fresh = useFreshOnOpen(me, editing);
  const current = me.data?.user ?? user;

  // biome-ignore lint/correctness/useExhaustiveDependencies: prefills once the opening read lands
  useEffect(() => {
    if (!editing || !fresh.ready) return;
    setFirst(current.first_name);
    setMiddle(current.middle_name ?? "");
    setLast(current.last_name);
  }, [editing, fresh.ready]);

  const stored = {
    first_name: current.first_name,
    middle_name: current.middle_name ?? "",
    last_name: current.last_name,
  };
  // PUT replaces the whole name, so every part is sent.
  const body = edited(stored, {
    first_name: first.trim(),
    middle_name: middle.trim(),
    last_name: last.trim(),
  });
  // First and last names are required, so only a changed one may not be blank.
  const blank = (["first_name", "last_name"] as const).some(
    (key) => body?.[key] === "" && stored[key] !== "",
  );

  function onSave() {
    if (!body) return setEditing(false);
    toastOnSuccess(update.mutateAsync({ path: { id: user.id }, body }), "Profile updated", () =>
      setEditing(false),
    );
  }

  return (
    <Card className="gap-0 divide-y divide-border py-0">
      <div className="flex items-center gap-4 px-6 py-5">
        <Avatar className="h-14 w-14 border-2 border-brand text-lg">
          <AvatarFallback>{initials(displayName(user))}</AvatarFallback>
        </Avatar>
        <div className="min-w-0">
          <div className="truncate text-base font-semibold">
            {displayName(user) || "Unknown user"}
          </div>
          <div className="truncate text-sm text-muted-foreground">
            {emailBeside(displayName(user), user.email)}
          </div>
        </div>
        <div className="ml-auto flex items-center gap-3">
          {user.status === "ACTIVE" ? <StatusBadge variant="active" /> : null}
          {canEdit ? (
            <Button
              variant="outline"
              size="sm"
              onClick={() => setEditing(true)}
              disabled={editing}
              focusableWhenDisabled
            >
              <Pencil className="h-3.5 w-3.5" />
              Edit name
            </Button>
          ) : null}
        </div>
      </div>

      {editing && !fresh.ready ? (
        <div className="px-6 py-5">
          {fresh.error ? (
            <ErrorState error={fresh.error} onRetry={() => me.refetch()} />
          ) : (
            <CardSkeleton />
          )}
        </div>
      ) : editing ? (
        <div className="px-6 py-5">
          <div className="grid gap-3 sm:grid-cols-3">
            <div>
              <Label htmlFor="settings-first-name">First name</Label>
              <Input
                id="settings-first-name"
                className="mt-1"
                maxLength={255}
                value={first}
                onChange={(e) => setFirst(e.target.value)}
              />
            </div>
            <div>
              <Label htmlFor="settings-middle-name">Middle name</Label>
              <Input
                id="settings-middle-name"
                className="mt-1"
                maxLength={255}
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
                maxLength={255}
                value={last}
                onChange={(e) => setLast(e.target.value)}
              />
            </div>
          </div>
          <div className="mt-4 flex items-center gap-2">
            <Button
              variant="brand"
              size="sm"
              onClick={onSave}
              pending={update.isPending}
              disabled={blank}
              focusableWhenDisabled
            >
              {update.isPending ? "Saving…" : "Save"}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setEditing(false)}
              disabled={update.isPending}
              focusableWhenDisabled
            >
              Cancel
            </Button>
            <span className="ml-auto flex items-center gap-1.5 text-xs text-muted-foreground">
              <Lock className="h-3 w-3" />
              {showsIdentities
                ? "Email and linked identities are managed by your identity provider."
                : "Email is managed by your identity provider."}
            </span>
          </div>
        </div>
      ) : null}

      <div className="grid grid-cols-[repeat(auto-fit,minmax(180px,1fr))] gap-4 px-6 py-5">
        <Field label="Username">
          <span className="font-mono text-[13px]">{user.email.split("@")[0] || "—"}</span>
        </Field>
        {organizationName ? <Field label="Organization">{organizationName}</Field> : null}
      </div>
    </Card>
  );
}
