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

import sourcesFixture from "@/features/core/audit/__fixtures__/sources.json";
import approveDone from "@/features/core/audit/__fixtures__/trace.approve.done.json";
import approveFailed from "@/features/core/audit/__fixtures__/trace.approve.failed.json";
import approveRetrying from "@/features/core/audit/__fixtures__/trace.approve.retrying.json";
import createdDone from "@/features/core/audit/__fixtures__/trace.created.done.json";
import deniedDone from "@/features/core/audit/__fixtures__/trace.denied.done.json";
import packetFailed from "@/features/core/audit/__fixtures__/trace.packet.failed.json";
import tracesListFixture from "@/features/core/audit/__fixtures__/traces.list.json";
import { http, HttpResponse } from "msw";

type TraceSummary = (typeof tracesListFixture)["traces"][number];

// The fixtures were captured at one moment. Shift them so they always sit
// inside the list's default window.
const captured = Date.parse(tracesListFixture.traces[0]?.started_at ?? new Date().toISOString());
const shift = Date.now() - 20 * 60_000 - captured;
const shifted = (iso: string) => new Date(Date.parse(iso) + shift).toISOString();

const allTraces: TraceSummary[] = tracesListFixture.traces.map((t) => ({
  ...t,
  started_at: shifted(t.started_at),
  ended_at: shifted(t.ended_at),
}));

const traceDetailsById: Record<string, unknown> = Object.fromEntries(
  [approveFailed, approveRetrying, approveDone, deniedDone, packetFailed, createdDone].map((d) => [
    d.trace_id,
    {
      ...d,
      tree: shiftTree(d.tree as TreeNode[]),
      deliveries: d.deliveries.map((x) => ({ ...x, next_run_at: shifted(x.next_run_at) })),
    },
  ]),
);

type TreeNode = { created_at: string; children: TreeNode[] } & Record<string, unknown>;
function shiftTree(nodes: TreeNode[]): TreeNode[] {
  return nodes.map((n) => ({
    ...n,
    created_at: shifted(n.created_at),
    children: shiftTree(n.children),
  }));
}

function filterList(url: URL) {
  const sources = url.searchParams.getAll("source");
  const statuses = url.searchParams.getAll("status");
  const from = url.searchParams.get("from");
  const to = url.searchParams.get("to");
  const q = url.searchParams.get("q")?.toLowerCase() ?? "";
  let filtered = allTraces;
  if (sources.length) filtered = filtered.filter((t) => sources.includes(t.source));
  if (statuses.length) filtered = filtered.filter((t) => statuses.includes(t.status));
  if (from) filtered = filtered.filter((t) => t.started_at >= from);
  if (to) filtered = filtered.filter((t) => t.started_at <= to);
  if (q) {
    filtered = filtered.filter(
      (t) =>
        t.trace_id.toLowerCase().startsWith(q) ||
        t.root_operation.toLowerCase().includes(q) ||
        JSON.stringify(traceDetailsById[t.trace_id] ?? "")
          .toLowerCase()
          .includes(q),
    );
  }
  const limit = Number(url.searchParams.get("limit") ?? "50");
  const offset = Number(url.searchParams.get("offset") ?? "0");
  return { traces: filtered.slice(offset, offset + limit), total: filtered.length, limit, offset };
}

export const tracesHandlers = [
  http.get("/api/v1/audit/traces", ({ request }) =>
    HttpResponse.json(filterList(new URL(request.url))),
  ),

  http.get("/api/v1/audit/traces/:traceId", ({ params }) => {
    const detail = traceDetailsById[String(params.traceId)];
    if (!detail) return HttpResponse.json({ error: "trace not found" }, { status: 404 });
    return HttpResponse.json(detail);
  }),

  http.get("/api/v1/audit/sources", () => HttpResponse.json(sourcesFixture)),
];
