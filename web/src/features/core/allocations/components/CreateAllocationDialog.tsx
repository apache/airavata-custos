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
import { toast } from "sonner";
import { useClusters } from "@/features/core/clusters/queries";
import { useProjects } from "@/features/core/projects/queries";
import type { ProjectResponse } from "@/generated/core/zod.gen";
import { useDebounce } from "@/shared/hooks/useDebounce";
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
import { SuggestInput } from "@/shared/ui/SuggestInput";
import { useCreateAllocation } from "../queries";

export type CreateAllocationDialogProps = {
  onOpenChange: (open: boolean) => void;
  // Fixes the project when opened from a project page.
  project?: Pick<ProjectResponse, "id" | "title" | "ref">;
};

const projectLabel = (p: Pick<ProjectResponse, "title" | "ref">) => `${p.ref}: ${p.title}`;

// Date inputs are local calendar days, so both directions stay in local time.
const isoDate = (d: Date) => d.toLocaleDateString("en-CA");
const startOfDay = (day: string) => new Date(`${day}T00:00:00`).toISOString();

export function CreateAllocationDialog({
  onOpenChange,
  project,
}: CreateAllocationDialogProps) {
  const create = useCreateAllocation();
  const clusters = useClusters();
  const [projectText, setProjectText] = React.useState("");
  const [chosen, setChosen] = React.useState<ProjectResponse | null>(null);
  const query = useDebounce(projectText.trim(), 200);
  const projects = useProjects({ q: query, limit: 8 });
  const [name, setName] = React.useState("");
  const [clusterId, setClusterId] = React.useState("");
  const [sus, setSus] = React.useState("");
  const [start, setStart] = React.useState(() => isoDate(new Date()));
  const [end, setEnd] = React.useState(() =>
    isoDate(new Date(Date.now() + 365 * 24 * 3600 * 1000)),
  );

  // Results for an older query would let Enter pick a project the typed text no longer names.
  const found = project || query !== projectText.trim() ? [] : (projects.data?.items ?? []);
  const picked = project ?? chosen;
  const cluster = clusterId || clusters.data?.[0]?.id || "";
  const ready = Boolean(
    picked && name.trim() && cluster && sus !== "" && start && end && start <= end,
  );

  return (
    <Dialog open onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add allocation</DialogTitle>
          <DialogDescription>Grant a project compute time on a cluster.</DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (!picked || !ready) return;
            create.mutate(
              {
                project_id: picked.id,
                name: name.trim(),
                compute_cluster_id: cluster,
                initial_su_amount: Number(sus),
                start_time: startOfDay(start),
                end_time: startOfDay(end),
              },
              {
                onSuccess: (created) => {
                  toast.success(`Allocation ${created.name} created.`);
                  onOpenChange(false);
                },
              },
            );
          }}
        >
          <div className="space-y-2">
            <Label htmlFor="new-allocation-project">Project</Label>
            {project ? (
              <p id="new-allocation-project" className="text-sm">
                {projectLabel(project)}
              </p>
            ) : (
              <SuggestInput
                id="new-allocation-project"
                value={projectText}
                onValueChange={(next) => {
                  setProjectText(next);
                  setChosen(found.find((p) => projectLabel(p) === next) ?? null);
                }}
                suggestions={found.map((p) => ({ value: projectLabel(p) }))}
                emptyMessage={
                  !projectText.trim()
                    ? "Type to search"
                    : query === projectText.trim() && !projects.isFetching
                      ? "No matching project."
                      : "Searching…"
                }
                placeholder="Search by title or ID"
              />
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="new-allocation-name">Name</Label>
            <Input
              id="new-allocation-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="space-y-2">
              <Label htmlFor="new-allocation-cluster">Cluster</Label>
              <select
                id="new-allocation-cluster"
                value={cluster}
                onChange={(e) => setClusterId(e.target.value)}
                className="h-9 w-full rounded-md border bg-background px-3 text-sm"
              >
                {(clusters.data ?? []).map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="new-allocation-sus">Initial SUs</Label>
              <Input
                id="new-allocation-sus"
                type="number"
                min={0}
                value={sus}
                onChange={(e) => setSus(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="new-allocation-start">Start date</Label>
              <Input
                id="new-allocation-start"
                type="date"
                value={start}
                onChange={(e) => setStart(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="new-allocation-end">End date</Label>
              <Input
                id="new-allocation-end"
                type="date"
                min={start}
                value={end}
                onChange={(e) => setEnd(e.target.value)}
                required
              />
            </div>
          </div>
          {create.error ? (
            <p className="text-destructive text-sm">{create.error.message}</p>
          ) : null}
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
