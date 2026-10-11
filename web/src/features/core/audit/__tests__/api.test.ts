/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import tracesListFixture from "@/features/core/audit/__fixtures__/traces.list.json";
import { listTraces } from "@/features/core/audit/api";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const fetchMock = vi.fn();

beforeEach(() => {
  vi.stubGlobal("fetch", fetchMock);
});

afterEach(() => {
  vi.unstubAllGlobals();
  fetchMock.mockReset();
});

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" },
  });
}

function requestedUrl(): string {
  const input = fetchMock.mock.calls[0]?.[0];
  return typeof input === "string" ? input : (input as Request).url;
}

describe("audit api", () => {
  // Make sure multi-value filters go on the wire as repeated params, sorted, so
  // the same filters always build the same URL and the same query cache key.
  it("listTraces sends sorted repeated params", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(tracesListFixture));
    await listTraces({
      source: ["core", "amie"],
      status: ["in_progress", "error"],
      from: "2026-10-01T00:00:00Z",
      q: "jsmith",
      limit: 50,
      offset: 50,
    });
    const url = new URL(requestedUrl(), "http://localhost");
    expect(url.pathname.endsWith("/audit/traces")).toBe(true);
    expect(url.searchParams.getAll("source")).toEqual(["amie", "core"]);
    expect(url.searchParams.getAll("status")).toEqual(["error", "in_progress"]);
    expect(url.searchParams.get("q")).toBe("jsmith");
    expect(url.searchParams.get("offset")).toBe("50");
  });

  // Make sure a response that does not match the wire shape is rejected at the
  // boundary instead of reaching the components half parsed.
  it("listTraces rejects a payload without delivery counts", async () => {
    const broken = {
      ...tracesListFixture,
      traces: [{ ...tracesListFixture.traces[0], deliveries: undefined }],
    };
    fetchMock.mockResolvedValueOnce(jsonResponse(broken));
    await expect(listTraces()).rejects.toThrow();
  });
});
