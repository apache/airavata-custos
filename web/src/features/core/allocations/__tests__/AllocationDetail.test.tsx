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

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const members = [
  { id: "m-1", membership_status: "ACTIVE" },
  { id: "m-2", membership_status: "INACTIVE" },
  { id: "m-3", membership_status: "ACTIVE" },
];

vi.mock("@/shared/casl/permissions", () => ({ usePermissions: () => ({}), requires: () => "" }));
vi.mock("next-auth/react", () => ({ useSession: () => ({ data: null }) }));
vi.mock("@/shared/layout/BreadcrumbLabelsProvider", () => ({ useBreadcrumbLabel: vi.fn() }));
vi.mock("@/features/core/projects/queries", () => ({ useProject: () => ({ data: undefined }) }));
vi.mock("@/shared/ui/ConfirmDialog", async (importOriginal) => ({
  ...(await importOriginal<typeof import("@/shared/ui/ConfirmDialog")>()),
  DeleteButton: () => null,
}));
vi.mock("@/shared/ui/TabsRouter", () => ({ TabsRouter: () => null }));
vi.mock("../components/AllocationDetailHeader", () => ({
  AllocationDetailHeader: ({ memberCount }: { memberCount: number }) => (
    <p>{memberCount} members</p>
  ),
}));
vi.mock("../queries", async (importOriginal) => ({
  ...(await importOriginal<typeof import("../queries")>()),
  useAllocation: () => ({
    data: { id: "alloc-001", name: "Alpha" },
    isLoading: false,
    error: null,
  }),
  useAllocationMembers: () => ({ data: members }),
}));

import { AllocationDetail } from "../components/AllocationDetail";

describe("<AllocationDetail />", () => {
  it("counts active members only", () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <AllocationDetail allocationId="alloc-001" />
      </QueryClientProvider>,
    );
    expect(screen.getByText("2 members")).toBeInTheDocument();
  });
});
