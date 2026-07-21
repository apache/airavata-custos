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
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import type { AccessRequest } from "@/features/core/access-requests/schemas";
import { accessRequestsHandlers, resetAccessRequests } from "@/mocks/handlers/access-requests";

let searchParams = "";

vi.mock("next-auth/react", () => ({
  useSession: () => ({
    data: { user: { name: "Jane Doe", email: "jane@example.edu" } },
    status: "authenticated",
  }),
}));

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(searchParams),
}));

import { NoAccessBody } from "../NoAccessBody";

const server = setupServer(...accessRequestsHandlers);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());
afterEach(() => {
  server.resetHandlers();
  resetAccessRequests();
  searchParams = "";
});

const myRequest = (overrides: Partial<AccessRequest> = {}): AccessRequest => ({
  id: "areq-me",
  oidc_sub: "sub-mock-caller",
  email: "jane@example.edu",
  name: "Jane Doe",
  institution: "Example University",
  event_code: "PEARC26",
  reason: "",
  status: "PENDING",
  approver_id: "",
  deny_reason: "",
  expires_at: null,
  created_user_id: "",
  timestamp: "2026-07-12T10:00:00Z",
  ...overrides,
});

function renderBody() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <NoAccessBody />
    </QueryClientProvider>,
  );
}

async function openForm() {
  renderBody();
  fireEvent.click(await screen.findByRole("button", { name: /request trial access/i }));
}

describe("<NoAccessBody />", () => {
  it("shows the entry state when the caller has no request yet", async () => {
    renderBody();
    expect(await screen.findByRole("button", { name: /request trial access/i })).toBeVisible();
    expect(screen.getByText(/request temporary trial access/i)).toBeInTheDocument();
  });

  it("prefills the form from the token and the ?event= param, resolving the code", async () => {
    searchParams = "event=PEARC26";
    await openForm();
    expect(screen.getByLabelText("Name")).toHaveValue("Jane Doe");
    expect(screen.getByLabelText("Email")).toHaveValue("jane@example.edu");
    expect(screen.getByLabelText("Event code")).toHaveValue("PEARC26");
    await screen.findByText(/pearc26-tutorial/i, undefined, { timeout: 2000 });
    // Institution is still empty, so submit stays blocked.
    expect(screen.getByRole("button", { name: /submit request/i })).toBeDisabled();
  });

  it("blocks submit on an unknown event code with an inline error", async () => {
    await openForm();
    fireEvent.change(screen.getByLabelText("Institution"), { target: { value: "Example U" } });
    fireEvent.change(screen.getByLabelText("Event code"), { target: { value: "NOPE" } });
    await screen.findByText(/unknown event code/i, undefined, { timeout: 2000 });
    expect(screen.getByRole("button", { name: /submit request/i })).toBeDisabled();
  });

  it("submits a valid request and flips to the pending state", async () => {
    searchParams = "event=PEARC26";
    await openForm();
    fireEvent.change(screen.getByLabelText("Institution"), { target: { value: "Example U" } });
    await screen.findByText(/pearc26-tutorial/i, undefined, { timeout: 2000 });
    // The prefilled username must clear the availability check before submit unlocks.
    await waitFor(
      () => expect(screen.getByRole("button", { name: /submit request/i })).toBeEnabled(),
      { timeout: 2000 },
    );
    fireEvent.click(screen.getByRole("button", { name: /submit request/i }));
    await screen.findByRole("heading", { name: /request received/i });
  });

  it("prefills the suggested username and marks it available", async () => {
    searchParams = "event=PEARC26";
    await openForm();
    await screen.findByText(/pearc26-tutorial/i, undefined, { timeout: 2000 });
    await waitFor(() => expect(screen.getByLabelText("Cluster username")).toHaveValue("nexus-mockcaller"), {
      timeout: 2000,
    });
    expect(await screen.findByLabelText(/username available/i, undefined, { timeout: 2000 })).toBeInTheDocument();
  });

  it("flags a taken username and blocks submit", async () => {
    searchParams = "event=PEARC26";
    await openForm();
    fireEvent.change(screen.getByLabelText("Institution"), { target: { value: "Example U" } });
    await screen.findByText(/pearc26-tutorial/i, undefined, { timeout: 2000 });
    fireEvent.change(screen.getByLabelText("Cluster username"), { target: { value: "taken" } });
    expect(await screen.findByText(/already taken/i, undefined, { timeout: 2000 })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /submit request/i })).toBeDisabled();
  });

  it("rejects a malformed username with a format hint", async () => {
    searchParams = "event=PEARC26";
    await openForm();
    fireEvent.change(screen.getByLabelText("Institution"), { target: { value: "Example U" } });
    await screen.findByText(/pearc26-tutorial/i, undefined, { timeout: 2000 });
    fireEvent.change(screen.getByLabelText("Cluster username"), { target: { value: "Bad Name" } });
    expect(await screen.findByText(/lowercase letters, digits/i, undefined, { timeout: 2000 })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /submit request/i })).toBeDisabled();
  });

  it("shows the pending state for an existing PENDING request", async () => {
    server.use(http.get("*/api/v1/access-requests/me", () => HttpResponse.json(myRequest())));
    renderBody();
    await screen.findByRole("heading", { name: /request received/i });
    expect(await screen.findByText(/pearc26-tutorial/i)).toBeInTheDocument();
    expect(screen.getByText(/jul 12, 2026/i)).toBeInTheDocument();
  });

  it("shows the deny reason verbatim and allows a new request", async () => {
    server.use(
      http.get("*/api/v1/access-requests/me", () =>
        HttpResponse.json(myRequest({ status: "DENIED", deny_reason: "Event is full." })),
      ),
    );
    renderBody();
    await screen.findByRole("heading", { name: /request declined/i });
    expect(screen.getByText("Event is full.")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /submit a new request/i }));
    expect(screen.getByLabelText("Institution")).toBeInTheDocument();
  });

  it("falls back to the pending state when the backend reports a pending conflict", async () => {
    let posted = false;
    server.use(
      http.get("*/api/v1/access-requests/me", () =>
        posted
          ? HttpResponse.json(myRequest())
          : HttpResponse.json({ error: "nf" }, { status: 404 }),
      ),
      http.post("*/api/v1/access-requests", () => {
        posted = true;
        return HttpResponse.json({ error: "a pending request already exists" }, { status: 409 });
      }),
    );
    searchParams = "event=PEARC26";
    await openForm();
    fireEvent.change(screen.getByLabelText("Institution"), { target: { value: "Example U" } });
    await screen.findByText(/pearc26-tutorial/i, undefined, { timeout: 2000 });
    await waitFor(
      () => expect(screen.getByRole("button", { name: /submit request/i })).toBeEnabled(),
      { timeout: 2000 },
    );
    fireEvent.click(screen.getByRole("button", { name: /submit request/i }));
    await screen.findByRole("heading", { name: /request received/i });
  });
});
