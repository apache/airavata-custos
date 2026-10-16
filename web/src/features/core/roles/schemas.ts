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
  type PrivilegeKey,
  privilegeKeySchema,
  roleDetailResponseSchema,
} from "@/features/core/identity/schemas";
import {
  type Role,
  zCreateRoleRequest,
  zGetRolesByIdHoldersResponse,
  zGetRolesResponse,
} from "@/generated/core/zod.gen";
import { z } from "zod";

export { privilegeKeySchema, roleDetailResponseSchema };

export const rolesResponseSchema = zGetRolesResponse.nullable().transform((value) => value ?? []);

export const roleHoldersResponseSchema = zGetRolesByIdHoldersResponse
  .nullable()
  .transform((value) => value ?? []);

export const privilegeCatalogResponseSchema = z
  .array(privilegeKeySchema)
  .nullable()
  .transform((value) => value ?? []);

// The create/update body plus the privilege set, which is synced by separate calls.
export const roleInputSchema = zCreateRoleRequest.extend({
  name: z.string().trim().min(1),
  description: z.string().trim(),
  privileges: z.array(privilegeKeySchema),
});

export type RoleInput = z.infer<typeof roleInputSchema>;

export type RoleRow = Role & {
  privileges: PrivilegeKey[];
  holderIds: string[];
  memberCount: number;
};
