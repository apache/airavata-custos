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

import type { AuditEvent, PendingDelivery } from "@/generated/core/types.gen";

// Matches maxAttempts in pkg/events/worker.go.
export const MAX_ATTEMPTS = 10;

export type DeliveryState = "succeeded" | "failed" | "retrying" | "waiting" | "not-running";

// A pending delivery whose connector is not running waits until it starts,
// so it is shown apart from one that is simply queued.
export function deliveryState(
  delivery: PendingDelivery,
  runningSubscribers?: Set<string>,
): DeliveryState {
  if (delivery.status === "SUCCEEDED") return "succeeded";
  if (delivery.status === "FAILED") return "failed";
  if (runningSubscribers && !runningSubscribers.has(delivery.subscriber)) return "not-running";
  return delivery.attempts > 0 ? "retrying" : "waiting";
}

export type HistoryStep =
  | { kind: "attempt"; at: string; attempt: number; ok: boolean; error?: string }
  | { kind: "retry"; at: string; actorId: string; previousAttempts: number }
  | { kind: "other"; at: string; label: string };

function parseDetails(details: string): Record<string, unknown> {
  try {
    const parsed: unknown = JSON.parse(details);
    return parsed && typeof parsed === "object" ? (parsed as Record<string, unknown>) : {};
  } catch {
    return {};
  }
}

export function historyStep(entry: AuditEvent): HistoryStep {
  const d = parseDetails(entry.details);
  switch (entry.event_type) {
    case "EVENT_DELIVERY_SUCCEEDED":
    case "EVENT_DELIVERY_FAILED":
      return {
        kind: "attempt",
        at: entry.event_time,
        attempt: Number(d.attempt) || 0,
        ok: entry.event_type === "EVENT_DELIVERY_SUCCEEDED",
        error: typeof d.error === "string" ? d.error : undefined,
      };
    case "EVENT_DELIVERY_RETRIED":
      return {
        kind: "retry",
        at: entry.event_time,
        actorId: typeof d.actor_id === "string" ? d.actor_id : "",
        previousAttempts: Number(d.previous_attempts) || 0,
      };
    default:
      return { kind: "other", at: entry.event_time, label: entry.event_type };
  }
}

// A retry starts the attempt count over, so the history is split into runs at
// each retry to keep the attempt numbers readable.
export function historyRuns(history?: AuditEvent[] | null): HistoryStep[][] {
  const runs: HistoryStep[][] = [[]];
  for (const entry of history ?? []) {
    const step = historyStep(entry);
    if (step.kind === "retry") runs.push([step]);
    else runs.at(-1)?.push(step);
  }
  return runs.filter((run) => run.length > 0);
}

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

// "3 min ago" for the past and "in 3 min" for the future.
export function relativeTime(iso: string, now: number = Date.now()): string {
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return iso;
  const diff = t - now;
  const ms = Math.abs(diff);
  if (ms < MINUTE) return diff > 0 ? "in less than a minute" : "just now";
  const text =
    ms < HOUR
      ? `${Math.floor(ms / MINUTE)} min`
      : ms < DAY
        ? `${Math.floor(ms / HOUR)}h`
        : `${Math.floor(ms / DAY)}d`;
  return diff > 0 ? `in ${text}` : `${text} ago`;
}

export function absoluteTime(iso: string | undefined): string {
  if (!iso) return "-";
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? iso : d.toLocaleString();
}
