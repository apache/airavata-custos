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

import type { TraceEvent, TraceSummary } from "@/generated/core/types.gen";
import { cn } from "@/lib/utils";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Skeleton } from "@/shared/ui/skeleton";
import { Box, ExternalLink, LayoutGrid, Package, Server, User, Users } from "lucide-react";
import * as React from "react";
import { useAuditEventsForTrace } from "../queries";
import { formatAbsoluteUtc, formatRelative, getEntityRefs } from "../utils";
import { CopyValue } from "./primitives/CopyValue";
import { SourcePill } from "./primitives/SourcePill";

export type TraceLinkedEntitiesTabProps = {
  trace: TraceSummary;
  spans: TraceEvent[];
};

type EntityKindCfg = {
  label: string;
  Icon: typeof Package;
  bg: string;
  fg: string;
  // null means there is no known portal route yet — render muted fallback copy.
  routeFor: ((id: string) => string) | null;
};

const KIND_CONFIG: Record<string, EntityKindCfg> = {
  packet: {
    label: "AMIE packet",
    Icon: Package,
    bg: "var(--tone-info-bg)",
    fg: "var(--tone-info-fg)",
    routeFor: (id) => `/admin/amie/packets/${encodeURIComponent(id)}`,
  },
  user: {
    label: "User",
    Icon: User,
    bg: "var(--muted)",
    fg: "var(--muted-foreground)",
    routeFor: (id) => `/admin/users/management?user=${encodeURIComponent(id)}`,
  },
  project: {
    label: "Project",
    Icon: LayoutGrid,
    bg: "var(--muted)",
    fg: "var(--muted-foreground)",
    routeFor: (id) => `/projects/${encodeURIComponent(id)}`,
  },
  co_person: {
    label: "CO person",
    Icon: Users,
    bg: "var(--tone-accent-bg)",
    fg: "var(--tone-accent-fg)",
    routeFor: null,
  },
  compute_allocation: {
    label: "Allocation",
    Icon: Box,
    bg: "var(--muted)",
    fg: "var(--muted-foreground)",
    routeFor: (id) => `/allocations/${encodeURIComponent(id)}`,
  },
  compute_cluster_user: {
    label: "Cluster account",
    Icon: Server,
    bg: "var(--tone-warn-bg)",
    fg: "var(--tone-warn-fg)",
    routeFor: (id) => `/admin/users/cluster-accounts?status=ALL&account=${encodeURIComponent(id)}`,
  },
  event_delivery: {
    label: "Event delivery",
    Icon: Box,
    bg: "var(--muted)",
    fg: "var(--muted-foreground)",
    routeFor: (id) => `/admin/events?delivery=${encodeURIComponent(id)}`,
  },
};

const FALLBACK_CFG: Omit<EntityKindCfg, "label"> = {
  Icon: Box,
  bg: "var(--muted)",
  fg: "var(--muted-foreground)",
  routeFor: null,
};

export function TraceLinkedEntitiesTab({ trace, spans }: TraceLinkedEntitiesTabProps) {
  const entityRefs = React.useMemo(() => getEntityRefs(spans), [spans]);
  const auditQuery = useAuditEventsForTrace(trace.trace_id);

  return (
    <div className="max-w-[920px] space-y-8">
      <section>
        <div className="mb-2 text-[11.5px] font-bold uppercase tracking-[0.04em] text-muted-foreground">
          LINKED ENTITIES
        </div>
        <p className="mb-4 text-[13px] text-muted-foreground">
          Entities referenced by spans in this trace.
        </p>
        {entityRefs.length === 0 ? (
          <div
            data-testid="entities-empty"
            className="rounded-[10px] border border-dashed border-[color:var(--border-strong)] px-4 py-3.5 text-[13px] text-muted-foreground"
          >
            No referenced entities found across spans.
          </div>
        ) : (
          <div
            data-testid="entity-cards"
            className="grid gap-4"
            style={{ gridTemplateColumns: "repeat(auto-fill, minmax(240px, 1fr))" }}
          >
            {entityRefs.map(({ entity_type = "", entity_id = "" }) => {
              const cfg = KIND_CONFIG[entity_type] ?? { ...FALLBACK_CFG, label: entity_type };
              const href = cfg.routeFor ? cfg.routeFor(entity_id) : null;
              const { Icon, label } = cfg;
              return (
                <div
                  key={`${entity_type}::${entity_id}`}
                  data-testid={`entity-card-${entity_type}`}
                  className="rounded-xl border border-[color:var(--border)] bg-[color:var(--card)] p-4 shadow-sm"
                >
                  <div className="mb-2.5 flex items-center gap-2.5">
                    <span
                      aria-hidden="true"
                      className="inline-flex h-[30px] w-[30px] items-center justify-center rounded-md"
                      style={{ background: cfg.bg, color: cfg.fg }}
                    >
                      <Icon className="h-4 w-4" />
                    </span>
                    <span className="text-[11.5px] font-bold uppercase tracking-[0.03em] text-muted-foreground">
                      {label}
                    </span>
                  </div>
                  <div className="mb-2.5">
                    <CopyValue value={entity_id} label={label} explicit />
                  </div>
                  {href ? (
                    <a
                      href={href}
                      className="inline-flex items-center gap-1 text-[12.5px] font-semibold text-[color:var(--brand)] hover:underline"
                    >
                      View {label.toLowerCase()}{" "}
                      <ExternalLink className="h-3 w-3" aria-hidden="true" />
                    </a>
                  ) : (
                    <span
                      className="text-[12.5px] text-muted-foreground"
                      title={`No portal route registered for ${label}`}
                    >
                      Route not yet available
                    </span>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </section>

      <section>
        <div className="mb-2 text-[11.5px] font-bold uppercase tracking-[0.04em] text-muted-foreground">
          AUDIT EVENTS
        </div>
        <AuditEventsTable
          loading={auditQuery.isLoading}
          error={auditQuery.error}
          events={auditQuery.data?.events}
          onRetry={() => auditQuery.refetch()}
        />
      </section>
    </div>
  );
}

function AuditEventsTable({
  loading,
  error,
  events = [],
  onRetry,
}: {
  loading: boolean;
  error: Error | null;
  events: TraceEvent[] | undefined;
  onRetry: () => void;
}) {
  const rows = React.useMemo(
    () => events.toSorted((a, b) => b.created_at.localeCompare(a.created_at)),
    [events],
  );

  if (loading) {
    return (
      <div data-testid="audit-events-loading" className="space-y-2">
        <Skeleton className="h-8 w-full" />
        <Skeleton className="h-8 w-full" />
        <Skeleton className="h-8 w-full" />
      </div>
    );
  }
  if (error) {
    return (
      <ErrorState
        message={error.message ?? "Could not load audit events."}
        onRetry={onRetry}
        retryLabel="Retry"
      />
    );
  }
  if (rows.length === 0) {
    return (
      <div
        data-testid="audit-events-empty"
        className="rounded-[10px] border border-dashed border-[color:var(--border-strong)] px-4 py-3.5 text-[13px] text-muted-foreground"
      >
        No audit events written under this trace.
      </div>
    );
  }
  return (
    <div className="overflow-hidden rounded-[10px] border border-[color:var(--border)]">
      <table className="w-full text-[13px]">
        <caption className="sr-only">Audit events under this trace</caption>
        <thead className="bg-[color:var(--muted-2)]">
          <tr>
            {["Created", "Source", "Event type", "Entity ID", "Summary"].map((h) => (
              <th
                key={h}
                className="px-3 py-2 text-left text-[11.5px] font-semibold uppercase tracking-[0.04em] text-muted-foreground"
              >
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((r, i) => (
            <tr
              key={`${r.span_id}::${i}`}
              className={cn(
                "border-t border-[color:var(--border)]",
                i % 2 === 1 ? "bg-[color:var(--muted-2)]" : "bg-[color:var(--card)]",
              )}
            >
              <td
                className="px-3 py-2 tabular-nums text-muted-foreground"
                title={formatAbsoluteUtc(r.created_at)}
              >
                {formatRelative(r.created_at)}
              </td>
              <td className="px-3 py-2">
                <SourcePill source={r.source} size="sm" />
              </td>
              <td className="px-3 py-2 font-mono text-[12.5px]">{r.event_type}</td>
              <td className="px-3 py-2 font-mono text-[12.5px]">{r.entity_id}</td>
              <td className="px-3 py-2 text-muted-foreground">{r.description}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
