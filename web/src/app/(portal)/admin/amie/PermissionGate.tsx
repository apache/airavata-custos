"use client";

import type * as React from "react";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { ErrorState } from "@/shared/ui/ErrorState";

export function AmiePermissionGate({ children }: { children: React.ReactNode }) {
  const ability = useAbility();
  if (!ability.can("read", "AMIE") && !ability.can("manage", "AMIE")) {
    return <ErrorState message="Not permitted." />;
  }
  return <>{children}</>;
}
