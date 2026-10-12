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

import { UserPicker, isAssignableUser } from "@/features/core/users/components/UserPicker";
import {
  type UserManagementRow,
  useMergeUsers,
  useUpdateUserStatus,
  useUserPages,
} from "@/features/core/users/queries";
import type { User, UserStatus } from "@/generated/core/types.gen";
import { enumLabel } from "@/shared/format";
import { Button } from "@/shared/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { confirmToast, toastOnSuccess } from "@/shared/ui/sonner";
import { displayName } from "@/shared/users-admin/displayName";
import * as React from "react";
import { DrawerSection } from "./DrawerSection";

// MERGED is set by a merge, so it is offered only as a merged user's current status.
const SETTABLE_STATUSES: UserStatus[] = ["PENDING", "ACTIVE", "INACTIVE", "SUSPENDED"];
const ENDS_SIGN_IN: Partial<Record<UserStatus, string>> = {
  SUSPENDED: "Suspend",
  INACTIVE: "Inactivate",
};

export function UserStatusSection({
  user,
  canWrite,
  isCurrentUser,
  onMerged,
}: {
  user: UserManagementRow;
  canWrite: boolean;
  isCurrentUser: boolean;
  onMerged: () => void;
}) {
  const updateStatus = useUpdateUserStatus((status) => isCurrentUser && status in ENDS_SIGN_IN);
  const triggerId = React.useId();
  const statuses: UserStatus[] =
    user.status === "MERGED" ? ["MERGED", ...SETTABLE_STATUSES] : SETTABLE_STATUSES;

  function handleStatus(value: string | null) {
    const status = statuses.find((s) => s === value);
    if (!status || status === user.status) return;
    const apply = () =>
      toastOnSuccess(
        updateStatus.mutateAsync({ path: { id: user.id }, body: { status } }),
        `Status set to ${enumLabel(status).toLowerCase()}`,
      );
    const verb = ENDS_SIGN_IN[status];
    if (!verb && user.status !== "ACTIVE" && user.status !== "MERGED") return void apply();
    const name = displayName(user);
    // The chosen option unmounts, so the confirmation's Cancel returns to the select instead.
    document.getElementById(triggerId)?.focus();
    confirmToast(
      verb
        ? isCurrentUser
          ? `${verb} yourself? You will be signed out and can no longer sign in.`
          : `${verb} ${name}? ${name} can no longer sign in.`
        : user.status === "MERGED"
          ? `Reactivate ${name}? ${name} was merged into another user.`
          : `Set ${isCurrentUser ? "yourself" : name} to ${enumLabel(status).toLowerCase()}?`,
      verb ?? (user.status === "MERGED" ? "Reactivate" : "Change status"),
      apply,
    );
  }

  return (
    <DrawerSection title="Status">
      <div className="flex flex-wrap items-center gap-2">
        {canWrite ? (
          <Select
            value={user.status}
            onValueChange={handleStatus}
            readOnly={updateStatus.isPending}
          >
            <SelectTrigger id={triggerId} aria-label="User status" className="h-8 w-36 px-3">
              <SelectValue>{enumLabel}</SelectValue>
            </SelectTrigger>
            <SelectContent>
              {statuses.map((status) => (
                <SelectItem key={status} value={status}>
                  {enumLabel(status)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : (
          <span className="text-sm text-foreground">{enumLabel(user.status)}</span>
        )}
        {canWrite && !isCurrentUser && isAssignableUser(user) ? (
          <MergeUserDialog user={user} onMerged={onMerged} />
        ) : null}
      </div>
    </DrawerSection>
  );
}

function MergeUserDialog({
  user,
  onMerged,
}: {
  user: UserManagementRow;
  onMerged: () => void;
}) {
  const merge = useMergeUsers();
  const [open, setOpen] = React.useState(false);
  const [survivorId, setSurvivorId] = React.useState("");
  // The picker's own pages, so naming the survivor costs no request.
  const survivor = useUserPages(open)
    .data?.pages.flatMap((page) => page.items ?? [])
    .find((u) => u.id === survivorId);

  function handleMerge() {
    if (!survivorId) return;
    // Two users of one name are told apart by email.
    const sameName = survivor && displayName(survivor) === displayName(user);
    const label = (u: User) => (sameName ? `${displayName(u)} (${u.email})` : displayName(u));
    const into = survivor ? label(survivor) : "the selected user";
    confirmToast(
      `Merge ${label(user)} into ${into}? Identities, cluster accounts, PI projects and memberships move over, and ${label(user)}'s cluster accounts on clusters where ${into} already has one are deleted. Roles and direct privileges stay with ${label(user)}, who is marked merged. ${label(user)} can no longer sign in. This cannot be undone.`,
      "Merge",
      () =>
        toastOnSuccess(
          merge.mutateAsync({ surviving_user_id: survivorId, retiring_user_id: user.id }),
          (survivor) => `Merged into ${label(survivor)}`,
          () => {
            setOpen(false);
            onMerged();
          },
        ),
    );
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <Button
        variant="outline"
        size="sm"
        onClick={() => {
          setSurvivorId("");
          setOpen(true);
        }}
      >
        Merge into…
      </Button>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Merge user</DialogTitle>
          <DialogDescription>
            Choose the user that {displayName(user)} is merged into.
          </DialogDescription>
        </DialogHeader>
        <UserPicker
          id="merge-survivor"
          label="Surviving user"
          enabled={open}
          single
          eligible={(u) => u.id !== user.id && isAssignableUser(u)}
          selected={new Set([survivorId])}
          onToggle={setSurvivorId}
        />
        <DialogFooter>
          <Button variant="outline" type="button" onClick={() => setOpen(false)}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            type="button"
            pending={merge.isPending}
            disabled={!survivorId}
            focusableWhenDisabled
            onClick={handleMerge}
          >
            {merge.isPending ? "Merging…" : "Merge"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
