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
import type { ComputeAllocation, ComputeAllocationChangeRequest } from "@/generated/core/types.gen";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

const edit = vi.hoisted(() => vi.fn(() => Promise.resolve()));
const submit = vi.hoisted(() => vi.fn(() => Promise.resolve()));
// The request as the server holds it when the drawer opens, and when Save reads it again.
const fresh = vi.hoisted(() => ({ data: undefined as unknown, saved: undefined as unknown }));
vi.mock("../queries", () => ({
  alreadyDecided: (status: string) => new Error(status),
  useRefreshChangeRequests: () => vi.fn(),
  useRefreshDecidedChangeRequest: () => vi.fn(),
  useSubmitChangeRequest: () => ({ mutateAsync: submit, isPending: false }),
  useUpdateChangeRequest: () => ({ mutateAsync: edit, isPending: false }),
  useChangeRequest: () => ({
    data: fresh.data,
    refetch: async () => ({ data: fresh.saved ?? fresh.data ?? pending }),
    dataUpdatedAt: Number.MAX_SAFE_INTEGER,
    errorUpdatedAt: 0,
    error: null,
  }),
}));
vi.mock("@/features/core/identity/queries", () => ({
  useCurrentUser: () => ({ user: { id: "u-me" } }),
}));
vi.mock("@/shared/ui/SideDrawer", () => ({
  SideDrawer: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));

import { ChangeRequestSubmitDrawer } from "../components/ChangeRequestSubmitDrawer";

const allocation: ComputeAllocation = {
  id: "01000005-0000-4000-8000-000001000005",
  project_id: "19000001-0000-4000-8000-000019000001",
  name: "Test Allocation",
  status: "ACTIVE",
  compute_cluster_id: "04000001-0000-4000-8000-000004000001",
  initial_su_amount: 1000,
  start_time: "2026-04-01T00:00:00.000Z",
  end_time: "2027-03-31T00:00:00.000Z",
  posix_group: null,
};

const pending: ComputeAllocationChangeRequest = {
  id: "cr-1",
  compute_allocation_id: "01000005-0000-4000-8000-000001000005",
  requester_id: "u-2",
  requested_su_amount: 1500,
  requested_status: "ACTIVE",
  change_status: "PENDING",
  reason: "Extending the allocation for the next phase.",
  timestamp: "2026-10-01T12:00:00Z",
};

// The submit button ignores a second press until the first has settled.
const settle = () => new Promise((resolve) => setTimeout(resolve, 60));

const drawer = (request?: ComputeAllocationChangeRequest, onOpenChange = vi.fn()) => (
  <ChangeRequestSubmitDrawer
    open
    onOpenChange={onOpenChange}
    allocation={allocation}
    request={request}
  />
);

describe("<ChangeRequestSubmitDrawer />", () => {
  it("files a new request as the signed-in user with the SU total and status asked for", async () => {
    render(drawer());
    expect(screen.getByLabelText("Requested status")).toHaveValue("ACTIVE");
    fireEvent.change(screen.getByLabelText("Requested SUs"), { target: { value: "2000" } });
    fireEvent.change(screen.getByLabelText(/Reason/), {
      target: { value: "Need more SUs for the final modeling run." },
    });
    fireEvent.click(screen.getByRole("button", { name: "Submit request" }));
    await waitFor(() =>
      expect(submit).toHaveBeenCalledWith({
        body: {
          compute_allocation_id: "01000005-0000-4000-8000-000001000005",
          requester_id: "u-me",
          requested_su_amount: 2000,
          requested_status: "ACTIVE",
          reason: "Need more SUs for the final modeling run.",
        },
      }),
    );
  });

  it("prefills from the fresh request, sends nothing untouched and never sends a decision", async () => {
    fresh.data = { ...pending, requested_su_amount: 1700 };
    const onOpenChange = vi.fn();
    render(drawer(pending, onOpenChange));
    expect(screen.getByLabelText("Requested SUs")).toHaveValue(1700);
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }));
    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
    expect(edit).not.toHaveBeenCalled();
    await settle();

    fireEvent.change(screen.getByLabelText("Requested SUs"), { target: { value: "1800" } });
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }));
    await waitFor(() =>
      expect(edit).toHaveBeenCalledWith({
        path: { id: "cr-1" },
        body: {
          compute_allocation_id: "01000005-0000-4000-8000-000001000005",
          requester_id: "u-2",
          timestamp: "2026-10-01T12:00:00Z",
          requested_su_amount: 1800,
          requested_status: "ACTIVE",
          reason: "Extending the allocation for the next phase.",
        },
      }),
    );
    await settle();

    edit.mockClear();
    fresh.saved = { ...pending, change_status: "APPROVED", approver_id: "u-3" };
    fireEvent.change(screen.getByLabelText("Requested SUs"), { target: { value: "1900" } });
    fireEvent.click(screen.getByRole("button", { name: "Save changes" }));
    await settle();
    expect(edit).not.toHaveBeenCalled();
    fresh.data = undefined;
    fresh.saved = undefined;
  });

  it("submits once on a double-click", async () => {
    edit.mockClear();
    render(drawer(pending));
    fireEvent.change(screen.getByLabelText("Requested SUs"), { target: { value: "1600" } });
    const save = screen.getByRole("button", { name: "Save changes" });
    fireEvent.click(save);
    fireEvent.click(save);
    await waitFor(() => expect(edit).toHaveBeenCalledTimes(1));
    await settle();
    expect(edit).toHaveBeenCalledTimes(1);
  });
});
