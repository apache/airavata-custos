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
  privilegeKeySchema,
  roleDetailResponseSchema,
  userIdentitiesResponseSchema,
  userRolesResponseSchema,
} from "@/features/core/identity/schemas";
import {
  zGetRolesResponse,
  zGetUsersResponse,
  zPostUsersBody,
  zUserPrivilege,
} from "@/generated/core/zod.gen";
import { z } from "zod";

export {
  privilegeKeySchema,
  roleDetailResponseSchema,
  userIdentitiesResponseSchema,
  userRolesResponseSchema,
};

export const userPrivilegeSchema = zUserPrivilege.extend({ privilege: privilegeKeySchema });

export const userListResponseSchema = zGetUsersResponse.transform((value) => ({
  ...value,
  items: value.items ?? [],
}));
export const rolesResponseSchema = zGetRolesResponse.nullable().transform((value) => value ?? []);
export const userPrivilegesResponseSchema = z
  .array(userPrivilegeSchema)
  .nullable()
  .transform((value) => value ?? []);

export const createUserPayloadSchema = zPostUsersBody.required({ email: true });

export type UserPrivilege = z.infer<typeof userPrivilegeSchema>;
export type UserListResponse = z.infer<typeof userListResponseSchema>;
export type RoleDetail = z.infer<typeof roleDetailResponseSchema>;
export type CreateUserPayload = z.infer<typeof createUserPayloadSchema>;
