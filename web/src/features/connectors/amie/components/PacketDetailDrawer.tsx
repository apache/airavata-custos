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

import { JsonBlock } from "@/features/core/audit/components/primitives/JsonBlock";
import { Timestamp } from "@/features/core/audit/components/primitives/Timestamp";
import type { PacketResponse } from "@/generated/amie/types.gen";
import { cn } from "@/lib/utils";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { enumLabel } from "@/shared/format";
import { ErrorState, QueryErrorState } from "@/shared/ui/ErrorState";
import { CenteredSpinner } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { keepFocus } from "@/shared/ui/focus";
import { Link } from "@/shared/ui/link";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/shared/ui/tabs";
import * as React from "react";
import { usePacket, usePacketAudits, usePacketEvents } from "../queries";
import { ageHoursOf } from "../utils";
import { PacketAuditList } from "./PacketAuditList";
import { PacketEventsTable } from "./PacketEventsTable";
import { PacketStatusBadge } from "./PacketStatusBadge";

// The payload indented when it parses, else as received.
function rawText(raw: string): string {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

export type PacketDetailDrawerProps = {
  packetId: string | undefined;
  onClose: () => void;
};

type StateStep = { label: string; reached: boolean; at: string | undefined };

// The last step names the state the packet settled in, after DECODED if it got that far;
// until then, the expected PROCESSED.
// Only PROCESSED has a time of its own: updated_at is not moved by a refusal, a hold or a failure.
function stateSteps(p: PacketResponse): StateStep[] {
  const settled = p.status !== "NEW" && p.status !== "DECODED";
  const decoded: StateStep = { label: "DECODED", reached: Boolean(p.decoded_at), at: p.decoded_at };
  return [
    { label: "NEW", reached: true, at: p.received_at },
    ...(settled && !decoded.reached ? [] : [decoded]),
    settled
      ? {
          label: p.status,
          reached: true,
          at: p.status === "PROCESSED" ? p.processed_at : undefined,
        }
      : { label: "PROCESSED", reached: false, at: undefined },
  ];
}

export function StateMachineBreadcrumb({ packet }: { packet: PacketResponse }) {
  const steps = stateSteps(packet);
  return (
    <ol className="flex flex-wrap items-center gap-2 text-xs" aria-label="State machine">
      {steps.map(({ label, reached, at }, i) => {
        return (
          <React.Fragment key={label}>
            <li
              className={cn(
                "rounded-md border px-2 py-1",
                reached
                  ? "border-brand/30 bg-brand-tint"
                  : "border-border bg-muted/30 text-muted-foreground",
              )}
            >
              <span className="font-medium">{enumLabel(label)}</span>
              {at ? (
                <Timestamp iso={at} className="ml-2 tabular-nums text-muted-foreground" />
              ) : null}
            </li>
            {i < steps.length - 1 ? <span aria-hidden="true">→</span> : null}
          </React.Fragment>
        );
      })}
    </ol>
  );
}

function HeldReplyNotice() {
  const ability = useAbility();
  return (
    <div className="rounded-md bg-[color:var(--tone-warn-bg)] p-3 text-sm text-[color:var(--tone-warn-fg)]">
      The reply to ACCESS is held on the cluster account.{" "}
      {ability.can("read", "Cluster") ? (
        <Link href="/admin/users/cluster-accounts?status=ALL" className="font-medium underline">
          Cluster accounts
        </Link>
      ) : null}
    </div>
  );
}

function OverviewTab({ packet }: { packet: PacketResponse }) {
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
        <dt className="text-muted-foreground">Failed attempts</dt>
        <dd>{packet.retries}</dd>
        <dt className="text-muted-foreground">Received</dt>
        <dd className="tabular-nums">
          <Timestamp iso={packet.received_at} />
        </dd>
        <dt className="text-muted-foreground">Decoded</dt>
        <dd className="tabular-nums">
          {packet.decoded_at ? <Timestamp iso={packet.decoded_at} /> : "—"}
        </dd>
        <dt className="text-muted-foreground">Processed</dt>
        <dd className="tabular-nums">
          {packet.processed_at ? <Timestamp iso={packet.processed_at} /> : "—"}
        </dd>
      </dl>
      {packet.status === "WAITING_APPROVAL" ? <HeldReplyNotice /> : null}
      {packet.status === "REFUSED" ? (
        <div className="rounded-md bg-muted p-3 text-sm text-muted-foreground">
          An admin denied the cluster account. ACCESS was sent a failure with the reason.
        </div>
      ) : null}
      {/* The server keeps the last error after a later attempt succeeds. */}
      {packet.last_error && packet.status !== "PROCESSED" ? (
        <div className="rounded-md border border-[color:var(--banner-error-border)] bg-[color:var(--banner-error-bg)] p-3 text-sm text-[color:var(--banner-error-fg)]">
          <strong>Last error:</strong> {packet.last_error}
        </div>
      ) : null}
    </div>
  );
}

export function PacketDetailDrawer({ packetId, onClose }: PacketDetailDrawerProps) {
  const query = usePacket(packetId);
  const packet = query.data;
  const events = usePacketEvents(packet?.id);
  const audits = usePacketAudits(packet?.id);
  const tabList = React.useRef<HTMLDivElement>(null);
  // The retried error state unmounts with its button; focus stays on the tab it reloaded.
  const reload = (q: { refetch: () => unknown }) => () => {
    keepFocus(() => tabList.current?.querySelector<HTMLElement>('[aria-selected="true"]') ?? null);
    q.refetch();
  };

  return (
    <SideDrawer
      open={packetId != null}
      recordKey={packetId}
      onOpenChange={(open) => (open ? null : onClose())}
      width="lg"
      title={packet ? `Packet ${packet.amie_id}` : "Packet"}
      description={packet ? `${packet.type} · ${packet.id}` : undefined}
    >
      {query.isLoading ? (
        <CenteredSpinner label="Loading packet" />
      ) : query.error ? (
        <QueryErrorState
          error={query.error}
          what="packet"
          onRetry={() => query.refetch()}
          as="h3"
        />
      ) : !packet ? null : (
        <Tabs defaultValue="overview">
          <TabsList variant="line" ref={tabList}>
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="raw">Raw JSON</TabsTrigger>
            <TabsTrigger value="timeline">Timeline</TabsTrigger>
            <TabsTrigger value="audit">Audit</TabsTrigger>
          </TabsList>
          <TabsContent value="overview" className="pt-4">
            <OverviewTab packet={packet} />
          </TabsContent>
          <TabsContent value="raw" className="pt-4">
            {packet.raw_json ? (
              <JsonBlock text={rawText(packet.raw_json)} label="Packet JSON" />
            ) : (
              <p className="text-sm text-muted-foreground">No raw payload available.</p>
            )}
          </TabsContent>
          <TabsContent value="timeline" className="pt-4">
            {events.error ? (
              <ErrorState error={events.error} onRetry={reload(events)} as="h3" />
            ) : (
              <PacketEventsTable events={events.data ?? []} isLoading={events.isLoading} />
            )}
          </TabsContent>
          <TabsContent value="audit" className="pt-4">
            {audits.error ? (
              <ErrorState error={audits.error} onRetry={reload(audits)} as="h3" />
            ) : (
              <PacketAuditList events={audits.data?.events ?? []} isLoading={audits.isLoading} />
            )}
          </TabsContent>
        </Tabs>
      )}
    </SideDrawer>
  );
}
