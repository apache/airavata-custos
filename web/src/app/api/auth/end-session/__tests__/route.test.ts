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

vi.mock("@/shared/auth/auth", () => ({
  auth: vi.fn(async () => ({ idToken: "id-token-abc" })),
}));

vi.mock("@/lib/env", () => ({
  serverEnv: {
    OIDC_ISSUER_URL: "https://idp.example.org",
    OIDC_CLIENT_ID: "client-123",
    NEXTAUTH_URL: "https://portal.example.org",
  },
}));

import { GET } from "../route";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

function discoveryResponse(body: unknown, ok = true): Response {
  return new Response(JSON.stringify(body), { status: ok ? 200 : 500 });
}

function request() {
  // Internal origin, as seen behind the reverse proxy; the public URL must
  // come from NEXTAUTH_URL, never from here.
  return new NextRequest("http://localhost:3000/api/auth/end-session?callbackUrl=/sign-in");
}

afterEach(() => {
  fetchMock.mockReset();
});

describe("end-session route", () => {
  it("round-trips through the IdP when discovery advertises end_session_endpoint", async () => {
    fetchMock.mockResolvedValueOnce(
      discoveryResponse({ end_session_endpoint: "https://idp.example.org/logout" }),
    );
    const res = await GET(request());
    const location = new URL(res.headers.get("location") as string);
    expect(location.origin + location.pathname).toBe("https://idp.example.org/logout");
    expect(location.searchParams.get("id_token_hint")).toBe("id-token-abc");
    expect(location.searchParams.get("post_logout_redirect_uri")).toBe(
      "https://portal.example.org/sign-in",
    );
  });

  it("redirects straight to the callback when the issuer has no end_session_endpoint", async () => {
    fetchMock.mockResolvedValueOnce(discoveryResponse({}));
    const res = await GET(request());
    expect(res.headers.get("location")).toBe("https://portal.example.org/sign-in");
  });

  it("redirects locally when discovery fails", async () => {
    fetchMock.mockRejectedValueOnce(new Error("network"));
    const res = await GET(request());
    expect(res.headers.get("location")).toBe("https://portal.example.org/sign-in");
  });

  it("clears the session cookie regardless of the redirect target", async () => {
    fetchMock.mockResolvedValueOnce(discoveryResponse({}));
    const res = await GET(request());
    const cookie = res.cookies.get("custos.session-token");
    expect(cookie?.value).toBe("");
    expect(cookie?.maxAge).toBe(0);
  });
});
