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
import { describe, expect, it, vi } from "vitest";
import { TabsRouter } from "../TabsRouter";

// Stub the App Router hooks — the component reads `?tab=` and replaces the
// route on change, neither of which we exercise here.
vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: vi.fn() }),
  useSearchParams: () => new URLSearchParams(),
}));

const baseTabs = [
  { value: "users", label: "Users & Roles", content: <p>Users panel</p> },
  { value: "credits", label: "Credits & Resources", content: <p>Credits panel</p> },
  { value: "audit", label: "Audit Log", content: <p>Audit panel</p> },
];

describe("TabsRouter", () => {
  it("renders all tabs and the default tab's content", () => {
    render(<TabsRouter tabs={baseTabs} defaultValue="users" />);
    expect(screen.getByRole("tab", { name: /Users & Roles/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /Credits & Resources/i })).toBeInTheDocument();
    expect(screen.getByRole("tab", { name: /Audit Log/i })).toBeInTheDocument();
  });

  it("renders a single rightSlot node beside the tab strip", () => {
    render(
      <TabsRouter
        tabs={baseTabs}
        defaultValue="users"
        rightSlot={<button type="button">Global action</button>}
      />,
    );
    expect(screen.getByRole("button", { name: /Global action/i })).toBeInTheDocument();
  });

  it("renders the per-tab rightSlot matching the active tab", () => {
    render(
      <TabsRouter
        tabs={baseTabs}
        defaultValue="users"
        rightSlot={{
          users: <button type="button">Add user</button>,
          credits: <button type="button">Request extension</button>,
          audit: undefined,
        }}
      />,
    );
    expect(screen.getByRole("button", { name: /Add user/i })).toBeInTheDocument();
    // Inactive tab slots are not rendered.
    expect(screen.queryByRole("button", { name: /Request extension/i })).not.toBeInTheDocument();
  });

  it("omits the right slot entirely when no rightSlot is passed", () => {
    const { container } = render(<TabsRouter tabs={baseTabs} defaultValue="users" />);
    // The list wrapper has `justify-between`; we only care that no extra slot
    // sibling renders next to the tab list.
    const list = container.querySelector("[role='tablist']") as HTMLElement;
    expect(list).not.toBeNull();
    expect(list.parentElement?.children).toHaveLength(1);
  });
});
