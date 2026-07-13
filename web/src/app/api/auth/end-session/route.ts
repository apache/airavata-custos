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

const SESSION_COOKIES = [
  "custos.session-token",
  "__Secure-custos.session-token",
];

// Cookies are deleted on the redirect response itself; sign-out must not
// depend on an IdP logout hop, which some issuers do not have.
function clearAuthCookies(request: NextRequest, res: NextResponse) {
  const requestCookies = request.cookies.getAll();
  for (const name of SESSION_COOKIES) {
    // Browsers reject __Secure- deletions that aren't marked Secure.
    const secure = name.startsWith("__Secure-");
    // Oversized sessions get chunked into name.0, name.1, ...
    const chunks = requestCookies
      .filter((c) => c.name.startsWith(`${name}.`))
      .map((c) => c.name);
    for (const target of [name, ...chunks]) {
      res.cookies.set({ name: target, value: "", path: "/", maxAge: 0, httpOnly: true, secure });
    }
  }
}

// Discover the end-session endpoint; some issuers don't have one.
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
  const callbackUrl = request.nextUrl.searchParams.get("callbackUrl") ?? "/";
  // Behind a reverse proxy request.nextUrl.origin is the internal address.
  const origin = serverEnv.NEXTAUTH_URL ?? request.nextUrl.origin;
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
  clearAuthCookies(request, res);
  return res;
}
