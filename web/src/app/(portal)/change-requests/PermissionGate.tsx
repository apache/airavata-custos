"use client";

import type { ReactNode } from "react";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { ErrorState } from "@/shared/ui/ErrorState";

export function ChangeRequestsPermissionGate({ children }: { children: ReactNode }) {
  const ability = useAbility();
  if (ability.cannot("read", "Allocation")) {
    return <ErrorState message="Not permitted." />;
  }
  return <>{children}</>;
}
