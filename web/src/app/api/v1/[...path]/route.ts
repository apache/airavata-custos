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

export const runtime = "nodejs";

type Context = { params: Promise<{ path: string[] }> };

async function proxy(request: NextRequest, ctx: Context) {
  const { path } = await ctx.params;
  // The signer is a separate service rooted at /api/v1; strip the "signer"
  // routing segment and forward the rest with the user's Bearer.
  const isSigner = path[0] === "signer";
  const baseUrl = isSigner
    ? serverEnv.CUSTOS_SIGNER_API_BASE_URL
    : serverEnv.CUSTOS_CORE_API_BASE_URL;
  const upstreamPath = isSigner ? `/api/v1/${path.slice(1).join("/")}` : `/${path.join("/")}`;
  const upstreamUrl = new URL(upstreamPath, baseUrl);
  upstreamUrl.search = request.nextUrl.search;

  const session = await getPortalSession();
  const bearer = pickBackendBearer(session);
  if (!bearer) {
    return NextResponse.json({ code: "missing_bearer", message: "Not authenticated" }, { status: 401 });
  }

  const headers = new Headers();
  const incomingType = request.headers.get("content-type");
  if (incomingType) headers.set("content-type", incomingType);
  const accept = request.headers.get("accept");
  if (accept) headers.set("accept", accept);
  headers.set("authorization", `Bearer ${bearer}`);

  const method = request.method;
  const body = method === "GET" || method === "HEAD" ? undefined : await request.text();

  const upstream = await fetch(upstreamUrl, {
    method,
    headers,
    body,
    cache: "no-store",
  });

  const responseHeaders = new Headers();
  const upstreamType = upstream.headers.get("content-type");
  if (upstreamType) responseHeaders.set("content-type", upstreamType);
  const traceId = upstream.headers.get("x-trace-id");
  if (traceId) responseHeaders.set("x-trace-id", traceId);

  return new NextResponse(await upstream.text(), {
    status: upstream.status,
    headers: responseHeaders,
  });
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
