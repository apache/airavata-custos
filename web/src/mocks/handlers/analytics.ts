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

import contexts from "@/features/core/analytics/__fixtures__/contexts.json";
import jobsFixture from "@/features/core/analytics/__fixtures__/jobs.json";
import summaries from "@/features/core/analytics/__fixtures__/usage-summary.json";
import { zJob } from "@/generated/analytics/zod.gen";
import { http, HttpResponse } from "msw";
import { z } from "zod";

const summaryById = new Map<string, unknown>(Object.entries(summaries));
const jobsById = z
  .record(z.string(), z.object({ callerId: z.string(), jobs: z.array(zJob) }))
  .parse(jobsFixture);

// The fixture bakes the privacy rule per allocation: the researcher project's
// allocation carries by_member: null, the PI project's carries the ranked list.
// Unknown ids 404 like the real membership-scoped endpoint.
export const analyticsHandlers = [
  http.get("*/api/v1/connectors/analytics/contexts", () => HttpResponse.json(contexts)),
  http.get("*/api/v1/connectors/analytics/allocations/:id/usage-summary", ({ params: { id } }) => {
    const summary = summaryById.get(String(id));
    if (!summary) {
      return HttpResponse.json({ error: "not found" }, { status: 404 });
    }
    return HttpResponse.json(summary);
  }),
  http.get("*/api/v1/connectors/analytics/allocations/:id/jobs", ({ params: { id }, request }) => {
    const entry = jobsById[String(id)];
    if (!entry) {
      return HttpResponse.json({ error: "not found" }, { status: 404 });
    }
    const url = new URL(request.url);
    const mine = url.searchParams.get("mine") === "true";
    const limit = Number(url.searchParams.get("limit") ?? 20);
    const offset = Number(url.searchParams.get("offset") ?? 0);
    const rows = mine ? entry.jobs.filter((j) => j.user_id === entry.callerId) : entry.jobs;
    return HttpResponse.json({ jobs: rows.slice(offset, offset + limit), total: rows.length });
  }),
];
