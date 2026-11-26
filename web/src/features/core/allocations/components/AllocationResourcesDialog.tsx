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
import * as React from "react";
import { type PickerOption, ResourcePicker } from "./ResourcePicker";

// Mounted only while open, so the picker starts from `initial` each time.
export function AllocationResourcesDialog({
  title,
  description,
  options,
  empty,
  initial,
  pending,
  error,
  onSave,
  onOpenChange,
}: {
  title: string;
  description?: React.ReactNode;
  options: PickerOption[];
  empty?: string | undefined;
  initial: string[];
  pending: boolean;
  error?: string | null | undefined;
  onSave: (ids: string[]) => void;
  onOpenChange: (open: boolean) => void;
}) {
  const [ids, setIds] = React.useState(initial);
  return (
    <Dialog open onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? <DialogDescription>{description}</DialogDescription> : null}
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            onSave(ids);
          }}
        >
          <ResourcePicker options={options} empty={empty} value={ids} onChange={setIds} />
          {error ? <p className="text-destructive text-sm">{error}</p> : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={pending}>
              {pending ? "Saving…" : "Save"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
