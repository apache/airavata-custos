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
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import fixture from "../__fixtures__/cluster-accounts.json";
import { ClusterAccountsTable, clusterAccountStatus } from "../components/ClusterAccountsTable";

const rows = zClusterAccountResponse.array().parse(fixture);

function byId(id: string): ClusterAccountResponse {
  const row = rows.find((r) => r.id === id);
  if (!row) throw new Error(`no fixture ${id}`);
  return row;
}

function rowOf(text: string): HTMLTableRowElement {
  const row = screen.getByText(text).closest("tr");
  if (!row) throw new Error(`no row for ${text}`);
  return row;
}

function renderTable(canReview: boolean, shown: ClusterAccountResponse[] = rows) {
  const onApprove = vi.fn();
  const onDeny = vi.fn();
  render(
    <ClusterAccountsTable
      rows={shown}
      isLoading={false}
      error={null}
      canReview={canReview}
      onApprove={onApprove}
      onDeny={onDeny}
      page={1}
      pageSize={25}
      total={shown.length}
      onPageChange={vi.fn()}
      emptyHeading="Nothing waiting for approval."
    />,
  );
  return { onApprove, onDeny };
}

describe("clusterAccountStatus", () => {
  // Make sure an approved account reads as provisioning until it exists on
  // the cluster, so the admin can tell approval from a working login.
  it("tells approved from active by provisioned_at", () => {
    expect(clusterAccountStatus(byId("ccu-approved-1")).label).toBe("Approved · provisioning");
    expect(clusterAccountStatus(byId("ccu-active-1")).label).toBe("Active");
    expect(clusterAccountStatus(byId("ccu-pending-1")).label).toBe("Pending approval");
    expect(clusterAccountStatus(byId("ccu-denied-1")).label).toBe("Denied");
  });
});

describe("ClusterAccountsTable", () => {
  // Make sure a pending account offers both decisions, a denied one can only
  // be approved, and an approved one cannot be reviewed again.
  it("offers the decisions each status allows", () => {
    renderTable(true);
    const pending = rowOf("jdoe");
    expect(pending.querySelector("button")).not.toBeNull();
    expect(pending.textContent).toContain("Deny");
    expect(pending.textContent).toContain("Approve");

    const denied = rowOf("sokafor");
    expect(denied.textContent).toContain("Approve");
    expect(denied.textContent).not.toContain("Deny");

    const approved = rowOf("wzhang");
    expect(approved.querySelector("button")).toBeNull();
  });

  // Make sure a reader without the write privilege sees the list but no
  // decision buttons.
  it("hides the decisions without the review privilege", () => {
    renderTable(false);
    expect(screen.queryByRole("button", { name: "Approve" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Deny" })).toBeNull();
    expect(screen.getByText("jdoe")).toBeInTheDocument();
  });

  it("hands the clicked account to the approve and deny callbacks", () => {
    const { onApprove, onDeny } = renderTable(true, [byId("ccu-pending-1")]);
    fireEvent.click(screen.getByRole("button", { name: "Approve" }));
    expect(onApprove).toHaveBeenCalledWith(expect.objectContaining({ id: "ccu-pending-1" }));
    fireEvent.click(screen.getByRole("button", { name: "Deny" }));
    expect(onDeny).toHaveBeenCalledWith(expect.objectContaining({ id: "ccu-pending-1" }));
  });

  // Make sure the reason an admin gave is visible on the denied row, since it
  // is the only place the portal shows it.
  it("shows the denial note on the row", () => {
    renderTable(false, [byId("ccu-denied-1")]);
    expect(screen.getByText("Not on the approved collaborator list.")).toBeInTheDocument();
  });
});
