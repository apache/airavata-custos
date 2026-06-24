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
import { Calendar } from "lucide-react";
import { describe, expect, it } from "vitest";
import { MetaItem, MetaRow } from "../MetaRow";

describe("MetaRow", () => {
  it("renders as a labeled group with horizontal layout", () => {
    const { container } = render(
      <MetaRow>
        <MetaItem label="PI" value="Eroma" />
      </MetaRow>,
    );
    const root = container.firstElementChild as HTMLElement;
    expect(root.tagName).toBe("DIV");
    expect(root.getAttribute("role")).toBe("group");
    expect(root.getAttribute("aria-label")).toBe("Metadata");
    expect(root.className).toContain("flex-wrap");
  });
});

describe("MetaItem", () => {
  it("default variant renders icon, label, and value", () => {
    render(<MetaItem icon={Calendar} label="End date" value="Nov 06, 2026" />);
    expect(screen.getByText(/End date/)).toBeInTheDocument();
    expect(screen.getByText("Nov 06, 2026")).toBeInTheDocument();
    // Lucide icon is an inline SVG.
    expect(document.querySelector("svg")).not.toBeNull();
  });

  it("default variant omits icon and label when not provided", () => {
    render(<MetaItem value="42" />);
    expect(screen.getByText("42")).toBeInTheDocument();
    expect(document.querySelector("svg")).toBeNull();
  });

  it("status variant renders a StatusBadge in success tone with a dot", () => {
    const { container } = render(<MetaItem variant="status" tone="success" value="Active" />);
    expect(screen.getByText("Active")).toBeInTheDocument();
    // StatusBadge "active" variant emits the dot span.
    expect(container.querySelector("[aria-hidden='true']")).not.toBeNull();
  });

  it("status variant maps danger tone to rejected badge", () => {
    render(<MetaItem variant="status" tone="danger" value="Failed" />);
    expect(screen.getByText("Failed")).toBeInTheDocument();
  });
});
