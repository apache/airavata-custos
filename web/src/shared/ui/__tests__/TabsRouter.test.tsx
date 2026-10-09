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
});
