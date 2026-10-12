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

const packetId = "9f3c2a71-0d4e-4b8a-9c11-5e2f7a6b8d90";
vi.mock("next/navigation", () => ({
  usePathname: () => `/admin/amie/packets/${packetId}`,
  useParams: () => ({ id: packetId }),
}));
vi.mock("../BreadcrumbLabelsProvider", () => ({ useBreadcrumbLabels: () => ({}) }));

import { Breadcrumbs } from "../Breadcrumbs";

describe("<Breadcrumbs />", () => {
  it("names each parent by its route label, leaves /admin unlinked and names a loading record by its kind", () => {
    render(<Breadcrumbs />);
    expect(screen.getByText("Admin").closest("a")).toBeNull();
    expect(screen.getByRole("link", { name: "AMIE" })).toHaveAttribute("href", "/admin/amie");
    expect(screen.getByRole("link", { name: "Inbox" })).toHaveAttribute(
      "href",
      "/admin/amie/packets",
    );
    expect(screen.getByText("Packet")).toHaveAttribute("aria-current", "page");
  });

  it("keeps only the parent and the current label below lg", () => {
    render(<Breadcrumbs />);
    expect(screen.getByRole("link", { name: "AMIE" }).closest("li")).toHaveClass(
      "hidden",
      "lg:flex",
    );
    expect(screen.getByText("Admin").closest("li")).toHaveClass("hidden", "lg:flex");
    expect(screen.getByRole("link", { name: "Inbox" }).closest("li")).not.toHaveClass("hidden");
    expect(screen.getByText("…").closest("li")).toHaveClass("lg:hidden");
  });
});
