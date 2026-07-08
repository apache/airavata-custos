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

import { type NextRequest, NextResponse } from "next/server";
import { serverEnv } from "@/lib/env";
import { auth } from "@/shared/auth/auth";

export const runtime = "nodejs";

// Cleared inline so the redirect lands on /sign-in already-signed-out,
// regardless of whether the IdP round-trips.
const SESSION_COOKIES = [
  "custos.session-token",
  "__Secure-custos.session-token",
];

function clearAuthCookies(res: NextResponse) {
  for (const name of SESSION_COOKIES) {
    res.cookies.set({ name, value: "", path: "/", maxAge: 0 });
  }
}

// Discover the end-session endpoint; some issuers (CILogon) don't have one.
async function discoverEndSessionEndpoint(issuer: string): Promise<string | null> {
  try {
    const res = await fetch(`${issuer.replace(/\/$/, "")}/.well-known/openid-configuration`, {
      cache: "no-store",
    });
    if (!res.ok) return null;
    const doc = (await res.json()) as { end_session_endpoint?: unknown };
    return typeof doc.end_session_endpoint === "string" ? doc.end_session_endpoint : null;
  } catch {
    return null;
  }
}

export async function GET(request: NextRequest) {
  const callbackUrl = request.nextUrl.searchParams.get("callbackUrl") ?? "/sign-in";
  const origin = request.nextUrl.origin;
  const postLogout = new URL(callbackUrl, origin).toString();

  const session = await auth();
  const idToken = session?.idToken ?? null;
  const endpoint = await discoverEndSessionEndpoint(serverEnv.OIDC_ISSUER_URL);

  let redirectTo = postLogout;
  if (endpoint) {
    const url = new URL(endpoint);
    url.searchParams.set("post_logout_redirect_uri", postLogout);
    if (idToken) {
      url.searchParams.set("id_token_hint", idToken);
    } else {
      url.searchParams.set("client_id", serverEnv.OIDC_CLIENT_ID);
    }
    redirectTo = url.toString();
  }

  const res = NextResponse.redirect(redirectTo);
  clearAuthCookies(res);
  return res;
}
