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

import { ClusterName } from "@/features/core/clusters/components/ClusterName";
import type { ComputeClusterUser } from "@/generated/core/types.gen";
import { Button } from "@/shared/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import { Label } from "@/shared/ui/label";
import * as React from "react";

export type ReviewAction = "approve" | "deny";

export type ReviewClusterAccountDialogProps = {
  account: ComputeClusterUser | null;
  // The account's user by name; omitted when this caller cannot name them.
  user?: React.ReactNode;
  action: ReviewAction;
  onOpenChange: (open: boolean) => void;
  onConfirm: (note: string) => void;
  isPending: boolean;
  // Raised inside the account drawer, so it names its own return target.
  finalFocus: () => HTMLElement | null;
};

const copy: Record<ReviewAction, { title: string; description: string; button: string }> = {
  approve: {
    title: "Approve cluster account",
    description:
      "This creates the account on the cluster and lets the user log in once provisioning finishes.",
    button: "Approve",
  },
  deny: {
    title: "Deny cluster account",
    description:
      "No account is created on the cluster. ACCESS is told no, with your note. Memberships are unchanged.",
    button: "Deny",
  },
};

export function ReviewClusterAccountDialog({
  account,
  user,
  action,
  onOpenChange,
  onConfirm,
  isPending,
  finalFocus,
}: ReviewClusterAccountDialogProps) {
  const [note, setNote] = React.useState("");
  const open = account !== null;

  React.useEffect(() => {
    if (!open) setNote("");
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange} finalFocus={finalFocus}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{copy[action].title}</DialogTitle>
          <DialogDescription>
            {copy[action].description}
            {action === "approve" && account?.approval_status === "DENIED"
              ? " Requests held on this account are answered."
              : null}
          </DialogDescription>
        </DialogHeader>
        {account ? (
          <dl className="grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1.5 rounded-lg bg-muted px-3 py-2 text-sm">
            {user ? (
              <>
                <dt className="text-muted-foreground">User</dt>
                <dd>{user}</dd>
              </>
            ) : null}
            <dt className="text-muted-foreground">Cluster</dt>
            <dd className="font-mono text-xs">
              <ClusterName id={account.compute_cluster_id} />
            </dd>
            <dt className="text-muted-foreground">Username</dt>
            <dd className="font-mono text-xs">{account.local_username}</dd>
          </dl>
        ) : null}
        {action === "deny" ? (
          <div className="space-y-2">
            <Label htmlFor="review-note">Note (required, kept in the audit log)</Label>
            <textarea
              id="review-note"
              rows={3}
              required
              className="w-full rounded-lg border border-input bg-transparent px-3 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
              value={note}
              onChange={(e) => setNote(e.target.value)}
            />
          </div>
        ) : (
          <p className="rounded-lg bg-[color:var(--tone-warn-bg)] px-3 py-2 text-sm text-[color:var(--tone-warn-fg)]">
            Your name is recorded as the approver in the audit log.
          </p>
        )}
        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            type="button"
            variant={action === "deny" ? "destructive" : "default"}
            pending={isPending}
            disabled={action === "deny" && note.trim() === ""}
            focusableWhenDisabled
            onClick={() => onConfirm(note.trim())}
          >
            {isPending ? `${copy[action].button}…` : copy[action].button}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
