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

import type { ProjectResponse } from "@/generated/core/zod.gen";
import { requires, usePermissions } from "@/shared/casl/permissions";
import { GuardedButton } from "@/shared/ui/GuardedButton";
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
import { useCreateAllocation } from "../queries";
import {
  AllocationFields,
  draftPayload,
  emptyAllocationDraft,
  isDraftReady,
} from "./AllocationFields";

type Project = Pick<ProjectResponse, "id" | "title" | "ref">;

// "Add allocation" link that opens the create dialog for one project.
export function AddAllocationButton({ project }: { project: Project }) {
  const can = usePermissions();
  const [open, setOpen] = React.useState(false);
  return (
    <>
      <GuardedButton
        variant="link"
        className="h-auto border-0 p-0"
        allowed={can.allocationsWrite && can.clustersRead}
        reason={requires("allocationsWrite", "clustersRead")}
        onClick={() => setOpen(true)}
      >
        Add allocation
      </GuardedButton>
      {open ? <CreateAllocationDialog onOpenChange={setOpen} project={project} /> : null}
    </>
  );
}

function CreateAllocationDialog({
  onOpenChange,
  project,
}: {
  onOpenChange: (open: boolean) => void;
  project: Project;
}) {
  const create = useCreateAllocation();
  const [draft, setDraft] = React.useState(emptyAllocationDraft);
  const ready = isDraftReady(draft);

  return (
    <Dialog open onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add allocation</DialogTitle>
          <DialogDescription>
            Grant {project.ref}: {project.title} compute time on a cluster.
          </DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (!ready) return;
            create.mutate(draftPayload(project.id, draft), {
              onSuccess: (created) => {
                toast.success(`Allocation ${created.name} created.`);
                onOpenChange(false);
              },
            });
          }}
        >
          <AllocationFields value={draft} onChange={setDraft} />
          {create.error ? <p className="text-destructive text-sm">{create.error.message}</p> : null}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={create.isPending || !ready}>
              {create.isPending ? "Creating…" : "Add allocation"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
