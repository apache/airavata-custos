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

import { FailedQueue } from "@/features/connectors/amie/components/FailedQueue";
import { PacketDetailDrawer } from "@/features/connectors/amie/components/PacketDetailDrawer";
import { usePackets } from "@/features/connectors/amie/queries";
import { useCanonicalParams, useDeclaredParams } from "@/shared/hooks/useShallowSearchParams";
import { ErrorState } from "@/shared/ui/ErrorState";
import { keepFocus, landingTarget } from "@/shared/ui/focus";
import * as React from "react";
import { usePacketParam, usePageParam } from "../usePacketParam";

const PAGE_SIZE = 50;

const PATH = "/admin/amie/failed";

export function FailedQueueContainer() {
  const [page, setPage] = usePageParam(PATH);
  const [drawerId, replacePacketParam] = usePacketParam(PATH);
  useDeclaredParams(["page", "packet"]);
  useCanonicalParams({ page: page > 1 ? String(page) : undefined, packet: drawerId });
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

  return (
    <>
      {over24hQuery.error ? (
        <ErrorState
          error={over24hQuery.error}
          heading="Couldn't count packets failed for over 24 hours"
          onRetry={() => {
            keepFocus(landingTarget);
            over24hQuery.refetch();
          }}
          className="py-4"
        />
      ) : null}
      <FailedQueue
        rows={rows}
        total={total}
        isLoading={failedQuery.isLoading}
        error={failedQuery.error}
        failedOver24h={over24hQuery.data?.total ?? 0}
        page={page}
        pageSize={PAGE_SIZE}
        onRowClick={(p) => replacePacketParam(p.id)}
        onPageChange={setPage}
        onRefresh={() => failedQuery.refetch()}
      />
      <PacketDetailDrawer packetId={drawerId} onClose={() => replacePacketParam(undefined)} />
    </>
  );
}
