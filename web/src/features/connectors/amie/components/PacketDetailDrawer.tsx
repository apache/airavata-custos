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

import { Tabs as TabsPrimitive } from "@base-ui/react/tabs";
import dynamic from "next/dynamic";
import Link from "next/link";
import * as React from "react";
import { cn } from "@/lib/utils";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CenteredSpinner } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { Button } from "@/shared/ui/button";
import type { Packet, PacketEvent } from "../types";
import { ageHoursOf, formatDate } from "../utils";
import { PacketEventsTable } from "./PacketEventsTable";
import { PacketStatusBadge } from "./PacketStatusBadge";

// react-json-view-lite + its stylesheet is ~110 kB; load only when the user
// opens the Raw JSON tab so the inbox first paint stays small.
const PacketRawJson = dynamic(() => import("./PacketRawJson"), {
  ssr: false,
  loading: () => <CenteredSpinner label="Loading JSON viewer" />,
});

function entityHref(type: string, id: string): string | null {
  if (type === "project") return `/allocations?project=${encodeURIComponent(id)}`;
  return null;
}

export type PacketDetailDrawerProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  packet: Packet | undefined;
  events: PacketEvent[];
  isLoading: boolean;
  eventsLoading: boolean;
  error: Error | null;
  canRetry: boolean;
  canResolve: boolean;
  onRetry: () => void;
  onResolve: (reason: string) => void;
  onRefresh: () => void;
};

const TAB_VALUES = ["overview", "raw", "timeline", "linked"] as const;
type TabValue = (typeof TAB_VALUES)[number];

const STATE_BREADCRUMB: Array<{
  key: string;
  label: string;
  whenAt: (p: Packet) => string | undefined;
}> = [
  { key: "received", label: "NEW", whenAt: (p) => p.received_at },
  { key: "decoded", label: "DECODED", whenAt: (p) => p.decoded_at },
  {
    key: "terminal",
    label: "PROCESSED | FAILED",
    whenAt: (p) =>
      p.status === "PROCESSED"
        ? p.processed_at
        : p.status === "FAILED"
          ? p.updated_at
          : undefined,
  },
];

function StateMachineBreadcrumb({ packet }: { packet: Packet }) {
  return (
    <ol className="flex flex-wrap items-center gap-2 text-xs" aria-label="State machine">
      {STATE_BREADCRUMB.map((step, i) => {
        const at = step.whenAt(packet);
        const reached = Boolean(at);
        return (
          <React.Fragment key={step.key}>
            <li
              className={cn(
                "rounded-md border px-2 py-1",
                reached
                  ? "border-brand/30 bg-brand-tint"
                  : "border-border bg-muted/30 text-muted-foreground",
              )}
            >
              <span className="font-medium">{step.label}</span>
              {reached ? (
                <time className="ml-2 tabular-nums text-muted-foreground" dateTime={at}>
                  {formatDate(at)}
                </time>
              ) : null}
            </li>
            {i < STATE_BREADCRUMB.length - 1 ? <span aria-hidden="true">→</span> : null}
          </React.Fragment>
        );
      })}
    </ol>
  );
}

function OverviewTab({ packet }: { packet: Packet }) {
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <PacketStatusBadge status={packet.status} ageHours={ageHoursOf(packet.received_at)} />
        <span className="font-mono text-sm">{packet.amie_id}</span>
        <span className="text-xs text-muted-foreground">{packet.type}</span>
      </div>
      <StateMachineBreadcrumb packet={packet} />
      <dl className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
        <dt className="text-muted-foreground">Source</dt>
        <dd>{packet.source}</dd>
        <dt className="text-muted-foreground">Retries</dt>
        <dd>{packet.retries}</dd>
        <dt className="text-muted-foreground">Received</dt>
        <dd className="tabular-nums">{formatDate(packet.received_at)}</dd>
        <dt className="text-muted-foreground">Decoded</dt>
        <dd className="tabular-nums">{formatDate(packet.decoded_at)}</dd>
        <dt className="text-muted-foreground">Processed</dt>
        <dd className="tabular-nums">{formatDate(packet.processed_at)}</dd>
        <dt className="text-muted-foreground">Updated</dt>
        <dd className="tabular-nums">{formatDate(packet.updated_at)}</dd>
      </dl>
      {packet.last_error ? (
        <div className="rounded-md border border-[color:var(--custos-red-200)] bg-[color:var(--custos-red-50)] p-3 text-sm text-[color:var(--custos-red-700)]">
          <strong>Last error:</strong> {packet.last_error}
        </div>
      ) : null}
      {packet.decoded_payload ? (
        <section aria-labelledby="amie-decoded-heading" className="space-y-2">
          <h3 id="amie-decoded-heading" className="text-sm font-semibold">
            Decoded payload
          </h3>
          <dl className="grid grid-cols-[max-content_1fr] gap-x-4 gap-y-1 rounded-md border bg-muted/20 p-3 text-xs">
            {Object.entries(packet.decoded_payload).map(([key, value]) => (
              <React.Fragment key={key}>
                <dt className="text-muted-foreground">{key}</dt>
                <dd className="break-all">
                  {typeof value === "object" ? JSON.stringify(value) : String(value)}
                </dd>
              </React.Fragment>
            ))}
          </dl>
        </section>
      ) : null}
    </div>
  );
}

function RawJsonTab({ packet }: { packet: Packet }) {
  if (!packet.raw_json) {
    return <p className="text-sm text-muted-foreground">No raw payload available.</p>;
  }
  return <PacketRawJson rawJson={packet.raw_json} />;
}

function LinkedEntityTab({ packet }: { packet: Packet }) {
  if (!packet.linked_entity) {
    return (
      <p className="text-sm text-muted-foreground">
        This packet is not yet linked to a domain entity. Use the reconciliation queue to link it
        manually.
      </p>
    );
  }
  const href = entityHref(packet.linked_entity.type, packet.linked_entity.id);
  return (
    <div className="space-y-2 text-sm">
      <p className="text-muted-foreground">Linked entity</p>
      <p>
        <span className="font-medium">{packet.linked_entity.type}</span> ·{" "}
        <span className="font-mono">
          {packet.linked_entity.display_id ?? packet.linked_entity.id}
        </span>
      </p>
      {href ? (
        <Link href={href} className="text-brand hover:underline">
          View in portal →
        </Link>
      ) : null}
    </div>
  );
}

export function PacketDetailDrawer({
  open,
  onOpenChange,
  packet,
  events,
  isLoading,
  eventsLoading,
  error,
  canRetry,
  canResolve,
  onRetry,
  onResolve,
  onRefresh,
}: PacketDetailDrawerProps) {
  const [tab, setTab] = React.useState<TabValue>("overview");
  const [resolveOpen, setResolveOpen] = React.useState(false);
  const [reason, setReason] = React.useState("");

  React.useEffect(() => {
    if (!open) {
      setTab("overview");
      setResolveOpen(false);
      setReason("");
    }
  }, [open]);

  return (
    <SideDrawer
      open={open}
      onOpenChange={onOpenChange}
      width="lg"
      title={packet ? `Packet ${packet.amie_id}` : "Packet"}
      description={packet ? `${packet.type} · ${packet.id}` : undefined}
    >
      {isLoading ? (
        <CenteredSpinner label="Loading packet" />
      ) : error ? (
        <ErrorState message={error.message ?? "Failed to load packet"} onRetry={onRefresh} />
      ) : !packet ? (
        <p className="text-sm text-muted-foreground">Packet not found.</p>
      ) : (
        <div className="space-y-5">
          <TabsPrimitive.Root
            value={tab}
            onValueChange={(v) => typeof v === "string" && setTab(v as TabValue)}
          >
            <TabsPrimitive.List className="flex gap-1 border-b border-border/80">
              {[
                { value: "overview", label: "Overview" },
                { value: "raw", label: "Raw JSON" },
                { value: "timeline", label: "Timeline" },
                { value: "linked", label: "Linked entity" },
              ].map((t) => (
                <TabsPrimitive.Tab
                  key={t.value}
                  value={t.value}
                  className={cn(
                    "relative -mb-px inline-flex items-center justify-center px-3 py-2 text-sm font-medium text-muted-foreground transition-colors",
                    "hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                    "data-[active]:border-b-2 data-[active]:border-brand data-[active]:text-foreground",
                  )}
                >
                  {t.label}
                </TabsPrimitive.Tab>
              ))}
            </TabsPrimitive.List>
            <TabsPrimitive.Panel value="overview" className="pt-4">
              <OverviewTab packet={packet} />
            </TabsPrimitive.Panel>
            <TabsPrimitive.Panel value="raw" className="pt-4">
              <RawJsonTab packet={packet} />
            </TabsPrimitive.Panel>
            <TabsPrimitive.Panel value="timeline" className="pt-4">
              <PacketEventsTable events={events} isLoading={eventsLoading} />
            </TabsPrimitive.Panel>
            <TabsPrimitive.Panel value="linked" className="pt-4">
              <LinkedEntityTab packet={packet} />
            </TabsPrimitive.Panel>
          </TabsPrimitive.Root>

          <div className="flex flex-wrap items-center justify-end gap-2 border-t border-border/60 pt-4">
            {canRetry ? (
              <Button variant="outline" onClick={onRetry}>
                Retry
              </Button>
            ) : null}
            {canResolve ? (
              resolveOpen ? (
                <form
                  className="flex flex-1 items-end gap-2"
                  onSubmit={(e) => {
                    e.preventDefault();
                    if (reason.trim().length < 3) return;
                    onResolve(reason);
                    setResolveOpen(false);
                    setReason("");
                  }}
                >
                  <label className="flex flex-1 flex-col gap-1 text-xs">
                    <span className="text-muted-foreground">Resolution reason</span>
                    <input
                      className="rounded-md border bg-background px-3 py-1.5 text-sm"
                      placeholder="Manual resolution — what was done?"
                      value={reason}
                      onChange={(e) => setReason(e.currentTarget.value)}
                      required
                      minLength={3}
                    />
                  </label>
                  <Button type="submit" disabled={reason.trim().length < 3}>
                    Confirm resolve
                  </Button>
                  <Button type="button" variant="ghost" onClick={() => setResolveOpen(false)}>
                    Cancel
                  </Button>
                </form>
              ) : (
                <Button variant="outline" onClick={() => setResolveOpen(true)}>
                  Resolve manually…
                </Button>
              )
            ) : null}
          </div>
        </div>
      )}
    </SideDrawer>
  );
}
