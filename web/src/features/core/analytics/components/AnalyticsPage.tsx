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
import { ChartColumn } from "lucide-react";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Skeleton } from "@/shared/ui/skeleton";
import { buildResourceSeries } from "../lib";
import type { AnalyticsAllocation, AnalyticsContext, UsageSummary } from "../schemas";
import { useAnalyticsContexts, useUsageSummary } from "../queries";
import { useAnalyticsSelection } from "../useAnalyticsSelection";
import { ContextSwitcher } from "./ContextSwitcher";
import { HERO_TILE_COUNT, HeroTiles } from "./HeroTiles";
import { JobsTable } from "./JobsTable";
import { MemberBreakdown } from "./MemberBreakdown";
import { ResourceBreakdown } from "./ResourceBreakdown";
import { UsageOverTimeBars } from "./UsageOverTimeBars";

const MANAGER_ROLES = new Set(["PI", "CO_PI", "ALLOCATION_MANAGER"]);

const ROLE_LABEL: Record<AnalyticsContext["role"], string> = {
  PI: "PI on this project",
  CO_PI: "Co-PI on this project",
  ALLOCATION_MANAGER: "Allocation manager",
  MEMBER: "Member",
};

export function AnalyticsPage() {
  const contextsQuery = useAnalyticsContexts();
  const contexts = React.useMemo(() => contextsQuery.data ?? [], [contextsQuery.data]);

  const { project, allocation, select } = useAnalyticsSelection(contexts);
  const summaryQuery = useUsageSummary(allocation?.id);

  const isManager = project ? MANAGER_ROLES.has(project.role) : false;

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="space-y-1">
          <h1 className="font-display text-[28px] font-bold leading-tight">Analytics</h1>
          <p className="text-sm text-muted-foreground">
            How your allocation is being used and how long it will last.
          </p>
        </div>
        {project && isManager ? <RoleChip role={project.role} /> : null}
      </header>

      {contextsQuery.isLoading ? (
        <div className="space-y-4">
          <Skeleton className="h-10 w-96" />
          <TileSkeletons />
        </div>
      ) : contextsQuery.error ? (
        <ErrorState
          message="We couldn't load your analytics."
          onRetry={() => contextsQuery.refetch()}
        />
      ) : contexts.length === 0 || !project || !allocation ? (
        <EmptyState
          icon={<ChartColumn className="h-8 w-8" />}
          heading="No allocations to report on yet"
          description="Analytics appear once you're a member of a project with a compute allocation."
        />
      ) : (
        <>
          <ContextSwitcher
            contexts={contexts}
            selectedProjectId={project.project_id}
            selectedAllocationId={allocation.id}
            onSelect={select}
          />
          <AnalyticsBody
            project={project}
            allocation={allocation}
            isManager={isManager}
            summary={summaryQuery.data}
            summaryLoading={summaryQuery.isLoading}
            summaryError={(summaryQuery.error as Error | null) ?? null}
            onRetrySummary={() => summaryQuery.refetch()}
          />
        </>
      )}
    </div>
  );
}

function AnalyticsBody({
  project,
  allocation,
  isManager,
  summary,
  summaryLoading,
  summaryError,
  onRetrySummary,
}: {
  project: AnalyticsContext;
  allocation: AnalyticsAllocation;
  isManager: boolean;
  summary: UsageSummary | undefined;
  summaryLoading: boolean;
  summaryError: Error | null;
  onRetrySummary: () => void;
}) {
  if (summaryLoading) {
    return <TileSkeletons />;
  }
  if (summaryError) {
    return (
      <ErrorState message="We couldn't load this allocation's usage." onRetry={onRetrySummary} />
    );
  }
  if (!summary) return null;

  const now = new Date();
  const callerUsed = summary.by_resource.reduce((a, r) => a + r.used_by_caller, 0);
  const seriesColorById = Object.fromEntries(
    buildResourceSeries(summary.by_resource).map((s) => [s.id, s.color]),
  );

  return (
    <div className="space-y-6">
      <HeroTiles allocation={allocation} callerUsed={callerUsed} now={now} />
      <UsageOverTimeBars summary={summary} />
      <ResourceBreakdown summary={summary} />
      {isManager && summary.by_member ? <MemberBreakdown members={summary.by_member} /> : null}
      <JobsTable
        allocationId={allocation.id}
        canManage={isManager}
        seriesColorById={seriesColorById}
      />
    </div>
  );
}

function TileSkeletons() {
  const keys = Array.from({ length: HERO_TILE_COUNT }, (_, i) => `hero-skeleton-${i}`);
  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {keys.map((k) => (
        <Skeleton key={k} className="h-28 w-full" />
      ))}
    </div>
  );
}

// Only rendered for managers; the chip explains why the extra cards appear.
function RoleChip({ role }: { role: AnalyticsContext["role"] }) {
  return (
    <span className="inline-flex items-center rounded-full bg-[color:var(--tone-info-bg)] px-3 py-1 text-xs font-medium text-[color:var(--tone-info-fg)]">
      {ROLE_LABEL[role]}
    </span>
  );
}
