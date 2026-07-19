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
import { zRole, zUser, zUserIdentity } from "@/generated/core/zod.gen";
import {
  privilegeKeySchema,
  userPrivilegeSchema as baseUserPrivilegeSchema,
  userRoleSchema as baseUserRoleSchema,
} from "@/features/core/identity/schemas";

export { privilegeKeySchema };

export const userSchema = zUser.required({ id: true, email: true });
export const roleSchema = zRole.required({ id: true, name: true });
export const userIdentitySchema = zUserIdentity.required({
  id: true,
  source: true,
  user_id: true,
});
export const userRoleSchema = baseUserRoleSchema.required({
  role_id: true,
  user_id: true,
});
export const userPrivilegeSchema = baseUserPrivilegeSchema
  .extend({ privilege: privilegeKeySchema })
  .required({ user_id: true });

export const userListResponseSchema = z.object({
  items: z.array(userSchema),
  total: z.number().int().nonnegative(),
});
export const rolesResponseSchema = z
  .array(roleSchema)
  .nullish()
  .transform((value) => value ?? []);
export const userRolesResponseSchema = z
  .array(userRoleSchema)
  .nullish()
  .transform((value) => value ?? []);
export const userIdentitiesResponseSchema = z
  .array(userIdentitySchema)
  .nullish()
  .transform((value) => value ?? []);
export const userPrivilegesResponseSchema = z
  .array(userPrivilegeSchema)
  .nullish()
  .transform((value) => value ?? []);
export const roleDetailResponseSchema = z.object({
  role: roleSchema.optional(),
  privileges: z
    .array(privilegeKeySchema)
    .nullish()
    .transform((value) => value ?? []),
});

export const grantRoleResponseSchema = userRoleSchema;

export type User = z.infer<typeof userSchema>;
export type Role = z.infer<typeof roleSchema>;
export type UserIdentity = z.infer<typeof userIdentitySchema>;
export type UserRole = z.infer<typeof userRoleSchema>;
export type UserPrivilege = z.infer<typeof userPrivilegeSchema>;
export type UserListResponse = z.infer<typeof userListResponseSchema>;
export type RoleDetail = z.infer<typeof roleDetailResponseSchema>;
