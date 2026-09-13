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

const auth = vi.hoisted(() => ({ bearer: "one.two.three" as string | null }));
vi.mock("@/lib/env", () => ({
  serverEnv: { CUSTOS_SIGNER_API_BASE_URL: "https://signer.example.org" },
}));
vi.mock("@/shared/auth/session", () => ({
  decodeSharedSession: vi.fn(async () => (auth.bearer ? {} : null)),
  pickBackendBearer: vi.fn(() => auth.bearer),
}));

import { GET, POST } from "./route";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

afterEach(() => {
  auth.bearer = "one.two.three";
  fetchMock.mockReset();
});

describe("signer BFF", () => {
  it("forwards a bearer, query, body, and trace id", async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ success: true }), {
        headers: { "content-type": "application/json", "x-trace-id": "trace-1" },
      }),
    );
    const response = await POST(
      new NextRequest("http://localhost/signer/api/v1/admin/certificates/42/revoke?audit=true", {
        method: "POST",
        headers: { "content-type": "application/json", "x-trace-id": "trace-in" },
        body: JSON.stringify({ reason: "compromised" }),
      }),
      { params: Promise.resolve({ path: ["admin", "certificates", "42", "revoke"] }) },
    );
    const [url, init] = fetchMock.mock.calls[0] ?? [];
    expect(String(url)).toBe(
      "https://signer.example.org/api/v1/admin/certificates/42/revoke?audit=true",
    );
    expect(new Headers(init?.headers).get("authorization")).toBe("Bearer one.two.three");
    expect(new Headers(init?.headers).get("x-trace-id")).toBe("trace-in");
    expect(init?.body).toBe(JSON.stringify({ reason: "compromised" }));
    expect(response.headers.get("x-trace-id")).toBe("trace-1");
  });

  it("rejects a request without a shared-session bearer", async () => {
    auth.bearer = null;
    const response = await GET(
      new NextRequest("http://localhost/signer/api/v1/admin/certificates"),
      {
        params: Promise.resolve({ path: ["admin", "certificates"] }),
      },
    );
    expect(response.status).toBe(401);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("sanitizes an unavailable upstream", async () => {
    fetchMock.mockRejectedValueOnce(new Error("secret upstream details"));
    const response = await GET(
      new NextRequest("http://localhost/signer/api/v1/admin/certificates"),
      {
        params: Promise.resolve({ path: ["admin", "certificates"] }),
      },
    );
    expect(response.status).toBe(503);
    expect(await response.json()).toEqual({
      code: "upstream_unavailable",
      message: "Signer service is unavailable",
    });
  });
});
