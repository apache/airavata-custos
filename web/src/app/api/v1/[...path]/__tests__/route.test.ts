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

import { NextRequest } from "next/server";
import { afterEach, describe, expect, it, vi } from "vitest";

import { responseBodyForStatus } from "../proxy-response";

vi.mock("@/lib/env", () => ({
  serverEnv: {
    CUSTOS_CORE_API_BASE_URL: "https://core.example.org",
  },
}));

vi.mock("@/shared/auth/session", () => ({
  getPortalSession: vi.fn(async () => ({ user: { email: "admin@custos.local" } })),
  pickBackendBearer: vi.fn(() => "access-token-abc"),
}));

import { POST } from "../route";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

const ctx = { params: Promise.resolve({ path: ["roles", "role-1", "privileges"] }) };

afterEach(() => {
  fetchMock.mockReset();
});

describe("responseBodyForStatus", () => {
  it.each([204, 205, 304])("returns null for bodyless status %s", (status) => {
    expect(responseBodyForStatus(status, "")).toBeNull();
  });

  it("preserves an ordinary response body", () => {
    expect(responseBodyForStatus(200, '{"ok":true}')).toBe('{"ok":true}');
  });
});

describe("api v1 proxy route", () => {
  it("proxies no-content backend responses without constructing a response body", async () => {
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 204 }));

    const response = await POST(
      new NextRequest("http://localhost:3000/api/v1/roles/role-1/privileges", {
        method: "POST",
        body: JSON.stringify({ privilege: "core:roles:manage" }),
        headers: { "content-type": "application/json" },
      }),
      ctx,
    );

    expect(response.status).toBe(204);
    expect(await response.text()).toBe("");
  });
});
