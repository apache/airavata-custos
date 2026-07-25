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
import { StatusBadge, type StatusBadgeVariant } from "@/shared/ui/StatusBadge";
import { useAccessStatus } from "../queries";
import type { AccessCheck, AccessCheckEvent, AccessUIState } from "../schemas";

export const CHECK_NAME: Record<string, string> = {
  SIGN_IN: "Cluster sign-in",
  JOB_SUBMISSION: "Job submission",
};

const CHECK_DESCRIPTION: Record<string, string> = {
  SIGN_IN: "Signing in over SSH.",
  JOB_SUBMISSION: "Running jobs charged to this allocation.",
};

export const CHECK_CHIP: Record<AccessUIState, { variant: StatusBadgeVariant; label: string }> = {
  ok: { variant: "active", label: "Ready" },
  setting_up: { variant: "pending", label: "Setting up" },
  retrying: { variant: "warning", label: "Retrying" },
  stuck: { variant: "deleted", label: "Not working" },
};

// The band's overall tone is the worst check's tone.
const BAND_ORDER: AccessUIState[] = ["stuck", "retrying", "setting_up", "ok"];

const BAND_BADGE: Record<AccessUIState, { variant: StatusBadgeVariant; label: string }> = {
  ok: { variant: "active", label: "Live" },
  setting_up: { variant: "pending", label: "Setting up" },
  retrying: { variant: "warning", label: "Reconnecting" },
  stuck: { variant: "deleted", label: "Needs attention" },
};

export function AllocationAccessBand({
  allocationId,
  allocationName,
}: {
  allocationId: string;
  allocationName: string;
}) {
  const query = useAccessStatus(allocationId);
  const [expanded, setExpanded] = React.useState(false);

  // Renders nothing until the first result: access is usually live, and a
  // placeholder would push the page around for the common case.
  if (!query.data) return null;

  const checks = [...query.data.checks].sort((a) => (a.type === "SIGN_IN" ? -1 : 1));
  // Both checks log STARTED at the same setup; the story needs it once, at
  // the moment setup began. Events arrive newest first, so keep the last.
  const lastStarted = query.data.events.reduce(
    (acc, e, i) => (e.event_type === "STARTED" ? i : acc),
    -1,
  );
  const events = query.data.events.filter(
    (e, i) => e.event_type !== "STARTED" || i === lastStarted,
  );
  const overall = BAND_ORDER.find((s) => checks.some((c) => c.ui_state === s)) ?? "ok";
  const showDetails = overall !== "ok" || expanded;

  return (
    <section
      aria-label="Access to this allocation"
      className="rounded-xl border border-border bg-card p-4"
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-3">
          <StatusBadge {...BAND_BADGE[overall]} />
          <span className="text-sm font-medium">{verdictHeading(overall, checks)}</span>
          <span className="text-xs text-muted-foreground">
            {allocationName} · checked {relativeTime(lastChecked(checks))}
          </span>
        </div>
        {overall === "ok" ? (
          <button
            type="button"
            className="text-xs font-medium text-muted-foreground underline-offset-2 hover:underline"
            onClick={() => setExpanded((v) => !v)}
          >
            {expanded ? "Hide checks" : "View checks"}
          </button>
        ) : null}
      </div>

      {showDetails ? (
        <div className="mt-4 space-y-4">
          <p className="text-sm text-muted-foreground">{verdictBody(overall, checks)}</p>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
            {checks.map((c) => (
              <CheckRow key={c.type} check={c} />
            ))}
          </div>
          {events.length > 0 ? <Timeline events={events} /> : null}
        </div>
      ) : null}
    </section>
  );
}

function CheckRow({ check }: { check: AccessCheck }) {
  const chip = CHECK_CHIP[check.ui_state];
  return (
    <div className="flex items-start justify-between gap-3 rounded-lg border border-border p-3">
      <div className="space-y-0.5">
        <div className="text-sm font-medium">{CHECK_NAME[check.type]}</div>
        <div className="text-xs text-muted-foreground">{CHECK_DESCRIPTION[check.type]}</div>
        {check.ui_state === "retrying" ? (
          <div className="text-xs text-[color:var(--tone-warn-fg)]">
            Retrying automatically. Usually clears within a few minutes.
          </div>
        ) : null}
        {check.ui_state === "stuck" && check.failing_since ? (
          <div className="text-xs text-[color:var(--tone-error-fg)]">
            Failing since {timeLabel(check.failing_since)}.
          </div>
        ) : null}
      </div>
      <StatusBadge {...chip} />
    </div>
  );
}

const EVENT_LABEL: Record<AccessCheckEvent["event_type"], (name: string) => string> = {
  STARTED: () => "Access setup started",
  ONLINE: (name) => `${name} came online`,
  FAILED: (name) => `${name} failed a check`,
  RECOVERED: (name) => `${name} came back online`,
  STUCK: (name) => `${name} still failing`,
};

function Timeline({ events }: { events: AccessCheckEvent[] }) {
  return (
    <div>
      <h3 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        How your access came online
      </h3>
      <ul className="mt-2 space-y-1.5 border-l border-border pl-4">
        {events.map((e, i) => (
          <li
            key={`${e.check_type}-${e.event_type}-${e.occurred_at}-${i}`}
            className="flex items-baseline justify-between gap-3 text-sm"
          >
            <span>{EVENT_LABEL[e.event_type](CHECK_NAME[e.check_type] ?? "Access")}</span>
            <span className="shrink-0 text-xs text-muted-foreground">
              {timeLabel(e.occurred_at)}
            </span>
          </li>
        ))}
      </ul>
    </div>
  );
}

function verdictHeading(overall: AccessUIState, checks: AccessCheck[]): string {
  switch (overall) {
    case "ok":
      return "Your access is live";
    case "setting_up":
      return "Setting up your access";
    case "retrying":
      return "Reconnecting your access";
    case "stuck":
      return `Your ${worstName(checks, "stuck").toLowerCase()} access needs attention`;
  }
}

function verdictBody(overall: AccessUIState, checks: AccessCheck[]): string {
  const signInOK = checks.some((c) => c.type === "SIGN_IN" && c.ui_state === "ok");
  switch (overall) {
    case "ok":
      return "You can sign in to the cluster and submit jobs against this allocation. Nothing for you to do.";
    case "setting_up":
      return signInOK
        ? "Your cluster sign-in is ready. Job submission is still being set up. This is automatic and usually takes a few minutes."
        : "Your access is being set up. This is automatic and usually takes a few minutes.";
    case "retrying": {
      const name = worstName(checks, "retrying").toLowerCase();
      return `A system check for ${name} is failing right now. This is almost always temporary. We're retrying automatically, and no action is needed from you.`;
    }
    case "stuck": {
      const name = worstName(checks, "stuck");
      const rest = signInOK && name !== "Cluster sign-in" ? " Your sign-in still works." : "";
      return `${name} has been failing its checks for longer than a normal hiccup.${rest} Contact support if it doesn't clear.`;
    }
  }
}

function worstName(checks: AccessCheck[], state: AccessUIState): string {
  const check = checks.find((c) => c.ui_state === state);
  return (check && CHECK_NAME[check.type]) || "Access";
}

function lastChecked(checks: AccessCheck[]): string {
  return checks.map((c) => c.last_checked_at).sort().at(-1) ?? "";
}

function relativeTime(iso: string): string {
  if (!iso) return "just now";
  const seconds = Math.max(0, Math.floor((Date.now() - Date.parse(iso)) / 1000));
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

function timeLabel(iso: string): string {
  const date = new Date(iso);
  const today = new Date();
  const sameDay = date.toDateString() === today.toDateString();
  const time = date.toLocaleTimeString(undefined, { hour: "numeric", minute: "2-digit" });
  if (sameDay) return `Today · ${time}`;
  return `${date.toLocaleDateString(undefined, { month: "short", day: "numeric" })} · ${time}`;
}
