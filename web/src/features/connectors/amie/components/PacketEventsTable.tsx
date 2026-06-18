"use client";

import { CenteredSpinner } from "@/shared/ui/Loading";
// ADR-0004: sanctioned cross-feature import — ViewTraceLink is the ONE primitive
// other features may pull from core/audit. See web/CLAUDE.md "Pitfalls" §4.
import { ViewTraceLink } from "@/features/core/audit/components/ViewTraceLink";
import type { PacketEvent } from "../types";
import { formatDate } from "../utils";

function eventIconLabel(event: PacketEvent): string {
  switch (event.event_type) {
    case "RECEIVED":
      return "[●]";
    case "DECODED":
      return "[◇]";
    case "HANDLED":
      return "[✓]";
    case "FAILED":
      return "[!]";
    case "RETRY":
    case "RETRY_SCHEDULED":
      return "[↻]";
    case "MANUAL_RESOLVE":
    case "MANUAL_LINK":
      return "[★]";
    default:
      return "[•]";
  }
}

export type PacketEventsTableProps = {
  events: PacketEvent[];
  isLoading: boolean;
};

export function PacketEventsTable({ events, isLoading }: PacketEventsTableProps) {
  if (isLoading) return <CenteredSpinner label="Loading timeline" />;
  if (events.length === 0) {
    return <p className="text-sm text-muted-foreground">No events recorded.</p>;
  }
  return (
    <ol className="space-y-3" data-testid="packet-events-table">
      {events.map((event) => (
        <li
          key={event.id}
          data-testid={`packet-event-${event.id}`}
          className="flex items-start gap-3 rounded-md border bg-card p-3 text-sm"
        >
          <span aria-hidden="true" className="font-mono text-xs text-muted-foreground">
            {eventIconLabel(event)}
          </span>
          <div className="flex-1">
            <div className="flex items-baseline justify-between gap-2">
              <span className="font-medium">{event.event_type}</span>
              <time
                dateTime={event.timestamp}
                className="text-xs tabular-nums text-muted-foreground"
              >
                {formatDate(event.timestamp)}
              </time>
            </div>
            <p className="text-xs text-muted-foreground">
              {event.actor} · {event.status}
              {event.duration_ms !== undefined ? ` · ${event.duration_ms}ms` : null}
            </p>
            {event.message ? <p className="mt-1 text-sm">{event.message}</p> : null}
            {event.trace_id ? (
              <div className="mt-1">
                <ViewTraceLink traceId={event.trace_id} variant="text" />
              </div>
            ) : null}
          </div>
        </li>
      ))}
    </ol>
  );
}
