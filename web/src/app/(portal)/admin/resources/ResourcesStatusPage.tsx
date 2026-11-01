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
import { useConnectors } from "@/features/core/connectors/queries";
import { relativeTime } from "@/features/core/events/utils";
import { ResourcesTab } from "@/features/core/resources/components/ResourcesTab";
import type { ConnectorCheck, ConnectorStatus } from "@/generated/core/zod.gen";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { TabsRouter } from "@/shared/ui/TabsRouter";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/shared/ui/select";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/shared/ui/tooltip";
import * as React from "react";
import { AllocationsTab, StorageFoldersTab, UsersTab } from "./RetentionTabs";

type Tone = "ok" | "warn" | "error" | "off";

const DOT: Record<Tone, string> = {
  ok: "bg-[color:var(--tone-ok-fg)]",
  warn: "bg-[color:var(--tone-warn-fg)]",
  error: "bg-[color:var(--tone-error-fg)]",
  off: "bg-[color:var(--custos-gray-400)]",
};

function StatusPill({
  tone,
  name,
  status,
  check,
  failedDeliveries,
}: {
  tone: Tone;
  name: string;
  status: string;
  check: ConnectorCheck | undefined;
  failedDeliveries?: number;
}) {
  return (
    <Tooltip>
      <TooltipTrigger className="inline-flex cursor-default items-center gap-2 rounded-full border border-border bg-card px-3 py-1 text-xs">
        <span aria-hidden className={`size-2 shrink-0 rounded-full ${DOT[tone]}`} />
        {name}
      </TooltipTrigger>
      <TooltipContent className="flex-col items-start gap-0.5">
        <span className="font-medium">{status}</span>
        {check ? <span>Checked {relativeTime(check.checked_at)}</span> : null}
        {failedDeliveries ? <span>{failedDeliveries} failed deliveries</span> : null}
        {check?.error ? <span className="opacity-80">{check.error}</span> : null}
      </TooltipContent>
    </Tooltip>
  );
}

function connectorStatus(c: ConnectorStatus): { tone: Tone; label: string } {
  if (!c.enabled) return { tone: "off", label: "Disabled" };
  if ((c.checks ?? []).some((check) => !check.ok)) return { tone: "error", label: "Failing" };
  if (c.failed_deliveries > 0) return { tone: "warn", label: "Failed deliveries" };
  return { tone: "ok", label: "Running" };
}

function latest(checks: ConnectorCheck[]): ConnectorCheck | undefined {
  return checks.reduce<ConnectorCheck | undefined>(
    (a, c) => (!a || c.checked_at > a.checked_at ? c : a),
    undefined,
  );
}

export function ResourcesStatusPage() {
  const clusters = useClusters();
  const connectors = useConnectors();
  const checks = (connectors.data ?? []).flatMap((c) => c.checks ?? []);
  const [picked, setPicked] = React.useState<string>();
  const clusterId = picked ?? clusters.data?.[0]?.id;

  return (
    <div className="space-y-8">
      <div className="grid grid-flow-col grid-cols-[1fr_auto] grid-rows-[auto_auto] items-start gap-x-4">
        <h1 className="font-display text-[28px] font-bold leading-tight">Resources</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Cluster reachability, connector health and the resources allocations are billed on.
        </p>
        {(clusters.data ?? []).length > 1 ? (
          <Select value={clusterId} onValueChange={(id) => setPicked(id ?? undefined)}>
            <SelectTrigger aria-label="Cluster" className="h-9 w-48 px-3">
              <SelectValue>
                {(id: string) => clusters.data?.find((c) => c.id === id)?.name ?? id}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {(clusters.data ?? []).map((c) => (
                <SelectItem key={c.id} value={c.id}>
                  {c.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : null}
      </div>

      {clusters.isLoading || connectors.isLoading ? (
        <CardSkeleton />
      ) : clusters.error || connectors.error ? (
        <ErrorState
          message={(clusters.error ?? connectors.error)?.message ?? ""}
          onRetry={() => {
            clusters.refetch();
            connectors.refetch();
          }}
        />
      ) : (
        <dl className="grid gap-x-8 gap-y-3 rounded-lg border border-border bg-muted/40 p-4 text-sm sm:grid-cols-[max-content_1fr]">
          <dt className="text-muted-foreground">Clusters</dt>
          <dd className="flex flex-wrap gap-2">
            {(clusters.data ?? []).map((cluster) => {
              const check = latest(checks.filter((c) => c.target === cluster.id));
              return (
                <StatusPill
                  key={cluster.id}
                  tone={check ? (check.ok ? "ok" : "error") : "off"}
                  name={cluster.name}
                  status={check ? (check.ok ? "Reachable" : "Unreachable") : "Not checked yet"}
                  check={check}
                />
              );
            })}
          </dd>

          <dt className="text-muted-foreground">Connectors</dt>
          <dd className="flex flex-wrap gap-2">
            {(connectors.data ?? []).map((c) => {
              const status = connectorStatus(c);
              return (
                <StatusPill
                  key={c.name}
                  tone={status.tone}
                  name={c.name}
                  status={status.label}
                  check={latest(c.checks ?? [])}
                  failedDeliveries={c.failed_deliveries}
                />
              );
            })}
          </dd>
        </dl>
      )}

      <TabsRouter
        defaultValue="resources"
        tabs={[
          { value: "resources", label: "Resources", content: <ResourcesTab /> },
          {
            value: "storage",
            label: "Storage",
            content: <StorageFoldersTab clusterId={clusterId} />,
          },
          {
            value: "users",
            label: "Users",
            content: <UsersTab clusterId={clusterId} />,
          },
          {
            value: "allocations",
            label: "Allocations",
            content: <AllocationsTab clusterId={clusterId} />,
          },
        ]}
      />
    </div>
  );
}
