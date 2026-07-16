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

import { act, fireEvent, render, screen, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { AccessRequest } from "../schemas";

let rows: AccessRequest[] = [];
const decideMutate = vi.fn();
const decideMutateAsync = vi.fn();

vi.mock("../queries", () => ({
  useAccessRequests: () => ({ data: rows, isLoading: false, error: null, refetch: vi.fn() }),
  useDecideAccessRequest: () => ({ mutate: decideMutate, mutateAsync: decideMutateAsync }),
}));

vi.mock("sonner", () => ({
  toast: Object.assign(vi.fn(), { error: vi.fn(), success: vi.fn() }),
}));

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(""),
}));

// The drawer pulls the allocation feature; keep the queue tests self-contained.
vi.mock("@/features/core/allocations/queries", () => ({
  useAllocation: () => ({ data: undefined }),
}));

import { toast } from "sonner";
import { AccessRequestsQueue } from "../components/AccessRequestsQueue";

function request(id: string, name: string, overrides: Partial<AccessRequest> = {}): AccessRequest {
  return {
    id,
    oidc_sub: `sub-${id}`,
    email: `${id}@example.edu`,
    name,
    institution: "Example University",
    event_code: "PEARC26",
    reason: "",
    status: "PENDING",
    approver_id: "",
    deny_reason: "",
    expires_at: null,
    created_user_id: "",
    timestamp: "2026-07-10T14:05:00Z",
    decided_at: null,
    allocation_id: "alloc-1",
    ...overrides,
  };
}

function rowFor(name: string): HTMLElement {
  const cell = screen.getByText(name);
  const tr = cell.closest("tr");
  if (!tr) throw new Error(`no row for ${name}`);
  return tr;
}

beforeEach(() => {
  rows = [
    request("areq-1", "Amara Osei"),
    request("areq-2", "Li Wei"),
    request("areq-3", "Marcus Rivera", {
      status: "APPROVED",
      approver_id: "user-admin-001",
      decided_at: "2026-07-11T08:00:00Z",
      expires_at: "2026-08-09T00:00:00Z",
    }),
  ];
  decideMutate.mockReset();
  decideMutateAsync
    .mockReset()
    .mockResolvedValue(request("areq-1", "Amara Osei", { status: "APPROVED" }));
  vi.mocked(toast).mockClear();
  vi.mocked(toast.error).mockClear();
});

afterEach(() => {
  vi.useRealTimers();
});

describe("<AccessRequestsQueue />", () => {
  it("selects only pending rows and shows the bulk bar count", () => {
    render(<AccessRequestsQueue />);

    // Decided rows have no checkbox.
    expect(within(rowFor("Marcus Rivera")).queryByRole("checkbox")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("checkbox", { name: "Select all pending" }));
    expect(screen.getByRole("button", { name: "Approve 2 selected" })).toBeInTheDocument();

    fireEvent.click(screen.getByRole("checkbox", { name: "Select Li Wei" }));
    expect(screen.getByRole("button", { name: "Approve 1 selected" })).toBeInTheDocument();
  });

  it("does not fire the PUT when undo is clicked inside the window", () => {
    vi.useFakeTimers();
    render(<AccessRequestsQueue />);

    fireEvent.click(within(rowFor("Amara Osei")).getByRole("button", { name: "Approve" }));
    expect(within(rowFor("Amara Osei")).getByText("Approved")).toBeInTheDocument();
    expect(decideMutate).not.toHaveBeenCalled();

    const [, opts] = vi.mocked(toast).mock.calls[0] as unknown as [
      string,
      { action: { onClick: () => void } },
    ];
    act(() => {
      opts.action.onClick();
    });
    act(() => {
      vi.advanceTimersByTime(6000);
    });

    expect(decideMutate).not.toHaveBeenCalled();
    expect(within(rowFor("Amara Osei")).getByText("Pending")).toBeInTheDocument();
  });

  it("fires the PUT only after the undo window expires", () => {
    vi.useFakeTimers();
    render(<AccessRequestsQueue />);

    fireEvent.click(within(rowFor("Amara Osei")).getByRole("button", { name: "Approve" }));
    act(() => {
      vi.advanceTimersByTime(4999);
    });
    expect(decideMutate).not.toHaveBeenCalled();

    act(() => {
      vi.advanceTimersByTime(1);
    });
    expect(decideMutate).toHaveBeenCalledTimes(1);
    expect(decideMutate.mock.calls[0]?.[0]).toEqual({
      id: "areq-1",
      body: { status: "APPROVED" },
    });
  });

  it("round-trips the deny reason through the popover", () => {
    render(<AccessRequestsQueue />);

    fireEvent.click(within(rowFor("Li Wei")).getByRole("button", { name: "Deny" }));
    fireEvent.change(screen.getByLabelText("Reason (optional)"), {
      target: { value: "Not a tutorial attendee" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Deny request" }));

    expect(decideMutate).toHaveBeenCalledTimes(1);
    expect(decideMutate.mock.calls[0]?.[0]).toEqual({
      id: "areq-2",
      body: { status: "DENIED", deny_reason: "Not a tutorial attendee" },
    });
  });

  it("bulk approve issues one PUT per selected row", async () => {
    render(<AccessRequestsQueue />);

    fireEvent.click(screen.getByRole("checkbox", { name: "Select all pending" }));
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Approve 2 selected" }));
    });

    expect(decideMutateAsync).toHaveBeenCalledTimes(2);
    expect(decideMutateAsync.mock.calls.map((c) => c[0]?.id)).toEqual(["areq-1", "areq-2"]);
  });
});
