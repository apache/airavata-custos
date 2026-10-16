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

import { PacketDetailDrawer } from "@/features/connectors/amie/components/PacketDetailDrawer";
import {
  PACKET_TYPES,
  type PacketFilters,
  PacketInboxTable,
} from "@/features/connectors/amie/components/PacketInboxTable";
import { usePacket, usePacketStats, usePackets } from "@/features/connectors/amie/queries";
import { PACKET_STATES } from "@/features/connectors/amie/utils";
import { useCanonicalParams, useDeclaredParams } from "@/shared/hooks/useShallowSearchParams";
import { useBreadcrumbLabel } from "@/shared/layout/BreadcrumbLabelsProvider";
import { normalizeQueryId } from "@/shared/layout/nav";
import { ErrorState } from "@/shared/ui/ErrorState";
import { focusLandingOf } from "@/shared/ui/focus";
import dynamic from "next/dynamic";
import { useRouter, useSearchParams } from "next/navigation";
import * as React from "react";
import { usePacketParam, usePageParam } from "../usePacketParam";

// Recharts is heavy; defer until the inbox has hydrated so first paint stays fast.
const PacketsTrendChart = dynamic(
  () =>
    import("@/features/connectors/amie/components/PacketsTrendChart").then(
      (m) => m.PacketsTrendChart,
    ),
  { ssr: false, loading: () => <ChartPlaceholder /> },
);

function ChartPlaceholder() {
  return <div className="h-40 rounded-md border bg-card" />;
}

const DEFAULT_PAGE_SIZE = 20;
const INBOX_PATH = "/admin/amie/packets";
const FILTER_DEFAULTS: PacketFilters = { status: "all", type: "all", q: "" };

// A URL value the selects do not offer reads as All.
function known(value: string | null, options: ReadonlyArray<string>): string {
  return value && options.includes(value) ? value : "all";
}

export function PacketInboxContainer({ initialPacketId }: { initialPacketId?: string } = {}) {
  const [packetParam, replacePacketParam] = usePacketParam(INBOX_PATH);
  const selectedId = initialPacketId ?? packetParam;
  const routePacket = usePacket(initialPacketId).data;
  useBreadcrumbLabel(initialPacketId, routePacket ? `Packet ${routePacket.amie_id}` : "Packet");

  const [page, setPage] = usePageParam(INBOX_PATH);
  const router = useRouter();
  const searchParams = useSearchParams();
  const filters: PacketFilters = {
    status: known(searchParams.get("status"), PACKET_STATES),
    type: known(searchParams.get("type"), PACKET_TYPES),
    q: normalizeQueryId(searchParams.get("q")) ?? FILTER_DEFAULTS.q,
  };
  useDeclaredParams(["status", "type", "q", "page", "packet"]);
  useCanonicalParams({
    status: filters.status !== "all" ? filters.status : undefined,
    type: filters.type !== "all" ? filters.type : undefined,
    q: filters.q || undefined,
    page: page > 1 ? String(page) : undefined,
    // A /packets/[id] route names its packet in the path.
    packet: initialPacketId ? undefined : packetParam,
  });
  const [selected, setSelected] = React.useState<Set<string>>(new Set());

  const packetsQuery = usePackets({
    status: filters.status !== "all" ? filters.status : undefined,
    type: filters.type !== "all" ? filters.type : undefined,
    q: filters.q || undefined,
    limit: DEFAULT_PAGE_SIZE,
    offset: (page - 1) * DEFAULT_PAGE_SIZE,
  });

  const statsQuery = usePacketStats({ window: "30d" });

  const rows = packetsQuery.data?.packets ?? [];
  const total = packetsQuery.data?.total ?? 0;

  function handleBulkExport() {
    const exported = rows.filter((r) => selected.has(r.id));
    const blob = new Blob([JSON.stringify(exported, null, 2)], { type: "application/json" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `amie-packets-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  }

  return (
    <div className="space-y-6">
      {statsQuery.error ? (
        <ErrorState error={statsQuery.error} onRetry={() => statsQuery.refetch()} />
      ) : statsQuery.isLoading ? (
        <ChartPlaceholder />
      ) : (
        <PacketsTrendChart buckets={statsQuery.data?.byDay ?? []} />
      )}
      <PacketInboxTable
        rows={rows}
        total={total}
        isLoading={packetsQuery.isLoading}
        error={packetsQuery.error}
        page={page}
        pageSize={DEFAULT_PAGE_SIZE}
        filters={filters}
        selected={selected}
        onSelectChange={setSelected}
        onFiltersChange={(next) => {
          const params = new URLSearchParams(searchParams.toString());
          for (const key of ["status", "type", "q"] as const) {
            if (next[key] && next[key] !== FILTER_DEFAULTS[key]) params.set(key, next[key]);
            else params.delete(key);
          }
          params.delete("page");
          router.replace(`${INBOX_PATH}?${params}`, { scroll: false });
          setSelected(new Set());
        }}
        onPageChange={(next) => {
          setPage(next);
          setSelected(new Set());
        }}
        onRowClick={(packet) => replacePacketParam(packet.id)}
        onBulkExport={handleBulkExport}
        onRetry={() => packetsQuery.refetch()}
      />

      <PacketDetailDrawer
        packetId={selectedId}
        onClose={() => {
          // A /packets/[id] route closes onto the inbox, a page of its own.
          if (initialPacketId) focusLandingOf(INBOX_PATH);
          replacePacketParam(undefined);
        }}
      />
    </div>
  );
}
