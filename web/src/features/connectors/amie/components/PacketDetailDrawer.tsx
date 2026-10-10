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

import type { PacketResponse } from "@/generated/amie/types.gen";
import dynamic from "next/dynamic";
import Link from "next/link";
import * as React from "react";
import { cn } from "@/lib/utils";
import { QueryErrorState } from "@/shared/ui/ErrorState";
import { CenteredSpinner } from "@/shared/ui/Loading";
import { SideDrawer } from "@/shared/ui/SideDrawer";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/shared/ui/tabs";
import { usePacket, usePacketAudits, usePacketEvents } from "../queries";
import { ageHoursOf, formatDate } from "../utils";
import { PacketAuditList } from "./PacketAuditList";
import { PacketEventsTable } from "./PacketEventsTable";
import { PacketStatusBadge } from "./PacketStatusBadge";
import { PacketActionButtons, type PacketActions } from "./packetActions";

// react-json-view-lite + its stylesheet is ~110 kB; load only when the user
// opens the Raw JSON tab so the inbox first paint stays small.
const PacketRawJson = dynamic(() => import("./PacketRawJson"), {
  ssr: false,
  loading: () => <CenteredSpinner label="Loading JSON viewer" />,
});

export type PacketDetailDrawerProps = PacketActions & {
  packetId: string | undefined;
  onClose: () => void;
};

const STATE_BREADCRUMB: Array<{
  label: string;
  whenAt: (p: PacketResponse) => string | undefined;
}> = [
  { label: "NEW", whenAt: (p) => p.received_at },
  { label: "DECODED", whenAt: (p) => p.decoded_at },
  {
    label: "PROCESSED | FAILED",
    whenAt: (p) =>
      p.status === "PROCESSED" ? p.processed_at : p.status === "FAILED" ? p.updated_at : undefined,
  },
];

function StateMachineBreadcrumb({ packet }: { packet: PacketResponse }) {
  return (
    <ol className="flex flex-wrap items-center gap-2 text-xs" aria-label="State machine">
      {STATE_BREADCRUMB.map((step, i) => {
        const at = step.whenAt(packet);
        const reached = Boolean(at);
        return (
          <React.Fragment key={step.label}>
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

function OverviewTab({ packet, ...actions }: PacketActions & { packet: PacketResponse }) {
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-3">
        <PacketStatusBadge status={packet.status} ageHours={ageHoursOf(packet.received_at)} />
        <span className="font-mono text-sm">{packet.amie_id}</span>
        <span className="text-xs text-muted-foreground">{packet.type}</span>
        <span className="ml-auto flex gap-2">
          <PacketActionButtons ids={[packet.id]} {...actions} />
        </span>
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
      {packet.status === "WAITING_APPROVAL" ? (
        <div className="rounded-md bg-[color:var(--tone-warn-bg)] p-3 text-sm text-[color:var(--tone-warn-fg)]">
          The reply to ACCESS is held until an admin reviews the cluster account.{" "}
          <Link href="/admin/users/cluster-accounts" className="font-medium underline">
            Review cluster accounts
          </Link>
        </div>
      ) : null}
      {packet.status === "REFUSED" ? (
        <div className="rounded-md bg-muted p-3 text-sm text-muted-foreground">
          An admin denied the cluster account. ACCESS was sent a failure with the reason.
        </div>
      ) : null}
      {packet.last_error ? (
        <div className="rounded-md border border-[color:var(--custos-red-200)] bg-[color:var(--custos-red-50)] p-3 text-sm text-[color:var(--custos-red-700)]">
          <strong>Last error:</strong> {packet.last_error}
        </div>
      ) : null}
    </div>
  );
}

export function PacketDetailDrawer({ packetId, onClose, ...actions }: PacketDetailDrawerProps) {
  const query = usePacket(packetId);
  const events = usePacketEvents(packetId);
  const audits = usePacketAudits(packetId);
  const packet = query.data;

  return (
    <SideDrawer
      open={packetId != null}
      onOpenChange={(open) => (open ? null : onClose())}
      width="lg"
      title={packet ? `Packet ${packet.amie_id}` : "Packet"}
      description={packet ? `${packet.type} · ${packet.id}` : undefined}
    >
      {query.isLoading ? (
        <CenteredSpinner label="Loading packet" />
      ) : query.error ? (
        <QueryErrorState error={query.error} what="packet" onRetry={() => query.refetch()} />
      ) : !packet ? (
        <p className="text-sm text-muted-foreground">Packet not found.</p>
      ) : (
        <Tabs defaultValue="overview">
          <TabsList variant="line">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="raw">Raw JSON</TabsTrigger>
            <TabsTrigger value="timeline">Timeline</TabsTrigger>
            <TabsTrigger value="audit">Audit</TabsTrigger>
          </TabsList>
          <TabsContent value="overview" className="pt-4">
            <OverviewTab packet={packet} {...actions} />
          </TabsContent>
          <TabsContent value="raw" className="pt-4">
            {packet.raw_json ? (
              <PacketRawJson rawJson={packet.raw_json} />
            ) : (
              <p className="text-sm text-muted-foreground">No raw payload available.</p>
            )}
          </TabsContent>
          <TabsContent value="timeline" className="pt-4">
            <PacketEventsTable events={events.data ?? []} isLoading={events.isLoading} />
          </TabsContent>
          <TabsContent value="audit" className="pt-4">
            <PacketAuditList events={audits.data?.events ?? []} isLoading={audits.isLoading} />
          </TabsContent>
        </Tabs>
      )}
    </SideDrawer>
  );
}
