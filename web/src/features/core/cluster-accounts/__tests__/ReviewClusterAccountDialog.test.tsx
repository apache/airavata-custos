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

import { zClusterAccountResponse } from "@/generated/core/zod.gen";
import { projectsHandlers } from "@/mocks/handlers/projects";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import fixture from "../__fixtures__/cluster-accounts.json";
import {
  type ReviewAction,
  ReviewClusterAccountDialog,
} from "../components/ReviewClusterAccountDialog";

const [pending] = zClusterAccountResponse.array().parse(fixture);
if (!pending) throw new Error("no cluster account fixture");

const server = setupServer(...projectsHandlers);
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

function renderDialog(action: ReviewAction, onConfirm = vi.fn()) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <ReviewClusterAccountDialog
        accounts={pending ? [pending] : []}
        action={action}
        onOpenChange={vi.fn()}
        onConfirm={onConfirm}
        isPending={false}
      />
    </QueryClientProvider>,
  );
  return onConfirm;
}

describe("ReviewClusterAccountDialog", () => {
  // Make sure the denial passes the typed reason through, since it ends up in the audit log.
  it("sends the note with a denial", () => {
    const onConfirm = renderDialog("deny");
    fireEvent.change(screen.getByLabelText(/Reason/), { target: { value: "  Not on the list " } });
    fireEvent.click(screen.getByRole("button", { name: "Deny" }));
    expect(onConfirm).toHaveBeenCalledWith("Not on the list");
  });

  // Make sure a denial cannot be sent without a reason, since the backend
  // refuses it and the reason is passed on to whoever requested the account.
  it("keeps Deny disabled until a reason is typed", () => {
    renderDialog("deny");
    expect(screen.getByRole("button", { name: "Deny" })).toBeDisabled();
    fireEvent.change(screen.getByLabelText(/Reason/), { target: { value: "   " } });
    expect(screen.getByRole("button", { name: "Deny" })).toBeDisabled();
    fireEvent.change(screen.getByLabelText(/Reason/), { target: { value: "Not on the list" } });
    expect(screen.getByRole("button", { name: "Deny" })).toBeEnabled();
  });

  // Make sure the approval names the account the admin is about to create, so
  // a wrong row is caught before it goes to the cluster.
  it("restates the account before approving", () => {
    renderDialog("approve");
    expect(screen.getByText("custos-jdoe")).toBeInTheDocument();
    expect(screen.getByText("cluster-a")).toBeInTheDocument();
    expect(screen.getByText("Genomic Sequencing - GPU Pool")).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent("recorded as the approver");
    expect(screen.queryByLabelText(/Reason/)).toBeNull();
  });
});
