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
import type { ClusterAccount } from "../schemas";

export type ReviewAction = "approve" | "deny";

export type ReviewClusterAccountDialogProps = {
  account: ClusterAccount | null;
  action: ReviewAction;
  onOpenChange: (open: boolean) => void;
  onConfirm: (note: string) => void;
  isPending: boolean;
  error?: string | null;
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
      "No account is created on the cluster. The user keeps portal access and can be approved later. The reason is passed on to whoever requested the account.",
    button: "Deny",
  },
};

export function ReviewClusterAccountDialog({
  account,
  action,
  onOpenChange,
  onConfirm,
  isPending,
  error,
}: ReviewClusterAccountDialogProps) {
  const [note, setNote] = React.useState("");
  const open = account !== null;

  React.useEffect(() => {
    if (!open) setNote("");
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{copy[action].title}</DialogTitle>
          <DialogDescription>{copy[action].description}</DialogDescription>
        </DialogHeader>
        {account ? (
          <dl className="grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1.5 rounded-lg bg-muted px-3 py-2 text-sm">
            <dt className="text-muted-foreground">User</dt>
            <dd>
              {account.display_name || account.email}
              {account.display_name ? (
                <span className="text-muted-foreground"> ({account.email})</span>
              ) : null}
            </dd>
            <dt className="text-muted-foreground">Cluster</dt>
            <dd className="font-mono text-xs">{account.cluster_name}</dd>
            <dt className="text-muted-foreground">Username</dt>
            <dd className="font-mono text-xs">{account.local_username}</dd>
          </dl>
        ) : null}
        {action === "deny" ? (
          <div className="space-y-2">
            <Label htmlFor="review-note">Reason (required, kept in the audit log)</Label>
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
          <output className="block rounded-lg bg-[color:var(--tone-warn-bg)] px-3 py-2 text-sm text-[color:var(--tone-warn-fg)]">
            Your name is recorded as the approver in the audit log.
          </output>
        )}
        {error ? (
          <p className="text-sm text-[color:var(--tone-error-fg)]">{error}</p>
        ) : null}
        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            type="button"
            variant={action === "deny" ? "destructive" : "default"}
            disabled={isPending || (action === "deny" && note.trim() === "")}
            onClick={() => onConfirm(note.trim())}
          >
            {isPending ? `${copy[action].button}…` : copy[action].button}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
