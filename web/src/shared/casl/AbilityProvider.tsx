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

import { usePrivileges } from "@/features/core/identity/queries";
import { routeLabel } from "@/shared/layout/nav";
import { DetailLoading } from "@/shared/ui/Loading";
import { AbilityProvider as CaslAbilityProvider, useAbility as useCaslAbility } from "@casl/react";
import { useSession } from "next-auth/react";
import { usePathname } from "next/navigation";
import { type ReactNode, useMemo } from "react";
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
  return useCaslAbility();
}

// Page content, whose reads are gated on abilities, waits for /user/privileges: the session list
// can be stale, and a read it still allows would 403.
export function PrivilegesLoaded({ children }: { children: ReactNode }) {
  const path = usePathname();
  if (!usePrivileges().isPending) return children;
  const parts = path.split("/");
  // The section's label is the h1 every page under it shows (AMIE, Users & permissions).
  // "/" redirects to Allocations, whose heading it shows meanwhile.
  const title =
    path === "/"
      ? routeLabel("/allocations")
      : parts.map((_, i) => routeLabel(parts.slice(0, i + 1).join("/"))).find(Boolean);
  return <DetailLoading title={title} />;
}
