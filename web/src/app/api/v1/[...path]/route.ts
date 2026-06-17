import { serverEnv } from "@/lib/env";
import { getPortalSession } from "@/shared/auth/session";
import { type NextRequest, NextResponse } from "next/server";

export const runtime = "nodejs";

type Context = { params: Promise<{ path: string[] }> };

type Backend = "core";

function backendFor(path: string[]): { backend: Backend; baseUrl: string; prefix: string } {
  // Single backend today; AMIE/etc. split by path prefix lands when their
  // base URLs are wired into serverEnv.
  void path;
  return { backend: "core", baseUrl: serverEnv.CUSTOS_CORE_API_BASE_URL, prefix: "" };
}

function isAdminPath(path: string[]) {
  return path[0] === "admin";
}

async function proxy(request: NextRequest, ctx: Context) {
  const { path } = await ctx.params;
  const { baseUrl, prefix } = backendFor(path);

  const upstreamUrl = new URL(`${prefix}/${path.join("/")}`, baseUrl);
  upstreamUrl.search = request.nextUrl.search;

  const headers = new Headers();
  const incomingType = request.headers.get("content-type");
  if (incomingType) headers.set("content-type", incomingType);
  const accept = request.headers.get("accept");
  if (accept) headers.set("accept", accept);

  if (isAdminPath(path)) {
    if (!serverEnv.CUSTOS_ADMIN_CLIENT_ID || !serverEnv.CUSTOS_ADMIN_CLIENT_SECRET) {
      return NextResponse.json(
        { message: "Admin proxy not configured", path: path.join("/") },
        { status: 500 },
      );
    }
    headers.set("X-Client-Id", serverEnv.CUSTOS_ADMIN_CLIENT_ID);
    headers.set("X-Client-Secret", serverEnv.CUSTOS_ADMIN_CLIENT_SECRET);
  } else {
    const session = await getPortalSession();
    if (!session?.accessToken) {
      return NextResponse.json({ message: "Not authenticated" }, { status: 401 });
    }
    headers.set("authorization", `Bearer ${session.accessToken}`);
  }

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
