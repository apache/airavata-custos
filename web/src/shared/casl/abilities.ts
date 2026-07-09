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

import { type MongoAbility, createMongoAbility } from "@casl/ability";
import type { Privilege } from "@/features/core/identity/types";

export type AppAbility = MongoAbility;
export type Ability = AppAbility;

export type PrivilegeRule = { action: string; subject: string };

// TODO(privileges): backend should declare (subject, action) pairs per key
// so clients can derive this map from /me instead of shipping it.
export const PRIVILEGE_ABILITY_MAP: Record<string, PrivilegeRule[]> = {
  "core:clusters:read": [{ action: "read", subject: "Cluster" }],
  "core:clusters:write": [{ action: "manage", subject: "Cluster" }],
  "core:allocations:read": [{ action: "read", subject: "Allocation" }],
  "core:allocations:write": [{ action: "manage", subject: "Allocation" }],
  "core:projects:read": [{ action: "read", subject: "Project" }],
  "core:projects:write": [{ action: "manage", subject: "Project" }],
  "core:users:read": [{ action: "read", subject: "User" }],
  "core:users:write": [{ action: "manage", subject: "User" }],
  "core:organizations:read": [{ action: "read", subject: "Organization" }],
  "core:organizations:write": [{ action: "manage", subject: "Organization" }],
  "core:traces:read": [
    { action: "read", subject: "Trace" },
    { action: "read", subject: "AuditEvent" },
  ],
  "core:privileges:grant": [{ action: "manage", subject: "PrivilegeGrant" }],
  "core:roles:manage": [{ action: "manage", subject: "Role" }],
  "amie:packets:read": [{ action: "read", subject: "AMIE" }],
  "amie:packets:write": [{ action: "manage", subject: "AMIE" }],
  "amie:replies:read": [{ action: "read", subject: "AMIE" }],
  "amie:replies:write": [{ action: "manage", subject: "AMIE" }],
  "amie:unmapped:read": [{ action: "read", subject: "AMIE" }],
  "amie:unmapped:write": [{ action: "manage", subject: "AMIE" }],
};

export function defineAbilitiesFor(privileges: Privilege[]): AppAbility {
  const rules = privileges.flatMap((p) => PRIVILEGE_ABILITY_MAP[p] ?? []);
  return createMongoAbility(rules);
}
