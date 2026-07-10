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
