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

import { useAllocation, useAllocationMembers } from "@/features/core/allocations/queries";
import { CopyValue } from "@/features/core/audit/components/primitives/CopyValue";
import { SourcePill } from "@/features/core/audit/components/primitives/SourcePill";
import { StatusPill } from "@/features/core/audit/components/primitives/StatusPill";
import { Timestamp } from "@/features/core/audit/components/primitives/Timestamp";
import { EVENT_LABELS, tokenLabel } from "@/features/core/audit/eventLabels";
import { useUser, userKeys } from "@/features/core/users/queries";
import { getUsersById } from "@/generated/core/sdk.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { enumLabel } from "@/shared/format";
import { queryUuid } from "@/shared/layout/nav";
import { CenteredSpinner } from "@/shared/ui/Loading";
import { displayName } from "@/shared/users-admin/displayName";
import { useQueries } from "@tanstack/react-query";
import type { PacketAudit } from "../utils";

// AMIE names the records a step touched as `allocation=<id> user=<id>`; each reads as its name, else is left out.
function KeyedDescription({ text }: { text: string }) {
  const ability = useAbility();
  const tokens = text.split(/\s+/).map((token) => {
    const [key = "", value] = token.split(/=(.*)/);
    return { token, key, value, id: queryUuid(value) };
  });
  const idOf = (key: string) => tokens.find((t) => t.key === key)?.id;
  const canReadUsers = ability.can("read", "User");
  const allocationId = ability.can("read", "Allocation") ? idOf("allocation") : undefined;
  const allocation = useAllocation(allocationId);
  const members = useAllocationMembers(canReadUsers ? undefined : allocationId);
  const user = useUser(canReadUsers ? idOf("user") : undefined);
  const names: Record<string, string | undefined> = {
    allocation: allocation.data?.name,
    user: user.data
      ? displayName(user.data)
      : members.data?.find((m) => m.user_id === idOf("user"))?.display_name,
  };
  const shown = tokens
    .flatMap(({ token, key, value, id }) => {
      if (value === undefined) return [tokenLabel(token)];
      const label = id ? names[key] : tokenLabel(value);
      return label ? [`${enumLabel(key)}: ${label}`] : [];
    })
    .join(" · ");
  return shown ? <p className="mt-1 break-words text-sm">{shown}</p> : null;
}

const UUIDS = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/gi;

// Prose that names users by bare id (`merged <id> into <id>: <reason>`): with every user read, each
// id reads as the name; otherwise only the text after the ids' clause is kept.
function ProseDescription({ text, ids }: { text: string; ids: string[] }) {
  const canReadUsers = useAbility().can("read", "User");
  const users = useQueries({
    queries: ids.map((id) => ({
      queryKey: [...userKeys.all, "detail", id],
      queryFn: () => getUsersById({ path: { id } }),
      enabled: canReadUsers,
    })),
  });
  if (canReadUsers && users.some((u) => u.isPending)) return null;
  const names = new Map(
    ids.flatMap((id, i) => {
      const user = users[i]?.data;
      return user ? [[id, displayName(user)] as const] : [];
    }),
  );
  const clauseEnd = text.lastIndexOf(": ");
  const shown =
    names.size === ids.length
      ? text.replace(UUIDS, (id) => names.get(id.toLowerCase()) ?? id)
      : clauseEnd < 0
        ? ""
        : text
            .slice(clauseEnd + 2)
            .replace(UUIDS, "")
            .trim();
  return shown ? <p className="mt-1 break-words text-sm">{shown}</p> : null;
}

function AuditDescription({ text }: { text: string }) {
  if (text.split(/\s+/).some((token) => queryUuid(token.split(/=(.*)/)[1]))) {
    return <KeyedDescription text={text} />;
  }
  const ids = [...new Set(text.match(UUIDS)?.map((id) => id.toLowerCase()))];
  if (ids.length > 0) return <ProseDescription text={text} ids={ids} />;
  const shown = tokenLabel(text);
  return shown ? <p className="mt-1 break-words text-sm">{shown}</p> : null;
}

export type PacketAuditListProps = {
  events: PacketAudit[];
  isLoading: boolean;
};

export function PacketAuditList({ events, isLoading }: PacketAuditListProps) {
  if (isLoading) return <CenteredSpinner label="Loading audit events" />;
  if (events.length === 0) {
    return <p className="text-sm text-muted-foreground">No audit events recorded.</p>;
  }
  return (
    <ol className="space-y-3">
      {/* The server lists events oldest first. */}
      {events.toReversed().map((event, i) => (
        <li
          key={`${event.span_id}-${i}`}
          className="flex min-w-0 items-start gap-3 rounded-md border bg-card p-3 text-sm"
        >
          <StatusPill
            status={event.status === "error" ? "error" : "ok"}
            size="sm"
            className="mt-0.5"
          />
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-baseline justify-between gap-x-2 gap-y-1">
              <span className="flex min-w-0 flex-wrap items-center gap-2 font-medium">
                <SourcePill source={event.source} size="sm" />
                <span className="min-w-0 break-words" title={event.event_type}>
                  {EVENT_LABELS[event.event_type] ?? tokenLabel(event.event_type)}
                </span>
              </span>
              <Timestamp
                iso={event.created_at}
                className="text-xs tabular-nums text-muted-foreground"
              />
            </div>
            {event.entity_type && event.entity_id ? (
              <p className="flex min-w-0 items-center gap-1 text-xs text-muted-foreground">
                <span className="shrink-0">{enumLabel(event.entity_type)}</span>
                <CopyValue value={event.entity_id} label={event.entity_type} className="min-w-0" />
              </p>
            ) : null}
            {event.description ? <AuditDescription text={event.description} /> : null}
          </div>
        </li>
      ))}
    </ol>
  );
}
