import { NextRequest, NextResponse } from "next/server";
import {
  API_BASE_URL,
  CLIENT_ID,
  CLIENT_SECRET,
  DEV_BEARER_TOKEN,
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
  auth: "bearer" | "client"
) {
  // Mirror the browser-facing /api/v1 route onto the local Signer backend while
  // injecting the auth scheme expected by each endpoint family.
  const { path } = await context.params;
  const upstreamUrl = new URL(`/api/v1/${path.join("/")}`, API_BASE_URL);
  upstreamUrl.search = request.nextUrl.search;

  const headers = new Headers();
  const contentType = request.headers.get("content-type");

  if (contentType) {
    headers.set("Content-Type", contentType);
  }

  if (auth === "bearer") {
    headers.set("Authorization", `Bearer ${DEV_BEARER_TOKEN}`);
  } else {
    headers.set("X-Client-Id", CLIENT_ID);
    headers.set("X-Client-Secret", CLIENT_SECRET);
  }

  const response = await fetch(upstreamUrl, {
    method: request.method,
    headers,
    body: request.method === "GET" ? undefined : await request.text(),
    // Certificate state should reflect the signer service, not a cached route.
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
