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

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { toast } from "sonner";
import { describe, expect, it, vi } from "vitest";

vi.mock("sonner", () => ({ toast: { success: vi.fn() } }));
vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({ can: (_: string, subject: string) => subject === "User" }),
}));
vi.mock("@/features/core/users/components/UserPicker", () => ({
  UserName: ({ id }: { id: string }) => (id === "u-1" ? "Pat Lee" : id),
}));

const listed = vi.hoisted(() => [] as unknown[]);
const decide = vi.hoisted(() => vi.fn(async () => ({})));

vi.mock("../queries", () => ({
  CHANGE_REQUEST_LIMIT: 200,
  useAllocation: () => ({ data: { name: "Genomic GPU Pool" } }),
  useAllocationMembers: () => ({}),
  useDecideChangeRequest: () => ({ mutateAsync: decide }),
  useChangeRequests: (status: unknown) => {
    listed.push(status);
    return {
      data: [
        {
          id: "cr-1",
          compute_allocation_id: "01000005-0000-4000-8000-000001000005",
          requester_id: "u-1",
          requested_su_amount: 5000,
          requested_status: "INACTIVE",
          change_status: "PENDING",
          timestamp: "2026-10-01T12:00:00Z",
        },
      ],
      isLoading: false,
    };
  },
}));

import { ChangeRequestApproverQueue } from "../components/ChangeRequestApproverQueue";

const APPROVE = /^Approve change request for 5,000 SU, Inactive, submitted /;

function renderQueue() {
  return render(
    <QueryClientProvider client={new QueryClient()}>
      <ChangeRequestApproverQueue canApprove approverId="admin" />
    </QueryClientProvider>,
  );
}

describe("<ChangeRequestApproverQueue />", () => {
  it("names the allocation and requester beside the requested SU total and status", () => {
    renderQueue();
    const row = screen.getByRole("button", { name: APPROVE }).closest("tr") as HTMLElement;
    expect(within(row).getByRole("link", { name: "Genomic GPU Pool" })).toHaveAttribute(
      "href",
      "/allocations/01000005-0000-4000-8000-000001000005",
    );
    expect(within(row).getByText("Pat Lee")).toBeInTheDocument();
    expect(within(row).getByText("5,000")).toBeInTheDocument();
    expect(within(row).getByText("Inactive")).toBeInTheDocument();
  });

  it("decides as the signed-in approver and names the request in the buttons", async () => {
    renderQueue();
    // The short id keeps two requests with the same change and minute apart.
    expect(screen.getByRole("button", { name: APPROVE })).toHaveAccessibleName(/, request cr-1$/);
    fireEvent.click(screen.getByRole("button", { name: APPROVE }));
    await waitFor(() => expect(toast.success).toHaveBeenCalledWith("Approved the change request"));
    expect(decide).toHaveBeenCalledWith({
      id: "cr-1",
      change_status: "APPROVED",
      approver_id: "admin",
    });
  });

  it("reads the status filter from the URL", async () => {
    window.history.replaceState(null, "", "/change-requests?status=APPROVED");
    renderQueue();
    await waitFor(() => expect(screen.getByLabelText("Status")).toHaveValue("APPROVED"));
    expect(listed.at(-1)).toBe("APPROVED");
    window.history.replaceState(null, "", "/");
  });

  it("drops a status it does not apply from the URL", async () => {
    window.history.replaceState(null, "", "/change-requests?status=bogus");
    renderQueue();
    await waitFor(() => expect(window.location.search).toBe(""));
    expect(listed.at(-1)).toBe("PENDING");
    window.history.replaceState(null, "", "/");
  });
});
