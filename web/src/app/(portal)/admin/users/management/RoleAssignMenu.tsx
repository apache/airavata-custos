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

import { useRoleDetails } from "@/features/core/users/queries";
import type { Role } from "@/generated/core/types.gen";
import { Button } from "@/shared/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/shared/ui/dialog";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import { confirmToast } from "@/shared/ui/sonner";
import { toggleId } from "@/shared/users-admin/permissions";
import { Pencil } from "lucide-react";
import { useState } from "react";
import { PrivilegeList } from "./PrivilegeList";

function setsEqual(a: Set<string>, b: Set<string>) {
  if (a.size !== b.size) return false;
  for (const value of a) {
    if (!b.has(value)) return false;
  }
  return true;
}

export function RoleAssignMenu({
  roles,
  heldRoleIds,
  onSave,
  triggerLabel,
  isCurrentUser,
  isPending,
}: {
  roles: Role[];
  heldRoleIds: Set<string>;
  onSave: (roleIds: string[], reason: string | undefined, onSaved: () => void) => void;
  triggerLabel: string;
  isCurrentUser: boolean;
  isPending: boolean;
}) {
  const [open, setOpen] = useState(false);
  const [draftIds, setDraftIds] = useState<Set<string>>(new Set());
  const [reason, setReason] = useState("");
  const details = useRoleDetails(roles.map((role) => role.id), open);
  const detailById = new Map(details.details.map((detail) => [detail.role.id, detail]));

  function handleOpenChange(next: boolean) {
    setOpen(next);
    if (next) {
      setDraftIds(new Set(heldRoleIds));
      setReason("");
    }
  }

  function handleSave() {
    const removedIds = [...heldRoleIds].filter((roleId) => !draftIds.has(roleId));
    const removesOwnAdminAccess =
      isCurrentUser &&
      removedIds.some((roleId) =>
        detailById
          .get(roleId)
          ?.privileges?.some(
            (key) => key === "core:roles:manage" || key === "core:privileges:grant",
          ),
      );
    const confirmationMessage = removesOwnAdminAccess
      ? "This may remove your own ability to manage roles or grant privileges. Continue with these changes?"
      : isCurrentUser && removedIds.length > 0 && details.isError
        ? "Some role privileges are unavailable, so these changes may remove your own access. Continue?"
        : null;
    const save = () => onSave([...draftIds], reason.trim() || undefined, () => setOpen(false));
    if (confirmationMessage) confirmToast(confirmationMessage, "Continue", save);
    else save();
  }

  const hasChanges = !setsEqual(draftIds, heldRoleIds);

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger
        render={
          <button
            type="button"
            className="inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground hover:text-foreground"
          />
        }
      >
        <Pencil className="size-3.5" />
        Edit roles
      </DialogTrigger>
      <DialogContent className="gap-5 sm:max-w-2xl">
        <DialogHeader className="-mx-4 -mb-5 gap-1 border-b border-border px-4 pt-2 pb-4">
          <DialogTitle className="text-lg font-semibold">Manage roles</DialogTitle>
          <DialogDescription>{triggerLabel}</DialogDescription>
        </DialogHeader>

        <div className="-mx-4 max-h-[28rem] overflow-y-auto">
          <ul className="space-y-3 px-6 pt-3 pb-4">
            {roles.map((role) => {
              const assigned = draftIds.has(role.id);
              const privileges = detailById.get(role.id)?.privileges ?? [];

              return (
                <li key={role.id} className="overflow-hidden rounded-lg border border-border">
                  <div className="flex items-start justify-between gap-4 bg-muted/60 px-4 py-3">
                    <div>
                      <p className="font-heading text-base font-semibold text-foreground">
                        {role.name}
                      </p>
                      {role.description && (
                        <p className="mt-1 text-sm text-muted-foreground">{role.description}</p>
                      )}
                    </div>
                    <Button
                      type="button"
                      variant={assigned ? "secondary" : "default"}
                      size="sm"
                      className="shrink-0"
                      onClick={() =>
                        setDraftIds((prev) => toggleId(prev, role.id))
                      }
                      disabled={isPending}
                    >
                      {assigned ? "Unassign" : "Assign"}
                    </Button>
                  </div>

                  <div className="p-4">
                    {details.isLoading && privileges.length === 0 ? (
                      <p className="text-sm text-muted-foreground">Loading privileges…</p>
                    ) : details.isError && privileges.length === 0 ? (
                      <p className="text-sm text-muted-foreground">Privileges unavailable.</p>
                    ) : (
                      <PrivilegeList privileges={privileges} />
                    )}
                  </div>
                </li>
              );
            })}
          </ul>
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="role-assign-reason">Reason (optional)</Label>
          <Input
            id="role-assign-reason"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="Record why these roles are being granted"
            maxLength={500}
            disabled={isPending}
          />
        </div>

        <DialogFooter className="-mt-5">
          <Button
            variant="outline"
            onClick={() => setOpen(false)}
            type="button"
            disabled={isPending}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            type="button"
            disabled={!hasChanges || isPending || details.isLoading}
          >
            {isPending ? "Saving…" : "Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
