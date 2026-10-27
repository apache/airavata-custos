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

import {
  zCallerRoleGrant,
  zGetMeResponse,
  zGetRolesByIdResponse,
  zGetRolesResponse,
  zGetUserPrivilegesResponse,
  zGetUsersByIdRolesResponse,
  zGetUsersByIdUserIdentitiesResponse,
  zRole,
  zUser,
  zUserPrivilege,
  zUserRole,
} from "@/generated/core/zod.gen";
import { z } from "zod";

// Any non-empty string. Connector registries extend the catalog at runtime,
// so the generated core-only enum can't be a tight bound.
export const privilegeKeySchema = z.string().min(1);
export const userPrivilegeSchema = zUserPrivilege.extend({ privilege: privilegeKeySchema });

export const roleWithPrivilegesSchema = z.object({
  role: zRole,
  privileges: z.array(privilegeKeySchema),
  // The caller's grant for this role (granted_at, granted_by, reason).
  grant: zUserRole,
});

export const privilegesResponseSchema = zGetUserPrivilegesResponse
  .extend({ privileges: z.array(privilegeKeySchema) })
  .transform((value) => value.privileges);

export const callerRoleGrantSchema = zCallerRoleGrant.extend({
  privileges: z.array(privilegeKeySchema),
});

// GET /me: profile plus effective privilege keys and held roles.
export const meResponseSchema = zGetMeResponse.extend({
  user: zUser,
  privileges: z.array(privilegeKeySchema),
  roles: z.array(callerRoleGrantSchema),
});

// Backend may send null instead of [] for empty lists; tolerate it.
export const userIdentitiesResponseSchema = zGetUsersByIdUserIdentitiesResponse.transform(
  (value) => value ?? [],
);

export const userRolesResponseSchema = zGetUsersByIdRolesResponse.transform((value) => value ?? []);

export const userPrivilegesResponseSchema = z
  .array(userPrivilegeSchema)
  .nullable()
  .transform((value) => value ?? []);

// GET /roles/{id} — a role and the privilege keys it carries.
export const roleDetailResponseSchema = zGetRolesByIdResponse.extend({
  privileges: z
    .array(privilegeKeySchema)
    .nullable()
    .transform((value) => value ?? []),
});

export const rolesResponseSchema = zGetRolesResponse.transform((value) => value ?? []);

export type PrivilegeKey = z.infer<typeof privilegeKeySchema>;
export type UserPrivilege = z.infer<typeof userPrivilegeSchema>;
export type RoleWithPrivileges = z.infer<typeof roleWithPrivilegesSchema>;
export type CallerRoleGrant = z.infer<typeof callerRoleGrantSchema>;
