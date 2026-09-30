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
import { approveClusterAccount, denyClusterAccount, listClusterAccounts } from "../api";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

function mockResponse(status: number, body?: unknown): Response {
  return new Response(body === undefined ? null : JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" },
  });
}

const account = {
  id: "ccu-1",
  compute_cluster_id: "cluster-a",
  user_id: "user-1",
  local_username: "jdoe",
  approval_status: "PENDING",
  provisioned_at: null,
  display_name: "Jane Doe",
  email: "jdoe@example.edu",
  cluster_name: "cluster-a",
};

afterEach(() => fetchMock.mockReset());

describe("cluster accounts API", () => {
  // Make sure the status filter reaches the backend and a null provisioned_at
  // is accepted, since a pending account has no provisioned time yet.
  it("lists accounts by approval status", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, { items: [account], total: 1 }));
    const result = await listClusterAccounts({ approval_status: "PENDING", limit: 25 });
    expect(result.items[0]?.id).toBe("ccu-1");
    const url = String(fetchMock.mock.calls[0]?.[0]);
    expect(url).toContain("/compute-cluster-users?");
    expect(url).toContain("approval_status=PENDING");
    expect(url).toContain("limit=25");
  });

  // Make sure a denial sends the note, since it is what the audit log keeps.
  it("posts the note when denying", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(200, { id: "ccu-1", approval_status: "DENIED", provisioned_at: null }),
    );
    await denyClusterAccount("ccu-1", "Not on the list");
    expect(String(fetchMock.mock.calls[0]?.[0])).toContain("/compute-cluster-users/ccu-1/deny");
    expect(fetchMock.mock.calls[0]?.[1]).toMatchObject({
      method: "POST",
      body: JSON.stringify({ note: "Not on the list" }),
    });
  });

  it("posts to the approve route", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(200, { id: "ccu-1", approval_status: "APPROVED", provisioned_at: null }),
    );
    const result = await approveClusterAccount("ccu-1");
    expect(result.approval_status).toBe("APPROVED");
    expect(String(fetchMock.mock.calls[0]?.[0])).toContain("/compute-cluster-users/ccu-1/approve");
    expect(fetchMock.mock.calls[0]?.[1]).toMatchObject({ method: "POST" });
  });
});
