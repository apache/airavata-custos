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

export type ApiFetchInit = Omit<RequestInit, "body"> & { body?: unknown };

export async function apiFetch<T = unknown>(path: string, init: ApiFetchInit = {}): Promise<T> {
  const normalized = path.startsWith("/") ? path : `/${path}`;
  const url = `/signer/api/v1${normalized}`;
  const headers = new Headers(init.headers);
  headers.set("accept", headers.get("accept") ?? "application/json");
  let body: BodyInit | undefined;
  if (init.body != null) {
    headers.set("content-type", headers.get("content-type") ?? "application/json");
    body = typeof init.body === "string" ? init.body : JSON.stringify(init.body);
  }

  const response = await fetch(url, { ...init, headers, body });
  const contentType = response.headers.get("content-type") ?? "";
  let parsed: unknown = null;
  if (response.status !== 204) {
    parsed = contentType.includes("application/json")
      ? await response.json()
      : await response.text();
  }
  if (!response.ok) {
    throw new ApiError(response.status, url, parsed, errorMessage(parsed));
  }
  return parsed as T;
}

function errorMessage(body: unknown): string | undefined {
  if (!body || typeof body !== "object") return undefined;
  for (const key of ["error", "message"] as const) {
    const value = (body as Record<string, unknown>)[key];
    if (typeof value === "string" && value) return value;
  }
  return undefined;
}
