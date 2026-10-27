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

import { type ClusterAccountResponse, zClusterAccountResponse } from "@/generated/core/zod.gen";
import { projectsHandlers } from "@/mocks/handlers/projects";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import fixture from "../__fixtures__/cluster-accounts.json";
import { ClusterAccountsTable } from "../components/ClusterAccountsTable";

const rows = zClusterAccountResponse.array().parse(fixture);
const pending = rows.filter((r) => r.approval_status === "PENDING");

const server = setupServer(...projectsHandlers);
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

function byId(id: string): ClusterAccountResponse {
  const row = rows.find((r) => r.id === id);
  if (!row) throw new Error(`no fixture ${id}`);
  return row;
}

function rowOf(text: string): HTMLTableRowElement {
  const row = screen.getByText(text).closest("tr");
  if (!row) throw new Error(`no row for ${text}`);
  return row;
}

function renderTable(canReview: boolean, shown: ClusterAccountResponse[] = rows, grouped = false) {
  const onApprove = vi.fn();
  const onDeny = vi.fn();
  const onApproveAll = vi.fn();
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <ClusterAccountsTable
        rows={shown}
        isLoading={false}
        error={null}
        canReview={canReview}
        onApprove={onApprove}
        onDeny={onDeny}
        onApproveAll={onApproveAll}
        grouped={grouped}
        page={1}
        pageSize={25}
        total={shown.length}
        onPageChange={vi.fn()}
        emptyHeading="Nothing waiting for approval."
      />
    </QueryClientProvider>,
  );
  return { onApprove, onDeny, onApproveAll };
}

describe("ClusterAccountsTable", () => {
  // Make sure only a pending account offers decisions; the rest show their approval status in that column.
  it("offers decisions on pending accounts and the status on the rest", () => {
    renderTable(true);
    const row = rowOf("custos-jdoe");
    expect(row.textContent).toContain("Approve");
    expect(row.textContent).toContain("Deny");

    const denied = rowOf("custos-lroe");
    expect(denied.querySelector("button")).toBeNull();
    expect(denied.textContent).toContain("Denied");

    expect(rowOf("custos-sdoe").textContent).toContain("Approved · provisioning");
    expect(rowOf("custos-jdoe2").textContent).toContain("Active");
  });

  // Make sure a reader without the write privilege sees the pending status but no decision buttons.
  it("hides the decisions without the review privilege", () => {
    renderTable(false, pending, true);
    expect(screen.queryByRole("button", { name: "Approve" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Deny" })).toBeNull();
    expect(screen.queryByRole("button", { name: /^Approve all/ })).toBeNull();
    expect(rowOf("custos-jdoe").textContent).toContain("Pending approval");
  });

  it("hands the clicked account to the approve and deny callbacks", () => {
    const { onApprove, onDeny } = renderTable(true, [byId("ccu-pending-1")]);
    fireEvent.click(screen.getByRole("button", { name: "Approve" }));
    expect(onApprove).toHaveBeenCalledWith(expect.objectContaining({ id: "ccu-pending-1" }));
    fireEvent.click(screen.getByRole("button", { name: "Deny" }));
    expect(onDeny).toHaveBeenCalledWith(expect.objectContaining({ id: "ccu-pending-1" }));
  });

  // Make sure the reason an admin gave is visible on the denied row, since it
  // is the only place the portal shows it.
  it("shows the denial note on the row", () => {
    renderTable(false, [byId("ccu-denied-1")]);
    expect(screen.getByText("Not on the approved collaborator list.")).toBeInTheDocument();
  });

  // Make sure each pending request sits under its own allocation's card, in server order.
  it("groups pending requests by allocation in server order", () => {
    const unlinked = pending.filter((r) => !r.allocation_id);
    const linked = pending.filter((r) => r.allocation_id);
    renderTable(true, [...linked, ...unlinked], true);

    const sections = screen.getAllByRole("region");
    const names = sections.map(
      (s) => document.getElementById(s.getAttribute("aria-labelledby") ?? "")?.textContent,
    );
    expect(names).toEqual([
      "Genomic Sequencing - GPU Pool",
      "Plasma Simulation - GPU",
      "No allocation",
    ]);
    expect(within(sections[0] as HTMLElement).getByText("custos-jroe")).toBeInTheDocument();
    expect(within(sections[2] as HTMLElement).getByText("custos-adoe")).toBeInTheDocument();
  });

  // Make sure Approve all shows only on an allocation with two or more requests, and hands over just those.
  it("offers Approve all only on a linked group with several requests", () => {
    const { onApproveAll } = renderTable(true, pending, true);
    const buttons = screen.getAllByRole("button", { name: /^Approve all/ });
    expect(buttons).toHaveLength(1);

    fireEvent.click(buttons[0] as HTMLElement);
    const ids = (onApproveAll.mock.calls[0]?.[0] as ClusterAccountResponse[]).map((r) => r.id);
    expect(ids).toEqual(["ccu-pending-1", "ccu-pending-2", "ccu-pending-3"]);
  });
});
