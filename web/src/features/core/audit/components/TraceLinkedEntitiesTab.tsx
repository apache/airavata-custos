"use client";

import { Box, ExternalLink, LayoutGrid, Package, Server, User, Users } from "lucide-react";
import * as React from "react";
import { cn } from "@/lib/utils";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Skeleton } from "@/shared/ui/skeleton";
import { useAuditEventsForTrace } from "../queries";
import type { AuditEventsResponse, Span, Trace } from "../types";
import { formatAbsoluteUtc, formatRelative, getEntityRefs } from "../utils";
import { CopyValue } from "./primitives/CopyValue";
import { SourcePill } from "./primitives/SourcePill";

export type TraceLinkedEntitiesTabProps = {
  trace: Trace;
  spans: Span[];
};

type EntityKindCfg = {
  Icon: typeof Package;
  bg: string;
  fg: string;
  // null means there is no known portal route yet — render muted fallback copy.
  routeFor: ((id: string) => string) | null;
};

const KIND_CONFIG: Record<string, EntityKindCfg> = {
  "AMIE packet": {
    Icon: Package,
    bg: "var(--tone-info-bg)",
    fg: "var(--tone-info-fg)",
    routeFor: (id) => `/admin/amie/packets/${encodeURIComponent(id)}`,
  },
  User: {
    Icon: User,
    bg: "var(--muted)",
    fg: "var(--muted-foreground)",
    routeFor: null,
  },
  Project: {
    Icon: LayoutGrid,
    bg: "var(--muted)",
    fg: "var(--muted-foreground)",
    routeFor: (id) => `/projects/${encodeURIComponent(id)}`,
  },
  "CO person": {
    Icon: Users,
    bg: "var(--tone-accent-bg)",
    fg: "var(--tone-accent-fg)",
    routeFor: null,
  },
  Allocation: {
    Icon: Box,
    bg: "var(--muted)",
    fg: "var(--muted-foreground)",
    routeFor: (id) => `/allocations/${encodeURIComponent(id)}`,
  },
  "Cluster account": {
    Icon: Server,
    bg: "var(--tone-warn-bg)",
    fg: "var(--tone-warn-fg)",
    routeFor: null,
  },
};

const FALLBACK_CFG: EntityKindCfg = {
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
            {entityRefs.map((ref) => {
              const cfg = KIND_CONFIG[ref.kind] ?? FALLBACK_CFG;
              const href = cfg.routeFor ? cfg.routeFor(ref.primaryId) : null;
              const { Icon } = cfg;
              return (
                <div
                  key={`${ref.kind}::${ref.primaryId}`}
                  data-testid={`entity-card-${ref.kind.toLowerCase().replace(/\s+/g, "-")}`}
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
                      {ref.kind}
                    </span>
                  </div>
                  <div className="mb-2.5">
                    <CopyValue value={ref.primaryId} label={ref.kind} explicit />
                  </div>
                  {href ? (
                    <a
                      href={href}
                      className="inline-flex items-center gap-1 text-[12.5px] font-semibold text-[color:var(--brand)] hover:underline"
                    >
                      View {ref.kind.toLowerCase()}{" "}
                      <ExternalLink className="h-3 w-3" aria-hidden="true" />
                    </a>
                  ) : (
                    <span
                      className="text-[12.5px] text-muted-foreground"
                      title={`No portal route registered for ${ref.kind}`}
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
          error={auditQuery.error as Error | null}
          data={auditQuery.data}
          onRetry={() => auditQuery.refetch()}
        />
      </section>
    </div>
  );
}

type AuditRow = {
  key: string;
  createdAt: string;
  source: "core" | "amie";
  eventType: string;
  entityId: string;
  summary: string;
};

function mergeAuditRows(data: AuditEventsResponse | undefined): AuditRow[] {
  if (!data) return [];
  const core: AuditRow[] = data.audit_events.map((e) => ({
    key: `core::${e.id}`,
    createdAt: e.event_time,
    source: "core",
    eventType: e.event_type,
    entityId: e.entity_id,
    summary: e.details,
  }));
  const amie: AuditRow[] = data.amie_audit_log.map((e) => ({
    key: `amie::${e.id}`,
    createdAt: e.created_at,
    source: "amie",
    eventType: e.action,
    entityId: e.entity_id ?? "",
    summary: e.summary ?? "",
  }));
  return [...core, ...amie].sort((a, b) => b.createdAt.localeCompare(a.createdAt));
}

function AuditEventsTable({
  loading,
  error,
  data,
  onRetry,
}: {
  loading: boolean;
  error: Error | null;
  data: AuditEventsResponse | undefined;
  onRetry: () => void;
}) {
  const rows = React.useMemo(() => mergeAuditRows(data), [data]);

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
              key={r.key}
              className={cn(
                "border-t border-[color:var(--border)]",
                i % 2 === 1 ? "bg-[color:var(--muted-2)]" : "bg-[color:var(--card)]",
              )}
            >
              <td
                className="px-3 py-2 tabular-nums text-muted-foreground"
                title={formatAbsoluteUtc(r.createdAt)}
              >
                {formatRelative(r.createdAt)}
              </td>
              <td className="px-3 py-2">
                <SourcePill source={r.source} size="sm" />
              </td>
              <td className="px-3 py-2 font-mono text-[12.5px]">{r.eventType}</td>
              <td className="px-3 py-2 font-mono text-[12.5px]">{r.entityId}</td>
              <td className="px-3 py-2 text-muted-foreground">{r.summary}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
