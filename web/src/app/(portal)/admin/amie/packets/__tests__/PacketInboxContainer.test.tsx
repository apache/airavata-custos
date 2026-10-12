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

const nav = vi.hoisted(() => ({
  replace: vi.fn(),
  search: "status=FAILED&q=900&page=3&packet=p-1",
}));
const usePackets = vi.hoisted(() =>
  vi.fn(() => ({ data: { packets: [], total: 0 }, isLoading: false, error: null })),
);

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: nav.replace }),
  useSearchParams: () => new URLSearchParams(nav.search),
}));
vi.mock("next/dynamic", () => ({ default: () => () => null }));
vi.mock("@/features/connectors/amie/queries", () => ({
  usePackets,
  usePacketStats: () => ({ data: undefined }),
  usePacket: () => ({ data: undefined }),
}));
vi.mock("@/features/connectors/amie/components/PacketDetailDrawer", () => ({
  PacketDetailDrawer: () => null,
}));
vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => false }) }));

import { PacketInboxContainer } from "../PacketInboxContainer";

describe("<PacketInboxContainer />", () => {
  // Make sure filters and page survive a reload, a page past the end offers the first, and a filter change resets the page.
  it("reads its filters and page from the URL and writes changes back, keeping the open packet", () => {
    render(<PacketInboxContainer />);
    expect(usePackets).toHaveBeenLastCalledWith(
      expect.objectContaining({ status: "FAILED", q: "900", offset: 40 }),
    );
    expect(screen.getByText("No packets on page 3")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "First page" }));
    expect(nav.replace).toHaveBeenLastCalledWith(
      "/admin/amie/packets?status=FAILED&q=900&packet=p-1",
      { scroll: false },
    );
    expect(screen.getByLabelText("Status")).toHaveValue("FAILED");

    fireEvent.change(screen.getByLabelText("Type"), { target: { value: "request_user_modify" } });
    expect(nav.replace).toHaveBeenCalledWith(
      "/admin/amie/packets?status=FAILED&q=900&packet=p-1&type=request_user_modify",
      { scroll: false },
    );
  });
});
