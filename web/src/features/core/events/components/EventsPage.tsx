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

import { RefreshCwIcon, UnplugIcon } from "lucide-react";
import * as React from "react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import { useAbility } from "@/shared/casl/AbilityProvider";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { Label } from "@/shared/ui/label";
import { DELIVERY_PAGE_SIZE } from "../api";
import { useDeliveries, useDelivery, useRetryDelivery, useSubscriptions } from "../queries";
import type { Delivery, DeliveryView } from "../types";
import { DeliveryDetailDrawer } from "./DeliveryDetailDrawer";
import { DeliveryTable } from "./DeliveryTable";

const VIEWS: DeliveryView[] = ["all", "waiting", "failed"];

const VIEW_INFO: Record<DeliveryView, { label: string; empty: string }> = {
  all: { label: "All", empty: "No events yet." },
  waiting: { label: "Waiting", empty: "Nothing is waiting. Every delivery has succeeded or failed." },
  failed: { label: "Failed", empty: "No failed deliveries." },
};

function parseView(value: string | null): DeliveryView {
  return value === "waiting" || value === "failed" ? value : "all";
}

function countLabel(rows: Delivery[] | undefined): string {
  if (!rows) return "";
  return rows.length >= DELIVERY_PAGE_SIZE ? `${DELIVERY_PAGE_SIZE}+` : String(rows.length);
}

function setParam(params: URLSearchParams, key: string, value: string | undefined) {
  const next = new URLSearchParams(params.toString());
  if (value) next.set(key, value);
  else next.delete(key);
  replaceShallowSearchParams(next);
}

export function EventsPage() {
  const params = useShallowSearchParams();
  const view = parseView(params.get("view"));
  const connector = params.get("connector") ?? "";
  const openId = params.get("delivery") ?? undefined;
  const [confirmRetryOnOpen, setConfirmRetryOnOpen] = React.useState(false);

  const ability = useAbility();
  const canRetry = ability.can("manage", "EventDelivery");

  // One query per tab, so every tab shows its count.
  const all = useDeliveries();
  const waiting = useDeliveries("PENDING");
  const failed = useDeliveries("FAILED");
  const byView: Record<DeliveryView, typeof all> = { all, waiting, failed };
  const current = byView[view];

  const subscriptions = useSubscriptions();
  const running = React.useMemo(() => {
    if (!subscriptions.data) return undefined;
    return new Set(subscriptions.data.filter((s) => s.loaded).map((s) => s.subscriber));
  }, [subscriptions.data]);
  const notRunning = React.useMemo(() => {
    const names = (subscriptions.data ?? []).filter((s) => !s.loaded).map((s) => s.subscriber);
    return [...new Set(names)];
  }, [subscriptions.data]);

  const connectors = React.useMemo(
    () => [...new Set((all.data ?? []).map((d) => d.subscriber))].sort(),
    [all.data],
  );
  const rows = (current.data ?? []).filter((d) => !connector || d.subscriber === connector);

  const detail = useDelivery(openId);
  const retry = useRetryDelivery();

  function open(delivery: Delivery, confirm = false) {
    setConfirmRetryOnOpen(confirm);
    setParam(params, "delivery", delivery.id);
  }

  async function handleRetry() {
    if (!openId) return;
    try {
      await retry.mutateAsync(openId);
      toast.success("Retry started", { description: "The delivery will run again shortly." });
      setConfirmRetryOnOpen(false);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Retry failed");
    }
  }

  function refreshAll() {
    all.refetch();
    waiting.refetch();
    failed.refetch();
    subscriptions.refetch();
  }

  return (
    <div className="space-y-5">
      <header className="space-y-1">
        <h1 className="text-2xl font-bold tracking-tight">Events</h1>
        <p className="text-sm text-muted-foreground">
          Each event is sent to every connector that listens for it. A failed delivery is retried up
          to 10 times, then waits here for an admin.
        </p>
      </header>

      {notRunning.length > 0 ? (
        <output className="flex items-start gap-2 rounded-md bg-[color:var(--tone-warn-bg)] px-4 py-3 text-sm text-[color:var(--tone-warn-fg)]">
          <UnplugIcon className="mt-0.5 size-4 shrink-0" aria-hidden />
          <span>
            Not running on this server: <span className="font-mono">{notRunning.join(", ")}</span>
            . Their deliveries wait until the connector starts.
          </span>
        </output>
      ) : null}

      <div className="flex flex-wrap items-end justify-between gap-3 border-b border-border/60">
        <nav aria-label="Event views">
          <ul className="-mb-px flex flex-wrap gap-1">
            {VIEWS.map((v) => {
              const active = v === view;
              const count = countLabel(byView[v].data);
              const failedTab = v === "failed" && count !== "" && count !== "0";
              return (
                <li key={v}>
                  <button
                    type="button"
                    aria-current={active ? "page" : undefined}
                    onClick={() => setParam(params, "view", v === "all" ? undefined : v)}
                    className={cn(
                      "inline-flex items-center gap-2 border-b-2 px-4 py-2 text-sm font-medium transition-colors",
                      active
                        ? "border-brand text-foreground"
                        : "border-transparent text-muted-foreground hover:text-foreground",
                    )}
                  >
                    {VIEW_INFO[v].label}
                    {count ? (
                      <span
                        className={cn(
                          "rounded-md px-1.5 text-xs tabular-nums",
                          failedTab
                            ? "bg-[color:var(--tone-error-bg)] text-[color:var(--tone-error-fg)]"
                            : "bg-muted text-muted-foreground",
                        )}
                      >
                        {count}
                      </span>
                    ) : null}
                  </button>
                </li>
              );
            })}
          </ul>
        </nav>
        <div className="flex items-end gap-2 pb-2">
          <div className="flex flex-col gap-1">
            <Label htmlFor="events-connector">Connector</Label>
            <select
              id="events-connector"
              value={connector}
              onChange={(e) => setParam(params, "connector", e.currentTarget.value || undefined)}
              className="rounded-md border bg-background px-3 py-1.5 text-sm"
            >
              <option value="">All connectors</option>
              {connectors.map((c) => (
                <option key={c} value={c}>
                  {c}
                </option>
              ))}
            </select>
          </div>
          <Button variant="outline" onClick={refreshAll} aria-label="Refresh">
            <RefreshCwIcon className="size-4" aria-hidden />
            Refresh
          </Button>
        </div>
      </div>

      {current.isLoading ? (
        <TableSkeleton />
      ) : current.error ? (
        <ErrorState message={current.error.message} onRetry={() => current.refetch()} />
      ) : (
        <DeliveryTable
          rows={rows}
          runningSubscribers={running}
          canRetry={canRetry}
          onOpen={(d) => open(d)}
          onRetry={(d) => open(d, true)}
          empty={
            <EmptyState
              heading={connector ? "No deliveries match this connector." : VIEW_INFO[view].empty}
            />
          }
        />
      )}

      <DeliveryDetailDrawer
        open={openId != null}
        onOpenChange={(isOpen) => {
          if (!isOpen) setParam(params, "delivery", undefined);
        }}
        delivery={detail.data}
        runningSubscribers={running}
        isLoading={detail.isLoading}
        error={detail.error}
        canRetry={canRetry}
        confirmRetryOnOpen={confirmRetryOnOpen}
        isRetrying={retry.isPending}
        onRetry={handleRetry}
        onRefresh={() => detail.refetch()}
      />
    </div>
  );
}
