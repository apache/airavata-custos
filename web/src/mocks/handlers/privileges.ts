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

import { http, HttpResponse } from "msw";
import type { Privilege } from "@/features/core/identity/types";

// Default to admin-grade so MSW-only browsing exercises the full UI; tests
// override per-case via server.use().
const ALL_PRIVILEGES: Privilege[] = [
  "amie:read",
  "amie:write",
  "hpc:read",
  "hpc:write",
  "signer:read",
  "signer:write",
  "privileges:grant",
  "roles:manage",
];

const MOCK_USER = {
  id: "msw-user",
  organization_id: "msw-org",
  first_name: "MSW",
  last_name: "User",
  email: "msw@custos.local",
  status: "ACTIVE",
  type: "CLUSTER_LOCAL",
};

export const privilegesHandlers = [
  http.get("*/api/v1/user/privileges", () =>
    HttpResponse.json({ privileges: ALL_PRIVILEGES }),
  ),
  http.get("*/api/v1/me", () =>
    HttpResponse.json({ user: MOCK_USER, privileges: ALL_PRIVILEGES }),
  ),
];
