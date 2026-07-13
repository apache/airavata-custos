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

import { z } from "zod";
import {
  zRole,
  zUser,
  zUserIdentity,
  zUserPrivilege,
  zUserRole,
} from "@/generated/core/zod.gen";

// Any non-empty string. Connector registries extend the catalog at runtime,
// so the generated core-only enum can't be a tight bound.
export const privilegeKeySchema = z.string().min(1);
// granted_by and reason are nullable (*string) on the backend, so they arrive
// as JSON null; the generated schema only allows undefined, so widen them.
const nullableGrantMeta = {
  granted_by: z.string().nullish(),
  reason: z.string().nullish(),
};
export const userPrivilegeSchema = zUserPrivilege.extend(nullableGrantMeta);
export const userRoleSchema = zUserRole.extend(nullableGrantMeta);
export const roleSchema = zRole;
export const userSchema = zUser;
export const userIdentitySchema = zUserIdentity;

export const roleWithPrivilegesSchema = z.object({
  role: zRole,
  privileges: z.array(privilegeKeySchema),
  // The caller's grant for this role (granted_at, granted_by, reason).
  grant: userRoleSchema,
});

export const callerPrivilegesSchema = z.object({
  privileges: z.array(privilegeKeySchema).optional(),
});

export const privilegesResponseSchema = callerPrivilegesSchema.transform(
  (value) => value.privileges ?? [],
);

export const callerRoleGrantSchema = z.object({
  role: zRole,
  privileges: z
    .array(privilegeKeySchema)
    .nullish()
    .transform((value) => value ?? []),
  granted_at: z.string(),
});

// GET /me: profile plus effective privilege keys and held roles.
export const meResponseSchema = z.object({
  user: zUser,
  privileges: z.array(privilegeKeySchema).optional(),
  roles: z
    .array(callerRoleGrantSchema)
    .nullish()
    .transform((value) => value ?? []),
});

// Backend may send null instead of [] for empty lists; tolerate it.
export const userIdentitiesResponseSchema = z
  .array(zUserIdentity)
  .nullish()
  .transform((value) => value ?? []);

export const userRolesResponseSchema = z
  .array(userRoleSchema)
  .nullish()
  .transform((value) => value ?? []);

export const userPrivilegesResponseSchema = z
  .array(userPrivilegeSchema)
  .nullish()
  .transform((value) => value ?? []);

// GET /roles/{id} — a role and the privilege keys it carries.
export const roleDetailResponseSchema = z.object({
  role: zRole.optional(),
  privileges: z.array(privilegeKeySchema).nullish().transform((v) => v ?? []),
});

export const userNameUpdateSchema = z.object({
  first_name: z.string().optional(),
  middle_name: z.string().optional(),
  last_name: z.string().optional(),
});

export type PrivilegeKey = z.infer<typeof privilegeKeySchema>;
export type UserProfile = z.infer<typeof zUser>;
export type UserIdentity = z.infer<typeof zUserIdentity>;
export type UserRole = z.infer<typeof userRoleSchema>;
export type UserPrivilege = z.infer<typeof userPrivilegeSchema>;
export type RoleWithPrivileges = z.infer<typeof roleWithPrivilegesSchema>;
export type CallerRoleGrant = z.infer<typeof callerRoleGrantSchema>;
export type UserNameUpdate = z.infer<typeof userNameUpdateSchema>;
