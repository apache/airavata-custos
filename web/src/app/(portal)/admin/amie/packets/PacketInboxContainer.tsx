"use client";

import dynamic from "next/dynamic";
import { useRouter, useSearchParams } from "next/navigation";
import * as React from "react";
import { toast } from "sonner";
import { PacketDetailDrawer } from "@/features/connectors/amie/components/PacketDetailDrawer";
import {
  type PacketFilters,
  PacketInboxTable,
} from "@/features/connectors/amie/components/PacketInboxTable";
import {
  usePacket,
  usePacketEvents,
  usePacketStats,
  usePackets,
  useResolvePacket,
  useRetryPacket,
} from "@/features/connectors/amie/queries";
import type { Packet } from "@/features/connectors/amie/types";
import { pluralize } from "@/features/connectors/amie/utils";
import { Button } from "@/shared/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/shared/ui/dialog";
import { Label } from "@/shared/ui/label";

// Recharts is heavy; defer until the inbox has hydrated so first paint stays fast.
const PacketsTrendChart = dynamic(
  () =>
    import("@/features/connectors/amie/components/PacketsTrendChart").then(
      (m) => m.PacketsTrendChart,
    ),
  { ssr: false, loading: () => <div className="h-40 rounded-md border bg-card" /> },
);

const DEFAULT_PAGE_SIZE = 20;

const DEFAULT_FILTERS: PacketFilters = {
  status: "all",
  type: "all",
  source: "all",
  q: "",
};

export function PacketInboxContainer({ initialPacketId }: { initialPacketId?: string } = {}) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const drawerIdFromUrl = initialPacketId ?? searchParams.get("packet") ?? undefined;

  const [page, setPage] = React.useState(1);
  const [filters, setFilters] = React.useState<PacketFilters>(DEFAULT_FILTERS);
  const [selected, setSelected] = React.useState<Set<string>>(new Set());
  const [selectedId, setSelectedId] = React.useState<string | undefined>(drawerIdFromUrl);
  const [markProcessedOpen, setMarkProcessedOpen] = React.useState(false);
  const [markProcessedReason, setMarkProcessedReason] = React.useState("");

  React.useEffect(() => {
    setSelectedId(drawerIdFromUrl);
  }, [drawerIdFromUrl]);

  const packetsQuery = usePackets({
    status: filters.status !== "all" ? filters.status : undefined,
    type: filters.type !== "all" ? filters.type : undefined,
    source: filters.source !== "all" ? filters.source : undefined,
    q: filters.q || undefined,
    limit: DEFAULT_PAGE_SIZE,
    offset: (page - 1) * DEFAULT_PAGE_SIZE,
  });

  const statsQuery = usePacketStats({ window: "30d" });
  const detailQuery = usePacket(selectedId);
  const eventsQuery = usePacketEvents(selectedId);
  const retryMutation = useRetryPacket();
  const resolveMutation = useResolvePacket();

  const rows = packetsQuery.data?.packets ?? [];
  const total = packetsQuery.data?.total ?? 0;

  function openDrawer(packet: Packet) {
    setSelectedId(packet.id);
    const params = new URLSearchParams(searchParams.toString());
    params.set("packet", packet.id);
    router.replace(`?${params.toString()}`, { scroll: false });
  }

  function closeDrawer() {
    setSelectedId(undefined);
    const params = new URLSearchParams(searchParams.toString());
    params.delete("packet");
    const next = params.toString();
    router.replace(next ? `?${next}` : "?", { scroll: false });
  }

  async function handleBulkRetry() {
    const ids = Array.from(selected);
    if (ids.length === 0) return;
    let queued = 0;
    for (const id of ids) {
      try {
        await retryMutation.mutateAsync(id);
        queued += 1;
      } catch {
        // continue — surface aggregate result below
      }
    }
    toast.success(`Queued ${queued} ${pluralize("retry", queued, "retries")}`);
    setSelected(new Set());
  }

  function handleBulkMarkProcessedRequest() {
    if (selected.size === 0) return;
    setMarkProcessedReason("");
    setMarkProcessedOpen(true);
  }

  async function handleBulkMarkProcessedConfirm() {
    const reason = markProcessedReason.trim();
    if (reason.length < 3) return;
    const ids = Array.from(selected);
    if (ids.length === 0) return;
    let resolved = 0;
    for (const id of ids) {
      try {
        await resolveMutation.mutateAsync({ id, reason });
        resolved += 1;
      } catch {
        // continue
      }
    }
    toast.success(`Marked ${resolved} ${pluralize("packet", resolved)} processed`);
    setSelected(new Set());
    setMarkProcessedOpen(false);
    setMarkProcessedReason("");
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

  async function handleDrawerRetry() {
    if (!detailQuery.data) return;
    try {
      await retryMutation.mutateAsync(detailQuery.data.id);
      toast.success("Retry queued");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Retry failed");
    }
  }

  async function handleDrawerResolve(reason: string) {
    if (!detailQuery.data) return;
    try {
      await resolveMutation.mutateAsync({ id: detailQuery.data.id, reason });
      toast.success("Packet resolved");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Resolve failed");
    }
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
        onPageChange={setPage}
        onRowClick={openDrawer}
        onBulkRetry={handleBulkRetry}
        onBulkMarkProcessed={handleBulkMarkProcessedRequest}
        onBulkExport={handleBulkExport}
        onRetry={() => packetsQuery.refetch()}
      />

      <Dialog open={markProcessedOpen} onOpenChange={setMarkProcessedOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Mark {selected.size} processed</DialogTitle>
            <DialogDescription>
              Surface a reason on the audit timeline of each packet. Minimum 3 characters.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-2">
            <Label htmlFor="amie-bulk-reason">Reason</Label>
            <textarea
              id="amie-bulk-reason"
              rows={3}
              className="w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50"
              placeholder="e.g. Linked manually after ACCESS retry window expired"
              value={markProcessedReason}
              onChange={(e) => setMarkProcessedReason(e.currentTarget.value)}
            />
          </div>
          <DialogFooter>
            <Button variant="ghost" onClick={() => setMarkProcessedOpen(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleBulkMarkProcessedConfirm}
              disabled={markProcessedReason.trim().length < 3}
            >
              Mark processed
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <PacketDetailDrawer
        open={selectedId != null}
        onOpenChange={(open) => {
          if (!open) closeDrawer();
        }}
        packet={detailQuery.data}
        events={eventsQuery.data ?? []}
        isLoading={detailQuery.isLoading}
        eventsLoading={eventsQuery.isLoading}
        error={detailQuery.error}
        canRetry
        canResolve
        onRetry={handleDrawerRetry}
        onResolve={handleDrawerResolve}
        onRefresh={() => {
          detailQuery.refetch();
          eventsQuery.refetch();
        }}
      />
    </div>
  );
}
