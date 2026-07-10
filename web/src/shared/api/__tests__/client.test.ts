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

import { ApiError, apiFetch } from "@/shared/api/client";
import { getLastTraceId, recordTraceId } from "@/shared/api/last-trace-id";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest";

const TRACE_ID = "a3b1c92d3f4e5a6b7c8d9e0f12345678";

const baseHandlers = [
  http.get("*/api/v1/healthz", () =>
    HttpResponse.json({ status: "ok" }, { headers: { "x-trace-id": TRACE_ID } }),
  ),
  http.get("*/api/v1/example", () => HttpResponse.json({ ok: true })),
  http.delete("*/api/v1/grant", async ({ request }) => HttpResponse.json(await request.json())),
  http.delete("*/api/v1/users/u1/privileges/last", () =>
    HttpResponse.json(
      { error: "cannot remove the last holder of core:roles:manage" },
      { status: 409 },
    ),
  ),
];

const server = setupServer(...baseHandlers);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());
afterEach(() => server.resetHandlers(...baseHandlers));
beforeEach(() => recordTraceId(null));

describe("apiFetch", () => {
  it("prepends /api/v1 and returns the JSON body", async () => {
    const body = await apiFetch<{ status: string }>("/healthz");
    expect(body).toEqual({ status: "ok" });
  });

  it("captures X-Trace-Id into the singleton when the header is present", async () => {
    await apiFetch("/healthz");
    expect(getLastTraceId()).toBe(TRACE_ID);
  });

  it("leaves the singleton untouched when the header is absent", async () => {
    recordTraceId("previous-id");
    await apiFetch("/example");
    expect(getLastTraceId()).toBe("previous-id");
  });

  it("sends a JSON body on DELETE requests", async () => {
    const body = await apiFetch<{ reason: string }>("/grant", {
      method: "DELETE",
      body: { reason: "no longer needed" },
    });
    expect(body).toEqual({ reason: "no longer needed" });
  });

  it("derives the ApiError message from the body's error field", async () => {
    const error = await apiFetch("/users/u1/privileges/last", { method: "DELETE" }).catch(
      (e) => e,
    );
    expect(error).toBeInstanceOf(ApiError);
    expect((error as ApiError).message).toBe(
      "cannot remove the last holder of core:roles:manage",
    );
    expect((error as ApiError).status).toBe(409);
  });
});
