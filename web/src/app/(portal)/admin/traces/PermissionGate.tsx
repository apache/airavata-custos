"use client";

import type * as React from "react";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { ErrorState } from "@/shared/ui/ErrorState";

export function TracePermissionGate({ children }: { children: React.ReactNode }) {
  const ability = useAbility();
  if (ability.cannot("read", "Trace")) {
    return <ErrorState message="Not permitted." />;
  }
  return <>{children}</>;
}
