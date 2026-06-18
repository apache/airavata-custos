"use client";

import { useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";
import { apiFetch } from "@/shared/api/client";
import { privilegesResponseSchema } from "./schemas";
import type { CurrentUser, Privilege } from "./types";

export const identityKeys = {
  all: ["identity"] as const,
  current: () => [...identityKeys.all, "current"] as const,
  privileges: () => [...identityKeys.all, "privileges"] as const,
};

export function useCurrentUser() {
  const { data: session, status } = useSession();
  const user: CurrentUser | null = session?.user
    ? {
        id: session.user.id ?? session.user.email ?? "",
        email: session.user.email ?? "",
        name: session.user.name ?? session.user.email ?? "",
        privileges: session.privileges ?? [],
      }
    : null;
  return { user, status };
}

export function usePrivileges() {
  const { status } = useSession();
  return useQuery({
    queryKey: identityKeys.privileges(),
    queryFn: async (): Promise<Privilege[]> => {
      const raw = await apiFetch("/user/privileges");
      return privilegesResponseSchema.parse(raw);
    },
    enabled: status === "authenticated",
    staleTime: 30_000,
    gcTime: 300_000,
    refetchOnWindowFocus: false,
  });
}
