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

import { recordTraceId } from "./last-trace-id";

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly path: string,
    public readonly body: unknown,
    message?: string,
  ) {
    super(message ?? `API ${status} on ${path}`);
    this.name = "ApiError";
  }
}

export type ApiFetchInit = Omit<RequestInit, "body"> & {
  body?: unknown;
};

const PORTAL_API_PREFIX = "/api/v1";

function buildUrl(path: string): string {
  if (path.startsWith("http://") || path.startsWith("https://")) return path;
  const normalized = path.startsWith("/") ? path : `/${path}`;
  if (normalized.startsWith(PORTAL_API_PREFIX)) return normalized;
  return `${PORTAL_API_PREFIX}${normalized}`;
}

async function parseBody(response: Response): Promise<unknown> {
  const contentType = response.headers.get("content-type") ?? "";
  if (response.status === 204) return null;
  if (contentType.includes("application/json")) {
    try {
      return await response.json();
    } catch {
      return null;
    }
  }
  try {
    return await response.text();
  } catch {
    return null;
  }
}

// The backend reports errors as {"error": "..."}; surface that text so dialogs
// show the guard message instead of a generic "API 409 on ...".
function errorMessageFromBody(body: unknown): string | undefined {
  if (body && typeof body === "object" && "error" in body) {
    const { error } = body as { error: unknown };
    if (typeof error === "string" && error.length > 0) return error;
  }
  return undefined;
}

export async function apiFetch<T = unknown>(path: string, init: ApiFetchInit = {}): Promise<T> {
  const url = buildUrl(path);
  const headers = new Headers(init.headers);
  let body: BodyInit | undefined;

  if (init.body !== undefined && init.body !== null) {
    if (
      typeof init.body === "string" ||
      init.body instanceof FormData ||
      init.body instanceof Blob
    ) {
      body = init.body as BodyInit;
    } else {
      if (!headers.has("content-type")) headers.set("content-type", "application/json");
      body = JSON.stringify(init.body);
    }
  }

  if (!headers.has("accept")) headers.set("accept", "application/json");

  const response = await fetch(url, { ...init, body, headers });
  // Backend HTTP middleware sets X-Trace-Id; capture so callers can deep-link.
  const traceId = response.headers.get("x-trace-id");
  if (traceId) recordTraceId(traceId);
  const parsed = await parseBody(response);

  if (!response.ok) {
    throw new ApiError(response.status, url, parsed, errorMessageFromBody(parsed));
  }
  return parsed as T;
}
