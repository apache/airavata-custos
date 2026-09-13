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
import { decodeSharedSession, pickBackendBearer } from "@/shared/auth/session";

export const runtime = "nodejs";
type Context = { params: Promise<{ path: string[] }> };

async function proxy(request: NextRequest, context: Context) {
  const session = await decodeSharedSession(request);
  const bearer = pickBackendBearer(session);
  if (!bearer) {
    return NextResponse.json(
      { code: "missing_bearer", message: "Not authenticated" },
      { status: 401 },
    );
  }

  const { path } = await context.params;
  const upstream = new URL(`/api/v1/${path.join("/")}`, serverEnv.CUSTOS_SIGNER_API_BASE_URL);
  upstream.search = request.nextUrl.search;
  const headers = new Headers({ authorization: `Bearer ${bearer}` });
  for (const name of ["accept", "content-type", "x-trace-id"]) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }
  const body =
    request.method === "GET" || request.method === "HEAD" ? undefined : await request.text();

  let response: Response;
  try {
    response = await fetch(upstream, { method: request.method, headers, body, cache: "no-store" });
  } catch {
    return NextResponse.json(
      { code: "upstream_unavailable", message: "Signer service is unavailable" },
      { status: 503 },
    );
  }

  const responseHeaders = new Headers();
  for (const name of ["content-type", "x-trace-id"]) {
    const value = response.headers.get(name);
    if (value) responseHeaders.set(name, value);
  }
  const responseBody = response.status === 204 ? null : await response.arrayBuffer();
  return new NextResponse(responseBody, { status: response.status, headers: responseHeaders });
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
