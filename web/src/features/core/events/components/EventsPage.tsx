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

import type { PendingDelivery } from "@/generated/core/types.gen";
import { cn } from "@/lib/utils";
import { listOf } from "@/shared/api/listOf";
import { useAbility } from "@/shared/casl/AbilityProvider";
import {
  setSearchParam,
  useCanonicalParams,
  useDeclaredParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { ROUTE_LABELS, queryUuid } from "@/shared/layout/nav";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { TableSkeleton } from "@/shared/ui/Loading";
import { Button } from "@/shared/ui/button";
import { keepFocus } from "@/shared/ui/focus";
import { Label } from "@/shared/ui/label";
import { toastOnSuccess } from "@/shared/ui/sonner";
import { RefreshCwIcon, UnplugIcon } from "lucide-react";
import * as React from "react";
import {
  DELIVERY_PAGE_SIZE,
  useDeliveries,
  useDelivery,
  useRetryDelivery,
  useSubscriptions,
} from "../queries";
import { DeliveryDetailDrawer } from "./DeliveryDetailDrawer";
import { DeliveryTable } from "./DeliveryTable";

// The list a tab shows. "waiting" is every delivery not finished yet.
type DeliveryView = "all" | "waiting" | "failed";

const VIEWS: DeliveryView[] = ["all", "waiting", "failed"];

const VIEW_INFO: Record<DeliveryView, { label: string; empty: string; detail?: string }> = {
  all: { label: "All", empty: "No events yet" },
  waiting: {
    label: "Waiting",
    empty: "Nothing is waiting",
    detail: "Every delivery has succeeded or failed.",
  },
  failed: { label: "Failed", empty: "No failed deliveries" },
};

function parseView(value: string | null): DeliveryView {
  return value === "waiting" || value === "failed" ? value : "all";
}

function countLabel(rows: PendingDelivery[]): string {
  return rows.length >= DELIVERY_PAGE_SIZE ? `${DELIVERY_PAGE_SIZE}+` : String(rows.length);
}

export function EventsPage() {
  const params = useShallowSearchParams();
  const view = parseView(params.get("view"));
  const openId = queryUuid(params.get("delivery"));
  const [confirmRetryOnOpen, setConfirmRetryOnOpen] = React.useState(false);

  const ability = useAbility();
  const canRetry = ability.can("write", "EventDelivery");

  // One query per tab, so every tab shows its count.
  const all = useDeliveries();
  const waiting = useDeliveries("PENDING");
  const failed = useDeliveries("FAILED");
  const byView: Record<DeliveryView, typeof all> = { all, waiting, failed };
  const current = byView[view];

  const subscriptions = useSubscriptions();
  const running = React.useMemo(() => {
    const items = subscriptions.data?.items;
    if (!items) return undefined;
    return new Set(items.flatMap((s) => (s.loaded && s.subscriber ? [s.subscriber] : [])));
  }, [subscriptions.data]);
  const notRunning = React.useMemo(() => {
    const items = subscriptions.data?.items ?? [];
    const names = items.flatMap((s) => (s.loaded || !s.subscriber ? [] : [s.subscriber]));
    return [...new Set(names)];
  }, [subscriptions.data]);

  const connectors = React.useMemo(
    () =>
      [
        ...new Set([
          ...(subscriptions.data?.items ?? []).map((s) => s.subscriber),
          ...(all.data?.items ?? []).map((d) => d.subscriber),
        ]),
      ].sort(),
    [all.data, subscriptions.data],
  );
  // An unknown ?connector= shows All once the connectors are known.
  const rawConnector = params.get("connector") ?? "";
  const known = all.isFetched && subscriptions.isFetched;
  const connector = !known || connectors.includes(rawConnector) ? rawConnector : "";
  const rows = (current.data?.items ?? []).filter((d) => !connector || d.subscriber === connector);
  useDeclaredParams(["view", "delivery", "connector"]);
  useCanonicalParams({
    view: view === "all" ? undefined : view,
    delivery: openId,
    ...(known && { connector: connector || undefined }),
  });

  const detail = useDelivery(openId);
  const retry = useRetryDelivery();

  function open(delivery: PendingDelivery, confirm = false) {
    setConfirmRetryOnOpen(confirm);
    setSearchParam(params, "delivery", delivery.id);
  }

  function handleRetry() {
    if (!openId) return;
    toastOnSuccess(retry.mutateAsync({ path: { id: openId } }), "Retry started", () =>
      setConfirmRetryOnOpen(false),
    );
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
        <h1 className="text-2xl font-bold tracking-tight">{ROUTE_LABELS["/admin/events"]}</h1>
        <p className="text-sm text-muted-foreground">
          Each event is sent to every connector that listens for it. A failed delivery is retried up
          to 10 times, then waits here for an admin.
        </p>
      </header>

      {subscriptions.error ? (
        <ErrorState
          error={subscriptions.error}
          heading="Couldn't check which connectors are running"
          onRetry={() => {
            keepFocus(() => document.getElementById("events-connector"));
            subscriptions.refetch();
          }}
          className="py-4"
        />
      ) : notRunning.length > 0 ? (
        <output className="flex items-start gap-2 rounded-md bg-[color:var(--tone-warn-bg)] px-4 py-3 text-sm text-[color:var(--tone-warn-fg)]">
          <UnplugIcon className="mt-0.5 size-4 shrink-0" aria-hidden />
          <span>
            Not running on this server: <span className="font-mono">{notRunning.join(", ")}</span>.
            Their deliveries wait until the connector starts.
          </span>
        </output>
      ) : null}

      <div className="flex flex-wrap items-end justify-between gap-3 border-b border-border/60">
        <nav aria-label="Event views">
          <ul className="-mb-px flex flex-wrap gap-1">
            {VIEWS.map((v) => {
              const active = v === view;
              const { data } = byView[v];
              const count = data ? countLabel(listOf(data.items)) : "—";
              const failedTab = v === "failed" && data != null && count !== "0";
              return (
                <li key={v}>
                  <button
                    type="button"
                    aria-current={active ? "page" : undefined}
                    onClick={() => setSearchParam(params, "view", v === "all" ? null : v)}
                    className={cn(
                      "inline-flex items-center gap-2 border-b-2 px-4 py-2 text-sm font-medium transition-colors",
                      active
                        ? "border-brand text-foreground"
                        : "border-transparent text-muted-foreground hover:text-foreground",
                    )}
                  >
                    {VIEW_INFO[v].label}
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
              onChange={(e) => setSearchParam(params, "connector", e.currentTarget.value)}
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
        <ErrorState error={current.error} onRetry={() => current.refetch()} />
      ) : (
        <DeliveryTable
          rows={rows}
          runningSubscribers={running}
          canRetry={canRetry}
          onOpen={(d) => open(d)}
          onRetry={(d) => open(d, true)}
          empty={
            connector ? (
              <EmptyState heading="No deliveries match this connector" />
            ) : (
              <EmptyState heading={VIEW_INFO[view].empty} description={VIEW_INFO[view].detail} />
            )
          }
        />
      )}

      <DeliveryDetailDrawer
        deliveryId={openId}
        onOpenChange={(isOpen) => {
          if (!isOpen) setSearchParam(params, "delivery");
        }}
        delivery={detail.data}
        runningSubscribers={running}
        isLoading={detail.isLoading}
        error={detail.error}
        canRetry={canRetry}
        confirmRetryOnOpen={confirmRetryOnOpen}
        isRetrying={retry.isPending && retry.variables?.path.id === openId}
        onRetry={handleRetry}
        onRefresh={() => detail.refetch()}
      />
    </div>
  );
}
