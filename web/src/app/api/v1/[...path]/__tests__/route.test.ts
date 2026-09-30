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
import { pickBackendBearer } from "@/shared/auth/session";

vi.mock("@/lib/env", () => ({
  serverEnv: {
    CUSTOS_CORE_API_BASE_URL: "https://core.example.org",
    CUSTOS_SIGNER_API_BASE_URL: "https://signer.example.org",
  },
}));

vi.mock("@/shared/auth/session", () => ({
  getPortalSession: vi.fn(async () => ({ user: { email: "admin@custos.local" } })),
  pickBackendBearer: vi.fn(() => "access-token-abc"),
}));

import { GET, POST } from "../route";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

const ctx = { params: Promise.resolve({ path: ["roles", "role-1", "privileges"] }) };

afterEach(() => {
  fetchMock.mockReset();
  vi.mocked(pickBackendBearer).mockReturnValue("access-token-abc");
  vi.unstubAllEnvs();
  vi.restoreAllMocks();
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
  it("routes the signer namespace to the signer API without changing Core routing", async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ certificates: [] }), {
        status: 200,
        headers: { "content-type": "application/json", "x-trace-id": "trace-signer-1" },
      }),
    );

    const response = await GET(
      new NextRequest("http://localhost:3000/api/v1/signer/admin/certificates?limit=20"),
      { params: Promise.resolve({ path: ["signer", "admin", "certificates"] }) },
    );

    expect(response.status).toBe(200);
    const [url, init] = fetchMock.mock.calls[0] ?? [];
    expect(String(url)).toBe("https://signer.example.org/api/v1/admin/certificates?limit=20");
    expect(new Headers(init?.headers).get("authorization")).toBe("Bearer access-token-abc");
    expect(response.headers.get("x-trace-id")).toBe("trace-signer-1");
  });

  it("forwards signer request bodies and query parameters", async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ success: true }), {
        status: 200,
        headers: { "content-type": "application/json" },
      }),
    );

    await POST(
      new NextRequest("http://localhost:3000/api/v1/signer/admin/certificates/42/revoke?audit=true", {
        method: "POST",
        body: JSON.stringify({ reason: "compromised" }),
        headers: { "content-type": "application/json", accept: "application/json" },
      }),
      { params: Promise.resolve({ path: ["signer", "admin", "certificates", "42", "revoke"] }) },
    );

    const [url, init] = fetchMock.mock.calls[0] ?? [];
    expect(String(url)).toBe(
      "https://signer.example.org/api/v1/admin/certificates/42/revoke?audit=true",
    );
    expect(init?.body).toBe(JSON.stringify({ reason: "compromised" }));
    expect(new Headers(init?.headers).get("content-type")).toBe("application/json");
  });

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
    expect(String(fetchMock.mock.calls[0]?.[0])).toBe(
      "https://core.example.org/roles/role-1/privileges",
    );
  });

  it("rejects unauthenticated requests before contacting an upstream", async () => {
    vi.mocked(pickBackendBearer).mockReturnValueOnce(null);

    const response = await GET(
      new NextRequest("http://localhost:3000/api/v1/signer/admin/certificates"),
      { params: Promise.resolve({ path: ["signer", "admin", "certificates"] }) },
    );

    expect(response.status).toBe(401);
    expect(await response.json()).toEqual({
      code: "missing_bearer",
      message: "Not authenticated",
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("returns a sanitized 503 when the selected upstream is unavailable", async () => {
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => undefined);
    fetchMock.mockRejectedValueOnce(new TypeError("fetch failed for secret upstream URL"));

    const response = await GET(
      new NextRequest("http://localhost:3000/api/v1/signer/admin/certificates?limit=20"),
      { params: Promise.resolve({ path: ["signer", "admin", "certificates"] }) },
    );

    expect(response.status).toBe(503);
    expect(await response.json()).toEqual({
      code: "upstream_unavailable",
      message: "Backend service is unavailable",
    });
    expect(consoleError).toHaveBeenCalledWith("API proxy upstream unavailable", {
      service: "signer",
      method: "GET",
      path: "/api/v1/admin/certificates",
    });
    expect(JSON.stringify(consoleError.mock.calls)).not.toContain("access-token-abc");
    expect(JSON.stringify(consoleError.mock.calls)).not.toContain("signer.example.org");
    expect(JSON.stringify(consoleError.mock.calls)).not.toContain("secret upstream URL");
  });

  it("fails explicitly if a hermetic signer E2E request reaches the proxy", async () => {
    vi.stubEnv("CUSTOS_E2E_FAIL_ON_SIGNER_PROXY_REQUEST", "true");
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => undefined);

    const response = await GET(
      new NextRequest("http://localhost:3000/api/v1/signer/admin/certificates"),
      { params: Promise.resolve({ path: ["signer", "admin", "certificates"] }) },
    );

    expect(response.status).toBe(500);
    expect(await response.json()).toEqual({
      code: "unexpected_signer_proxy_request",
      message: "Signer E2E request escaped its Playwright route",
    });
    expect(fetchMock).not.toHaveBeenCalled();
    expect(consoleError).toHaveBeenCalledWith(
      "Unexpected signer proxy request during hermetic E2E",
      { method: "GET", path: "/api/v1/signer/admin/certificates" },
    );
  });
});
