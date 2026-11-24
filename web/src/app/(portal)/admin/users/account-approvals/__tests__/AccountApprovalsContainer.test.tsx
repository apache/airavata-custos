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

import fixture from "@/features/core/cluster-accounts/__fixtures__/cluster-accounts.json";
import { clustersHandlers } from "@/mocks/handlers/clusters";
import { projectsHandlers } from "@/mocks/handlers/projects";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import { AccountApprovalsContainer } from "../AccountApprovalsContainer";

const toast = vi.hoisted(() => ({ success: vi.fn(), error: vi.fn() }));
vi.mock("sonner", () => ({ toast }));
vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => true }) }));
vi.mock("@/shared/hooks/useShallowSearchParams", () => ({
  useShallowSearchParams: () => new URLSearchParams(),
  replaceShallowSearchParams: vi.fn(),
}));

const server = setupServer(...clustersHandlers, ...projectsHandlers);
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const pending = fixture.filter((a) => a.approval_status === "PENDING");

async function approveAll(johnReply: { status: number; error: string }) {
  server.use(
    http.get("*/api/v1/compute-cluster-users", () =>
      HttpResponse.json({ items: pending, total: pending.length }),
    ),
    http.post<{ id: string }>("*/api/v1/compute-cluster-users/:id/approve", ({ params }) =>
      params.id === "ccu-pending-2"
        ? HttpResponse.json({ error: johnReply.error }, { status: johnReply.status })
        : undefined,
    ),
  );
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <AccountApprovalsContainer />
    </QueryClientProvider>,
  );
  fireEvent.click(await screen.findByRole("button", { name: /^Approve all/ }));
  fireEvent.click(screen.getByRole("button", { name: "Approve 3" }));
}

describe("AccountApprovalsContainer", () => {
  // Make sure one failed approval in Approve all keeps only that account listed, with its reason.
  it("reports the rest approved and keeps the failed one listed", async () => {
    await approveAll({ status: 502, error: "cluster unreachable" });

    await waitFor(() => expect(toast.error).toHaveBeenCalledWith("Approved 2 of 3. 1 failed."));
    const list = screen.getByRole("list", { name: "Accounts to approve" });
    expect(within(list).getAllByRole("listitem")).toHaveLength(1);
    expect(within(list).getByText("custos-jroe")).toBeInTheDocument();
    expect(within(list).getByText("cluster unreachable")).toBeInTheDocument();
  });

  // Make sure an account someone else approved first (409) counts as approved, not failed.
  it("counts an already approved account as done", async () => {
    await approveAll({ status: 409, error: "already approved" });

    await waitFor(() =>
      expect(toast.success).toHaveBeenCalledWith("Approved 3 cluster accounts on cluster-a."),
    );
    expect(screen.queryByRole("list", { name: "Accounts to approve" })).toBeNull();
  });
});
