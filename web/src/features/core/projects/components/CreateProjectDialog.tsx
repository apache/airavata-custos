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

import {
  type AllocationDraft,
  AllocationFields,
  draftPayload,
  emptyAllocationDraft,
  isDraftReady,
} from "@/features/core/allocations/components/AllocationFields";
import { useCreateAllocation } from "@/features/core/allocations/queries";
import { ClusterName } from "@/features/core/clusters/components/ClusterName";
import { useResourceSummaries } from "@/features/core/resources/queries";
import { useUserEmailSearch } from "@/features/core/users/queries";
import { SuggestInput } from "@/shared/ui/SuggestInput";
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
import { useMutation } from "@tanstack/react-query";
import { Server } from "lucide-react";
import { useRouter } from "next/navigation";
import * as React from "react";
import { toast } from "sonner";
import { useCreateProject } from "../queries";

export type CreateProjectDialogProps = {
  onOpenChange: (open: boolean) => void;
};

const STEPS = [
  "Name the project and its PI.",
  "Add its allocations; more can be added later.",
  "Review, then confirm. Nothing is created before this.",
];

const day = (d: string) =>
  new Date(`${d}T00:00:00`).toLocaleDateString(undefined, { dateStyle: "medium" });

export function CreateProjectDialog({ onOpenChange }: CreateProjectDialogProps) {
  const router = useRouter();
  const createProject = useCreateProject();
  const createAllocation = useCreateAllocation();
  const [step, setStep] = React.useState(0);
  const [title, setTitle] = React.useState("");
  const [piEmail, setPiEmail] = React.useState("");
  const pi = useUserEmailSearch(piEmail, true);
  const [drafts, setDrafts] = React.useState<AllocationDraft[]>([]);
  const [draft, setDraft] = React.useState(emptyAllocationDraft);
  // One mutation spans every create, so Confirm stays disabled until all of them settle.
  const submit = useMutation({
    mutationFn: async (piId: string) => {
      const project = await createProject.mutateAsync({
        title: title.trim(),
        origination: "internal",
        project_pi_id: piId,
      });
      const results = await Promise.allSettled(
        drafts.map((d) => createAllocation.mutateAsync(draftPayload(project.id, d))),
      );
      const failed = results.filter((r) => r.status === "rejected").length;
      if (failed) toast.error(`Project ${project.title} created; ${failed} allocation(s) failed.`);
      else toast.success(`Project ${project.title} created with ${drafts.length} allocation(s).`);
      router.push(`/projects/${project.id}`);
    },
  });

  return (
    <Dialog open onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>New project</DialogTitle>
          <DialogDescription>
            Step {step + 1} of {STEPS.length} · {STEPS[step]}
          </DialogDescription>
        </DialogHeader>

        {step === 0 ? (
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="new-project-title">Title</Label>
              <Input
                id="new-project-title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                maxLength={200}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="new-project-pi">PI</Label>
              <SuggestInput
                id="new-project-pi"
                type="email"
                value={piEmail}
                onValueChange={setPiEmail}
                suggestions={pi.suggestions}
                emptyMessage={pi.emptyMessage}
                placeholder="pi@example.edu"
              />
            </div>
          </div>
        ) : step === 1 ? (
          <div className="space-y-4">
            <DraftList drafts={drafts} onRemove={(i) => setDrafts(drafts.toSpliced(i, 1))} />
            <div className="space-y-4 rounded-lg border p-4">
              <AllocationFields value={draft} onChange={setDraft} />
              <div className="flex justify-end">
                <Button
                  type="button"
                  variant="outline"
                  disabled={!isDraftReady(draft)}
                  onClick={() => {
                    setDrafts([...drafts, draft]);
                    setDraft({ ...emptyAllocationDraft(), clusterId: draft.clusterId });
                  }}
                >
                  Add allocation
                </Button>
              </div>
            </div>
          </div>
        ) : (
          <dl className="grid grid-cols-[max-content_1fr] gap-x-6 gap-y-2 text-sm">
            <dt className="text-muted-foreground">Title</dt>
            <dd>{title.trim()}</dd>
            <dt className="text-muted-foreground">PI</dt>
            <dd>{piEmail}</dd>
            <dt className="text-muted-foreground">Allocations</dt>
            <dd>
              <DraftList drafts={drafts} />
            </dd>
          </dl>
        )}

        {submit.error ? <p className="text-destructive text-sm">{submit.error.message}</p> : null}
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => (step === 0 ? onOpenChange(false) : setStep(step - 1))}
          >
            {step === 0 ? "Cancel" : "Back"}
          </Button>
          {step < 2 ? (
            <Button
              type="button"
              disabled={step === 0 && (!title.trim() || !pi.match)}
              onClick={() => setStep(step + 1)}
            >
              Next
            </Button>
          ) : (
            <Button
              type="button"
              disabled={submit.isPending}
              onClick={() => pi.match && submit.mutate(pi.match.id)}
            >
              {submit.isPending ? "Creating…" : "Confirm and create"}
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

function DraftLine({ draft }: { draft: AllocationDraft }) {
  const resources = useResourceSummaries().data ?? [];
  const names = resources.filter((r) => draft.resourceIds.includes(r.id)).map((r) => r.name);
  return (
    <span className="text-xs text-muted-foreground">
      <ClusterName id={draft.clusterId} /> · {Number(draft.sus).toLocaleString()} SUs ·{" "}
      {day(draft.start)} – {day(draft.end)}
      {names.length ? ` · ${names.join(", ")}` : ""}
    </span>
  );
}

function DraftList({
  drafts,
  onRemove,
}: {
  drafts: AllocationDraft[];
  onRemove?: (i: number) => void;
}) {
  if (drafts.length === 0)
    return <p className="text-sm text-muted-foreground">No allocations added.</p>;
  return (
    <ul className="divide-y rounded-md border">
      {drafts.map((d, i) => (
        <li key={i} className="flex items-center gap-3 px-3 py-2">
          <Server className="size-4 text-muted-foreground" aria-hidden />
          <div className="flex min-w-0 flex-1 flex-col">
            <span className="text-sm font-medium">{d.name}</span>
            <DraftLine draft={d} />
          </div>
          {onRemove ? (
            <Button type="button" variant="ghost" size="sm" onClick={() => onRemove(i)}>
              Remove
            </Button>
          ) : null}
        </li>
      ))}
    </ul>
  );
}
