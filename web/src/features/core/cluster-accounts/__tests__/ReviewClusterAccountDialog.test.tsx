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

import { clusterAccounts } from "@/mocks/handlers/clusters";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/features/core/clusters/components/ClusterName", () => ({
  ClusterName: () => "ClusterA",
}));

import { ReviewClusterAccountDialog } from "../components/ReviewClusterAccountDialog";

const [pending = null] = clusterAccounts;

describe("ReviewClusterAccountDialog", () => {
  // Make sure the denial passes the typed note through, since it ends up in
  // the audit log.
  it("sends the note with a denial", () => {
    const onConfirm = vi.fn();
    render(
      <ReviewClusterAccountDialog
        account={pending}
        action="deny"
        onOpenChange={vi.fn()}
        onConfirm={onConfirm}
        isPending={false}
        finalFocus={() => null}
      />,
    );
    fireEvent.change(screen.getByLabelText(/Note/), { target: { value: "  Not on the list " } });
    fireEvent.click(screen.getByRole("button", { name: "Deny" }));
    expect(onConfirm).toHaveBeenCalledWith("Not on the list");
  });

  // Make sure a denial cannot be sent without a note, since the backend
  // refuses it and the note is passed on to whoever requested the account.
  it("keeps Deny disabled until a note is typed", () => {
    render(
      <ReviewClusterAccountDialog
        account={pending}
        action="deny"
        onOpenChange={vi.fn()}
        onConfirm={vi.fn()}
        isPending={false}
        finalFocus={() => null}
      />,
    );
    expect(screen.getByRole("button", { name: "Deny" })).toHaveAttribute("aria-disabled", "true");
    fireEvent.change(screen.getByLabelText(/Note/), { target: { value: "   " } });
    expect(screen.getByRole("button", { name: "Deny" })).toHaveAttribute("aria-disabled", "true");
    fireEvent.change(screen.getByLabelText(/Note/), { target: { value: "Not on the list" } });
    expect(screen.getByRole("button", { name: "Deny" })).toHaveAttribute("aria-disabled", "false");
  });

  // Make sure the approval names the account the admin is about to create, so
  // a wrong row is caught before it goes to the cluster.
  it("restates the account before approving", () => {
    render(
      <ReviewClusterAccountDialog
        account={pending}
        action="approve"
        onOpenChange={vi.fn()}
        onConfirm={vi.fn()}
        isPending={false}
        finalFocus={() => null}
      />,
    );
    expect(screen.getByText("jdoe")).toBeInTheDocument();
    expect(screen.getByText("ClusterA")).toBeInTheDocument();
    expect(screen.getByText(/recorded as the approver/)).toBeInTheDocument();
    expect(screen.queryByLabelText(/Note/)).toBeNull();
  });
});
