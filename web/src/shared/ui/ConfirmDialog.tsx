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

import type { UseMutationResult } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import * as React from "react";
import { toast } from "sonner";
import { Button } from "./button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "./dialog";
import { GuardedButton } from "./GuardedButton";

export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  children,
  confirmLabel,
  pending,
  error,
  onConfirm,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  children: React.ReactNode;
  confirmLabel: string;
  pending: boolean;
  error: Error | null;
  onConfirm: () => void;
}) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{children}</DialogDescription>
        </DialogHeader>
        {error ? <p className="text-destructive text-sm">{error.message}</p> : null}
        <DialogFooter>
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button type="button" variant="destructive" disabled={pending} onClick={onConfirm}>
            {confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

// Deletes a record after confirmation, then leaves its now-missing page.
export function DeleteButton({
  noun,
  id,
  name,
  mutation,
  redirectTo,
  allowed,
  reason,
  requestOnly,
  children,
}: {
  noun: "Allocation" | "Project";
  id: string;
  name: string;
  mutation: UseMutationResult<void, Error, string>;
  redirectTo: string;
  allowed: boolean;
  reason: string;
  requestOnly?: boolean;
  children: React.ReactNode;
}) {
  const router = useRouter();
  const verb = requestOnly ? "Request deletion" : "Delete";
  const [open, setOpen] = React.useState(false);
  return (
    <>
      <GuardedButton
        variant="destructive"
        allowed={allowed}
        reason={reason}
        onClick={() => {
          mutation.reset();
          setOpen(true);
        }}
      >
        {verb}
      </GuardedButton>
      <ConfirmDialog
        open={open}
        onOpenChange={setOpen}
        title={requestOnly ? `Request deletion of ${name}?` : `Delete ${name}?`}
        confirmLabel={requestOnly ? verb : `Delete ${noun.toLowerCase()}`}
        pending={mutation.isPending}
        error={mutation.error}
        onConfirm={() =>
          mutation.mutate(id, {
            onSuccess: () => {
              if (requestOnly) {
                toast.success(`Deletion of ${noun.toLowerCase()} ${name} awaits admin approval.`);
                setOpen(false);
                return;
              }
              toast.success(`${noun} ${name} deleted.`);
              router.push(redirectTo);
            },
          })
        }
      >
        {children}
      </ConfirmDialog>
    </>
  );
}
