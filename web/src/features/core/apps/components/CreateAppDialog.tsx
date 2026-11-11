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
import { StatusBadge } from "@/shared/ui/StatusBadge";
import { useAppConnectors, useCreateApp } from "../queries";
import { TokenReveal } from "./TokenReveal";

const APP_KEY = /^[a-z0-9]+(-[a-z0-9]+)*$/;
const APP_KEY_MAX_LENGTH = 50;

export type CreateAppDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

export function CreateAppDialog({ open, onOpenChange }: CreateAppDialogProps) {
  const connectors = useAppConnectors();
  const create = useCreateApp();
  const { reset } = create;
  const [key, setKey] = React.useState("");
  const [name, setName] = React.useState("");
  const [description, setDescription] = React.useState("");
  const [actions, setActions] = React.useState<string[]>([]);

  React.useEffect(() => {
    if (!open) {
      setKey("");
      setName("");
      setDescription("");
      setActions([]);
      reset();
    }
  }, [open, reset]);

  const keyIsValid = APP_KEY.test(key) && key.length <= APP_KEY_MAX_LENGTH;
  const canSubmit = keyIsValid && name.trim() !== "" && actions.length > 0 && !create.isPending;
  const result = create.data;

  function toggleAction(action: string, checked: boolean) {
    setActions((current) => (checked ? [...current, action] : current.filter((a) => a !== action)));
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        // The token is shown once, so only Done closes this state.
        if (!create.isPending && !result) onOpenChange(next);
      }}
    >
      <DialogContent>
        {result ? (
          <TokenReveal
            appName={result.name}
            token={result.token}
            onDone={() => onOpenChange(false)}
          />
        ) : (
          <>
            <DialogHeader>
              <DialogTitle>Create app</DialogTitle>
              <DialogDescription>
                An app is a trusted tool that calls Custos with its own token.
              </DialogDescription>
            </DialogHeader>
            <form
              className="space-y-4"
              onSubmit={(e) => {
                e.preventDefault();
                if (canSubmit) create.mutate({ key, name: name.trim(), description, actions });
              }}
            >
              <div className="space-y-2">
                <Label htmlFor="app-key">App key</Label>
                <Input
                  id="app-key"
                  value={key}
                  onChange={(e) => setKey(e.target.value)}
                  aria-describedby="app-key-help"
                  className="font-mono"
                />
                <p id="app-key-help" className="text-xs text-muted-foreground">
                  Lowercase letters, digits and hyphens, for example notebook-hub. It cannot be
                  changed later.
                </p>
              </div>
              <div className="space-y-2">
                <Label htmlFor="app-name">Name</Label>
                <Input id="app-name" value={name} onChange={(e) => setName(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="app-description">Description</Label>
                <Input
                  id="app-description"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                />
              </div>
              <div className="space-y-3">
                <p className="text-sm font-medium">Actions</p>
                {connectors.data?.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No connector offers app actions.</p>
                ) : null}
                {connectors.data?.map((c) => (
                  <fieldset key={c.name} className="space-y-2">
                    <legend className="flex items-center gap-2 text-sm">
                      {c.name}
                      {c.needs_user_grant ? (
                        <StatusBadge variant="pending" label="Needs user authorization" />
                      ) : null}
                    </legend>
                    {c.actions?.map((action) => (
                      <label
                        key={action}
                        className="flex items-center gap-2 pl-4 font-mono text-xs"
                      >
                        <input
                          type="checkbox"
                          checked={actions.includes(action)}
                          onChange={(e) => toggleAction(action, e.target.checked)}
                        />
                        {action}
                      </label>
                    ))}
                  </fieldset>
                ))}
              </div>
              {create.error ? (
                <p role="alert" className="text-sm text-destructive">
                  {create.error.message}
                </p>
              ) : null}
              <DialogFooter>
                <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
                  Cancel
                </Button>
                <Button type="submit" disabled={!canSubmit}>
                  Create app
                </Button>
              </DialogFooter>
            </form>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
