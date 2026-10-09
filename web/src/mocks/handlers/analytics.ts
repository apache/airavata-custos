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
import contextsFixture from "@/features/core/analytics/__fixtures__/contexts.json";
import jobsFixture from "@/features/core/analytics/__fixtures__/jobs.json";
import summariesFixture from "@/features/core/analytics/__fixtures__/usage-summary.json";
import type {
  GetConnectorsAnalyticsAllocationsByIdJobsData,
  GetConnectorsAnalyticsAllocationsByIdUsageSummaryData,
  ProjectContext,
} from "@/generated/analytics/types.gen";
import { zJob, zProjectContext, zUsageSummary } from "@/generated/analytics/zod.gen";
import { z } from "zod";
import { notFound, page } from "../paging";

const contexts: ProjectContext[] = z.array(zProjectContext).parse(contextsFixture);
const summaries = z.record(z.string(), zUsageSummary).parse(summariesFixture);
// Per allocation: the caller the "mine" filter scopes to, and every job.
const jobsByAllocation = z
  .record(z.string(), z.object({ callerId: z.string(), jobs: z.array(zJob) }))
  .parse(jobsFixture);

// The fixture bakes the privacy rule per allocation: the researcher project's
// allocation omits by_member, the PI project's carries the ranked list.
// Unknown ids 404 like the real membership-scoped endpoint.
export const analyticsHandlers = [
  http.get("*/api/v1/connectors/analytics/contexts", () => HttpResponse.json(contexts)),
  http.get<GetConnectorsAnalyticsAllocationsByIdUsageSummaryData["path"]>(
    "*/api/v1/connectors/analytics/allocations/:id/usage-summary",
    ({ params }) => {
      const summary = summaries[params.id];
      return summary ? HttpResponse.json(summary) : notFound("allocation");
    },
  ),
  http.get<GetConnectorsAnalyticsAllocationsByIdJobsData["path"]>(
    "*/api/v1/connectors/analytics/allocations/:id/jobs",
    ({ params, request }) => {
      const entry = jobsByAllocation[params.id];
      if (!entry) return notFound("allocation");
      const url = new URL(request.url);
      const mine = url.searchParams.get("mine") === "true";
      const rows = mine ? entry.jobs.filter((j) => j.user_id === entry.callerId) : entry.jobs;
      const { items, total } = page(url, rows, 20);
      return HttpResponse.json({ jobs: items, total });
    },
  ),
];
