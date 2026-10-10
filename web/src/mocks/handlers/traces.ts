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
import type {
  GetAuditEventsResponse,
  GetAuditSourcesResponse,
  GetAuditTracesByTraceIdData,
  GetAuditTracesByTraceIdResponse,
  TraceSummary,
} from "@/generated/core/types.gen";
import {
  zGetAuditEventsResponse,
  zGetAuditSourcesResponse,
  zGetAuditTracesByTraceIdResponse,
  zTraceSummary,
} from "@/generated/core/zod.gen";
import { z } from "zod";
import { notFound, page } from "../paging";

// Shift every fixture timestamp so the latest trace ends now, as a live 30d window would.
const shift = Date.now() - Math.max(...tracesListFixture.traces.map((t) => Date.parse(t.ended_at)));
const rebase = (fixture: unknown): unknown =>
  JSON.parse(JSON.stringify(fixture), (_key, v) =>
    typeof v === "string" && /^\d{4}-\d\d-\d\dT[\d:.]+Z$/.test(v)
      ? new Date(Date.parse(v) + shift).toISOString()
      : v,
  );

const allTraces: TraceSummary[] = z.array(zTraceSummary).parse(rebase(tracesListFixture.traces));

const traceDetails: GetAuditTracesByTraceIdResponse[] = z
  .array(zGetAuditTracesByTraceIdResponse)
  .parse(rebase([failedFixture, successFixture, httpFixture, inProgressFixture]));

const auditEventsById: Record<string, GetAuditEventsResponse> = z
  .record(z.string(), zGetAuditEventsResponse)
  .parse(rebase(auditEventsFixture));

const sourceList: GetAuditSourcesResponse = zGetAuditSourcesResponse.parse(sourcesFixture);

// Mirrors the store: from/to match any event's created_at.
function filterList(url: URL) {
  const sources = url.searchParams.getAll("source");
  const statuses = url.searchParams.getAll("status");
  const from = url.searchParams.get("from");
  const to = url.searchParams.get("to");
  const q = url.searchParams.get("q")?.toLowerCase() ?? "";
  const filtered = allTraces.filter(
    ({ source, status, started_at, ended_at, trace_id, root_operation }) =>
      (!sources.length || sources.includes(source)) &&
      (!statuses.length || statuses.includes(status)) &&
      (!from || Date.parse(ended_at) >= Date.parse(from)) &&
      (!to || Date.parse(started_at) <= Date.parse(to)) &&
      (!q || trace_id.startsWith(q) || root_operation.toLowerCase().includes(q)),
  );
  const { items: traces, ...rest } = page(url, filtered);
  return { traces, ...rest };
}

export const tracesHandlers = [
  http.get("/api/v1/audit/traces", ({ request }) =>
    HttpResponse.json(filterList(new URL(request.url))),
  ),

  http.get<GetAuditTracesByTraceIdData["path"]>("/api/v1/audit/traces/:trace_id", ({ params }) => {
    const detail = traceDetails.find((d) => d.trace_id === params.trace_id);
    if (!detail) return notFound("trace");
    return HttpResponse.json(detail);
  }),

  http.get("/api/v1/audit/events", ({ request }) => {
    const traceId = new URL(request.url).searchParams.get("trace_id") ?? "";
    return HttpResponse.json({ events: auditEventsById[traceId]?.events ?? [] });
  }),

  http.get("/api/v1/audit/sources", () => HttpResponse.json(sourceList)),
];
