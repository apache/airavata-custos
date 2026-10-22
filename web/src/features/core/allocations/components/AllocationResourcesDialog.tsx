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

import type { ComputeAllocation, ComputeAllocationResource } from "@/generated/core/zod.gen";
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
import { toast } from "sonner";
import { useSetAllocationResources } from "../queries";
import { ResourcePicker } from "./ResourcePicker";

// Mounted only while open, so the picker starts from the current grants each time.
export function AllocationResourcesDialog({
  allocation,
  granted,
  onOpenChange,
}: {
  allocation: ComputeAllocation;
  granted: ComputeAllocationResource[];
  onOpenChange: (open: boolean) => void;
}) {
  const save = useSetAllocationResources(allocation.id);
  const [resourceIds, setResourceIds] = React.useState(() => granted.map((r) => r.id));
  return (
    <Dialog open onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Resources</DialogTitle>
          <DialogDescription>
            Usage on these resources is charged to the allocation&apos;s{" "}
            {allocation.initial_su_amount.toLocaleString()} SUs at each resource&apos;s rate.
          </DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            save.mutate(resourceIds, {
              onSuccess: () => {
                toast.success("Resources saved.");
                onOpenChange(false);
              },
            });
          }}
        >
          <ResourcePicker
            clusterId={allocation.compute_cluster_id}
            value={resourceIds}
            onChange={setResourceIds}
          />
          {save.error ? <p className="text-destructive text-sm">{save.error.message}</p> : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={save.isPending}>
              {save.isPending ? "Saving…" : "Save"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
