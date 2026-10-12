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

import { isGone, useOnGone } from "@/shared/api/useOnGone";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { useShallowSearchParams } from "@/shared/hooks/useShallowSearchParams";
import { useBreadcrumbLabel } from "@/shared/layout/BreadcrumbLabelsProvider";
import { ROUTE_LABELS } from "@/shared/layout/nav";
import { QueryErrorState } from "@/shared/ui/ErrorState";
import { DetailLoading } from "@/shared/ui/Loading";
import { TabsRouter } from "@/shared/ui/TabsRouter";
import { useAllocation, useAllocationMembers, useRefreshAllocationLists } from "../queries";
import { AllocationChangeRequestsTab } from "./AllocationChangeRequestsTab";
import { AllocationDetailHeader } from "./AllocationDetailHeader";
import { AllocationHistoryTab } from "./AllocationHistoryTab";
import { AllocationMembersTab } from "./AllocationMembersTab";
import { AllocationOverviewTab } from "./AllocationOverviewTab";
import { AllocationUsageTab } from "./AllocationUsageTab";
import { useDrawerParams } from "./useDrawerParam";

export type AllocationDetailProps = {
  allocationId: string;
};

export function AllocationDetail({ allocationId }: AllocationDetailProps) {
  const ability = useAbility();
  const allocationQuery = useAllocation(allocationId);
  useOnGone(isGone(allocationQuery.error), useRefreshAllocationLists());
  // Sub-reads wait for the allocation, so a hidden one does not fan out into more 404s.
  const membersQuery = useAllocationMembers(allocationQuery.data ? allocationId : undefined);
  useBreadcrumbLabel(allocationId, allocationQuery.data?.name ?? "Allocation");
  // Change requests and diffs require the allocations read privilege; the page itself is scoped.
  const canRead = ability.can("read", "Allocation");
  // Every list drawer opens for readers only.
  const tab = useShallowSearchParams().get("tab");
  useDrawerParams(canRead ? tab : null);

  if (allocationQuery.isLoading) return <DetailLoading title={ROUTE_LABELS["/allocations"]} />;
  if (allocationQuery.error) {
    return (
      <QueryErrorState
        error={allocationQuery.error}
        what="allocation"
        onRetry={() => allocationQuery.refetch()}
        as="h1"
        page
      />
    );
  }
  const allocation = allocationQuery.data;
  if (!allocation) return null;

  const canManage = ability.can("write", "Allocation");

  return (
    <section className="space-y-6">
      <AllocationDetailHeader allocation={allocation} memberCount={membersQuery.data?.length} />
      <TabsRouter
        defaultValue="overview"
        tabs={[
          {
            value: "overview",
            label: "Overview",
            content: (
              <AllocationOverviewTab
                allocation={allocation}
                canManage={canManage}
                canRead={canRead}
              />
            ),
          },
          {
            value: "members",
            label: "Members",
            content: (
              <AllocationMembersTab
                allocation={allocation}
                canManage={canManage}
                canRead={canRead}
              />
            ),
          },
          {
            value: "change-requests",
            label: "Change requests",
            content: <AllocationChangeRequestsTab allocation={allocation} canSubmit={canManage} />,
          },
          {
            value: "usage",
            label: "Usage",
            content: (
              <AllocationUsageTab allocation={allocation} canRead={canRead} canManage={canManage} />
            ),
          },
          {
            value: "history",
            label: "History",
            content: <AllocationHistoryTab allocation={allocation} canManage={canManage} />,
          },
        ].filter((tab) => canRead || (tab.value !== "change-requests" && tab.value !== "history"))}
      />
    </section>
  );
}
