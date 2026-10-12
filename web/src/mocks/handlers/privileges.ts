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

import { zPrivilegeKey } from "@/generated/core/zod.gen";
import { http, HttpResponse } from "msw";

// Connectors register these at runtime; the core spec's PrivilegeKey enum
// does not list them, so the effective set stays plain strings.
export const CONNECTOR_PRIVILEGES = [
  "amie:packets:read",
  "amie:packets:write",
  "amie:replies:read",
  "amie:replies:write",
  "amie:unmapped:read",
  "amie:unmapped:write",
  "temp-account:accounts:read",
  "temp-account:accounts:write",
];

// Default to admin-grade so MSW-only browsing exercises the full UI; tests
// override per-case via server.use().
const ALL_PRIVILEGES = [...zPrivilegeKey.options, ...CONNECTOR_PRIVILEGES];

// Test seam: e2e scopes privileges per persona via a non-httpOnly cookie.
// Unset (MSW-only browsing, unit tests) falls back to full access.
export function effectivePrivileges(): string[] {
  if (typeof document === "undefined") return ALL_PRIVILEGES;
  const match = document.cookie.split("; ").find((c) => c.startsWith("custos.test-privileges="));
  if (!match) return ALL_PRIVILEGES;
  const list = decodeURIComponent(match.slice("custos.test-privileges=".length))
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
  return list.length > 0 ? list : ALL_PRIVILEGES;
}

export const privilegesHandlers = [
  http.get("*/api/v1/user/privileges", () =>
    HttpResponse.json({ privileges: effectivePrivileges() }),
  ),
];
