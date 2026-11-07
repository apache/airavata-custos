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

import type { ComputeAllocation, ProjectResponse } from "@/generated/core/zod.gen";
import { StatusIcon } from "@/shared/ui/StatusIcon";
import { Server } from "lucide-react";
import Link from "next/link";
import { useAllocationUsedSU, useAllocationsByProject } from "../queries";
import { AddAllocationButton } from "./AddAllocationButton";

function RemainingSU({ allocation }: { allocation: ComputeAllocation }) {
  const used = useAllocationUsedSU(allocation.id).data;
  return (
    <span className="tabular-nums" title={`of ${allocation.initial_su_amount.toLocaleString()}`}>
      {used === undefined ? "…" : (allocation.initial_su_amount - used).toLocaleString()}
    </span>
  );
}

// One project's allocations as a list, ending with the add control.
export function AllocationList({
  project,
}: {
  project: Pick<ProjectResponse, "id" | "title" | "ref">;
}) {
  const { data, error } = useAllocationsByProject(project.id);
  return (
    <ul className="space-y-1 text-sm">
      {error ? <li className="text-destructive">{error.message}</li> : null}
      {(data?.items ?? []).map((row) => (
        <li key={row.id} className="flex flex-wrap items-center gap-x-2 gap-y-1">
          <StatusIcon
            icon={Server}
            active={row.status === "ACTIVE"}
            label={row.status.toLowerCase()}
          />
          <Link
            href={`/projects/${row.project_id}/${row.id}`}
            className="font-medium text-foreground hover:underline"
          >
            {row.name}
          </Link>
          <span className="text-xs text-muted-foreground">
            <RemainingSU allocation={row} /> SUs left · ends{" "}
            {new Date(row.end_time).toLocaleDateString(undefined, { dateStyle: "medium" })}
          </span>
        </li>
      ))}
      <li className="flex items-center gap-2">
        <span aria-hidden className="w-4 text-center text-muted-foreground">
          +
        </span>
        <AddAllocationButton project={project} />
      </li>
    </ul>
  );
}
