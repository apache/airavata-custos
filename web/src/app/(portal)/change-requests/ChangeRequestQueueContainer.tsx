"use client";

import { useAbility } from "@/shared/casl/AbilityProvider";
import { ChangeRequestApproverQueue } from "@/features/core/allocations/components/ChangeRequestApproverQueue";
import { useCurrentUser } from "@/features/core/identity/queries";

export function ChangeRequestQueueContainer() {
  const ability = useAbility();
  const { user } = useCurrentUser();
  const canApprove = ability.can("manage", "Allocation");
  return (
    <ChangeRequestApproverQueue
      canApprove={canApprove}
      approverId={user?.id ?? "anonymous"}
    />
  );
}
