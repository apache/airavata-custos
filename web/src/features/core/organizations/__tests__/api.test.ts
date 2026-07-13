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

import { afterEach, describe, expect, it, vi } from "vitest";
import { createOrganization, listOrganizations } from "../api";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

function mockResponse(status: number, body: unknown): Response {
  const text = typeof body === "string" ? body : JSON.stringify(body);
  return new Response(text, { status, headers: { "content-type": "application/json" } });
}

afterEach(() => {
  fetchMock.mockReset();
});

const validOrg = {
  id: "org-gatech",
  name: "Georgia Institute of Technology",
  originated_id: "GATECH",
};

describe("listOrganizations", () => {
  it("round-trips limit/offset and returns the envelope", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, { items: [validOrg], total: 1 }));
    const result = await listOrganizations({ limit: 50, offset: 50 });
    expect(result.total).toBe(1);
    expect(result.items).toHaveLength(1);
    const calledUrl = fetchMock.mock.calls[0]?.[0] as string;
    expect(calledUrl).toContain("limit=50");
    expect(calledUrl).toContain("offset=50");
  });

  it("omits the query string when no params are given", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, { items: [], total: 0 }));
    await listOrganizations();
    const calledUrl = fetchMock.mock.calls[0]?.[0] as string;
    expect(calledUrl).not.toContain("?");
  });

  it("rejects on schema mismatch (total missing)", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, { items: [validOrg] }));
    await expect(listOrganizations()).rejects.toThrow();
  });
});

describe("createOrganization", () => {
  it("POSTs the validated payload and parses the response", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(201, validOrg));
    const out = await createOrganization({ name: "Georgia Institute of Technology" });
    expect(out.id).toBe("org-gatech");
    expect(fetchMock.mock.calls[0]?.[0]).toContain("/organizations");
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe("POST");
  });

  it("rejects an empty name before calling fetch", async () => {
    await expect(createOrganization({ name: "" })).rejects.toThrow();
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
