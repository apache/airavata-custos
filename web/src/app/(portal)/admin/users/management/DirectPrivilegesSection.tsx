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

import { XIcon } from "lucide-react";
import * as React from "react";
import { usePrivilegeCatalog } from "@/features/core/roles/queries";
import {
  useDirectPrivileges,
  useGrantPrivilege,
  useRevokePrivilege,
} from "@/features/core/users/queries";
import { Button } from "@/shared/ui/button";
import { Input } from "@/shared/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { DrawerSection } from "./DrawerSection";

// Direct grants sit beside role grants; the backend forbids self-revoking core:privileges:grant.
export function DirectPrivilegesSection({
  userId,
  email,
  isCurrentUser,
}: {
  userId: string;
  email: string;
  isCurrentUser: boolean;
}) {
  const grants = useDirectPrivileges(userId, true);
  const catalog = usePrivilegeCatalog(true);
  const grant = useGrantPrivilege(userId);
  const revoke = useRevokePrivilege(userId);
  const [selected, setSelected] = React.useState("");
  const [reason, setReason] = React.useState("");
  const held = new Set((grants.data ?? []).map((g) => g.privilege));
  const grantable = (catalog.data ?? []).filter((key) => !held.has(key));

  function handleGrant() {
    const privilege = grantable.find((key) => key === selected);
    if (!privilege) return;
    grant.mutate(
      { privilege, reason: reason.trim() || undefined },
      toastOnSuccess(`Granted ${privilege}`, () => {
        setSelected("");
        setReason("");
      }),
    );
  }

  return (
    <DrawerSection
      title="Direct Grants"
      isLoading={grants.isLoading}
      isError={grants.isError}
      empty={held.size === 0 && "No direct grants."}
      footer={
        <div className="mt-3 flex flex-wrap items-center gap-2">
          <Select value={selected} onValueChange={(value) => setSelected(value ?? "")}>
            <SelectTrigger aria-label="Privilege to grant" className="h-8 w-56 px-3">
              <SelectValue placeholder="Grant a privilege" />
            </SelectTrigger>
            <SelectContent>
              {grantable.map((key) => (
                <SelectItem key={key} value={key}>
                  {key}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Input
            aria-label="Grant reason"
            placeholder="Reason (optional)"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            className="h-8 w-40"
          />
          <Button size="sm" disabled={!selected || grant.isPending} onClick={handleGrant}>
            Grant
          </Button>
        </div>
      }
    >
      <ul className="space-y-1">
        {(grants.data ?? []).map((g) => {
          const key = g.privilege;
          const locked = isCurrentUser && key === "core:privileges:grant";
          return (
            <li key={g.id} className="flex items-center justify-between gap-2 text-sm">
              <span className="break-all font-mono text-foreground" title={g.reason ?? undefined}>
                {key}
              </span>
              <Button
                variant="ghost"
                size="icon-xs"
                aria-label={`Revoke ${key}`}
                disabled={locked || revoke.isPending}
                title={locked ? "You cannot revoke this from yourself" : undefined}
                onClick={() =>
                  confirmToast(`Revoke ${key} from ${email}?`, "Revoke", () =>
                    revoke.mutate(key, toastOnSuccess(`Revoked ${key}`)),
                  )
                }
              >
                <XIcon />
              </Button>
            </li>
          );
        })}
      </ul>
    </DrawerSection>
  );
}
