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
import type { App } from "../api";
import { useUpdateApp } from "../queries";

export type EditAppDialogProps = {
  app: App;
  onClose: () => void;
};

export function EditAppDialog({ app, onClose }: EditAppDialogProps) {
  const update = useUpdateApp();
  const [name, setName] = React.useState(app.name);
  const [description, setDescription] = React.useState(app.description);
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
          <DialogTitle>Edit {app.name}</DialogTitle>
          <DialogDescription>
            The app key and actions cannot be changed. The token keeps working.
          </DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (canSubmit) {
              update.mutate({ id: app.id, name: name.trim(), description }, { onSuccess: onClose });
            }
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="edit-app-name">Name</Label>
            <Input id="edit-app-name" value={name} onChange={(e) => setName(e.target.value)} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="edit-app-description">Description</Label>
            <Input
              id="edit-app-description"
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
