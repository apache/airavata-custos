/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import type { ListStatus, Step, StepNode, TraceDelivery, TraceDetail, TraceSummary } from "./types";

// Must match the backend's retry limit.
export const MAX_DELIVERY_TRIES = 10;

// Rows the event worker writes for a delivery try carry this entity type.
export const DELIVERY_ENTITY = "event_delivery";

// ---------- formatting ----------

export function shortHex(value: string | undefined | null, length = 8): string {
  if (!value) return "";
  return value.length > length ? value.slice(0, length) : value;
}

// Offset of a step from the trace start, short enough for a row.
export function formatOffset(ms: number): string {
  if (!Number.isFinite(ms) || ms < 0) return "";
  if (ms < 60_000) return `+${(ms / 1000).toFixed(3)}s`;
  if (ms < 3_600_000) return `+${Math.floor(ms / 60_000)}m`;
  return `+${Math.floor(ms / 3_600_000)}h`;
}

export function formatAbsoluteUtc(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return `${d.toISOString().replace("T", " ").replace("Z", "")} UTC`;
}

const MIN_MS = 60 * 1000;
const HOUR_MS = 60 * MIN_MS;
const DAY_MS = 24 * HOUR_MS;

export function formatRelative(iso: string, now: number = Date.now()): string {
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return iso;
  const ageMs = Math.max(0, now - t);
  if (ageMs < MIN_MS) return "just now";
  if (ageMs < HOUR_MS) return `${Math.floor(ageMs / MIN_MS)} min ago`;
  if (ageMs < DAY_MS) return `${Math.floor(ageMs / HOUR_MS)}h ago`;
  return `${Math.floor(ageMs / DAY_MS)}d ago`;
}

// "in 4m" for a time ahead, "now" once it has passed.
export function formatUntil(iso: string, now: number = Date.now()): string {
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return iso;
  const ms = t - now;
  if (ms <= 0) return "now";
  if (ms < MIN_MS) return `in ${Math.ceil(ms / 1000)}s`;
  if (ms < HOUR_MS) return `in ${Math.ceil(ms / MIN_MS)}m`;
  return `in ${Math.ceil(ms / HOUR_MS)}h`;
}

// A code-shaped name has no whitespace and a `.`, `:` or `_` in it, so event
// types and span names render in mono while plain titles stay in sans.
export function isCodeShaped(name: string): boolean {
  if (/\s/.test(name)) return false;
  return /[.:_]/.test(name);
}

// ---------- status words ----------

export function listStatus(t: TraceSummary): ListStatus {
  if (t.status === "error") return "failed";
  if (t.status === "in_progress") return t.deliveries.attempts > 0 ? "retrying" : "waiting";
  return "done";
}

export function listStatusLabel(t: TraceSummary): string {
  const s = listStatus(t);
  if (s === "retrying") return `Retrying ${t.deliveries.attempts} of ${MAX_DELIVERY_TRIES}`;
  return STATUS_WORDS[s];
}

export const STATUS_WORDS: Record<ListStatus, string> = {
  failed: "Failed",
  retrying: "Retrying",
  waiting: "Waiting",
  done: "Done",
};

export function deliveryListStatus(d: TraceDelivery): ListStatus {
  if (d.status === "FAILED") return "failed";
  if (d.status === "SUCCEEDED") return "done";
  return d.attempts > 0 ? "retrying" : "waiting";
}

export function deliveryLabel(d: TraceDelivery): string {
  switch (deliveryListStatus(d)) {
    case "failed":
      return `Failed after ${d.attempts} ${d.attempts === 1 ? "try" : "tries"}`;
    case "retrying":
      return `Retrying, try ${d.attempts} of ${MAX_DELIVERY_TRIES}`;
    case "waiting":
      return "Waiting";
    default:
      return d.attempts > 1 ? `Done after ${d.attempts} tries` : "Done";
  }
}

// The source pill for a connector, from its subscriber name.
function connectorSource(subscriber: string): string {
  return subscriber.split("-")[0] ?? subscriber;
}

// ---------- names ----------

const PLAIN_NAMES: Record<string, string> = {
  CLUSTER_ACCOUNT_APPROVED: "Account approved",
  CLUSTER_ACCOUNT_DENIED: "Account denied",
  CLUSTER_ADMIN_ACCOUNT_CREATED: "Admin account created",
  PACKET_RECEIVED: "Packet received",
  REPLY_SENT: "Reply sent",
  REPLY_HELD: "Reply held for approval",
  CREATE_PERSON: "Person created",
  CREATE_ACCOUNT: "Account created",
  CREATE_PROJECT: "Project created",
  CREATE_ALLOCATION: "Allocation created",
  CREATE_MEMBERSHIP: "Membership created",
  TRANSACTION_COMPLETE: "Transaction complete",
  NOTIFICATION_SENT: "Email sent",
  USER_BOOTSTRAPPED: "Admin user bootstrapped",
  ROLE_BOOTSTRAPPED: "Role bootstrapped",
  EVENT_DELIVERY_SUCCEEDED: "Delivery succeeded",
  EVENT_DELIVERY_FAILED: "Delivery failed",
  EVENT_DELIVERY_RETRIED: "Delivery retried by an admin",
  ComanageProvisioningStarted: "Provisioning started",
  ComanageProvisioningFailed: "Provisioning failed",
  ComanageClusterAccountAttached: "Cluster account attached",
  PosixUsernameBuildFailed: "Username could not be built",
};

export function plainName(eventType: string): string | null {
  return PLAIN_NAMES[eventType] ?? null;
}

// ---------- details ----------

export type ParsedDetails = { entries: Array<[string, string]> } | { text: string } | null;

// Details come as JSON, as `key=value` pairs, or as free text.
export function parseDetails(description: string | undefined): ParsedDetails {
  const text = description?.trim() ?? "";
  if (!text) return null;
  if (text.startsWith("{")) {
    try {
      const obj = JSON.parse(text) as unknown;
      if (obj && typeof obj === "object" && !Array.isArray(obj)) {
        const entries = Object.entries(obj as Record<string, unknown>).map(
          ([k, v]): [string, string] => [k, typeof v === "string" ? v : JSON.stringify(v)],
        );
        return { entries };
      }
    } catch {
      // not JSON after all
    }
  }
  const pairs = text.match(/\b[\w.]+=[^\s]+/g);
  if (pairs && pairs.length > 0 && text.replace(/\b[\w.]+=[^\s]+/g, "").trim() === "") {
    const entries = pairs.map((p): [string, string] => {
      const i = p.indexOf("=");
      return [p.slice(0, i), p.slice(i + 1)];
    });
    return { entries };
  }
  if (pairs && pairs.length > 0) {
    // `step=x err=some text with spaces`: the last value runs to the end.
    const entries: Array<[string, string]> = [];
    const re = /\b([\w.]+)=/g;
    const keys: Array<{ key: string; start: number; end: number }> = [];
    let m: RegExpExecArray | null = re.exec(text);
    while (m) {
      keys.push({ key: m[1] ?? "", start: m.index, end: m.index + m[0].length });
      m = re.exec(text);
    }
    for (let i = 0; i < keys.length; i++) {
      const cur = keys[i];
      const next = keys[i + 1];
      if (!cur) continue;
      const value = text.slice(cur.end, next ? next.start : undefined).trim();
      entries.push([cur.key, value]);
    }
    if (entries.length > 0) return { entries };
  }
  return { text };
}

function detailValue(step: Step, key: string): string | undefined {
  const parsed = parseDetails(step.description);
  if (!parsed || !("entries" in parsed)) return undefined;
  return parsed.entries.find(([k]) => k === key)?.[1];
}

export function usernameOf(step: Step): string | undefined {
  return detailValue(step, "local_username") ?? detailValue(step, "username");
}

// The error text a failed step carries, for the row and the panel.
export function errorText(step: Step): string | undefined {
  if (step.status !== "error") return undefined;
  const parsed = parseDetails(step.description);
  if (!parsed) return undefined;
  if ("text" in parsed) return parsed.text;
  const err = parsed.entries.find(([k]) => k === "error" || k === "err")?.[1];
  if (err) {
    const stepName = parsed.entries.find(([k]) => k === "step")?.[1];
    return stepName ? `${err} (step ${stepName})` : err;
  }
  return parsed.entries.map(([k, v]) => `${k}=${v}`).join(" ");
}

// ---------- flow model ----------

export type FlowStep = {
  step: Step;
  depth: number;
  parent: FlowStep | null;
  children: FlowStep[];
};

export type DeliveryGroup = {
  key: string;
  delivery: TraceDelivery;
  tries: FlowStep[];
};

export type Hop = {
  key: string;
  publisher: FlowStep | null;
  eventType: string;
  deliveries: DeliveryGroup[];
};

export type FlowGroup =
  | { kind: "origin"; key: string; root: FlowStep; siblings: FlowStep[] }
  | { kind: "hop"; key: string; hop: Hop }
  | { kind: "orphans"; key: string; roots: FlowStep[] };

export type FlowModel = {
  groups: FlowGroup[];
  hopsBySpan: Map<string, Hop[]>;
  steps: Map<string, FlowStep>;
  // Try row id, or the span a try row shares with the connector's own rows, to the delivery id.
  tryOf: Map<string, string>;
  ownerOfSpan: Map<string, string>;
  spanRows: Map<string, FlowStep[]>;
  stepCount: number;
  startedAt: string;
  failingStepId: string | null;
  errorStepIds: string[];
  pathKeys: Set<string>;
  truncated: boolean;
};

export type FlowRow =
  | {
      kind: "group";
      key: string;
      depth: number;
      title: string;
      source: string;
      status: ListStatus | "ok" | "error" | "muted";
      delivery?: TraceDelivery;
      expanded: boolean;
      onPath: boolean;
    }
  | {
      kind: "hop";
      key: string;
      depth: number;
      eventType: string;
      deliveries: TraceDelivery[];
      onPath: boolean;
    }
  | {
      kind: "step";
      key: string;
      depth: number;
      step: Step;
      node: FlowStep;
      onPath: boolean;
      failing: boolean;
      try?: { n: number; outcome: "failed" | "succeeded" | "retried" };
    }
  | { kind: "fold"; key: string; depth: number; count: number; onPath: boolean }
  | { kind: "ghost"; key: string; depth: number; text: string }
  | { kind: "truncated"; key: string; depth: number };

function hopKey(spanId: string, eventType: string): string {
  return `h:${spanId}:${eventType}`;
}

export function deliveryKey(id: string): string {
  return `d:${id}`;
}

export function originKey(rootStepId: string): string {
  return `o:${rootStepId}`;
}

export const ORPHANS_KEY = "orphans";

function foldKey(deliveryId: string): string {
  return `f:${deliveryId}`;
}

function tryNumber(step: Step, index: number): number {
  const n = Number(detailValue(step, "attempt"));
  return Number.isFinite(n) && n > 0 ? n : index + 1;
}

function tryOutcome(step: Step): "failed" | "succeeded" | "retried" {
  if (step.event_type.endsWith("SUCCEEDED")) return "succeeded";
  if (step.event_type.endsWith("RETRIED")) return "retried";
  return "failed";
}

// Joins the tree and the deliveries into groups: the origin steps, a hop per
// published event with one group per delivery, and the steps whose parent
// wrote no row. The failing step and the rail to it are decided here too.
export function buildFlow(detail: TraceDetail): FlowModel {
  const steps = new Map<string, FlowStep>();
  const spanFirst = new Map<string, FlowStep>();
  const roots: FlowStep[] = [];
  const walk = (n: StepNode, parent: FlowStep | null, depth: number) => {
    const { children: kids, ...step } = n;
    const node: FlowStep = { step, depth, parent, children: [] };
    steps.set(step.id, node);
    if (!spanFirst.has(step.span_id)) spanFirst.set(step.span_id, node);
    if (parent) parent.children.push(node);
    else roots.push(node);
    for (const c of kids) walk(c, node, depth + 1);
  };
  for (const r of detail.tree) walk(r, null, 0);

  const deliveriesById = new Map(detail.deliveries.map((d) => [d.id, d]));
  const tryOf = new Map<string, string>();
  const ownerOfSpan = new Map<string, string>();
  const spanRows = new Map<string, FlowStep[]>();
  const triesByDelivery = new Map<string, FlowStep[]>();
  for (const node of steps.values()) {
    const { step } = node;
    const list = spanRows.get(step.span_id) ?? [];
    list.push(node);
    spanRows.set(step.span_id, list);
    if (
      step.entity_type === DELIVERY_ENTITY &&
      step.entity_id &&
      deliveriesById.has(step.entity_id)
    ) {
      tryOf.set(step.id, step.entity_id);
      ownerOfSpan.set(step.span_id, step.entity_id);
      const tries = triesByDelivery.get(step.entity_id) ?? [];
      tries.push(node);
      triesByDelivery.set(step.entity_id, tries);
    }
  }
  // A connector with no span of its own writes its rows in the delivery's
  // span, beside the try row, so those rows belong to the delivery too.
  const owned = (node: FlowStep): boolean =>
    tryOf.has(node.step.id) || ownerOfSpan.has(node.step.span_id);

  const hopsBySpan = new Map<string, Hop[]>();
  const rootHops: Hop[] = [];
  const hopByKey = new Map<string, Hop>();
  for (const d of detail.deliveries) {
    const key = hopKey(d.span_id, d.event_type);
    let hop = hopByKey.get(key);
    if (!hop) {
      const publisher = spanFirst.get(d.span_id) ?? null;
      hop = { key, publisher, eventType: d.event_type, deliveries: [] };
      hopByKey.set(key, hop);
      if (publisher) {
        const list = hopsBySpan.get(d.span_id) ?? [];
        list.push(hop);
        hopsBySpan.set(d.span_id, list);
      } else {
        rootHops.push(hop);
      }
    }
    hop.deliveries.push({
      key: deliveryKey(d.id),
      delivery: d,
      tries: triesByDelivery.get(d.id) ?? [],
    });
  }

  // Roots in one span are one group. The first root's parent, when it wrote no
  // row, is the request span, so roots under the same parent are origins too.
  // A root under some other missing parent is an orphan.
  const groups: FlowGroup[] = [];
  const orphans: FlowStep[] = [];
  const origins = new Map<string, FlowStep[]>();
  const firstRoot = roots.find((r) => !owned(r));
  for (const root of roots) {
    if (owned(root)) continue;
    const parent = root.step.parent_span_id;
    if (parent && !spanFirst.has(parent) && parent !== firstRoot?.step.parent_span_id) {
      orphans.push(root);
      continue;
    }
    const same = origins.get(root.step.span_id);
    if (same) same.push(root);
    else {
      const siblings: FlowStep[] = [];
      origins.set(root.step.span_id, siblings);
      groups.push({ kind: "origin", key: originKey(root.step.id), root, siblings });
    }
  }
  for (const hop of rootHops) groups.push({ kind: "hop", key: hop.key, hop });
  if (orphans.length > 0) groups.push({ kind: "orphans", key: ORPHANS_KEY, roots: orphans });

  const startedAt = [...steps.values()].reduce(
    (min, n) => (min === "" || n.step.created_at < min ? n.step.created_at : min),
    "",
  );

  const model: FlowModel = {
    groups,
    hopsBySpan,
    steps,
    tryOf,
    ownerOfSpan,
    spanRows,
    stepCount: steps.size,
    startedAt,
    failingStepId: null,
    errorStepIds: [],
    pathKeys: new Set(),
    truncated: detail.truncated,
  };

  // Error steps in display order. A failed try inside a delivery that later
  // succeeded is history. A try row only counts when its delivery wrote nothing else.
  const deliveryOf = (node: FlowStep): string | undefined => {
    for (let cur: FlowStep | null = node; cur; cur = cur.parent) {
      const id = tryOf.get(cur.step.id) ?? ownerOfSpan.get(cur.step.span_id);
      if (id) return id;
    }
    return undefined;
  };
  const everything = new Set<string>();
  for (const g of groups) everything.add(g.key);
  for (const d of detail.deliveries) everything.add(deliveryKey(d.id));
  for (const d of detail.deliveries) everything.add(foldKey(d.id));
  const display = flattenFlow(model, everything);
  const candidates = display.flatMap((row) => {
    if (row.kind !== "step" || row.step.status !== "error") return [];
    const deliveryId = deliveryOf(row.node);
    if (deliveryId && deliveriesById.get(deliveryId)?.status === "SUCCEEDED") return [];
    return [row];
  });
  const nonTry = new Set(
    candidates.filter((r) => !tryOf.has(r.step.id)).map((r) => deliveryOf(r.node)),
  );
  model.errorStepIds = candidates
    .filter((r) => !tryOf.has(r.step.id) || !nonTry.has(deliveryOf(r.node)))
    .map((r) => r.step.id);
  model.failingStepId = model.errorStepIds[0] ?? null;

  if (model.failingStepId) {
    const failing = steps.get(model.failingStepId);
    for (let cur: FlowStep | null = failing ?? null; cur; cur = cur.parent) {
      // A try row shows after the connector's rows, so the path skips it
      // unless the try row is the failing step itself.
      if (!tryOf.has(cur.step.id) || cur === failing) model.pathKeys.add(cur.step.id);
      const deliveryId = tryOf.get(cur.step.id) ?? ownerOfSpan.get(cur.step.span_id);
      const delivery = deliveryId ? deliveriesById.get(deliveryId) : undefined;
      if (delivery) {
        model.pathKeys.add(deliveryKey(delivery.id));
        model.pathKeys.add(hopKey(delivery.span_id, delivery.event_type));
      }
      if (!cur.parent && !owned(cur)) {
        const node = cur;
        const group = groups.find(
          (g) => g.kind === "origin" && (g.root === node || g.siblings.includes(node)),
        );
        if (group) model.pathKeys.add(group.key);
      }
    }
  }
  return model;
}

// Every group open for a short trace. For a long one, only the groups on the
// rail, so the failing step is in view when the drawer opens.
export function defaultExpanded(model: FlowModel): Set<string> {
  const short = model.stepCount < 40;
  const open = new Set<string>();
  for (const key of expandableKeys(model)) {
    if (key.startsWith("f:")) continue;
    if (short || key.startsWith("o:") || model.pathKeys.has(key)) open.add(key);
  }
  open.delete(ORPHANS_KEY);
  return open;
}

// Flattens the groups into the rows the flow shows, honoring what is expanded.
export function flattenFlow(
  model: FlowModel,
  expanded: Set<string>,
  now: number = Date.now(),
): FlowRow[] {
  const rows: FlowRow[] = [];
  const onPath = (key: string) => model.pathKeys.has(key);
  const owned = (node: FlowStep) =>
    model.tryOf.has(node.step.id) || model.ownerOfSpan.has(node.step.span_id);

  const emitSubtree = (node: FlowStep, depth: number) => {
    rows.push({
      kind: "step",
      key: node.step.id,
      depth,
      step: node.step,
      node,
      onPath: onPath(node.step.id),
      failing: node.step.id === model.failingStepId,
    });
    emitRuns(
      node.children.filter((c) => !owned(c)),
      depth + 1,
    );
  };

  // Rows of one span stay together in time order, and the events that span
  // published come after its last row.
  const emitRuns = (nodes: FlowStep[], depth: number) => {
    let run: FlowStep[] = [];
    const flush = () => {
      for (const n of run) emitSubtree(n, depth);
      const hops = run[0] ? model.hopsBySpan.get(run[0].step.span_id) : undefined;
      if (hops && run.some((n) => hops[0]?.publisher === n)) {
        for (const hop of hops) emitHop(hop, depth);
      }
      run = [];
    };
    for (const n of nodes) {
      if (run[0] && run[0].step.span_id !== n.step.span_id) flush();
      run.push(n);
    }
    if (run.length) flush();
  };

  const emitDelivery = (dg: DeliveryGroup, depth: number) => {
    const { delivery, tries } = dg;
    const isOpen = expanded.has(dg.key);
    rows.push({
      kind: "group",
      key: dg.key,
      depth,
      title: delivery.subscriber,
      source: connectorSource(delivery.subscriber),
      status: deliveryListStatus(delivery),
      delivery,
      expanded: isOpen,
      onPath: onPath(dg.key),
    });
    if (!isOpen) return;
    const foldOpen = expanded.has(foldKey(delivery.id));
    const shown: Array<{ node: FlowStep; index: number } | { fold: number }> = [];
    if (tries.length > 3 && !foldOpen) {
      const first = tries[0];
      const last = tries[tries.length - 1];
      if (first) shown.push({ node: first, index: 0 });
      shown.push({ fold: tries.length - 2 });
      if (last) shown.push({ node: last, index: tries.length - 1 });
    } else {
      tries.forEach((node, index) => shown.push({ node, index }));
    }
    for (const item of shown) {
      if ("fold" in item) {
        rows.push({
          kind: "fold",
          key: foldKey(delivery.id),
          depth: depth + 1,
          count: item.fold,
          onPath: false,
        });
        continue;
      }
      // The connector's rows for this try: the ones sharing the try's span,
      // then the ones under the try row, each with its own subtree.
      emitRuns(
        (model.spanRows.get(item.node.step.span_id) ?? []).filter((s) => s !== item.node),
        depth + 1,
      );
      emitRuns(item.node.children, depth + 1);
      rows.push({
        kind: "step",
        key: item.node.step.id,
        depth: depth + 1,
        step: item.node.step,
        node: item.node,
        onPath: onPath(item.node.step.id),
        failing: item.node.step.id === model.failingStepId,
        try: { n: tryNumber(item.node.step, item.index), outcome: tryOutcome(item.node.step) },
      });
    }
    if (delivery.status === "PENDING") {
      const text =
        delivery.attempts === 0
          ? tries.length === 0
            ? `Waiting for the first try, ${formatUntil(delivery.next_run_at, now)}`
            : "Waiting"
          : `Next try ${formatUntil(delivery.next_run_at, now)}, ${MAX_DELIVERY_TRIES - delivery.attempts} left`;
      rows.push({ kind: "ghost", key: `w:${delivery.id}`, depth: depth + 1, text });
    }
  };

  const emitHop = (hop: Hop, depth: number) => {
    rows.push({
      kind: "hop",
      key: hop.key,
      depth,
      eventType: hop.eventType,
      deliveries: hop.deliveries.map((d) => d.delivery),
      onPath: onPath(hop.key),
    });
    for (const dg of hop.deliveries) emitDelivery(dg, depth + 1);
  };

  for (const g of model.groups) {
    if (g.kind === "origin") {
      const isOpen = expanded.has(g.key);
      const members = [g.root, ...g.siblings];
      rows.push({
        kind: "group",
        key: g.key,
        depth: 0,
        title: originTitle(g.root.step),
        source: g.root.step.source,
        status: members.some((m) => originStatus(m, model) === "error") ? "error" : "ok",
        expanded: isOpen,
        onPath: onPath(g.key),
      });
      if (isOpen) emitRuns(members, 1);
    } else if (g.kind === "hop") {
      emitHop(g.hop, 0);
    } else {
      const isOpen = expanded.has(g.key);
      const n = g.roots.reduce((sum, r) => sum + 1 + countSteps(r, model), 0);
      rows.push({
        kind: "group",
        key: g.key,
        depth: 0,
        title: `${n} ${n === 1 ? "step" : "steps"} without a parent`,
        source: "",
        status: "muted",
        expanded: isOpen,
        onPath: false,
      });
      if (isOpen) for (const r of g.roots) emitSubtree(r, 1);
    }
  }
  if (model.truncated) rows.push({ kind: "truncated", key: "truncated", depth: 0 });

  // The rail runs through every row between a group's header and the row
  // where it leaves the group, so rows there read as on the path too.
  let inside = false;
  for (const row of rows) {
    if (row.kind === "group") {
      inside = row.onPath;
      continue;
    }
    if (row.kind === "ghost" || row.kind === "truncated") continue;
    if (inside && !row.onPath) row.onPath = true;
    if (row.kind === "hop" && row.onPath) inside = false;
    if (row.kind === "step" && row.failing) inside = false;
  }
  return rows;
}

// Which rail segments carry the failing path. For each row, one flag per
// level 1..depth: true when the header owning that rail is on the path and
// the path continues at or below this row inside that header's rows.
function onPathOf(row: FlowRow): boolean {
  return "onPath" in row && row.onPath;
}

export function pathRails(rows: FlowRow[]): boolean[][] {
  // ponytail: O(rows * depth) scan, rows are capped at 500 by the backend.
  const lastOnPath = rows.map((row, i) => {
    let last = -1;
    for (let j = i + 1; j < rows.length; j++) {
      const later = rows[j];
      if (!later || later.depth <= row.depth) break;
      if (onPathOf(later)) last = j;
    }
    return last;
  });
  const owners: number[] = [];
  return rows.map((row, i) => {
    owners.length = row.depth;
    const flags: boolean[] = [];
    for (let level = 1; level <= row.depth; level++) {
      const owner = owners[level - 1];
      const ownerRow = owner === undefined ? undefined : rows[owner];
      flags.push(
        ownerRow !== undefined && onPathOf(ownerRow) && (lastOnPath[owner ?? -1] ?? -1) >= i,
      );
    }
    owners[row.depth] = i;
    return flags;
  });
}

// Keys of the groups and folds that can open and close.
export function expandableKeys(model: FlowModel): string[] {
  const keys: string[] = [];
  const deliveries = (hops: Hop[]) => {
    for (const hop of hops) {
      for (const dg of hop.deliveries) {
        keys.push(dg.key);
        if (dg.tries.length > 3) keys.push(foldKey(dg.delivery.id));
      }
    }
  };
  for (const g of model.groups) {
    if (g.kind === "hop") deliveries([g.hop]);
    else keys.push(g.key);
  }
  for (const hops of model.hopsBySpan.values()) deliveries(hops);
  return keys;
}

// The groups and folds that must be open for the row with this key to show.
export function containersOf(model: FlowModel, key: string): string[] {
  const groups = expandableKeys(model).filter((k) => !k.startsWith("f:"));
  const find = (expanded: Set<string>) => {
    const stack: string[] = [];
    for (const row of flattenFlow(model, expanded)) {
      stack.length = row.depth;
      if (row.key === key) return stack.filter(Boolean);
      if (row.kind === "group") stack[row.depth] = row.key;
    }
    return null;
  };
  const found = find(new Set(groups));
  if (found) return found;
  // Hidden by a fold, so that fold opens too. A try sits right under its delivery.
  const withFolds = find(new Set(expandableKeys(model)));
  if (!withFolds) return [];
  const delivery = withFolds.findLast((k) => k.startsWith("d:"));
  return delivery ? [...withFolds, foldKey(delivery.slice(2))] : withFolds;
}

// Steps under a node, not counting rows that belong to a delivery.
function countSteps(node: FlowStep, model: FlowModel): number {
  let n = 0;
  for (const c of node.children) {
    if (model.tryOf.has(c.step.id) || model.ownerOfSpan.has(c.step.span_id)) continue;
    n += 1 + countSteps(c, model);
  }
  return n;
}

function originStatus(root: FlowStep, model: FlowModel): "ok" | "error" {
  const visit = (node: FlowStep): boolean => {
    if (node.step.status === "error") return true;
    return node.children.some(
      (c) => !model.tryOf.has(c.step.id) && !model.ownerOfSpan.has(c.step.span_id) && visit(c),
    );
  };
  return visit(root) ? "error" : "ok";
}

// "Account approved for jsmith", or the event type when there is no plain name.
export function originTitle(step: Step): string {
  const name = plainName(step.event_type) ?? step.event_type;
  const username = usernameOf(step);
  return username ? `${name} for ${username}` : name;
}

// ---------- entities ----------

export type EntityLink = { kind: string; id: string; href: string | null };

const ENTITY_KINDS: Record<string, { kind: string; href: ((id: string) => string) | null }> = {
  packet: { kind: "Packet", href: (id) => `/admin/amie/packets/${encodeURIComponent(id)}` },
  project: { kind: "Project", href: (id) => `/projects/${encodeURIComponent(id)}` },
  compute_allocation: {
    kind: "Allocation",
    href: (id) => `/allocations/${encodeURIComponent(id)}`,
  },
  compute_allocation_membership: { kind: "Membership", href: null },
  compute_cluster_user: { kind: "Cluster account", href: null },
  user: { kind: "User", href: null },
  organization: { kind: "Organization", href: null },
};

// The things the trace is about, in the order they first appear. Delivery
// rows name the delivery, which is not an entity the admin opens from here.
export function entityLinks(detail: TraceDetail): EntityLink[] {
  const out: EntityLink[] = [];
  const seen = new Set<string>();
  const walk = (n: StepNode) => {
    const cfg = n.entity_type ? ENTITY_KINDS[n.entity_type] : undefined;
    if (cfg && n.entity_id && !seen.has(`${n.entity_type}:${n.entity_id}`)) {
      seen.add(`${n.entity_type}:${n.entity_id}`);
      out.push({ kind: cfg.kind, id: n.entity_id, href: cfg.href ? cfg.href(n.entity_id) : null });
    }
    for (const c of n.children) walk(c);
  };
  for (const r of detail.tree) walk(r);
  return out;
}

export function countRows(nodes: StepNode[]): number {
  return nodes.reduce((n, node) => n + 1 + countRows(node.children), 0);
}

// The first step of the trace, which names it in the list and the drawer.
export function rootStep(detail: TraceDetail): Step | null {
  const first = detail.tree[0];
  if (!first) return null;
  const { children: _children, ...step } = first;
  return step;
}
