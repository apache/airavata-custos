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

import type { ClusterAccountResponse } from "@/generated/core/types.gen";
import { clusterAccounts } from "@/mocks/handlers/clusters";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ClusterAccountsTable, clusterAccountStatus } from "../components/ClusterAccountsTable";

const rows = clusterAccounts;
const byId = (id: string) => rows.find((r) => r.id === id) ?? expect.unreachable(id);

function renderTable(shown: ClusterAccountResponse[] = rows) {
  render(
    <ClusterAccountsTable
      rows={shown}
      isLoading={false}
      error={null}
      page={1}
      pageSize={25}
      total={shown.length}
      onPageChange={vi.fn()}
      emptyHeading="Nothing waiting for approval"
    />,
  );
}

describe("clusterAccountStatus", () => {
  // Make sure an approved account reads as provisioning until it exists on
  // the cluster, so the admin can tell approval from a working login.
  it("tells approved from active by provisioned_at", () => {
    expect(clusterAccountStatus(byId("03000006-0000-4000-8000-000003000006")).label).toBe(
      "Approved · provisioning",
    );
    expect(clusterAccountStatus(byId("03000005-0000-4000-8000-000003000005")).label).toBe("Active");
    expect(clusterAccountStatus(byId("03000011-0000-4000-8000-000003000011")).label).toBe(
      "Pending approval",
    );
    expect(clusterAccountStatus(byId("03000010-0000-4000-8000-000003000010")).label).toBe("Denied");
  });
});

describe("ClusterAccountsTable", () => {
  // Make sure the reason an admin gave is visible on the denied row, since it
  // is the only place the portal shows it.
  it("shows the denial note on the row", () => {
    renderTable([byId("03000010-0000-4000-8000-000003000010")]);
    expect(screen.getByText("Not on the approved collaborator list.")).toBeInTheDocument();
  });
});
