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

import { apiFetch } from "@/shared/api/client";
import {
  meResponseSchema,
  privilegesResponseSchema,
  roleDetailResponseSchema,
  userIdentitiesResponseSchema,
  userPrivilegesResponseSchema,
  userRolesResponseSchema,
  userSchema,
} from "./schemas";
import type {
  RoleWithPrivileges,
  UserIdentity,
  UserNameUpdate,
  UserPrivilege,
  UserProfile,
} from "./schemas";
import type { Privilege } from "./types";

export async function getPrivileges(): Promise<Privilege[]> {
  const raw = await apiFetch("/user/privileges");
  return privilegesResponseSchema.parse(raw);
}

export type Me = { user: UserProfile; privileges: Privilege[] };

export async function getMe(): Promise<Me> {
  const parsed = meResponseSchema.parse(await apiFetch("/me"));
  return { user: parsed.user, privileges: parsed.privileges ?? [] };
}

export async function getMyIdentities(userId: string): Promise<UserIdentity[]> {
  return userIdentitiesResponseSchema.parse(
    await apiFetch(`/users/${userId}/user-identities`),
  );
}

// No user-scoped roles-with-privileges endpoint exists; compose it from the
// caller's role grants and each role's privilege detail.
export async function getMyRolesWithPrivileges(
  userId: string,
): Promise<RoleWithPrivileges[]> {
  const grants = userRolesResponseSchema.parse(
    await apiFetch(`/users/${userId}/roles`),
  );
  const details = await Promise.all(
    grants.map(async (grant) => {
      const detail = roleDetailResponseSchema.parse(
        await apiFetch(`/roles/${grant.role_id}`),
      );
      if (!detail.role) return null;
      return { role: detail.role, privileges: detail.privileges, grant };
    }),
  );
  return details.filter((d): d is RoleWithPrivileges => d !== null);
}

export async function getMyDirectPrivileges(
  userId: string,
): Promise<UserPrivilege[]> {
  return userPrivilegesResponseSchema.parse(
    await apiFetch(`/users/${userId}/privileges`),
  );
}

export async function updateMyName(
  userId: string,
  name: UserNameUpdate,
): Promise<UserProfile> {
  return userSchema.parse(
    await apiFetch(`/users/${userId}`, { method: "PUT", body: name }),
  );
}
