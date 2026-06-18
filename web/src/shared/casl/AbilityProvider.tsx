"use client";

import { AbilityProvider as CaslAbilityProvider, useAbility as useCaslAbility } from "@casl/react";
import { useSession } from "next-auth/react";
import { type ReactNode, useMemo } from "react";
import { usePrivileges } from "@/features/core/identity/queries";
import { type AppAbility, defineAbilitiesFor } from "./abilities";

export function AbilityProvider({ children }: { children: ReactNode }) {
  const { data: session } = useSession();
  const { data: privileges } = usePrivileges();
  // /user/privileges wins once it arrives; until then drive abilities from the
  // session-embedded list so the sidebar doesn't flicker empty on first paint.
  const effective = privileges ?? session?.privileges ?? [];
  const ability = useMemo(() => defineAbilitiesFor(effective), [effective]);
  return <CaslAbilityProvider value={ability}>{children}</CaslAbilityProvider>;
}

export function useAbility(): AppAbility {
  return useCaslAbility() as AppAbility;
}
