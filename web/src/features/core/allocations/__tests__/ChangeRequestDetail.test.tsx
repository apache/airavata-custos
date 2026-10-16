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

const subReads = vi.hoisted(() => ({
  events: vi.fn(() => ({})),
  latest: vi.fn(() => ({})),
  request: vi.fn(),
  allocation: vi.fn((): { isLoading: boolean; data?: { name: string } } => ({ isLoading: true })),
  refreshQueue: vi.fn(),
}));

vi.mock("../queries", async () => {
  const { ApiError } = await import("@/shared/api/client");
  const none = () => ({});
  return {
    useAllocation: subReads.allocation,
    useAllocationMembers: none,
    useChangeRequest: subReads.request.mockImplementation(() => ({
      isLoading: false,
      error: new ApiError(404, "/compute-allocation-change-requests/missing", {
        error: "not found",
      }),
    })),
    useChangeRequestEvent: none,
    useChangeRequestEvents: subReads.events,
    useCreateChangeRequestEvent: none,
    useDecideChangeRequest: none,
    useDeleteChangeRequest: none,
    useDeleteChangeRequestEvent: none,
    useLatestChangeRequestEvent: subReads.latest,
    useRefreshChangeRequests: () => subReads.refreshQueue,
  };
});
vi.mock("@/features/core/identity/queries", () => ({ useCurrentUser: () => ({}) }));
vi.mock("@/features/core/users/components/UserPicker", () => ({ UserName: () => null }));
vi.mock("../components/ChangeRequestSubmitDrawer", () => ({
  ChangeRequestSubmitDrawer: () => null,
}));
vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => true }) }));
vi.mock("next/navigation", () => ({
  useRouter: () => ({}),
  useSearchParams: () => new URLSearchParams(),
  usePathname: () => "/change-requests/missing",
}));

import { ChangeRequestDetail } from "../components/ChangeRequestDetail";

describe("<ChangeRequestDetail />", () => {
  it("shows a missing change request as unavailable, refreshes the queue and skips its events", () => {
    render(<ChangeRequestDetail changeRequestId="missing" />);
    expect(screen.getByText("Not found, or you don't have access.")).toBeInTheDocument();
    expect(subReads.refreshQueue).toHaveBeenCalledTimes(1);
    expect(subReads.events).toHaveBeenCalledWith(undefined);
    expect(subReads.latest).toHaveBeenCalledWith(undefined);
  });

  it("keeps the route heading and announces the wait while the request loads", () => {
    subReads.request.mockReturnValue({ isLoading: true });
    render(<ChangeRequestDetail changeRequestId="cr-1" />);
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent("Change requests");
    expect(screen.getByRole("status", { name: "Loading" })).toBeInTheDocument();
  });

  it("titles the request by its allocation's name, never the raw id", () => {
    subReads.request.mockReturnValue({
      isLoading: false,
      data: {
        id: "cr-1",
        compute_allocation_id: "01000005-0000-4000-8000-000001000005",
        change_status: "APPROVED",
        requested_status: "ACTIVE",
      },
    });
    subReads.allocation.mockReturnValue({ isLoading: false, data: { name: "Protein folding" } });
    render(<ChangeRequestDetail changeRequestId="cr-1" />);
    expect(screen.getByRole("heading", { level: 1 })).toHaveTextContent(
      "Change request for Protein folding",
    );
  });

  it("lists timeline events inside the Timeline region", () => {
    subReads.request.mockReturnValue({
      isLoading: false,
      data: {
        id: "cr-1",
        compute_allocation_id: "01000005-0000-4000-8000-000001000005",
        change_status: "PENDING",
        requested_status: "ACTIVE",
      },
    });
    subReads.events.mockReturnValue({
      data: [
        {
          id: "evt-1",
          event_type: "CREATED",
          timestamp: "2026-06-01T10:00:00Z",
        },
      ],
    });
    render(<ChangeRequestDetail changeRequestId="cr-1" />);
    const event = screen.getByRole("button", { name: /Created/ });
    // The drawer returns focus here once the event's own row is deleted.
    expect(screen.getByRole("region", { name: "Timeline" })).toContainElement(event);
  });

  it("refuses an event type the server records up front", () => {
    subReads.request.mockReturnValue({
      isLoading: false,
      data: {
        id: "cr-1",
        compute_allocation_id: "01000005-0000-4000-8000-000001000005",
        change_status: "PENDING",
        requested_status: "ACTIVE",
      },
    });
    render(<ChangeRequestDetail changeRequestId="cr-1" />);
    const type = screen.getByLabelText<HTMLInputElement>("Event type");
    fireEvent.change(type, { target: { value: " approved " } });
    expect(type.validationMessage).toBe("APPROVED is recorded by the system.");
    fireEvent.change(type, { target: { value: "COMMENT" } });
    expect(type.validity.valid).toBe(true);
  });
});
