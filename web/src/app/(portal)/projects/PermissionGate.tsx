"use client";

import type { ReactNode } from "react";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { ErrorState } from "@/shared/ui/ErrorState";

export function ProjectsPermissionGate({ children }: { children: ReactNode }) {
  const ability = useAbility();
  if (ability.cannot("read", "Project")) {
    return <ErrorState message="Not permitted." />;
  }
  return <>{children}</>;
}
