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
import { CreateOrganizationDialog } from "../components/CreateOrganizationDialog";

describe("CreateOrganizationDialog", () => {
  it("keeps submit disabled until a name is entered", () => {
    const onSubmit = vi.fn();
    render(
      <CreateOrganizationDialog
        open
        onOpenChange={() => {}}
        onSubmit={onSubmit}
        isPending={false}
      />,
    );
    const submit = screen.getByRole("button", { name: "Create organization" });
    expect(submit).toBeDisabled();
    fireEvent.change(screen.getByLabelText("Name"), {
      target: { value: "Georgia Institute of Technology" },
    });
    expect(submit).toBeEnabled();
  });

  it("submits name and optional originated_id", () => {
    const onSubmit = vi.fn();
    render(
      <CreateOrganizationDialog
        open
        onOpenChange={() => {}}
        onSubmit={onSubmit}
        isPending={false}
      />,
    );
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "Purdue University" } });
    fireEvent.change(screen.getByLabelText("Originated ID"), { target: { value: "PURDUE" } });
    fireEvent.click(screen.getByRole("button", { name: "Create organization" }));
    expect(onSubmit).toHaveBeenCalledWith({ name: "Purdue University", originated_id: "PURDUE" });
  });

  it("renders the backend error verbatim", () => {
    render(
      <CreateOrganizationDialog
        open
        onOpenChange={() => {}}
        onSubmit={() => {}}
        isPending={false}
        error="organization already exists"
      />,
    );
    expect(screen.getByText("organization already exists")).toBeInTheDocument();
  });
});
