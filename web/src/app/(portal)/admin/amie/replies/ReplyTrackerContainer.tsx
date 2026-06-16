"use client";

import * as React from "react";
import { toast } from "sonner";
import { ReplyTracker } from "@/features/connectors/amie/components/ReplyTracker";
import { useReplies, useRetryReply } from "@/features/connectors/amie/queries";
import type { ReplyStatus } from "@/features/connectors/amie/types";

export function ReplyTrackerContainer() {
  const [statusFilter, setStatusFilter] = React.useState<ReplyStatus | "all">("all");
  const repliesQuery = useReplies({
    status: statusFilter !== "all" ? statusFilter : undefined,
    limit: 200,
  });
  const retryMutation = useRetryReply();
  const rows = repliesQuery.data?.replies ?? [];

  async function handleRetry(id: string) {
    try {
      await retryMutation.mutateAsync(id);
      toast.success("Reply queued for resend");
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Retry failed");
    }
  }

  return (
    <ReplyTracker
      rows={rows}
      total={repliesQuery.data?.total ?? rows.length}
      isLoading={repliesQuery.isLoading}
      error={repliesQuery.error}
      statusFilter={statusFilter}
      onStatusChange={setStatusFilter}
      onRetry={handleRetry}
      onRefresh={() => repliesQuery.refetch()}
    />
  );
}
