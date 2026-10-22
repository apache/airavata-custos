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

import { useProject } from "@/features/core/projects/queries";
import type { ClusterAccountResponse } from "@/generated/core/zod.gen";
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
import { AllocationRef } from "./AllocationRef";

export type ReviewAction = "approve" | "deny";

export type ReviewClusterAccountDialogProps = {
  accounts: ClusterAccountResponse[];
  action: ReviewAction;
  bulk?: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (note: string) => void;
  isPending: boolean;
  error?: string | null;
  failures?: Record<string, string>;
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

const DETAILS =
  "grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1.5 rounded-lg bg-muted px-3 py-2 text-sm";
const NOTICE =
  "block rounded-lg bg-[color:var(--tone-warn-bg)] px-3 py-2 text-sm text-[color:var(--tone-warn-fg)]";

export function ReviewClusterAccountDialog({
  accounts,
  action,
  bulk = false,
  onOpenChange,
  onConfirm,
  isPending,
  error,
  failures = {},
}: ReviewClusterAccountDialogProps) {
  const [note, setNote] = React.useState("");
  const account = accounts[0];
  const open = account !== undefined;
  const { data: project } = useProject(account?.project_id);

  React.useEffect(() => {
    if (!open) setNote("");
  }, [open]);

  const n = accounts.length;
  const title = bulk ? `Approve ${n} cluster account${n === 1 ? "" : "s"}?` : copy[action].title;
  const button = bulk ? `Approve ${n}` : copy[action].button;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>
            {bulk && account
              ? `Accounts are created on ${account.cluster_name} for everyone below in ${account.allocation_name}. Your name is recorded as the approver for each one.`
              : copy[action].description}
          </DialogDescription>
        </DialogHeader>
        {account && bulk ? (
          <>
            <dl className={DETAILS}>
              <dt className="text-muted-foreground">Allocation</dt>
              <dd>
                <AllocationRef account={account} />
              </dd>
              <dt className="text-muted-foreground">Project</dt>
              <dd>{project?.title}</dd>
              <dt className="text-muted-foreground">PI</dt>
              <dd>{project ? (project.project_pi_display_name ?? project.project_pi_id) : null}</dd>
              <dt className="text-muted-foreground">Cluster</dt>
              <dd className="font-mono text-xs">{account.cluster_name}</dd>
            </dl>
            <ul
              aria-label="Accounts to approve"
              className="max-h-80 divide-y divide-border overflow-y-auto rounded-lg border border-border text-sm"
            >
              {accounts.map((a) => (
                <li key={a.id} className="px-3 py-2">
                  <div className="flex items-center justify-between gap-3">
                    <span>{a.display_name || a.email}</span>
                    <span className="font-mono text-xs text-muted-foreground">
                      {a.local_username}
                    </span>
                  </div>
                  {failures[a.id] ? (
                    <p className="text-xs text-[color:var(--tone-error-fg)]">{failures[a.id]}</p>
                  ) : null}
                </li>
              ))}
            </ul>
          </>
        ) : account ? (
          <dl className={DETAILS}>
            <dt className="text-muted-foreground">User</dt>
            <dd>
              {account.display_name || account.email}
              {account.display_name ? (
                <span className="text-muted-foreground"> ({account.email})</span>
              ) : null}
            </dd>
            {account.allocation_id ? (
              <>
                <dt className="text-muted-foreground">Allocation</dt>
                <dd>
                  <AllocationRef account={account} />
                </dd>
              </>
            ) : null}
            <dt className="text-muted-foreground">Cluster</dt>
            <dd className="font-mono text-xs">{account.cluster_name}</dd>
            <dt className="text-muted-foreground">Username</dt>
            <dd className="font-mono text-xs">{account.local_username}</dd>
          </dl>
        ) : null}
        {action === "deny" ? (
          <>
            <p className={NOTICE}>
              If ACCESS is waiting on this person, it is told no for every request, not only this
              allocation. Approving later still creates the account, but ACCESS is not told again.
            </p>
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
          </>
        ) : bulk ? null : (
          <output className={NOTICE}>
            Your name is recorded as the approver in the audit log.
          </output>
        )}
        {error ? <p className="text-sm text-[color:var(--tone-error-fg)]">{error}</p> : null}
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
            {isPending ? `${button}…` : button}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
