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
  type Role,
  type UserRole,
  zCallerRoleGrant,
  zGetMeResponse,
  zGetRolesByIdResponse,
  zGetUserPrivilegesResponse,
  zGetUsersByIdPrivilegesResponse,
  zGetUsersByIdRolesResponse,
  zGetUsersByIdUserIdentitiesResponse,
  zUser,
} from "@/generated/core/zod.gen";
import { z } from "zod";

// Any non-empty string. Connector registries extend the catalog at runtime,
// so the generated core-only enum can't be a tight bound.
export const privilegeKeySchema = z.string().min(1);
const privilegeKeysSchema = z
  .array(privilegeKeySchema)
  .nullable()
  .transform((value) => value ?? []);

export const privilegesResponseSchema = zGetUserPrivilegesResponse
  .extend({ privileges: z.array(privilegeKeySchema) })
  .transform((value) => value.privileges);

export const callerRoleGrantSchema = zCallerRoleGrant.extend({ privileges: privilegeKeysSchema });

// GET /me: profile plus effective privilege keys and held roles.
export const meResponseSchema = zGetMeResponse.extend({
  user: zUser,
  privileges: privilegeKeysSchema,
  roles: z
    .array(callerRoleGrantSchema)
    .nullable()
    .transform((value) => value ?? []),
});

// Backend may send null instead of [] for empty lists; tolerate it.
export const userIdentitiesResponseSchema = zGetUsersByIdUserIdentitiesResponse
  .nullable()
  .transform((value) => value ?? []);

export const userRolesResponseSchema = zGetUsersByIdRolesResponse
  .nullable()
  .transform((value) => value ?? []);

export const userPrivilegesResponseSchema = zGetUsersByIdPrivilegesResponse
  .nullable()
  .transform((value) => value ?? []);

// GET /roles/{id} — a role and the privilege keys it carries.
export const roleDetailResponseSchema = zGetRolesByIdResponse.extend({
  privileges: privilegeKeysSchema,
});

export type PrivilegeKey = z.infer<typeof privilegeKeySchema>;
// The caller's grant for a role, joined with the role and its privilege keys.
export type RoleWithPrivileges = { role: Role; privileges: PrivilegeKey[]; grant: UserRole };
export type CallerRoleGrant = z.infer<typeof callerRoleGrantSchema>;
