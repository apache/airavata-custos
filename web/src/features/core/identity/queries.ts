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
