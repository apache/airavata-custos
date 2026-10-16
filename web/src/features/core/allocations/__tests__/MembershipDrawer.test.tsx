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

const update = { mutateAsync: vi.fn(async () => ({})), isPending: false };
const idle = { mutateAsync: vi.fn(), isPending: false };
const updateOverride = { mutateAsync: vi.fn(async () => ({})), isPending: false };
const membership = {
  id: "m-1",
  compute_allocation_id: "01000005-0000-4000-8000-000001000005",
  user_id: "u-1",
  membership_status: "ACTIVE",
  start_time: "2026-01-01T00:00:00Z",
  end_time: "2026-12-31T00:00:00Z",
};
const overrides = vi.hoisted(() => ({ data: [] as unknown[] }));
const fetched = vi.hoisted(() => ({ error: null as Error | null }));
const refreshMembers = vi.hoisted(() => vi.fn());
vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => true }) }));
vi.mock("@/features/core/resources/queries", () => ({
  useResources: () => ({ data: [], isSuccess: true }),
}));
vi.mock("../queries", () => ({
  useRefreshAllocationMembers: () => refreshMembers,
  // The refetch on open has landed.
  useMembership: () => ({
    data: fetched.error ? undefined : membership,
    isLoading: false,
    error: fetched.error,
    refetch: vi.fn(),
    dataUpdatedAt: Number.MAX_SAFE_INTEGER,
    errorUpdatedAt: Number.MAX_SAFE_INTEGER,
  }),
  useUpdateMembership: () => update,
  useSetMembershipStatus: () => idle,
  useMembershipOverrides: () => ({
    data: overrides.data,
    isLoading: false,
    error: null,
    refetch: vi.fn(),
    dataUpdatedAt: Number.MAX_SAFE_INTEGER,
  }),
  useAllocationResources: () => ({ data: [], isSuccess: true }),
  useCreateOverride: () => idle,
  useDeleteOverride: () => idle,
  useUpdateOverride: () => updateOverride,
}));

import { ApiError } from "@/shared/api/client";
import { MembershipDrawer } from "../components/MembershipDrawer";

const props = {
  allocationId: "01000005-0000-4000-8000-000001000005",
  membershipId: "m-1",
  name: "Paula Ivey",
  canManage: true,
  onClose: () => {},
};

describe("<MembershipDrawer />", () => {
  it("sends the membership with the new dates, leaving its status to the server", () => {
    render(<MembershipDrawer {...props} />);
    fireEvent.change(screen.getByLabelText("End"), { target: { value: "2027-06-30" } });
    fireEvent.submit(screen.getByLabelText("End").closest("form") as HTMLFormElement);
    expect(update.mutateAsync).toHaveBeenCalledWith({
      path: { id: "m-1" },
      body: {
        compute_allocation_id: "01000005-0000-4000-8000-000001000005",
        user_id: "u-1",
        start_time: "2026-01-01T00:00:00Z",
        end_time: "2027-06-30T00:00:00Z",
      },
    });
  });

  it("sends the whole override with the new grant", () => {
    const override = {
      id: "o-1",
      compute_allocation_membership_id: "m-1",
      compute_allocation_resource_id: "r-1",
      override_resource_amount: 4,
      override_resource_time: 60,
    };
    overrides.data = [override];
    render(<MembershipDrawer {...props} />);
    fireEvent.click(screen.getByRole("button", { name: "Edit override for Unknown resource" }));
    fireEvent.change(screen.getByLabelText("Resource amount"), { target: { value: "8" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(updateOverride.mutateAsync).toHaveBeenCalledWith({
      path: { id: "o-1" },
      body: { ...override, override_resource_amount: 8 },
    });
    overrides.data = [];
  });

  it("says an override another tab deleted no longer exists, with Save disabled", () => {
    overrides.data = [{ id: "o-1", override_resource_amount: 4, override_resource_time: 60 }];
    const drawer = () => <MembershipDrawer {...props} />;
    const { rerender } = render(drawer());
    fireEvent.click(screen.getByRole("button", { name: "Edit override for Unknown resource" }));
    overrides.data = [];
    rerender(drawer());
    expect(screen.getByText("This override no longer exists")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Save" })).toHaveAttribute("aria-disabled", "true");
    expect(screen.queryByRole("button", { name: "Try again" })).not.toBeInTheDocument();
  });

  it("refreshes the members list when the membership was deleted elsewhere", () => {
    fetched.error = new ApiError(404, "/", { error: "record not found" });
    render(<MembershipDrawer {...props} />);
    expect(screen.getByText("This membership no longer exists")).toBeInTheDocument();
    expect(refreshMembers).toHaveBeenCalled();
    fetched.error = null;
  });

  it("shows another allocation's membership as gone, without controls", () => {
    render(<MembershipDrawer {...props} allocationId="01000006-0000-4000-8000-000001000006" />);
    expect(screen.getByText("This membership no longer exists")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Deactivate" })).not.toBeInTheDocument();
  });
});
