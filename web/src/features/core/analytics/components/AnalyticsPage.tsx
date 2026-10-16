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

import type { Allocation, UsageSummary } from "@/generated/analytics/types.gen";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { ROUTE_LABELS } from "@/shared/layout/nav";
import { EmptyState } from "@/shared/ui/EmptyState";
import { ErrorState } from "@/shared/ui/ErrorState";
import { Skeleton } from "@/shared/ui/skeleton";
import { ChartColumn } from "lucide-react";
import * as React from "react";
import { buildResourceSeries } from "../lib";
import { useAnalyticsContexts, useUsageSummary } from "../queries";
import { useAnalyticsSelection } from "../useAnalyticsSelection";
import { ContextSwitcher } from "./ContextSwitcher";
import { HERO_TILE_COUNT, HeroTiles } from "./HeroTiles";
import { JobsTable } from "./JobsTable";
import { MemberBreakdown } from "./MemberBreakdown";
import { ResourceBreakdown } from "./ResourceBreakdown";
import { UsageOverTimeBars } from "./UsageOverTimeBars";

const MANAGER_ROLES = new Set(["PI", "CO_PI", "ALLOCATION_MANAGER"]);

const ROLE_LABEL: Record<string, string> = {
  PI: "PI on this project",
  CO_PI: "Co-PI on this project",
  ALLOCATION_MANAGER: "Allocation manager on this project",
};

export function AnalyticsPage() {
  const contextsQuery = useAnalyticsContexts();
  const contexts = React.useMemo(() => contextsQuery.data ?? [], [contextsQuery.data]);

  const { project, allocation, select, unavailable } = useAnalyticsSelection(
    contexts,
    contextsQuery.isSuccess,
  );
  const summaryQuery = useUsageSummary(allocation?.id);

  const isManager = project ? MANAGER_ROLES.has(project.role) : false;
  // Allocation readers reach usage from the allocation itself; analytics covers memberships only.
  const readsAllocations = useAbility().can("read", "Allocation");

  return (
    <div className="space-y-6">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="space-y-1">
          <h1 className="font-display text-[28px] font-bold leading-tight">
            {ROUTE_LABELS["/analytics"]}
          </h1>
          <p className="text-sm text-muted-foreground">
            How your allocation is being used and how long it will last.
          </p>
        </div>
        {project && isManager ? <RoleChip role={project.role} /> : null}
      </header>

      {unavailable ? (
        <output className="block text-sm text-muted-foreground">
          That project or allocation is not available to you.
        </output>
      ) : null}

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
      ) : !contexts.some((c) => c.allocations?.length) ? (
        <EmptyState
          icon={<ChartColumn className="h-8 w-8" />}
          heading="No allocations to report on yet"
          description={
            readsAllocations
              ? "You are not a member of any allocation. Open an allocation to see its usage."
              : "Analytics appear once you're a member of a project with a compute allocation."
          }
        />
      ) : (
        <>
          <ContextSwitcher
            contexts={contexts}
            selectedAllocationId={allocation?.id}
            onSelect={select}
          />
          {allocation ? (
            <AnalyticsBody
              allocation={allocation}
              isManager={isManager}
              summary={summaryQuery.data}
              summaryLoading={summaryQuery.isLoading}
              summaryError={summaryQuery.error}
              onRetrySummary={() => summaryQuery.refetch()}
            />
          ) : (
            <EmptyState
              icon={<ChartColumn className="h-8 w-8" />}
              heading={`${project?.project_name ?? "This project"} has no allocations`}
              description="Pick an allocation from another project above."
            />
          )}
        </>
      )}
    </div>
  );
}

function AnalyticsBody({
  allocation,
  isManager,
  summary,
  summaryLoading,
  summaryError,
  onRetrySummary,
}: {
  allocation: Allocation;
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
  const byResource = summary.by_resource ?? [];
  const callerUsed = byResource.reduce((a, r) => a + r.used_by_caller, 0);
  const seriesColorById = Object.fromEntries(
    buildResourceSeries(byResource).map((s) => [s.id, s.color]),
  );

  return (
    <div className="space-y-6">
      <HeroTiles allocation={allocation} callerUsed={callerUsed} now={now} />
      <UsageOverTimeBars summary={summary} />
      {isManager && summary.by_member ? (
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <ResourceBreakdown summary={summary} />
          <MemberBreakdown members={summary.by_member} total={summary.total} />
        </div>
      ) : (
        <ResourceBreakdown summary={summary} />
      )}
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
function RoleChip({ role }: { role: string }) {
  return (
    <span className="inline-flex items-center rounded-full bg-[color:var(--tone-info-bg)] px-3 py-1 text-xs font-medium text-[color:var(--tone-info-fg)]">
      {ROLE_LABEL[role]}
    </span>
  );
}
