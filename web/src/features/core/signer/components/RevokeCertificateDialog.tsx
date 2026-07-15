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
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";

export type RevokeCertificateDialogProps = {
  serialNumber: number;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (reason: string) => void;
  isPending: boolean;
  error?: string | null;
};

export function RevokeCertificateDialog({
  serialNumber,
  open,
  onOpenChange,
  onSubmit,
  isPending,
  error,
}: RevokeCertificateDialogProps) {
  const [reason, setReason] = React.useState("");

  React.useEffect(() => {
    if (!open) setReason("");
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Revoke certificate?</DialogTitle>
          <DialogDescription>
            This revokes certificate serial {serialNumber}. This cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            const trimmed = reason.trim();
            if (!trimmed) return;
            onSubmit(trimmed);
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="revoke-reason">Reason</Label>
            <Input
              id="revoke-reason"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              placeholder="Why is this certificate being revoked?"
              required
            />
          </div>
          {error ? (
            <p className="text-sm text-[color:var(--custos-red-700)]">{error}</p>
          ) : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" variant="destructive" disabled={isPending || !reason.trim()}>
              {isPending ? "Revoking…" : "Confirm revoke"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
