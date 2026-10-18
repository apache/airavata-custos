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
import type { ComputeAllocationResourceSummary, ComputeCluster } from "@/generated/core/zod.gen";
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
import { useEffectiveRate, useSaveResource } from "../queries";

// SLURM TRES names the usage monitor bills against.
const RESOURCE_TYPES = ["cpu", "gres/gpu"] as const;

export type ResourceDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  clusters: ComputeCluster[];
  resource: ComputeAllocationResourceSummary | null;
};

export function ResourceDialog({ open, onOpenChange, clusters, resource }: ResourceDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{resource ? `Edit ${resource.name}` : "Add resource"}</DialogTitle>
          <DialogDescription>
            A partition or TRES that allocations can be granted on, and its SU rate.
          </DialogDescription>
        </DialogHeader>
        <ResourceForm clusters={clusters} resource={resource} onDone={() => onOpenChange(false)} />
      </DialogContent>
    </Dialog>
  );
}

function ResourceForm({
  clusters,
  resource,
  onDone,
}: {
  clusters: ComputeCluster[];
  resource: ComputeAllocationResourceSummary | null;
  onDone: () => void;
}) {
  const save = useSaveResource();
  const current = useEffectiveRate(resource?.rate_count ? resource.id : undefined).data?.rate;
  const [name, setName] = React.useState(resource?.name ?? "");
  const [type, setType] = React.useState<string>(resource?.resource_type ?? RESOURCE_TYPES[0]);
  const [clusterId, setClusterId] = React.useState(resource?.compute_cluster_id ?? "");
  // Null until typed, so a rate that loads after opening still shows.
  const [rate, setRate] = React.useState<string | null>(null);
  const cluster = clusterId || clusters[0]?.id || "";
  const rateText = rate ?? (current === undefined ? "" : String(current));
  const newRate = rateText !== "" && Number(rateText) !== current ? Number(rateText) : undefined;
  const ready = Boolean(name.trim() && cluster && (rateText === "" ? resource : Number(rateText) >= 0));

  return (
    <form
      className="space-y-4"
      onSubmit={(e) => {
        e.preventDefault();
        save.mutate(
          {
            id: resource?.id,
            resource: {
              name: name.trim(),
              resource_type: type,
              compute_cluster_id: cluster,
            },
            rate: newRate,
          },
          {
            onSuccess: (r) => {
              toast.success(`Resource ${r.name} ${resource ? "saved" : "added"}.`);
              onDone();
            },
          },
        );
      }}
    >
      <div className="space-y-2">
        <Label htmlFor="resource-name">Name</Label>
        <Input
          id="resource-name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="gpu-a100"
        />
      </div>
      <div className="grid grid-cols-2 gap-3">
        <div className="space-y-2">
          <Label htmlFor="resource-type">Type</Label>
          <select
            id="resource-type"
            value={type}
            onChange={(e) => setType(e.target.value)}
            className="h-9 w-full rounded-md border bg-background px-3 text-sm"
          >
            {[...new Set([...RESOURCE_TYPES, type])].map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <Label htmlFor="resource-rate">SU per unit·h</Label>
          <Input
            id="resource-rate"
            type="number"
            min={0}
            step="any"
            value={rateText}
            onChange={(e) => setRate(e.target.value)}
          />
        </div>
      </div>
      {clusters.length > 1 ? (
        <div className="space-y-2">
          <Label htmlFor="resource-cluster">Cluster</Label>
          <select
            id="resource-cluster"
            value={cluster}
            onChange={(e) => setClusterId(e.target.value)}
            className="h-9 w-full rounded-md border bg-background px-3 text-sm"
          >
            {clusters.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
          </select>
        </div>
      ) : null}
      {save.error ? <p className="text-destructive text-sm">{save.error.message}</p> : null}
      <DialogFooter>
        <Button type="button" variant="outline" onClick={onDone}>
          Cancel
        </Button>
        <Button type="submit" disabled={save.isPending || !ready}>
          {save.isPending ? "Saving…" : resource ? "Save" : "Add resource"}
        </Button>
      </DialogFooter>
    </form>
  );
}
