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

import { type ClusterAccountResponse, zClusterAccountResponse } from "@/generated/core/zod.gen";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import fixture from "../__fixtures__/cluster-accounts.json";
import {
  ClusterAccountBadge,
  ClusterAccountDecisions,
  clusterAccountStatus,
} from "../components/ClusterAccountBadge";

const rows = zClusterAccountResponse.array().parse(fixture);

function byId(id: string): ClusterAccountResponse {
  const row = rows.find((r) => r.id === id);
  if (!row) throw new Error(`no fixture ${id}`);
  return row;
}

function decisionsFor(id: string): string[] {
  const { unmount } = render(
    <ClusterAccountDecisions account={byId(id)} onApprove={vi.fn()} onDeny={vi.fn()} />,
  );
  const names = screen.queryAllByRole("button").map((b) => b.textContent ?? "");
  unmount();
  return names;
}

describe("cluster accounts", () => {
  // Make sure an approved account reads as provisioning until it exists on
  // the cluster, so the admin can tell approval from a working login.
  it("tells approved from active by provisioned_at", () => {
    expect(clusterAccountStatus(byId("ccu-approved-1")).label).toBe("Approved · provisioning");
    expect(clusterAccountStatus(byId("ccu-active-1")).label).toBe("Active");
    expect(clusterAccountStatus(byId("ccu-pending-1")).label).toBe("Pending approval");
    expect(clusterAccountStatus(byId("ccu-denied-1")).label).toBe("Denied");
  });

  // Make sure a pending account offers both decisions, a denied one can only
  // be approved, and an approved one cannot be reviewed again.
  it("offers the decisions each status allows", () => {
    expect(decisionsFor("ccu-pending-1")).toEqual(["Deny", "Approve"]);
    expect(decisionsFor("ccu-denied-1")).toEqual(["Approve"]);
    expect(decisionsFor("ccu-approved-1")).toEqual([]);
  });

  // Make sure the reason an admin gave stays visible, since nothing else shows it.
  it("names the denial note on the account icon", () => {
    render(<ClusterAccountBadge name="ClusterA" account={byId("ccu-denied-1")} />);
    expect(
      screen.getByRole("img", { name: "Denied: Not on the approved collaborator list." }),
    ).toBeInTheDocument();
  });
});
