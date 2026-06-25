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
import { GroupByChip, GroupByChipGroup } from "../GroupByChip";

const options = [
  { value: "resource", label: "Resource" },
  { value: "project", label: "Project" },
  { value: "user", label: "User" },
];

describe("GroupByChip", () => {
  it("renders trigger with label and current value", () => {
    render(
      <GroupByChip
        label="Group by"
        value="resource"
        options={options}
        onChange={vi.fn()}
      />,
    );
    const trigger = screen.getByRole("button", { name: /Group by/i });
    expect(trigger).toBeInTheDocument();
    expect(trigger).toHaveTextContent("Resource");
  });

  it("clicking an option fires onChange", () => {
    const onChange = vi.fn();
    render(
      <GroupByChip label="Group by" value="resource" options={options} onChange={onChange} />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Group by/i }));
    // Menu items render with role=menuitem.
    const projectItem = screen.getByRole("menuitem", { name: "Project" });
    fireEvent.click(projectItem);
    expect(onChange).toHaveBeenCalledWith("project");
  });

  it("marks current value with data-current attribute", () => {
    render(
      <GroupByChip label="Group by" value="project" options={options} onChange={vi.fn()} />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Group by/i }));
    const projectItem = screen.getByRole("menuitem", { name: "Project" });
    expect(projectItem.getAttribute("data-current")).not.toBeNull();
    const userItem = screen.getByRole("menuitem", { name: "User" });
    expect(userItem.getAttribute("data-current")).toBeNull();
  });
});

describe("GroupByChipGroup", () => {
  it("wraps children in an accessible group", () => {
    const { container } = render(
      <GroupByChipGroup>
        <GroupByChip
          label="Group by"
          value="resource"
          options={options}
          onChange={vi.fn()}
        />
      </GroupByChipGroup>,
    );
    const root = container.firstElementChild as HTMLElement;
    expect(root.getAttribute("role")).toBe("group");
    expect(root.getAttribute("aria-label")).toBe("Group by");
    expect(root.className).toContain("gap-2");
  });
});
