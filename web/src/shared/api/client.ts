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

import type { CreateClientConfig } from "@/generated/core/client.gen";
import { z } from "zod";

const errorBodySchema = z.object({
  error: z.string().optional(),
  code: z.string().optional(),
});
type ApiErrorBody = z.infer<typeof errorBodySchema> | string | null;

export class ApiError extends Error {
  constructor(
    public readonly status: number,
    public readonly path: string,
    public readonly body: ApiErrorBody,
    message?: string,
  ) {
    super(message ?? `The request failed (HTTP ${status})`);
    this.name = "ApiError";
  }
}

// The machine-readable code of an ApiError body ({"code": "user_disabled"}), else null.
export function errorCode(error: unknown): string | null {
  if (!(error instanceof ApiError) || !error.body || typeof error.body !== "object") return null;
  return error.body.code ?? null;
}

// A cause master returns for a request, worded for the user: matched on status and, when given,
// on master's error text, which is never shown.
export type KnownError = { status: number; when?: RegExp; copy: string };

// What a failed request tells the user: the first known cause it matches, else a generic refusal
// (4xx) or retry prompt (5xx, network). A plain Error is the portal's own worded refusal.
export function errorCopy(error: unknown, known: readonly KnownError[] = []): string {
  if (!(error instanceof ApiError))
    return error instanceof Error && error.name === "Error"
      ? error.message
      : "Something went wrong. Try again.";
  const text = typeof error.body === "string" ? error.body : (error.body?.error ?? "");
  const cause = known.find((k) => k.status === error.status && (k.when?.test(text) ?? true));
  if (cause) return cause.copy;
  return error.status < 500 ? "This change was refused." : "Something went wrong. Try again.";
}

function parseErrorBody(text: string): ApiErrorBody {
  try {
    return errorBodySchema.parse(JSON.parse(text));
  } catch {
    return text || null;
  }
}

// Under the mock backend, requests wait for the worker; pages still render on the server, so a
// server redirect stays an HTTP redirect.
const mockWorker =
  process.env.NEXT_PUBLIC_PORTAL_USE_MSW === "true" && typeof window !== "undefined"
    ? import("@/mocks/browser")
        .then(({ worker }) =>
          worker.start({
            onUnhandledRequest: "bypass",
            serviceWorker: { url: "/mockServiceWorker.js" },
          }),
        )
        .catch((err) => console.error("MSW worker failed to start", err))
    : undefined;

// Every generated SDK client fetches through here; non-2xx responses throw ApiError.
const portalFetch: typeof fetch = async (input, init) => {
  await mockWorker;
  const response = await fetch(input, init);
  if (response.ok) return response;
  const url = input instanceof Request ? input.url : String(input);
  const body = parseErrorBody(await response.text());
  // The backend reports {"error": "..."}, the sign-in refusals with a "code" beside it.
  const message = typeof body === "object" ? body?.error : undefined;
  throw new ApiError(response.status, url, body, message || undefined);
};

// GETs a portal API path (e.g. "/audit/traces") as JSON through the same error handling.
export async function apiFetch<T = unknown>(path: string): Promise<T> {
  const response = await portalFetch(`/api/v1${path}`, { headers: { accept: "application/json" } });
  return (await response.json()) as T;
}

// URL parsing resolves "." and ".." as dot segments even when escaped, so a request with such a path
// value would reach another route; it is refused before it is sent.
const refuseDotSegments = async (data: unknown) => {
  const { url, path } = data as { url: string; path?: Record<string, unknown> };
  const dot = Object.values(path ?? {}).find((value) => value === "." || value === "..");
  if (dot !== undefined) throw new ApiError(400, url, null, `"${dot}" is not a valid ID`);
};

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  fetch: portalFetch,
  requestValidator: refuseDotSegments,
});
