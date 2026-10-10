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

import {
  getMe,
  getRolesById,
  getUserPrivileges,
  getUsersByIdPrivileges,
  getUsersByIdRoles,
  getUsersByIdUserIdentities,
  putUsersById,
} from "@/generated/core/sdk.gen";
import type {
  CallerRoleGrant,
  PrivilegeKey,
  UserPrivilege,
  UserRole,
} from "@/generated/core/types.gen";
import { useInvalidating } from "@/shared/api/useInvalidating";
import { skipToken, useQuery } from "@tanstack/react-query";
import { useSession } from "next-auth/react";

export const identityKeys = {
  all: ["identity"] as const,
  privileges: () => [...identityKeys.all, "privileges"] as const,
  me: () => [...identityKeys.all, "me"] as const,
  identities: (userId: string) => [...identityKeys.all, "identities", userId] as const,
  access: (userId: string) => [...identityKeys.all, "access", userId] as const,
};

export function useCurrentUser() {
  const user = useSession().data?.user;
  return { user: user ? { id: user.id ?? user.email ?? "" } : null };
}

export function usePrivileges() {
  const { status } = useSession();
  return useQuery({
    queryKey: identityKeys.privileges(),
    queryFn: async () => (await getUserPrivileges()).privileges ?? [],
    enabled: status === "authenticated",
  });
}

export function useMe() {
  const { status } = useSession();
  return useQuery({
    queryKey: identityKeys.me(),
    queryFn: () => getMe(),
    enabled: status === "authenticated",
  });
}

export function useMyIdentities(userId: string | undefined) {
  return useQuery({
    queryKey: identityKeys.identities(userId ?? ""),
    queryFn: userId ? () => getUsersByIdUserIdentities({ path: { id: userId } }) : skipToken,
  });
}

// A held role joined with its grant attribution, which /me omits.
export type HeldRole = CallerRoleGrant & Pick<UserRole, "granted_by">;

export type MyAccess = {
  roles: HeldRole[];
  direct: UserPrivilege[];
  privileges: PrivilegeKey[];
  provenance: boolean;
};

// Direct-grant reads are admin-gated; without them roles come from /me and
// unattributed keys stay unlabeled instead of claiming "Direct grant".
export function useMyAccess(
  userId: string | undefined,
  effective: PrivilegeKey[],
  heldRoles: CallerRoleGrant[],
) {
  const provenance =
    effective.includes("core:roles:manage") && effective.includes("core:privileges:grant");
  const heldKey = heldRoles.map((g) => g.role.id).join(",");
  return useQuery({
    queryKey: [...identityKeys.access(userId ?? ""), provenance, heldKey],
    queryFn: userId
      ? async (): Promise<MyAccess> => {
          if (!provenance) {
            return { roles: heldRoles, direct: [], privileges: effective, provenance: false };
          }
          const path = { id: userId };
          const [grants, direct] = await Promise.all([
            getUsersByIdRoles({ path }),
            getUsersByIdPrivileges({ path }),
          ]);
          // No user-scoped roles-with-privileges endpoint exists; join each grant
          // with its role's privilege detail.
          const roles = await Promise.all(
            grants.map(async ({ role_id, granted_at, granted_by }) => ({
              ...(await getRolesById({ path: { id: role_id } })),
              granted_at,
              granted_by,
            })),
          );
          return { roles, direct, privileges: effective, provenance };
        }
      : skipToken,
  });
}

export function useUpdateMyName() {
  return useInvalidating(putUsersById<true>, identityKeys.me());
}
