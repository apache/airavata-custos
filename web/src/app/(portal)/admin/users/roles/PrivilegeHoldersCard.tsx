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

import { Timestamp } from "@/features/core/audit/components/primitives/Timestamp";
import { usePrivilegeCatalog } from "@/features/core/roles/queries";
import { UserName } from "@/features/core/users/components/UserPicker";
import {
  useGrantPrivilege,
  usePrivilegeHolders,
  useRevokePrivilege,
  useUser,
} from "@/features/core/users/queries";
import type { PrivilegeKey, UserPrivilege } from "@/generated/core/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { formatDate } from "@/shared/format";
import { ErrorState } from "@/shared/ui/ErrorState";
import { idText, text } from "@/shared/ui/FormDialog";
import { Button } from "@/shared/ui/button";
import { keepFocus } from "@/shared/ui/focus";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { displayName } from "@/shared/users-admin/displayName";
import { idInput } from "@/shared/users-admin/idInput";
import { useSession } from "next-auth/react";
import * as React from "react";

type Revoke = ReturnType<typeof useRevokePrivilege>;

// Direct grants only; holders through a role show on the role cards above. Without users:read the
// holders are only counted, and a grant or revoke names its user by ID.
export function PrivilegeHoldersCard() {
  const canReadUsers = useAbility().can("read", "User");
  const currentUserId = useSession().data?.custosUserId;
  const catalogQuery = usePrivilegeCatalog(true);
  const catalog = catalogQuery.data ?? [];
  const [selected, setSelected] = React.useState("");
  const key = catalog.find((k) => k === selected);
  const holders = usePrivilegeHolders(key);
  const revoke = useRevokePrivilege();
  const list = React.useRef<HTMLUListElement>(null);
  const userInput = React.useRef<HTMLInputElement>(null);
  const rows = holders.data ?? [];

  function confirmRevoke(userId: string, privilege: PrivilegeKey, name: string) {
    confirmToast(
      `Revoke ${privilege} from ${userId === currentUserId ? "yourself" : name}?`,
      "Revoke",
      () =>
        toastOnSuccess(revoke.mutateAsync({ userId, key: privilege }), `Revoked ${privilege}`, () =>
          keepFocus(() => list.current?.querySelector("button") ?? userInput.current),
        ),
    );
  }

  return (
    <section className="space-y-3 rounded-md border bg-card p-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-sm font-semibold">Direct privilege holders</h2>
        <Select value={selected} onValueChange={(value) => setSelected(value ?? "")}>
          <SelectTrigger aria-label="Privilege" className="h-9 w-64 px-3">
            <SelectValue placeholder="Pick a privilege" />
          </SelectTrigger>
          <SelectContent>
            {catalog.map((k) => (
              <SelectItem key={k} value={k}>
                {k}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>
      {catalogQuery.isError ? (
        <ErrorState error={catalogQuery.error} onRetry={() => catalogQuery.refetch()} as="h3" />
      ) : !key ? null : (
        <>
          {holders.isLoading ? (
            <output className="block text-sm text-muted-foreground">Loading holders…</output>
          ) : holders.isError ? (
            <ErrorState error={holders.error} onRetry={() => holders.refetch()} as="h3" />
          ) : rows.length === 0 ? (
            <p className="text-sm text-muted-foreground">No user holds {key} directly.</p>
          ) : canReadUsers ? (
            <ul ref={list} className="space-y-1.5 text-sm">
              {rows.map((grant) => (
                <HolderRow
                  key={grant.id}
                  grant={grant}
                  locked={grant.user_id === currentUserId && key === "core:privileges:grant"}
                  revoke={revoke}
                  onRevoke={confirmRevoke}
                />
              ))}
            </ul>
          ) : (
            <p className="text-sm text-muted-foreground">
              {rows.length === 1 ? "1 user holds" : `${rows.length} users hold`} {key} directly.
            </p>
          )}
          <GrantForm
            privilege={key}
            userInput={userInput}
            revoke={canReadUsers ? undefined : revoke}
            onRevoke={(userId) => confirmRevoke(userId, key, "this user")}
          />
        </>
      )}
    </section>
  );
}

function HolderRow({
  grant,
  locked,
  revoke,
  onRevoke,
}: {
  grant: UserPrivilege;
  locked: boolean;
  revoke: Revoke;
  onRevoke: (userId: string, privilege: PrivilegeKey, name: string) => void;
}) {
  const user = useUser(grant.user_id).data;
  const name = user ? displayName(user) : "this user";
  return (
    <li className="flex items-center justify-between gap-2">
      <span className="font-medium text-foreground">
        <UserName id={grant.user_id} />
      </span>
      <span className="flex items-center gap-2">
        <span className="truncate text-xs text-muted-foreground">
          {grant.reason}
          {grant.reason ? " · " : null}
          <Timestamp iso={grant.granted_at}>{formatDate(grant.granted_at)}</Timestamp>
        </span>
        <Button
          variant="ghost"
          size="xs"
          aria-label={`Revoke ${grant.privilege} from ${name}`}
          pending={revoke.isPending && revoke.variables.userId === grant.user_id}
          disabled={locked}
          focusableWhenDisabled
          title={locked ? "You cannot revoke this from yourself." : undefined}
          onClick={() => onRevoke(grant.user_id, grant.privilege, name)}
        >
          Revoke
        </Button>
      </span>
    </li>
  );
}

// Revoke by ID is offered only when the holders cannot be listed by name.
function GrantForm({
  privilege,
  userInput,
  revoke,
  onRevoke,
}: {
  privilege: PrivilegeKey;
  userInput: React.RefObject<HTMLInputElement | null>;
  revoke: Revoke | undefined;
  onRevoke: (userId: string) => void;
}) {
  const grant = useGrantPrivilege();
  const form = React.useRef<HTMLFormElement>(null);
  return (
    <form
      ref={form}
      className="flex flex-wrap items-end gap-2 border-t border-border pt-3"
      onSubmit={(e) => {
        e.preventDefault();
        const data = new FormData(e.currentTarget);
        const reason = text(data, "reason");
        toastOnSuccess(
          grant.mutateAsync({
            userId: idText(data, "user"),
            privilege,
            reason: reason || undefined,
          }),
          `Granted ${privilege}`,
          () => form.current?.reset(),
        );
      }}
    >
      <div className="flex flex-col gap-1">
        <Label htmlFor="holder-user-id">User ID</Label>
        <Input
          ref={userInput}
          id="holder-user-id"
          name="user"
          required
          className="w-80 font-mono"
          {...idInput("user")}
        />
      </div>
      <div className="flex flex-col gap-1">
        <Label htmlFor="holder-reason">Reason</Label>
        <Input id="holder-reason" name="reason" placeholder="Optional" className="w-40" />
      </div>
      <Button type="submit" size="sm" pending={grant.isPending}>
        Grant
      </Button>
      {revoke ? (
        <Button
          type="button"
          size="sm"
          variant="outline"
          pending={revoke.isPending}
          onClick={() => {
            if (form.current?.reportValidity())
              onRevoke(idText(new FormData(form.current), "user"));
          }}
        >
          Revoke
        </Button>
      ) : null}
    </form>
  );
}
