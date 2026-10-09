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

import * as React from "react";
import { FailedQueue } from "@/features/connectors/amie/components/FailedQueue";
import { PacketDetailDrawer } from "@/features/connectors/amie/components/PacketDetailDrawer";
import { usePackets } from "@/features/connectors/amie/queries";
import { usePageClamp } from "@/shared/hooks/usePageClamp";
import { usePacketActions } from "../usePacketActions";

const PAGE_SIZE = 50;

export function FailedQueueContainer() {
  const [page, setPage] = React.useState(1);
  const failedQuery = usePackets({
    status: "FAILED",
    limit: PAGE_SIZE,
    offset: (page - 1) * PAGE_SIZE,
  });
  // Computed once per mount so the query key stays stable.
  const [over24hCutoff] = React.useState(() => new Date(Date.now() - 24 * 3600_000).toISOString());
  const over24hQuery = usePackets({ status: "FAILED", to: over24hCutoff, limit: 1 });
  const rows = failedQuery.data?.packets ?? [];
  const total = failedQuery.data?.total ?? 0;
  // Retried and resolved packets leave the queue.
  usePageClamp(page, setPage, failedQuery.data?.total, PAGE_SIZE);

  const [selected, setSelected] = React.useState<Set<string>>(new Set());
  const actions = usePacketActions(() => setSelected(new Set()));
  const [drawerId, setDrawerId] = React.useState<string | undefined>(undefined);

  return (
    <>
      <FailedQueue
        rows={rows}
        total={total}
        isLoading={failedQuery.isLoading}
        error={failedQuery.error}
        failedOver24h={over24hQuery.data?.total ?? 0}
        page={page}
        pageSize={PAGE_SIZE}
        onRowClick={(p) => setDrawerId(p.id)}
        onPageChange={(next) => {
          setPage(next);
          setSelected(new Set());
        }}
        onRefresh={() => failedQuery.refetch()}
        selected={selected}
        onSelectChange={setSelected}
        {...actions}
      />
      <PacketDetailDrawer
        packetId={drawerId}
        onClose={() => setDrawerId(undefined)}
        {...actions}
      />
    </>
  );
}
