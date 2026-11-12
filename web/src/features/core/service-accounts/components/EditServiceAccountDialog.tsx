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
import type { ServiceAccount } from "../api";
import { useUpdateServiceAccount } from "../queries";

export type EditServiceAccountDialogProps = {
  account: ServiceAccount;
  onClose: () => void;
};

export function EditServiceAccountDialog({ account, onClose }: EditServiceAccountDialogProps) {
  const update = useUpdateServiceAccount();
  const [name, setName] = React.useState(account.name);
  const [description, setDescription] = React.useState(account.description);
  const canSubmit = name.trim() !== "" && !update.isPending;

  return (
    <Dialog
      open
      onOpenChange={(next) => {
        if (!next && !update.isPending) onClose();
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit {account.name}</DialogTitle>
          <DialogDescription>
            The key and privileges cannot be changed. The token keeps working.
          </DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (canSubmit) {
              update.mutate(
                { id: account.id, name: name.trim(), description },
                { onSuccess: onClose },
              );
            }
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="edit-service-account-name">Name</Label>
            <Input
              id="edit-service-account-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-service-account-description">Description</Label>
            <Input
              id="edit-service-account-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
            />
          </div>
          {update.error ? (
            <p role="alert" className="text-sm text-destructive">
              {update.error.message}
            </p>
          ) : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose} disabled={update.isPending}>
              Cancel
            </Button>
            <Button type="submit" disabled={!canSubmit}>
              Save
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
