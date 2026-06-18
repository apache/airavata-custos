"use client";

import * as React from "react";
import { toast } from "sonner";
import { FailedQueue } from "@/features/connectors/amie/components/FailedQueue";
import { PacketDetailDrawer } from "@/features/connectors/amie/components/PacketDetailDrawer";
import {
  usePacket,
  usePacketEvents,
  usePackets,
  useResolvePacket,
  useRetryPacket,
} from "@/features/connectors/amie/queries";
import type { Packet } from "@/features/connectors/amie/types";
import { pluralize } from "@/features/connectors/amie/utils";

function failedOver24h(rows: Packet[]): number {
  const cutoff = Date.now() - 24 * 60 * 60 * 1000;
  return rows.filter((r) => {
    const t = Date.parse(r.received_at);
    return !Number.isNaN(t) && t < cutoff;
  }).length;
}

export function FailedQueueContainer() {
  const failedQuery = usePackets({ status: "FAILED", limit: 100 });
  const rows = failedQuery.data?.packets ?? [];
  const total = failedQuery.data?.total ?? 0;

  const [selected, setSelected] = React.useState<Set<string>>(new Set());
  const [drawerId, setDrawerId] = React.useState<string | undefined>(undefined);

  const detailQuery = usePacket(drawerId);
  const eventsQuery = usePacketEvents(drawerId);
  const retryMutation = useRetryPacket();
  const resolveMutation = useResolvePacket();

  async function handleRetry(id: string) {
    try {
      await retryMutation.mutateAsync(id);
      toast.success("Retry queued");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Retry failed");
    }
  }

  async function handleResolveRow(packet: Packet) {
    const reason = window.prompt(`Resolve ${packet.amie_id} — reason?`);
    if (!reason || reason.trim().length < 3) return;
    try {
      await resolveMutation.mutateAsync({ id: packet.id, reason: reason.trim() });
      toast.success(`Resolved ${packet.amie_id}`);
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Resolve failed");
    }
  }

  async function handleBulkRetry() {
    const ids = Array.from(selected);
    let queued = 0;
    for (const id of ids) {
      try {
        await retryMutation.mutateAsync(id);
        queued += 1;
      } catch {
        // continue
      }
    }
    toast.success(`Queued ${queued} ${pluralize("retry", queued, "retries")}`);
    setSelected(new Set());
    failedQuery.refetch();
  }

  return (
    <>
      <FailedQueue
        rows={rows}
        total={total}
        isLoading={failedQuery.isLoading}
        error={failedQuery.error}
        failedOver24h={failedOver24h(rows)}
        selected={selected}
        onSelectChange={setSelected}
        onRowClick={(p) => setDrawerId(p.id)}
        onRetryRow={handleRetry}
        onResolveRow={handleResolveRow}
        onBulkRetry={handleBulkRetry}
        onRefresh={() => failedQuery.refetch()}
      />
      <PacketDetailDrawer
        open={drawerId != null}
        onOpenChange={(open) => {
          if (!open) setDrawerId(undefined);
        }}
        packet={detailQuery.data}
        events={eventsQuery.data ?? []}
        isLoading={detailQuery.isLoading}
        eventsLoading={eventsQuery.isLoading}
        error={detailQuery.error}
        canRetry
        canResolve
        onRetry={() => detailQuery.data && handleRetry(detailQuery.data.id)}
        onResolve={(reason) =>
          detailQuery.data
            ? resolveMutation
                .mutateAsync({ id: detailQuery.data.id, reason })
                .then(() => toast.success("Packet resolved"))
                .catch((err: unknown) =>
                  toast.error(err instanceof Error ? err.message : "Resolve failed"),
                )
            : undefined
        }
        onRefresh={() => {
          detailQuery.refetch();
          eventsQuery.refetch();
        }}
      />
    </>
  );
}
