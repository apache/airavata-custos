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

import { MoreVertical } from "lucide-react";
import * as React from "react";
import { AllocationResourcesDialog } from "@/features/core/allocations/components/AllocationResourcesDialog";
import { useQoSOptions } from "@/features/core/clusters/queries";
import { ApiError } from "@/shared/api/client";
import { useAbility } from "@/shared/casl/AbilityProvider";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/shared/ui/dropdown-menu";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import type { Conflict } from "../api";
import { asRecord, conflictLines } from "../format";
import { useResolveConflict } from "../queries";
import { QoSForm } from "./QoSForm";

const qosNames = (value: unknown): string[] => asRecord<{ qos?: string[] }>(value)?.qos ?? [];

type SlurmChoice = "custos" | "slurm" | "manual";

const CHOICE_LABEL: Record<SlurmChoice, string> = {
  custos: "Use Custos",
  slurm: "Use SLURM",
  manual: "Do manually…",
};

function effectOf(conflict: Conflict, choice: SlurmChoice): string {
  const { name } = conflict;
  if (choice === "manual") return "Pick each field";
  if (choice === "custos") {
    return conflict.custos == null
      ? `Deletes ${name} from SLURM`
      : `Pushes ${conflictLines(conflict.custos).join(", ")} to SLURM`;
  }
  return `Saves ${conflictLines(conflict.slurm).join(", ")} in Custos`;
}

type SlurmActionsProps = {
  clusterId: string;
  // Selects the write privilege that gates the menu.
  scope: "cluster" | "allocation";
  // Conflicts loading or Slurm unreachable: hides status and choices, keeps items.
  syncUnknown?: boolean;
  // Empty means in sync.
  conflicts?: Conflict[];
  // Row actions, listed under the choices.
  items?: React.ReactNode;
};

export function SlurmActions({
  clusterId,
  scope,
  syncUnknown = false,
  conflicts = [],
  items,
}: SlurmActionsProps) {
  const ability = useAbility();
  const resolve = useResolveConflict(clusterId);
  const [manualConflict, setManualConflict] = React.useState<Conflict | null>(null);
  const [notice, setNotice] = React.useState<string | null>(null);
  const isQoSList =
    manualConflict?.kind === "allocation_qos" || manualConflict?.kind === "member_qos";
  const manualQoS = isQoSList
    ? [...new Set([...qosNames(manualConflict.custos), ...qosNames(manualConflict.slurm)])]
    : [];
  const qosOptions = useQoSOptions(clusterId, manualQoS);

  const canWrite = ability.can("manage", scope === "cluster" ? "Cluster" : "Allocation");
  const active = syncUnknown ? [] : conflicts;
  const grouped = active.length > 1;
  const hasChoices = active.some((c) => (c.choices ?? []).length > 0);

  function openManual(target: Conflict) {
    setNotice(null);
    setManualConflict(target);
  }

  function submit(target: Conflict, choice: SlurmChoice, manual?: unknown) {
    setNotice(null);
    resolve.mutate(
      {
        key: target.key,
        choice,
        seen_slurm: target.slurm,
        ...(manual === undefined ? {} : { manual }),
      },
      {
        onSuccess: () => setManualConflict(null),
        onError: (error) => {
          const stale = error instanceof ApiError && error.status === 409;
          // A 409 voids the pick and closes; other errors keep the dialog.
          if (stale) setManualConflict(null);
          setNotice(stale ? "Changed on SLURM; refreshed" : error.message);
        },
      },
    );
  }

  return (
    <div>
      <div className="flex items-center gap-2">
        {syncUnknown ? (
          <span className="text-xs text-muted-foreground">—</span>
        ) : conflicts.length > 0 ? (
          <StatusBadge
            variant="warning"
            label={`Out of sync${conflicts.length > 1 ? ` · ${conflicts.length}` : ""}`}
          />
        ) : (
          <StatusBadge variant="active" label="In sync" />
        )}
        {canWrite && (hasChoices || items) ? (
          <DropdownMenu>
            <DropdownMenuTrigger
              render={(props) => (
                <button
                  {...props}
                  type="button"
                  aria-label="Actions"
                  className="rounded px-1.5 py-0.5 text-muted-foreground hover:bg-muted"
                >
                  <MoreVertical className="h-4 w-4" />
                </button>
              )}
            />
            <DropdownMenuContent align="end" className="w-64">
              {active.map((c, i) => (
                <React.Fragment key={c.key}>
                  {i > 0 ? <DropdownMenuSeparator /> : null}
                  {grouped ? (
                    <div className="px-2 py-1 text-xs font-medium text-muted-foreground">
                      {c.label}
                    </div>
                  ) : null}
                  {((c.choices ?? []) as SlurmChoice[]).map((choice) => (
                    <DropdownMenuItem
                      key={choice}
                      disabled={resolve.isPending}
                      onClick={() => (choice === "manual" ? openManual(c) : submit(c, choice))}
                    >
                      <div>
                        <div>{CHOICE_LABEL[choice]}</div>
                        <div className="text-xs text-muted-foreground">{effectOf(c, choice)}</div>
                      </div>
                    </DropdownMenuItem>
                  ))}
                </React.Fragment>
              ))}
              {hasChoices && items ? <DropdownMenuSeparator /> : null}
              {items}
            </DropdownMenuContent>
          </DropdownMenu>
        ) : null}
      </div>
      {notice && !manualConflict ? (
        <div role="alert" className="mt-1 text-xs text-muted-foreground">
          {notice}
        </div>
      ) : null}
      {manualConflict && isQoSList ? (
        <AllocationResourcesDialog
          title="Allocation QoS"
          options={qosOptions}
          empty="No QoS on this cluster."
          initial={manualQoS}
          pending={resolve.isPending}
          error={notice}
          onSave={(qos) => submit(manualConflict, "manual", qos)}
          onOpenChange={(open) => !open && setManualConflict(null)}
        />
      ) : manualConflict ? (
        <QoSForm
          title={`Do manually · ${manualConflict.name}`}
          compare={{ custos: manualConflict.custos, slurm: manualConflict.slurm }}
          pending={resolve.isPending}
          error={notice}
          onCancel={() => setManualConflict(null)}
          onSubmit={(manual) => submit(manualConflict, "manual", manual)}
        />
      ) : null}
    </div>
  );
}
