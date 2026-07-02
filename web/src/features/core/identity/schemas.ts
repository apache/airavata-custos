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
  zUserPrivilege,
  zUserRole,
} from "@/generated/core/zod.gen";

// Any non-empty string. Connector registries extend the catalog at runtime,
// so the generated core-only enum can't be a tight bound.
export const privilegeKeySchema = z.string().min(1);
export const userPrivilegeSchema = zUserPrivilege;
export const userRoleSchema = zUserRole;
export const roleSchema = zRole;

export const roleWithPrivilegesSchema = z.object({
  role: zRole,
  privileges: z.array(privilegeKeySchema),
});

export const callerPrivilegesSchema = z.object({
  privileges: z.array(privilegeKeySchema).optional(),
});

export const privilegesResponseSchema = callerPrivilegesSchema.transform(
  (value) => value.privileges ?? [],
);

export type PrivilegeKey = z.infer<typeof privilegeKeySchema>;
