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
import { toast } from "sonner";
import { UserPicker, fullNameFor } from "@/features/core/users/components/UserPicker";
import {
  type UserManagementRow,
  useMergeUsers,
  useUpdateUserStatus,
} from "@/features/core/users/queries";
import type { UserStatus } from "@/generated/core/types.gen";
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
import { DrawerSection } from "./DrawerSection";

// MERGED is set only by a merge, never picked directly.
const SETTABLE_STATUSES: UserStatus[] = ["PENDING", "ACTIVE", "INACTIVE", "SUSPENDED"];

export function UserStatusSection({
  user,
  canWrite,
  onMerged,
}: {
  user: UserManagementRow;
  canWrite: boolean;
  onMerged: () => void;
}) {
  const updateStatus = useUpdateUserStatus();
  const merged = user.status === "MERGED";

  function handleStatus(value: string | null) {
    const status = SETTABLE_STATUSES.find((s) => s === value);
    if (!status || status === user.status) return;
    updateStatus.mutate(
      { path: { id: user.id }, body: { status } },
      toastOnSuccess(`Status set to ${status}`),
    );
  }

  return (
    <DrawerSection title="Status">
      <div className="flex flex-wrap items-center gap-2">
        {canWrite && !merged ? (
          <>
            <Select value={user.status} onValueChange={handleStatus}>
              <SelectTrigger
                aria-label="User status"
                className="h-8 w-36 px-3"
                disabled={updateStatus.isPending}
              >
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {SETTABLE_STATUSES.map((status) => (
                  <SelectItem key={status} value={status}>
                    {status}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <MergeUserDialog user={user} onMerged={onMerged} />
          </>
        ) : (
          <span className="text-sm text-foreground">{user.status}</span>
        )}
      </div>
    </DrawerSection>
  );
}

function MergeUserDialog({ user, onMerged }: { user: UserManagementRow; onMerged: () => void }) {
  const merge = useMergeUsers();
  const [open, setOpen] = React.useState(false);
  const [survivorId, setSurvivorId] = React.useState("");

  function handleMerge() {
    if (!survivorId) return;
    confirmToast(
      `Merge ${fullNameFor(user)} into the selected user? Identities, cluster accounts, PI projects and memberships move over and ${fullNameFor(user)} is marked MERGED. This cannot be undone.`,
      "Merge",
      () =>
        merge.mutate(
          { surviving_user_id: survivorId, retiring_user_id: user.id },
          {
            onSuccess: (survivor) => {
              toast.success(`Merged into ${fullNameFor(survivor)}`);
              setOpen(false);
              onMerged();
            },
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
            Choose the user that {fullNameFor(user)} is merged into.
          </DialogDescription>
        </DialogHeader>
        <UserPicker
          id="merge-survivor"
          label="Surviving user"
          enabled={open}
          single
          excludeId={user.id}
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
            disabled={!survivorId || merge.isPending}
            onClick={handleMerge}
          >
            {merge.isPending ? "Merging…" : "Merge"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
