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

import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const denied = vi.hoisted(() => ({ subjects: [] as string[] }));
const mutateAsync = vi.hoisted(() => vi.fn());

const ALLOCATION_ID = "01000001-0000-4000-8000-000001000001";

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({
    can: (_action: string, subject: string) => !denied.subjects.includes(subject),
  }),
}));
vi.mock("@/features/core/clusters/queries", () => ({ useClusters: () => ({ data: [] }) }));
vi.mock("@/features/core/allocations/queries", () => ({
  useAllocations: () => ({
    data: { items: [{ id: ALLOCATION_ID, name: "Genomic Sequencing - GPU Pool" }], total: 1 },
  }),
}));
vi.mock("@/features/core/users/queries", () => ({
  useCreateUser: () => ({ mutateAsync, isPending: false }),
}));

import { CreateUserDialog } from "../CreateUserDialog";

async function openDialog() {
  render(<CreateUserDialog />);
  fireEvent.click(screen.getByRole("button", { name: "Create user" }));
  fireEvent.change(await screen.findByLabelText("Email"), {
    target: { value: "jkellett@ncsa.illinois.edu" },
  });
  fireEvent.change(screen.getByLabelText("First name"), { target: { value: "Jane" } });
}

const submit = () =>
  fireEvent.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Create user" }));

describe("<CreateUserDialog />", () => {
  beforeEach(() => {
    denied.subjects = [];
    mutateAsync.mockReset().mockResolvedValue({});
  });

  it("offers the admin flags only to a roles:manage holder", async () => {
    for (const subjects of [[], ["Role"]]) {
      denied.subjects = subjects;
      const { unmount } = render(<CreateUserDialog />);
      fireEvent.click(screen.getByRole("button", { name: "Create user" }));
      expect(await screen.findByLabelText("Email")).toBeInTheDocument();
      expect(Boolean(screen.queryByLabelText("Portal admin"))).toBe(subjects.length === 0);
      unmount();
    }
  });

  it("leaves a researcher's username and allocation to the backend when blank", async () => {
    await openDialog();
    submit();
    expect(mutateAsync).toHaveBeenCalledWith({
      portal_admin: false,
      cluster_admin: false,
      email: "jkellett@ncsa.illinois.edu",
      first_name: "Jane",
      last_name: "",
      username: undefined,
      allocation_id: undefined,
      compute_cluster_id: undefined,
    });
  });

  it("joins a researcher to the picked allocation, which fixes the cluster", async () => {
    await openDialog();
    fireEvent.change(screen.getByLabelText("Cluster username"), { target: { value: "jkellett" } });
    fireEvent.change(screen.getByLabelText("Allocation"), { target: { value: ALLOCATION_ID } });
    submit();
    expect(mutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        username: "jkellett",
        allocation_id: ALLOCATION_ID,
        compute_cluster_id: undefined,
      }),
    );
  });

  it("onboards a portal-only admin without a cluster account or allocation", async () => {
    await openDialog();
    fireEvent.click(screen.getByLabelText("Portal admin"));
    expect(screen.queryByLabelText("Cluster username")).not.toBeInTheDocument();
    expect(screen.queryByLabelText("Allocation")).not.toBeInTheDocument();
    submit();
    expect(mutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        portal_admin: true,
        cluster_admin: false,
        allocation_id: undefined,
      }),
    );
  });

  it("keeps a cluster admin's username but no allocation", async () => {
    await openDialog();
    fireEvent.click(screen.getByLabelText("Cluster admin"));
    expect(screen.queryByLabelText("Allocation")).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Cluster username"), { target: { value: "jkadmin" } });
    submit();
    expect(mutateAsync).toHaveBeenCalledWith(
      expect.objectContaining({
        cluster_admin: true,
        username: "jkadmin",
        allocation_id: undefined,
      }),
    );
  });
});
