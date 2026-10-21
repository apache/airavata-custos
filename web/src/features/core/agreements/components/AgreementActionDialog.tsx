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

import { Button } from "@/shared/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";

export type AgreementAction = "publish" | "remove" | "delete";

export const ACTION_COPY: Record<
  AgreementAction,
  (version: string) => { title: string; body: string; confirm: string; done: string }
> = {
  publish: (v) => ({
    title: `Publish version ${v}?`,
    body: "Users must accept this version at their next request. It replaces the current version.",
    confirm: `Publish version ${v}`,
    done: `Version ${v} published`,
  }),
  remove: (v) => ({
    title: `Remove version ${v}?`,
    body: "The row is kept. Users are no longer asked to accept it. The newest remaining published version becomes the one users must accept.",
    confirm: "Remove",
    done: `Version ${v} removed`,
  }),
  delete: (v) => ({
    title: `Delete draft ${v}?`,
    body: "This cannot be undone.",
    confirm: "Delete draft",
    done: `Draft ${v} deleted`,
  }),
};

export type AgreementActionDialogProps = {
  action: AgreementAction | null;
  version: string;
  isPending: boolean;
  error: string | null;
  onCancel: () => void;
  onConfirm: () => void;
};

export function AgreementActionDialog({
  action,
  version,
  isPending,
  error,
  onCancel,
  onConfirm,
}: AgreementActionDialogProps) {
  const copy = action ? ACTION_COPY[action](version) : null;
  return (
    <Dialog open={copy !== null} onOpenChange={(open) => !open && onCancel()}>
      <DialogContent>
        {copy ? (
          <>
            <DialogHeader>
              <DialogTitle>{copy.title}</DialogTitle>
              <DialogDescription>{copy.body}</DialogDescription>
            </DialogHeader>
            {error ? (
              <p role="alert" className="text-sm text-destructive">
                {error}
              </p>
            ) : null}
            <DialogFooter>
              <Button variant="outline" onClick={onCancel} disabled={isPending}>
                Cancel
              </Button>
              <Button
                variant={action === "publish" ? "default" : "destructive"}
                onClick={onConfirm}
                disabled={isPending}
              >
                {copy.confirm}
              </Button>
            </DialogFooter>
          </>
        ) : null}
      </DialogContent>
    </Dialog>
  );
}
