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

import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { TabsRouter } from "../TabsRouter";

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(window.location.search),
}));

const tabs = [
  { value: "users", label: "Users", content: <p>Users panel</p> },
  { value: "audit", label: "Audit Log", content: <p>Audit panel</p> },
];

describe("TabsRouter", () => {
  it.each([
    ["?tab=bogus&q=1", "?q=1", "Users panel"],
    ["?tab=users&q=1", "?q=1", "Users panel"],
    ["?tab=audit", "?tab=audit", "Audit panel"],
  ])("canonicalises %s to %s", async (from, to, panel) => {
    window.history.replaceState(null, "", `/admin${from}`);
    render(<TabsRouter tabs={tabs} defaultValue="users" />);
    expect(screen.getByText(panel)).toBeInTheDocument();
    await waitFor(() => expect(window.location.search).toBe(to));
  });
});
