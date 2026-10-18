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

import { useSession } from "next-auth/react";
import { requires, usePermissions } from "@/shared/casl/permissions";
import { useBreadcrumbLabel } from "@/shared/layout/BreadcrumbLabelsProvider";
import { ErrorState } from "@/shared/ui/ErrorState";
import { DeleteButton } from "@/shared/ui/ConfirmDialog";
import { CardSkeleton } from "@/shared/ui/Loading";
import { TabsRouter } from "@/shared/ui/TabsRouter";
import { useAllocation, useAllocationMembers, useDeleteAllocation } from "../queries";
import { AllocationChangeRequestsTab } from "./AllocationChangeRequestsTab";
import { AllocationDetailHeader } from "./AllocationDetailHeader";
import { AllocationHistoryTab } from "./AllocationHistoryTab";
import { AllocationMembersTab } from "./AllocationMembersTab";
import { AllocationOverviewTab } from "./AllocationOverviewTab";
import { AllocationUsageTab } from "./AllocationUsageTab";

export type AllocationDetailProps = {
  allocationId: string;
};

export function AllocationDetail({ allocationId }: AllocationDetailProps) {
  const can = usePermissions();
  const { data: session } = useSession();
  const allocationQuery = useAllocation(allocationId);
  const membersQuery = useAllocationMembers(allocationId);
  const remove = useDeleteAllocation();
  useBreadcrumbLabel(allocationId, allocationQuery.data?.name);

  if (allocationQuery.isLoading) return <CardSkeleton />;
  if (allocationQuery.error) {
    return (
      <ErrorState
        message={allocationQuery.error.message}
        onRetry={() => allocationQuery.refetch()}
      />
    );
  }
  const allocation = allocationQuery.data;
  if (!allocation) return null;

  const memberCount = membersQuery.data?.length ?? 0;

  return (
    <section className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <AllocationDetailHeader allocation={allocation} memberCount={memberCount} />
        <DeleteButton
          noun="Allocation"
          id={allocation.id}
          name={allocation.name}
          mutation={remove}
          redirectTo="/allocations"
          allowed={can.allocationsWrite}
          reason={requires("allocationsWrite")}
        >
          Its members, resources and usage records are deleted with it. This cannot be undone.
        </DeleteButton>
      </div>
      <TabsRouter
        defaultValue="overview"
        tabs={[
          {
            value: "overview",
            label: "Overview",
            content: (
              <div className="space-y-6">
                <AllocationOverviewTab allocation={allocation} />
                <AllocationMembersTab
                  allocation={allocation}
                  canEditAccess={can.allocationsWrite}
                  canEditRoles={can.projectsWrite}
                  canSearchUsers={can.usersRead}
                />
              </div>
            ),
          },
          {
            value: "change-requests",
            label: "Change requests",
            content: (
              <AllocationChangeRequestsTab
                allocation={allocation}
                canSubmit={can.allocationsWrite}
                requesterId={session?.user?.id ?? ""}
              />
            ),
          },
          {
            value: "usage",
            label: "Usage",
            content: <AllocationUsageTab allocation={allocation} />,
          },
          {
            value: "history",
            label: "History",
            content: <AllocationHistoryTab allocation={allocation} />,
          },
        ]}
      />
    </section>
  );
}
