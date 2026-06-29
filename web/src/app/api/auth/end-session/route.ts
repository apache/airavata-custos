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

// next-auth v5 session cookie names — cleared inline so the redirect lands on
// /sign-in already-signed-out, regardless of whether Keycloak round-trips.
const NEXT_AUTH_COOKIES = [
  "authjs.session-token",
  "__Secure-authjs.session-token",
];

function clearAuthCookies(res: NextResponse) {
  for (const name of NEXT_AUTH_COOKIES) {
    res.cookies.set({ name, value: "", path: "/", maxAge: 0 });
  }
}

export async function GET(request: NextRequest) {
  const callbackUrl = request.nextUrl.searchParams.get("callbackUrl") ?? "/sign-in";
  const origin = request.nextUrl.origin;
  const postLogout = new URL(callbackUrl, origin).toString();

  const session = await auth();
  const idToken = session?.idToken ?? null;
  const endSession = new URL(
    `${serverEnv.OIDC_ISSUER_URL.replace(/\/$/, "")}/protocol/openid-connect/logout`,
  );
  endSession.searchParams.set("post_logout_redirect_uri", postLogout);
  if (idToken) {
    endSession.searchParams.set("id_token_hint", idToken);
  } else {
    endSession.searchParams.set("client_id", serverEnv.OIDC_CLIENT_ID);
  }

  const res = NextResponse.redirect(endSession.toString());
  clearAuthCookies(res);
  return res;
}
