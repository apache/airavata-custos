/**
 * Browser → Signer API proxy.
 *
 * All calls to /api/v1/* from the portal client are server-handled here and
 * forwarded to the signer service. The proxy exists to keep the signer's
 * client credentials and the user's session token server-side: nothing
 * sensitive ever reaches the browser bundle.
 *
 * GET endpoints (e.g. /userinfo, /certificates, /certificates/{serial}) are
 * scoped to the signed-in user and require a session bearer. POST endpoints
 * (currently /revoke) act on behalf of the portal as an OAuth client and
 * use signer-issued client credentials instead.
 */
import { NextRequest, NextResponse } from "next/server";
import { auth } from "../../../../../auth";
import {
  API_BASE_URL,
  CLIENT_ID,
  CLIENT_SECRET,
} from "../../../../lib/serverConfig";

type Context = {
  params: Promise<{
    path: string[];
  }>;
};

export async function GET(request: NextRequest, context: Context) {
  return proxySignerRequest(request, context, "bearer");
}

export async function POST(request: NextRequest, context: Context) {
  return proxySignerRequest(request, context, "client");
}

async function proxySignerRequest(
  request: NextRequest,
  context: Context,
  authScheme: "bearer" | "client"
) {
  const { path } = await context.params;
  const upstreamUrl = new URL(`/api/v1/${path.join("/")}`, API_BASE_URL);
  upstreamUrl.search = request.nextUrl.search;

  const headers = new Headers();
  const contentType = request.headers.get("content-type");

  if (contentType) {
    headers.set("Content-Type", contentType);
  }

  if (authScheme === "bearer") {
    // Read the session server-side and forward the issuer-issued access
    // token to the signer. A missing session means an anonymous request,
    // which the signer would reject anyway — short-circuit with a clearer
    // 401 so the client doesn't have to decode an upstream HTML error.
    const session = await auth();

    if (!session?.accessToken) {
      return NextResponse.json(
        { message: "Not authenticated" },
        { status: 401 }
      );
    }

    headers.set("Authorization", `Bearer ${session.accessToken}`);
  } else {
    // Client-credentials path. Used for endpoints the signer treats as
    // portal-initiated rather than user-initiated (e.g. revocations).
    headers.set("X-Client-Id", CLIENT_ID);
    headers.set("X-Client-Secret", CLIENT_SECRET);
  }

  const response = await fetch(upstreamUrl, {
    method: request.method,
    headers,
    body: request.method === "GET" ? undefined : await request.text(),
    // Certificate state changes outside the request lifetime (issuance,
    // revocation), so route caching would hand back stale data.
    cache: "no-store",
  });

  const body = await response.text();
  const responseHeaders = new Headers();
  const upstreamContentType = response.headers.get("content-type");

  if (upstreamContentType) {
    responseHeaders.set("Content-Type", upstreamContentType);
  }

  return new NextResponse(body, {
    status: response.status,
    headers: responseHeaders,
  });
}
