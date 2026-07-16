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

import type { RoleRow } from "./types";

// Seed roles for the mock portal. Permissions are granted per left-nav
// section (see permissions.ts) — source once a real "list roles" call backs
// this page: GET /roles + GET /roles/{id}.
export const ROLE_SUPER_ADMIN: RoleRow = {
  id: "role-super-admin",
  name: "Super Admin",
  description: "Full administrative access across the portal.",
  is_system: true,
  permissions: [
    "amie:packets:read",
    "amie:packets:write",
    "amie:replies:read",
    "amie:replies:write",
    "amie:unmapped:read",
    "amie:unmapped:write",
    "core:access-requests:read",
    "core:access-requests:write",
    "core:allocations:read",
    "core:allocations:write",
    "core:clusters:read",
    "core:clusters:write",
    "core:organizations:read",
    "core:organizations:write",
    "core:projects:read",
    "core:projects:write",
    "core:traces:read",
    "core:users:read",
    "core:users:write",
  ],
};
export const ROLE_OPERATOR: RoleRow = {
  id: "role-operator",
  name: "Operator",
  description: "Day-to-day allocation, project, and AMIE operations (read + write).",
  is_system: false,
  permissions: [
    "amie:packets:read",
    "amie:packets:write",
    "amie:replies:read",
    "amie:replies:write",
    "amie:unmapped:read",
    "amie:unmapped:write",
    "core:allocations:read",
    "core:allocations:write",
    "core:projects:read",
    "core:projects:write",
  ],
};
export const ROLE_AUDITOR: RoleRow = {
  id: "role-auditor",
  name: "Auditor",
  description: "Read-only access across allocations, projects, tracing, and AMIE.",
  is_system: false,
  permissions: [
    "core:allocations:read",
    "core:projects:read",
    "core:traces:read",
    "amie:packets:read",
    "amie:replies:read",
    "amie:unmapped:read",
  ],
};

export const INITIAL_ROLES: RoleRow[] = [ROLE_SUPER_ADMIN, ROLE_OPERATOR, ROLE_AUDITOR];
