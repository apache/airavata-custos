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

import type { Allocation, ProjectContext } from "@/generated/analytics/types.gen";
import { cn } from "@/lib/utils";
import { formatSU } from "@/shared/format";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";
import { Check, ChevronsUpDown } from "lucide-react";
import { type UrgencyBand, pctRemaining, suBand } from "../lib";

function remaining({ initial_su_amount, used_su_amount }: Allocation): number {
  return Math.max(0, initial_su_amount - used_su_amount);
}

function balance(a: Allocation): string {
  return `${formatSU(remaining(a))} SU left`;
}

// Draw attention to low balances only; healthy ones stay muted.
function metaClass(band: UrgencyBand): string {
  if (band === "red") return "text-[color:var(--tone-error-fg)]";
  if (band === "amber") return "text-[color:var(--tone-warn-fg)]";
  return "text-muted-foreground";
}

type ContextSwitcherProps = {
  contexts: ProjectContext[];
  selectedAllocationId: string | undefined;
  onSelect: (projectId: string, allocationId: string) => void;
};

export function ContextSwitcher({
  contexts,
  selectedAllocationId,
  onSelect,
}: ContextSwitcherProps) {
  const rows = contexts
    .flatMap((c) => (c.allocations ?? []).map((a) => ({ project: c, allocation: a })))
    .sort((a, b) => a.allocation.name.localeCompare(b.allocation.name));
  const selected = rows.find((r) => r.allocation.id === selectedAllocationId);
  const projectName = selected?.project.project_name ?? "";
  const projectLabel = projectName ? (
    <span className="min-w-0 max-w-full truncate text-sm text-muted-foreground" title={projectName}>
      {projectName}
    </span>
  ) : null;

  // Nothing to switch: render a static label.
  if (selected && rows.length === 1) {
    return (
      <div className="flex flex-wrap items-center gap-3">
        <div
          className="min-w-0 max-w-full truncate rounded-md border border-border bg-card px-3 py-2 text-sm font-medium"
          title={selected.allocation.name}
        >
          {`${selected.allocation.name} · ${balance(selected.allocation)}`}
        </div>
        {projectLabel}
      </div>
    );
  }

  return (
    <div className="flex flex-wrap items-center gap-3">
      <DropdownMenu>
        <DropdownMenuTrigger
          render={(props) => (
            <button
              {...props}
              type="button"
              className="inline-flex w-[22rem] max-w-full items-center gap-2 rounded-md border border-border bg-card px-3 py-2 text-sm font-medium outline-none hover:bg-accent focus-visible:ring-2 focus-visible:ring-ring"
            >
              <span className="min-w-0 flex-1 truncate text-left" title={selected?.allocation.name}>
                {selected?.allocation.name ?? "Select an allocation"}
              </span>
              {selected ? (
                <span className="shrink-0 whitespace-nowrap font-normal tabular-nums text-muted-foreground">
                  {balance(selected.allocation)}
                </span>
              ) : null}
              <ChevronsUpDown
                className="h-4 w-4 shrink-0 text-muted-foreground"
                aria-hidden="true"
              />
            </button>
          )}
        />
        <DropdownMenuContent
          align="start"
          className="w-[30rem] max-w-[calc(100vw-2rem)] p-0 shadow-xl ring-border"
        >
          {rows.map(({ project, allocation }) => {
            const isSelected = allocation.id === selectedAllocationId;
            const pct = pctRemaining(allocation.initial_su_amount, allocation.used_su_amount);
            const band = suBand(pct);
            return (
              <DropdownMenuItem
                key={allocation.id}
                onClick={() => onSelect(project.project_id, allocation.id)}
                className="gap-3 rounded-none border-b border-border px-3.5 py-2.5 last:border-b-0"
              >
                <Check
                  className={cn("h-4 w-4 shrink-0", isSelected ? "opacity-100" : "opacity-0")}
                  aria-hidden="true"
                />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-sm" title={allocation.name}>
                    {allocation.name}
                  </div>
                  <div
                    className="truncate text-xs text-muted-foreground"
                    title={project.project_name}
                  >
                    {project.project_name}
                  </div>
                </div>
                <span
                  className={cn(
                    "ml-auto shrink-0 whitespace-nowrap text-xs tabular-nums",
                    metaClass(band),
                  )}
                >
                  {Math.round(pct)}% · {formatSU(remaining(allocation))} SU
                </span>
              </DropdownMenuItem>
            );
          })}
        </DropdownMenuContent>
      </DropdownMenu>
      {projectLabel}
    </div>
  );
}
