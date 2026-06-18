"use client";

import { useSession } from "next-auth/react";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { TabsRouter } from "@/shared/ui/TabsRouter";
import { useAllocation, useAllocationMembers } from "../queries";
import { AllocationChangeRequestsTab } from "./AllocationChangeRequestsTab";
import { AllocationDetailHeader } from "./AllocationDetailHeader";
import { AllocationMembersTab } from "./AllocationMembersTab";
import { AllocationOverviewTab } from "./AllocationOverviewTab";
import { AllocationUsageTab } from "./AllocationUsageTab";

export type AllocationDetailProps = {
  allocationId: string;
};

export function AllocationDetail({ allocationId }: AllocationDetailProps) {
  const ability = useAbility();
  const { data: session } = useSession();
  const allocationQuery = useAllocation(allocationId);
  const membersQuery = useAllocationMembers(allocationId);

  if (allocationQuery.isLoading) return <CardSkeleton />;
  if (allocationQuery.error) {
    return (
      <ErrorState
        message={(allocationQuery.error as Error).message}
        onRetry={() => allocationQuery.refetch()}
      />
    );
  }
  const allocation = allocationQuery.data;
  if (!allocation) return null;

  const canManage = ability.can("manage", "Allocation");
  const memberCount = membersQuery.data?.length ?? 0;

  return (
    <section className="space-y-6">
      <AllocationDetailHeader allocation={allocation} memberCount={memberCount} />
      <TabsRouter
        defaultValue="overview"
        tabs={[
          {
            value: "overview",
            label: "Overview",
            content: <AllocationOverviewTab allocation={allocation} />,
          },
          {
            value: "members",
            label: "Members",
            content: <AllocationMembersTab allocation={allocation} canManage={canManage} />,
          },
          {
            value: "change-requests",
            label: "Change requests",
            content: (
              <AllocationChangeRequestsTab
                allocation={allocation}
                canSubmit={canManage}
                requesterId={session?.user?.id ?? ""}
              />
            ),
          },
          {
            value: "usage",
            label: "Usage",
            content: <AllocationUsageTab allocation={allocation} />,
          },
        ]}
      />
    </section>
  );
}
