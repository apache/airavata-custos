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

import { useClusters } from "@/features/core/clusters/queries";
import { Input } from "@/shared/ui/input";
import { Label } from "@/shared/ui/label";
import * as React from "react";
import { ResourcePicker } from "./ResourcePicker";

export type AllocationDraft = {
  name: string;
  clusterId: string;
  sus: string;
  start: string;
  end: string;
  resourceIds: string[];
};

// Date inputs are local calendar days, so both directions stay in local time.
const isoDate = (d: Date) => d.toLocaleDateString("en-CA");
const startOfDay = (day: string) => new Date(`${day}T00:00:00`).toISOString();

export const emptyAllocationDraft = (): AllocationDraft => ({
  name: "",
  clusterId: "",
  sus: "",
  start: isoDate(new Date()),
  end: isoDate(new Date(Date.now() + 365 * 24 * 3600 * 1000)),
  resourceIds: [],
});

export const isDraftReady = (d: AllocationDraft) =>
  Boolean(d.name.trim() && d.clusterId && d.sus !== "" && d.start && d.end && d.start <= d.end);

export const draftPayload = (projectId: string, d: AllocationDraft) => ({
  project_id: projectId,
  name: d.name.trim(),
  compute_cluster_id: d.clusterId,
  initial_su_amount: Number(d.sus),
  start_time: startOfDay(d.start),
  end_time: startOfDay(d.end),
  resourceIds: d.resourceIds,
});

// Fields of one allocation; an unset cluster falls back to the first one listed.
export function AllocationFields({
  value,
  onChange,
}: {
  value: AllocationDraft;
  onChange: (next: AllocationDraft) => void;
}) {
  const clusters = useClusters().data ?? [];
  const first = clusters[0]?.id;
  React.useEffect(() => {
    if (!value.clusterId && first) onChange({ ...value, clusterId: first });
  }, [value, first, onChange]);
  const set = (patch: Partial<AllocationDraft>) => onChange({ ...value, ...patch });
  return (
    <>
      <div className="space-y-2">
        <Label htmlFor="allocation-name">Name</Label>
        <Input
          id="allocation-name"
          value={value.name}
          onChange={(e) => set({ name: e.target.value })}
        />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-2">
          <Label htmlFor="allocation-cluster">Cluster</Label>
          <select
            id="allocation-cluster"
            value={value.clusterId}
            onChange={(e) => set({ clusterId: e.target.value, resourceIds: [] })}
            className="h-9 w-full rounded-md border bg-background px-3 text-sm"
          >
            {clusters.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <Label htmlFor="allocation-sus">Initial SUs</Label>
          <Input
            id="allocation-sus"
            type="number"
            min={0}
            value={value.sus}
            onChange={(e) => set({ sus: e.target.value })}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="allocation-start">Start date</Label>
          <Input
            id="allocation-start"
            type="date"
            value={value.start}
            onChange={(e) => set({ start: e.target.value })}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="allocation-end">End date</Label>
          <Input
            id="allocation-end"
            type="date"
            min={value.start}
            value={value.end}
            onChange={(e) => set({ end: e.target.value })}
          />
        </div>
      </div>
      <div className="space-y-2">
        <Label>Resources</Label>
        <ResourcePicker
          clusterId={value.clusterId}
          value={value.resourceIds}
          onChange={(resourceIds) => set({ resourceIds })}
        />
      </div>
    </>
  );
}
