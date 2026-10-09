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

import dynamic from "next/dynamic";
import { useRouter, useSearchParams } from "next/navigation";
import * as React from "react";
import { PacketDetailDrawer } from "@/features/connectors/amie/components/PacketDetailDrawer";
import {
  type PacketFilters,
  PacketInboxTable,
} from "@/features/connectors/amie/components/PacketInboxTable";
import { usePacketStats, usePackets } from "@/features/connectors/amie/queries";
import { usePageClamp } from "@/shared/hooks/usePageClamp";
import { usePacketActions } from "../usePacketActions";

// Recharts is heavy; defer until the inbox has hydrated so first paint stays fast.
const PacketsTrendChart = dynamic(
  () =>
    import("@/features/connectors/amie/components/PacketsTrendChart").then(
      (m) => m.PacketsTrendChart,
    ),
  { ssr: false, loading: () => <div className="h-40 rounded-md border bg-card" /> },
);

const DEFAULT_PAGE_SIZE = 20;
const INBOX_PATH = "/admin/amie/packets";

export function PacketInboxContainer({ initialPacketId }: { initialPacketId?: string } = {}) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const selectedId = initialPacketId ?? searchParams.get("packet") ?? undefined;

  const [page, setPage] = React.useState(1);
  const [filters, setFilters] = React.useState<PacketFilters>({ status: "all", type: "all", q: "" });
  const [selected, setSelected] = React.useState<Set<string>>(new Set());

  const packetsQuery = usePackets({
    status: filters.status !== "all" ? filters.status : undefined,
    type: filters.type !== "all" ? filters.type : undefined,
    q: filters.q || undefined,
    limit: DEFAULT_PAGE_SIZE,
    offset: (page - 1) * DEFAULT_PAGE_SIZE,
  });

  const actions = usePacketActions(() => setSelected(new Set()));
  const statsQuery = usePacketStats({ window: "30d" });

  const rows = packetsQuery.data?.packets ?? [];
  const total = packetsQuery.data?.total ?? 0;
  usePageClamp(page, setPage, packetsQuery.data?.total, DEFAULT_PAGE_SIZE);

  // Always targets the inbox path so a /packets/[id] deep link doesn't keep its stale id.
  function replacePacketParam(id: string | undefined) {
    const params = new URLSearchParams(searchParams.toString());
    if (id) params.set("packet", id);
    else params.delete("packet");
    const next = params.toString();
    router.replace(next ? `${INBOX_PATH}?${next}` : INBOX_PATH, { scroll: false });
  }

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
      <PacketsTrendChart buckets={statsQuery.data?.byDay ?? []} />
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
          setFilters(next);
          setPage(1);
          setSelected(new Set());
        }}
        onPageChange={(next) => {
          setPage(next);
          setSelected(new Set());
        }}
        onRowClick={(packet) => replacePacketParam(packet.id)}
        onBulkExport={handleBulkExport}
        {...actions}
        onRetry={() => packetsQuery.refetch()}
      />

      <PacketDetailDrawer
        packetId={selectedId}
        onClose={() => replacePacketParam(undefined)}
        {...actions}
      />
    </div>
  );
}
