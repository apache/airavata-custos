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

import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(window.location.search),
}));
vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => true }) }));
vi.mock("@/features/core/cluster-accounts/queries", () => ({
  useClusterAccounts: () => ({ data: { items: [], total: 0 } }),
  useApproveClusterAccount: () => ({}),
  useDenyClusterAccount: () => ({}),
  useClusterAccount: (id?: string) => ({
    data: id ? { id, user_id: "u-1", approval_status: "PENDING" } : undefined,
  }),
}));
vi.mock("@/features/core/cluster-accounts/components/ClusterAccountsTable", () => ({
  ClusterAccountsTable: ({ onRowClick }: { onRowClick: (row: { id: string }) => void }) => (
    <button
      type="button"
      onClick={() => onRowClick({ id: "00000004-0000-4000-8000-000000000004" })}
    >
      row
    </button>
  ),
}));
vi.mock("@/features/core/cluster-accounts/components/ClusterAccountDetail", () => ({
  ClusterAccountDetail: ({ accountId, canWrite }: { accountId: string; canWrite: boolean }) => (
    <p>
      detail {accountId}
      {canWrite ? " writable" : ""}
    </p>
  ),
  useAccountUser: () => undefined,
}));
vi.mock("@/features/core/cluster-accounts/components/CreateClusterAccountDialog", () => ({
  CreateClusterAccountDialog: () => null,
}));
vi.mock("@/features/core/cluster-accounts/components/ReviewClusterAccountDialog", () => ({
  ReviewClusterAccountDialog: () => null,
}));

import { ClusterAccountsContainer } from "../ClusterAccountsContainer";

describe("<ClusterAccountsContainer />", () => {
  it("keeps the open account in ?account=", () => {
    window.history.replaceState(null, "", "/admin/users/cluster-accounts");
    render(<ClusterAccountsContainer />);
    fireEvent.click(screen.getByRole("button", { name: "row" }));
    expect(
      screen.getByText("detail 00000004-0000-4000-8000-000000000004 writable"),
    ).toBeInTheDocument();
    expect(window.location.search).toBe("?account=00000004-0000-4000-8000-000000000004");

    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    expect(window.location.search).toBe("");
  });

  it("offers Approve and Deny on a pending account", () => {
    window.history.replaceState(
      null,
      "",
      "/admin/users/cluster-accounts?account=00000004-0000-4000-8000-000000000004",
    );
    render(<ClusterAccountsContainer />);
    expect(screen.getByRole("button", { name: "Approve" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Deny" })).toBeInTheDocument();
  });
});
