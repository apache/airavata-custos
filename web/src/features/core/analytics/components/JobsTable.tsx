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
import { cn } from "@/lib/utils";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Skeleton } from "@/shared/ui/skeleton";
import { useAllocationJobs } from "../queries";
import { CHART_OTHER_COLOR, formatCredits, formatDateTime, formatNative } from "../lib";
import type { AnalyticsJob } from "../schemas";

const PAGE_SIZE = 10;

type JobsTableProps = {
  allocationId: string;
  // Managers can toggle Everyone/Mine and see the "who ran it" column; members
  // are always scoped to their own jobs by the backend.
  canManage: boolean;
  seriesColorById: Record<string, string>;
};

export function JobsTable({ allocationId, canManage, seriesColorById }: JobsTableProps) {
  const [mine, setMine] = React.useState(false);
  const effectiveMine = canManage ? mine : true;
  const showUser = canManage && !effectiveMine;

  const jobsQuery = useAllocationJobs(allocationId, { mine: effectiveMine, limit: PAGE_SIZE });
  const data = jobsQuery.data;

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          <h2 className="text-sm font-semibold">Recent jobs</h2>
          {canManage ? (
            <div className="inline-flex rounded-md border border-border p-0.5 text-xs">
              <button
                type="button"
                onClick={() => setMine(false)}
                className={cn(
                  "rounded px-2.5 py-1 font-medium",
                  !mine ? "bg-accent text-accent-foreground" : "text-muted-foreground",
                )}
              >
                Everyone
              </button>
              <button
                type="button"
                onClick={() => setMine(true)}
                className={cn(
                  "rounded px-2.5 py-1 font-medium",
                  mine ? "bg-accent text-accent-foreground" : "text-muted-foreground",
                )}
              >
                Mine
              </button>
            </div>
          ) : null}
        </div>
        <span className="text-xs text-muted-foreground">
          {showUser ? "all jobs charged to this allocation" : "your jobs"} · last 30 days
        </span>
      </div>

      {jobsQuery.isLoading ? (
        <Skeleton className="h-40 w-full" />
      ) : jobsQuery.error ? (
        <ErrorState message="We couldn't load the jobs." onRetry={() => jobsQuery.refetch()} />
      ) : !data || data.jobs.length === 0 ? (
        <p className="py-6 text-sm text-muted-foreground">
          No jobs yet. Jobs appear here as soon as work is charged to this allocation.
        </p>
      ) : (
        <>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left text-[11px] uppercase tracking-wide text-muted-foreground">
                  <th className="py-2 pr-3 font-semibold">When</th>
                  <th className="py-2 pr-3 font-semibold">Job</th>
                  {showUser ? <th className="py-2 pr-3 font-semibold">User</th> : null}
                  <th className="py-2 pr-3 font-semibold">Resource</th>
                  <th className="py-2 pr-3 font-semibold">Amount</th>
                  <th className="py-2 text-right font-semibold">Credits</th>
                </tr>
              </thead>
              <tbody>
                {data.jobs.map((job) => (
                  <JobRow
                    key={job.id}
                    job={job}
                    showUser={showUser}
                    color={seriesColorById[job.resource_id] ?? CHART_OTHER_COLOR}
                  />
                ))}
              </tbody>
            </table>
          </div>
          <div className="mt-3 flex flex-wrap justify-between gap-2 text-xs text-muted-foreground">
            <span>
              Showing {data.jobs.length} of {data.total} jobs
            </span>
            <span>Recent activity · older charges live in the accounting ledger</span>
          </div>
        </>
      )}
    </div>
  );
}

function JobRow({
  job,
  showUser,
  color,
}: {
  job: AnalyticsJob;
  showUser: boolean;
  color: string;
}) {
  return (
    <tr className="border-b border-border last:border-b-0">
      <td className="py-2.5 pr-3 whitespace-nowrap">{formatDateTime(job.calculated_time)}</td>
      <td className="py-2.5 pr-3 font-mono text-xs text-muted-foreground">{job.job_id}</td>
      {showUser ? <td className="py-2.5 pr-3">{job.user_name || job.user_id}</td> : null}
      <td className="py-2.5 pr-3">
        <span className="inline-flex items-center gap-1.5 whitespace-nowrap">
          <span
            className="h-2 w-2 shrink-0 rounded-sm"
            style={{ background: color }}
            aria-hidden="true"
          />
          {job.resource_name || job.resource_type}
        </span>
      </td>
      <td className="py-2.5 pr-3 whitespace-nowrap">
        {formatNative(job.used_raw)} {job.native_unit}
      </td>
      <td className="py-2.5 text-right tabular-nums">{formatCredits(job.used)}</td>
    </tr>
  );
}
