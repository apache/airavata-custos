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

import { canonicalPath } from "@/shared/layout/nav";
import { type NextRequest, NextResponse } from "next/server";

// Inlined from shared/auth/auth.ts to keep next-auth out of the bundle.
const SESSION_COOKIES = ["__Secure-custos.session-token", "custos.session-token"];

// The portal layout reads the requested path from this header to return a signed-out user to it.
export const PATH_HEADER = "x-portal-path";

// Signed-out users get the landing page at the root; the check is presence-only, an
// invalid cookie falls through to the portal's own auth gate.
export function middleware(request: NextRequest) {
  const { pathname, search } = request.nextUrl;
  if (pathname !== "/") {
    const canonical = canonicalPath(pathname);
    // A path no route matches, so the 404 page renders on the server.
    if (canonical === null) return NextResponse.rewrite(new URL("/_malformed-id", request.url));
    if (canonical !== pathname) {
      const url = request.nextUrl.clone();
      url.pathname = canonical;
      return NextResponse.redirect(url, 308);
    }
    const headers = new Headers(request.headers);
    headers.set(PATH_HEADER, pathname + search);
    return NextResponse.next({ request: { headers } });
  }
  // Match chunked cookies (name.0, name.1, ...) too.
  const hasSession = request.cookies
    .getAll()
    .some(({ name }) =>
      SESSION_COOKIES.some((base) => name === base || name.startsWith(`${base}.`)),
    );
  if (!hasSession) {
    const landing = new URL("/landing/index.html", request.url);
    landing.search = request.nextUrl.search;
    return NextResponse.rewrite(landing);
  }
  return NextResponse.next();
}

export const config = { matcher: "/((?!api/|_next/).*)" };
