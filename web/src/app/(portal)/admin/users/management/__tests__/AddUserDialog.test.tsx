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
import { AddUserDialog } from "../AddUserDialog";

function renderDialog(overrides: Partial<Parameters<typeof AddUserDialog>[0]> = {}) {
  const onSubmit = vi.fn();
  render(
    <AddUserDialog
      open
      onOpenChange={() => {}}
      onSubmit={onSubmit}
      isPending={false}
      canCreateAdmins
      allocations={[
        {
          id: "alloc-1",
          project_id: "proj-1",
          name: "CHM-240013",
          status: "ACTIVE",
          compute_cluster_id: "cluster-1",
          initial_su_amount: 1000,
          start_time: "2026-01-01T00:00:00Z",
          end_time: "2027-01-01T00:00:00Z",
        },
      ]}
      {...overrides}
    />,
  );
  return { onSubmit };
}

describe("AddUserDialog", () => {
  it("submits a researcher without a username so the backend generates one", () => {
    const { onSubmit } = renderDialog();
    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "jkellett@ncsa.illinois.edu" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add user" }));
    expect(onSubmit).toHaveBeenCalledWith({
      email: "jkellett@ncsa.illinois.edu",
      first_name: "",
      last_name: "",
    });
  });

  it("sends a typed username as is", () => {
    const { onSubmit } = renderDialog();
    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "jkellett@ncsa.illinois.edu" },
    });
    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "jkellett" } });
    fireEvent.click(screen.getByRole("button", { name: "Add user" }));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ username: "jkellett" }),
    );
  });

  it("submits a portal admin without a cluster account", () => {
    const { onSubmit } = renderDialog();
    fireEvent.click(screen.getByRole("radio", { name: /Admin/ }));
    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "sandrade@sdsc.edu" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add admin user" }));
    expect(onSubmit).toHaveBeenCalledWith({
      email: "sandrade@sdsc.edu",
      first_name: "",
      last_name: "",
      portal_admin: true,
    });
  });

  it("disables the admin type for non super admins", () => {
    renderDialog({ canCreateAdmins: false });
    expect(screen.getByRole("radio", { name: /Admin/ })).toBeDisabled();
  });

  it("sends cluster_admin when cluster admin is ticked", () => {
    const { onSubmit } = renderDialog();
    fireEvent.click(screen.getByRole("radio", { name: /Admin/ }));

    const clusterAdmin = screen
      .getByText(/Admin access on the cluster itself/)
      .closest("label")
      ?.querySelector("input");
    expect(clusterAdmin).toBeEnabled();

    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "sandrade@sdsc.edu" },
    });
    fireEvent.click(clusterAdmin as HTMLInputElement);
    fireEvent.click(screen.getByRole("button", { name: "Add admin user" }));

    expect(onSubmit).toHaveBeenCalledWith({
      email: "sandrade@sdsc.edu",
      first_name: "",
      last_name: "",
      portal_admin: true,
      cluster_admin: true,
    });
  });

  it("leaves cluster_admin off when the box is not ticked", () => {
    const { onSubmit } = renderDialog();
    fireEvent.click(screen.getByRole("radio", { name: /Admin/ }));
    fireEvent.change(screen.getByLabelText("Email"), {
      target: { value: "sandrade@sdsc.edu" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add admin user" }));

    expect(onSubmit).toHaveBeenCalledWith(
      expect.not.objectContaining({ cluster_admin: expect.anything() }),
    );
  });
});
