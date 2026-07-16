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
  privilegeKeySchema,
  roleDetailResponseSchema,
  roleSchema,
  userRoleSchema,
} from "@/features/core/identity/schemas";

export { privilegeKeySchema, roleDetailResponseSchema, roleSchema };

export const rolesResponseSchema = z
  .array(roleSchema)
  .nullish()
  .transform((value) => value ?? []);

export const roleHoldersResponseSchema = z
  .array(userRoleSchema)
  .nullish()
  .transform((value) => value ?? []);

export const privilegeCatalogResponseSchema = z
  .array(privilegeKeySchema)
  .nullish()
  .transform((value) => value ?? []);

export const roleInputSchema = z.object({
  name: z.string().trim().min(1),
  description: z.string().trim(),
  privileges: z.array(privilegeKeySchema),
});

export type PrivilegeKey = z.infer<typeof privilegeKeySchema>;
export type Role = z.infer<typeof roleSchema>;
export type RoleInput = z.infer<typeof roleInputSchema>;
export type UserRole = z.infer<typeof roleHoldersResponseSchema>[number];

export type RoleRow = Role & {
  privileges: PrivilegeKey[];
  holderIds: string[];
  memberCount: number;
};
