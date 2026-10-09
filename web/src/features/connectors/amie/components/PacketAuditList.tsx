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

import { SourcePill } from "@/features/core/audit/components/primitives/SourcePill";
import { StatusPill } from "@/features/core/audit/components/primitives/StatusPill";
import type { PacketAuditsResponse } from "@/generated/amie/types.gen";
import { CenteredSpinner } from "@/shared/ui/Loading";
import { formatDate } from "../utils";

export type PacketAuditListProps = {
  events: PacketAuditsResponse["events"];
  isLoading: boolean;
};

export function PacketAuditList({ events, isLoading }: PacketAuditListProps) {
  if (isLoading) return <CenteredSpinner label="Loading audit events" />;
  if (events.length === 0) {
    return <p className="text-sm text-muted-foreground">No audit events recorded.</p>;
  }
  return (
    <ol className="space-y-3">
      {events.map((event) => (
        <li key={event.span_id} className="flex items-start gap-3 rounded-md border bg-card p-3 text-sm">
          <StatusPill tone={event.status === "error" ? "error" : "ok"} dotOnly className="mt-1.5" />
          <div className="min-w-0 flex-1">
            <div className="flex items-baseline justify-between gap-2">
              <span className="flex items-center gap-2 font-mono font-medium">
                <SourcePill source={event.source} size="sm" />
                {event.event_type}
              </span>
              <time dateTime={event.created_at} className="text-xs tabular-nums text-muted-foreground">
                {formatDate(event.created_at)}
              </time>
            </div>
            {event.entity_id ? (
              <p className="text-xs text-muted-foreground">
                {event.entity_type}={event.entity_id}
              </p>
            ) : null}
            {event.description ? <p className="mt-1 break-words text-sm">{event.description}</p> : null}
          </div>
        </li>
      ))}
    </ol>
  );
}
