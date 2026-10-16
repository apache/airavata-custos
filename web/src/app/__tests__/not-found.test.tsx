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
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

vi.mock("@/shared/auth/auth", () => ({ auth: async () => ({ custosUserId: "u-1" }) }));
vi.mock("@/shared/layout/PortalLayout", () => ({
  PortalLayout: ({ children, breadcrumbs }: { children: ReactNode; breadcrumbs?: boolean }) => (
    <div data-testid="portal" data-breadcrumbs={String(breadcrumbs ?? true)}>
      {children}
    </div>
  ),
}));

import NotFound from "../not-found";

describe("not-found page", () => {
  it("renders inside the portal chrome without breadcrumbs, with an h1 and a link home", async () => {
    render(await NotFound());
    expect(screen.getByRole("heading", { level: 1, name: "Page not found" })).toBeInTheDocument();
    expect(screen.getByTestId("portal")).toHaveAttribute("data-breadcrumbs", "false");
    expect(screen.getByRole("link", { name: "Back to home" })).toHaveAttribute("href", "/allocations");
  });
});
