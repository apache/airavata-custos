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
import type { CreateOrganizationPayload } from "../schemas";

export type CreateOrganizationDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (payload: CreateOrganizationPayload) => void;
  isPending: boolean;
  error?: string | null;
};

export function CreateOrganizationDialog({
  open,
  onOpenChange,
  onSubmit,
  isPending,
  error,
}: CreateOrganizationDialogProps) {
  const [name, setName] = React.useState("");
  const [originatedId, setOriginatedId] = React.useState("");

  React.useEffect(() => {
    if (!open) {
      setName("");
      setOriginatedId("");
    }
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create organization</DialogTitle>
          <DialogDescription>
            Register an institution or resource provider.
          </DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (!name.trim()) return;
            const originated = originatedId.trim();
            onSubmit({ name: name.trim(), originated_id: originated || undefined });
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="create-org-name">Name</Label>
            <Input
              id="create-org-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Georgia Institute of Technology"
              required
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="create-org-originated-id">Originated ID</Label>
            <Input
              id="create-org-originated-id"
              value={originatedId}
              onChange={(e) => setOriginatedId(e.target.value)}
              placeholder="Optional, e.g. ACCESS org code"
            />
          </div>
          {error ? (
            <p className="text-sm text-[color:var(--custos-red-700)]">{error}</p>
          ) : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isPending || !name.trim()}>
              {isPending ? "Creating…" : "Create organization"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
