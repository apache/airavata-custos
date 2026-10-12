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

import { ApiError } from "@/shared/api/client";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(window.location.search),
}));
vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({ can: (_: string, subject: string) => subject !== "Organization" }),
}));
const confirm = vi.hoisted(() => vi.fn());
vi.mock("@/shared/ui/sonner", () => ({ confirmToast: confirm, toastOnSuccess: vi.fn() }));
const MISSING = "00000003-0000-4000-8000-000000000003";
const NO_MEMBERSHIPS = "00000004-0000-4000-8000-000000000004";
const NOT_VIRTUAL = "00000005-0000-4000-8000-000000000005";
const apiError = (status: number, error: string) => new ApiError(status, "/temp", { error }, error);
vi.mock("@/features/core/users/queries", () => ({
  useUser: (id?: string) =>
    id === MISSING
      ? { error: apiError(404, "not found") }
      : {
          data: id && {
            id,
            first_name: "Tia",
            last_name: "Temp",
            email: "tia@example.org",
            type: "VIRTUAL",
          },
        },
}));
vi.mock("@/features/connectors/temp-account/queries", () => ({
  useRemoveTempAccount: () => ({}),
  useAssignTempAllocation: () => ({}),
  useUpdateTempAllocation: () => ({}),
  useTempMemberships: (userId?: string) =>
    userId === MISSING || userId === NO_MEMBERSHIPS
      ? { error: apiError(404, "membership not found for user") }
      : userId === NOT_VIRTUAL
        ? { error: apiError(400, "User type must be VIRTUAL") }
        : { data: userId ? [] : undefined, error: null },
}));
vi.mock("@/features/core/allocations/queries", () => ({
  useAllocation: () => ({}),
  useAllocations: () => ({}),
}));
vi.mock("@/features/core/allocations/components/AllocationName", () => ({
  AllocationName: ({ id }: { id: string }) => <>{id}</>,
}));
vi.mock("@/features/connectors/temp-account/components/CreateTempAccountDialog", () => ({
  CreateTempAccountDialog: () => <button type="button">Create</button>,
}));

import { TempAccountsContainer } from "../TempAccountsContainer";

describe("<TempAccountsContainer />", () => {
  it("restores the opened user from ?user= and offers create without Organization read", () => {
    window.history.replaceState(
      null,
      "",
      "/admin/temp-accounts?user=00000002-0000-4000-8000-000000000002",
    );
    render(<TempAccountsContainer />);
    expect(screen.getByLabelText("User ID")).toHaveValue("00000002-0000-4000-8000-000000000002");
    expect(screen.getByRole("button", { name: "Remove account" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create" })).toBeInTheDocument();
    window.history.replaceState(null, "", "/admin/temp-accounts");
  });

  it("offers no account actions for an unknown user id", () => {
    render(<TempAccountsContainer />);
    fireEvent.change(screen.getByLabelText("User ID"), {
      target: { value: MISSING },
    });
    fireEvent.click(screen.getByRole("button", { name: "Open" }));
    expect(screen.getByRole("heading", { name: "No user has that ID" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Assign allocation" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Remove account" })).toBeNull();
  });

  it("names the temporary account in the remove confirmation", () => {
    render(<TempAccountsContainer />);
    fireEvent.change(screen.getByLabelText("User ID"), {
      target: { value: "00000002-0000-4000-8000-000000000002" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Open" }));
    fireEvent.click(screen.getByRole("button", { name: "Remove account" }));
    expect(confirm).toHaveBeenCalledWith(
      "Remove the temporary account of Tia Temp?",
      "Remove",
      expect.any(Function),
    );
  });

  it("offers Assign to a temporary account the connector finds no memberships for", () => {
    render(<TempAccountsContainer />);
    fireEvent.change(screen.getByLabelText("User ID"), { target: { value: NO_MEMBERSHIPS } });
    fireEvent.click(screen.getByRole("button", { name: "Open" }));
    expect(screen.getByRole("heading", { name: "No allocation memberships" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Assign allocation" })).toBeInTheDocument();
  });

  it("takes a non-temporary user from the connector's refusal", () => {
    render(<TempAccountsContainer />);
    fireEvent.change(screen.getByLabelText("User ID"), { target: { value: NOT_VIRTUAL } });
    fireEvent.click(screen.getByRole("button", { name: "Open" }));
    expect(screen.getByRole("heading", { name: "Not a temporary account" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Remove account" })).toBeNull();
  });
});
