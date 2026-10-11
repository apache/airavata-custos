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

import { http, HttpResponse } from "msw";
import auditEventsFixture from "@/features/core/audit/__fixtures__/audit-events.json";
import sourcesFixture from "@/features/core/audit/__fixtures__/sources.json";
import failedFixture from "@/features/core/audit/__fixtures__/trace.amie.failed.json";
import successFixture from "@/features/core/audit/__fixtures__/trace.amie.success.json";
import httpFixture from "@/features/core/audit/__fixtures__/trace.http.json";
import inProgressFixture from "@/features/core/audit/__fixtures__/trace.in-progress.json";
import tracesListFixture from "@/features/core/audit/__fixtures__/traces.list.json";

type TraceSummary = {
  trace_id: string;
  root_operation: string;
  source: string;
  status: string;
  started_at: string;
  ended_at: string;
  event_count: number;
};

const allTraces: TraceSummary[] = (tracesListFixture as { traces: TraceSummary[] }).traces.map(
  (t) => ({ ...t }),
);

const traceDetailsById: Record<string, unknown> = {
  [failedFixture.trace_id]: failedFixture,
  [successFixture.trace_id]: successFixture,
  [httpFixture.trace_id]: httpFixture,
  [inProgressFixture.trace_id]: inProgressFixture,
};

const auditEventsById = auditEventsFixture as Record<string, { events: unknown[] }>;

function filterList(
  rows: TraceSummary[],
  url: URL,
): { rows: TraceSummary[]; total: number; limit: number; offset: number } {
  const sources = url.searchParams.getAll("source");
  const statuses = url.searchParams.getAll("status");
  const from = url.searchParams.get("from");
  const to = url.searchParams.get("to");
  const q = url.searchParams.get("q")?.toLowerCase() ?? "";
  let filtered = rows;
  if (sources.length) filtered = filtered.filter((t) => sources.includes(t.source));
  if (statuses.length) filtered = filtered.filter((t) => statuses.includes(t.status));
  if (from) filtered = filtered.filter((t) => t.started_at >= from);
  if (to) filtered = filtered.filter((t) => t.started_at <= to);
  if (q) {
    filtered = filtered.filter(
      (t) =>
        t.trace_id.toLowerCase().startsWith(q) ||
        t.root_operation.toLowerCase().includes(q),
    );
  }
  const limit = Number(url.searchParams.get("limit") ?? "50");
  const offset = Number(url.searchParams.get("offset") ?? "0");
  const page = filtered.slice(offset, offset + limit);
  return { rows: page, total: filtered.length, limit, offset };
}

export const tracesHandlers = [
  http.get("/api/v1/audit/traces", ({ request }) => {
    const url = new URL(request.url);
    const { rows, total, limit, offset } = filterList(allTraces, url);
    return HttpResponse.json({ traces: rows, total, limit, offset });
  }),

  http.get("/api/v1/audit/traces/:traceId", ({ params }) => {
    const traceId = String(params.traceId);
    const detail = traceDetailsById[traceId];
    if (!detail) return HttpResponse.json({ error: "trace not found" }, { status: 404 });
    return HttpResponse.json(detail);
  }),

  http.get("/api/v1/audit/events", ({ request }) => {
    const url = new URL(request.url);
    const traceId = url.searchParams.get("trace_id") ?? "";
    const bucket = auditEventsById[traceId];
    return HttpResponse.json({ events: bucket?.events ?? [] });
  }),

  http.get("/api/v1/audit/sources", () => {
    return HttpResponse.json(sourcesFixture);
  }),
];
