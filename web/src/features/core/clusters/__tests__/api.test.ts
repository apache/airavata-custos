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
import { getCluster, listClusters } from "../api";

const fetchMock = vi.fn<typeof fetch>();
vi.stubGlobal("fetch", fetchMock);

function mockResponse(status: number, body: unknown): Response {
  const text = typeof body === "string" ? body : JSON.stringify(body);
  return new Response(text, { status, headers: { "content-type": "application/json" } });
}

afterEach(() => {
  fetchMock.mockReset();
});

const cluster = { id: "cluster-a", name: "ClusterA" };

describe("listClusters", () => {
  it("parses a bare array of clusters", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, [cluster]));
    const out = await listClusters();
    expect(out).toHaveLength(1);
    expect(out[0]?.name).toBe("ClusterA");
  });

  it("returns an empty array when the backend sends null", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, "null"));
    await expect(listClusters()).resolves.toEqual([]);
  });
});

describe("getCluster", () => {
  it("parses a single cluster", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, cluster));
    const out = await getCluster("cluster-a");
    expect(out.id).toBe("cluster-a");
  });
});
