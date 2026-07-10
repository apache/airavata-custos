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
import { getEffectiveRate, listResourceRates, listResourceSummaries } from "../api";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

function mockResponse(status: number, body: unknown): Response {
  const text = typeof body === "string" ? body : JSON.stringify(body);
  return new Response(text, { status, headers: { "content-type": "application/json" } });
}

afterEach(() => {
  fetchMock.mockReset();
});

const summary = {
  id: "res-a-cpu",
  name: "ClusterA CPU",
  resource_type: "CPU",
  resource_amount: 1000000,
  compute_cluster_id: "cluster-a",
  allocation_count: 2,
  total_allocated: 50000,
  total_used_su: 12500,
  rate_count: 3,
};
const rate = {
  id: "rate-1",
  compute_allocation_resource_id: "res-a-cpu",
  rate: 0.05,
  start_time: "2020-01-01T00:00:00Z",
  end_time: "2035-01-01T00:00:00Z",
};

describe("listResourceSummaries", () => {
  it("parses a bare array of summary rows", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, [summary]));
    const out = await listResourceSummaries();
    expect(out[0]?.name).toBe("ClusterA CPU");
    expect(out[0]?.rate_count).toBe(3);
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).toContain("/compute-allocation-resources/summary");
  });

  it("returns an empty array when the backend sends null", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, "null"));
    await expect(listResourceSummaries()).resolves.toEqual([]);
  });
});

describe("listResourceRates", () => {
  it("parses a bare array of rates", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, [rate]));
    const out = await listResourceRates("res-a-cpu");
    expect(out[0]?.rate).toBe(0.05);
  });

  it("returns an empty array when the backend sends null", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, "null"));
    await expect(listResourceRates("res-a-cpu")).resolves.toEqual([]);
  });
});

describe("getEffectiveRate", () => {
  it("parses a single effective rate and appends the at query when given", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, rate));
    await getEffectiveRate("res-a-cpu", "2026-06-01T00:00:00Z");
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).toContain("/rates/effective?at=");
  });
});
