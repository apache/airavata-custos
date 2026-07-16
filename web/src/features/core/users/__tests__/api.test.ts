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

import { afterEach, describe, expect, it, vi } from "vitest";
import { assignUserRole, listUsers, removeUserRole } from "../api";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

function mockResponse(status: number, body?: unknown): Response {
  return new Response(body === undefined ? null : JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" },
  });
}

afterEach(() => fetchMock.mockReset());

describe("users API", () => {
  it("sends limit and offset when listing users", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(200, {
        items: [{ id: "user-1", email: "user@example.org" }],
        total: 1,
      }),
    );
    const result = await listUsers({ limit: 25, offset: 50 });
    expect(result.total).toBe(1);
    const url = String(fetchMock.mock.calls[0]?.[0]);
    expect(url).toContain("limit=25");
    expect(url).toContain("offset=50");
  });

  it("posts the role id when assigning a role", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(201, { user_id: "user-1", role_id: "role-1" }));
    await assignUserRole("user-1", "role-1");
    expect(fetchMock.mock.calls[0]?.[1]).toMatchObject({
      method: "POST",
      body: JSON.stringify({ role_id: "role-1" }),
    });
  });

  it("includes the reason in the assignment body when provided", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(201, { user_id: "user-1", role_id: "role-1" }));
    await assignUserRole("user-1", "role-1", "Onboarding");
    expect(fetchMock.mock.calls[0]?.[1]).toMatchObject({
      method: "POST",
      body: JSON.stringify({ role_id: "role-1", reason: "Onboarding" }),
    });
  });

  it("omits the reason when the value is blank", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(201, { user_id: "user-1", role_id: "role-1" }));
    await assignUserRole("user-1", "role-1", "   ");
    expect(fetchMock.mock.calls[0]?.[1]).toMatchObject({
      method: "POST",
      body: JSON.stringify({ role_id: "role-1" }),
    });
  });

  it("uses the assignment resource when removing a role", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(204));
    await removeUserRole("user-1", "role-1");
    expect(fetchMock.mock.calls[0]?.[0]).toContain("/users/user-1/roles/role-1");
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe("DELETE");
  });
});
