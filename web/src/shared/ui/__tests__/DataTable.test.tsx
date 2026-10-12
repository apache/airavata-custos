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

import { render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { DataTable } from "../DataTable";

describe("<DataTable />", () => {
  it("names an action column whose header is empty", async () => {
    const { container } = render(
      <DataTable
        caption="Items"
        columns={[
          { key: "name", header: "Name", cell: (r: { id: string }) => r.id },
          { key: "actions", header: "", cell: () => <button type="button">Edit</button> },
        ]}
        rows={[{ id: "a" }]}
        rowKey={(r) => r.id}
      />,
    );
    const result = await axe.run(container, { runOnly: ["empty-table-header"] });
    expect(result.violations).toEqual([]);
    expect(result.passes).toHaveLength(1);
  });

  // Keyboard users can scroll a table wider than the screen even when no cell is focusable (axe
  // scrollable-region-focusable); a table that fits is no extra Tab stop.
  it("makes its scroll area a Tab stop only while it overflows", () => {
    const table = () => (
      <DataTable
        caption="Project members"
        columns={[{ key: "name", header: "Name", cell: (r: { id: string }) => r.id }]}
        rows={[{ id: "a" }]}
        rowKey={(r) => r.id}
      />
    );
    const { unmount } = render(table());
    expect(screen.getByRole("region", { name: "Project members" })).not.toHaveAttribute("tabindex");
    unmount();
    const width = vi.spyOn(HTMLElement.prototype, "scrollWidth", "get").mockReturnValue(900);
    render(table());
    expect(screen.getByRole("region", { name: "Project members" })).toHaveAttribute(
      "tabindex",
      "0",
    );
    width.mockRestore();
  });

  it("hides the pager on a page past the end", () => {
    render(
      <DataTable
        caption="Items"
        columns={[{ key: "name", header: "Name", cell: (r: { id: string }) => r.id }]}
        rows={[]}
        rowKey={(r) => r.id}
        pagination={{ page: 999, pageSize: 25, total: 267, onPageChange: () => {} }}
      />,
    );
    expect(screen.queryByRole("button", { name: "Next" })).toBeNull();
    expect(screen.queryByText(/Showing/)).toBeNull();
  });
});
