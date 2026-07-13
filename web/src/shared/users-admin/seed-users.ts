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

import type { UserIdentity } from "@/generated/core/types.gen";
import { ROLE_AUDITOR, ROLE_OPERATOR, ROLE_SUPER_ADMIN } from "./roles-catalog";
import type { UserRow } from "./types";

function identity(
  user_id: string,
  source: string,
  external_id: string,
  email?: string,
): UserIdentity {
  return {
    id: `${user_id}-${source}`,
    user_id,
    source,
    external_id,
    email,
    created_at: "2026-01-15T09:00:00Z",
  };
}

export const INITIAL_USERS: UserRow[] = [
  {
    id: "u1",
    first_name: "Dev",
    last_name: "Admin",
    email: "admin@custos.local",
    status: "ACTIVE",
    roles: [ROLE_SUPER_ADMIN, ROLE_OPERATOR, ROLE_AUDITOR],
    identities: [
      identity("u1", "access", "dev-admin-access-001", "admin@access-ci.org"),
      identity("u1", "cilogon", "cilogon-uid-1001", "admin@cilogon.org"),
      identity("u1", "orcid", "0000-0001-1111-0001"),
      identity("u1", "nairr", "nairr-uid-1001"),
    ],
  },
  {
    id: "u2",
    first_name: "Rachel",
    last_name: "Gao",
    email: "rgao@access-ci.org",
    status: "ACTIVE",
    roles: [ROLE_OPERATOR],
    identities: [
      identity("u2", "access", "rgao-access-001", "rgao@access-ci.org"),
      identity("u2", "orcid", "0000-0002-1825-0097"),
    ],
  },
  {
    id: "u3",
    first_name: "James",
    last_name: "Okonkwo",
    email: "jokonkwo@university.edu",
    status: "ACTIVE",
    roles: [ROLE_AUDITOR],
    identities: [identity("u3", "cilogon", "cilogon-uid-3003", "jokonkwo@cilogon.org")],
  },
  {
    id: "u4",
    first_name: "Priya",
    last_name: "Sharma",
    email: "psharma@hpc-lab.org",
    status: "SUSPENDED",
    roles: [ROLE_OPERATOR, ROLE_AUDITOR],
    identities: [
      identity("u4", "access", "psharma-access-004", "psharma@hpc-lab.org"),
      identity("u4", "cilogon", "cilogon-uid-4004"),
      identity("u4", "nairr", "nairr-uid-4004"),
    ],
  },
  {
    id: "u5",
    first_name: "Daniel",
    last_name: "Wu",
    email: "dwu@custos-hpc.io",
    status: "INACTIVE",
    identities: [],
    roles: [],
  },
];
