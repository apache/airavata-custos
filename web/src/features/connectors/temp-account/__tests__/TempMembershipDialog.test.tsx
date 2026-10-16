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

import type { ComputeAllocationMembership } from "@/generated/core/types.gen";
import type { UseQueryResult } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const membership: ComputeAllocationMembership = {
  id: "m-1",
  user_id: "u-1",
  compute_allocation_id: "01000005-0000-4000-8000-000001000005",
  start_time: "2026-07-31T20:00:00-04:00",
  end_time: "2027-08-01T00:00:00Z",
  membership_status: "ACTIVE",
};
const update = vi.hoisted(() => vi.fn(async () => ({})));
vi.mock("../queries", () => ({
  useAssignTempAllocation: () => ({ isPending: false }),
  useUpdateTempAllocation: () => ({ isPending: false, mutateAsync: update }),
}));
vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => false }) }));
vi.mock("@/features/core/allocations/queries", () => ({
  useAllocation: () => ({}),
  useAllocations: () => ({}),
}));
const memberships = {
  refetch: vi.fn(),
  dataUpdatedAt: Number.MAX_SAFE_INTEGER,
  data: [membership],
} as unknown as UseQueryResult<ComputeAllocationMembership[]>;

import { TempMembershipDialog } from "../components/TempMembershipDialog";

describe("<TempMembershipDialog />", () => {
  // The suite runs west of UTC, where slicing the offset timestamp would prefill Jul 31.
  it("prefills the UTC calendar day of a timestamp rendered in another offset", () => {
    render(
      <TempMembershipDialog
        userId="u-1"
        membership={membership}
        memberships={memberships}
        onClose={() => {}}
      />,
    );
    expect(screen.getByLabelText("Start")).toHaveValue("2026-08-01");
    expect(screen.getByLabelText("End")).toHaveValue("2027-08-01");
  });

  it("updates with the whole membership and the new window", async () => {
    render(
      <TempMembershipDialog
        userId="u-1"
        membership={membership}
        memberships={memberships}
        onClose={() => {}}
      />,
    );
    fireEvent.change(screen.getByLabelText("End"), { target: { value: "2027-09-01" } });
    fireEvent.click(screen.getByRole("button", { name: "Update" }));
    await waitFor(() =>
      expect(update).toHaveBeenCalledWith({
        id: "m-1",
        user_id: "u-1",
        compute_allocation_id: "01000005-0000-4000-8000-000001000005",
        start_time: "2026-08-01T00:00:00Z",
        end_time: "2027-09-01T00:00:00Z",
        membership_status: "ACTIVE",
      }),
    );
  });
});
