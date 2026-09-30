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

import { serverEnv } from "@/lib/env";
import { getPortalSession, pickBackendBearer } from "@/shared/auth/session";
import { type NextRequest, NextResponse } from "next/server";
import { responseBodyForStatus } from "./proxy-response";

export const runtime = "nodejs";

type Context = { params: Promise<{ path: string[] }> };

async function proxy(request: NextRequest, ctx: Context) {
  const { path } = await ctx.params;
  const isSigner = path[0] === "signer";
  if (isSigner && process.env.CUSTOS_E2E_FAIL_ON_SIGNER_PROXY_REQUEST === "true") {
    console.error("Unexpected signer proxy request during hermetic E2E", {
      method: request.method,
      path: request.nextUrl.pathname,
    });
    return NextResponse.json(
      {
        code: "unexpected_signer_proxy_request",
        message: "Signer E2E request escaped its Playwright route",
      },
      { status: 500 },
    );
  }
  const upstreamBase = isSigner
    ? serverEnv.CUSTOS_SIGNER_API_BASE_URL
    : serverEnv.CUSTOS_CORE_API_BASE_URL;
  const upstreamPath = isSigner ? `/api/v1/${path.slice(1).join("/")}` : `/${path.join("/")}`;
  const upstreamUrl = new URL(upstreamPath, upstreamBase);
  upstreamUrl.search = request.nextUrl.search;

  const session = await getPortalSession();
  const bearer = pickBackendBearer(session);
  if (!bearer) {
    return NextResponse.json(
      { code: "missing_bearer", message: "Not authenticated" },
      { status: 401 },
    );
  }

  const headers = new Headers();
  const incomingType = request.headers.get("content-type");
  if (incomingType) headers.set("content-type", incomingType);
  const accept = request.headers.get("accept");
  if (accept) headers.set("accept", accept);
  headers.set("authorization", `Bearer ${bearer}`);

  const method = request.method;
  const body = method === "GET" || method === "HEAD" ? undefined : await request.text();

  let upstream: Response;
  try {
    upstream = await fetch(upstreamUrl, {
      method,
      headers,
      body,
      cache: "no-store",
    });
  } catch {
    console.error("API proxy upstream unavailable", {
      service: isSigner ? "signer" : "core",
      method,
      path: upstreamPath,
    });
    return NextResponse.json(
      { code: "upstream_unavailable", message: "Backend service is unavailable" },
      { status: 503 },
    );
  }

  const responseHeaders = new Headers();
  const upstreamType = upstream.headers.get("content-type");
  if (upstreamType) responseHeaders.set("content-type", upstreamType);
  const traceId = upstream.headers.get("x-trace-id");
  if (traceId) responseHeaders.set("x-trace-id", traceId);

  const upstreamBody = await upstream.text();
  return new NextResponse(responseBodyForStatus(upstream.status, upstreamBody), {
    status: upstream.status,
    headers: responseHeaders,
  });
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
