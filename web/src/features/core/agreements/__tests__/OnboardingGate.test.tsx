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
import { fireEvent, render, screen } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import { OnboardingGate } from "../components/OnboardingGate";

vi.mock("next/navigation", () => ({ useRouter: () => ({ refresh: vi.fn() }) }));
vi.mock("next-auth/react", () => ({
  useSession: () => ({
    data: { user: { name: "John Doe", email: "jdoe@example.edu" } },
    update: vi.fn(),
  }),
}));

const server = setupServer(
  http.get("*/api/v1/me", () =>
    HttpResponse.json({
      consent: {
        pending: [
          { id: "agr-1", title: "User Agreement", version_label: "1.0", body_md: "Be nice." },
        ],
      },
      onboarding: { required: true, username: "custos-jdoe", username_editable: true },
    }),
  ),
  http.get("*/api/v1/me/onboarding/username", ({ request }) => {
    const u = new URL(request.url).searchParams.get("u");
    const available = u !== "taken-user";
    return HttpResponse.json({
      valid: true,
      available,
      suggestion: available ? "" : "taken-user2",
    });
  }),
);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());
afterEach(() => server.resetHandlers());

async function renderGate() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <OnboardingGate />
    </QueryClientProvider>,
  );
  await screen.findByRole("heading", { name: "Set up your account" });
}

describe("<OnboardingGate />", () => {
  // Checks that Continue needs both an available username and the agreement box checked.
  it("keeps Continue disabled until the username is available and the box is checked", async () => {
    await renderGate();
    const cont = screen.getByRole("button", { name: /continue/i });
    expect(await screen.findByText("Available")).toBeInTheDocument();
    expect(cont).toBeDisabled();
    fireEvent.click(screen.getByLabelText(/I have read and agree/));
    expect(cont).toBeEnabled();
    fireEvent.change(screen.getByLabelText("Username"), { target: { value: "taken-user" } });
    expect(await screen.findByText("Taken")).toBeInTheDocument();
    expect(cont).toBeDisabled();
  });

  // Checks that a taken username offers the suggestion and one click puts it in the field.
  it("offers the suggestion for a taken username", async () => {
    await renderGate();
    const field = screen.getByLabelText("Username");
    fireEvent.change(field, { target: { value: "taken-user" } });
    fireEvent.click(await screen.findByRole("button", { name: "Use taken-user2" }));
    expect(field).toHaveValue("taken-user2");
    expect(await screen.findByText("Available")).toBeInTheDocument();
  });

  // Checks that a user with no cluster account sees no username field and can continue when no agreement is published.
  it("lets a user without a cluster account continue", async () => {
    server.use(
      http.get("*/api/v1/me", () =>
        HttpResponse.json({
          consent: { pending: [] },
          onboarding: { required: true, username: "", username_editable: false },
        }),
      ),
    );
    await renderGate();
    expect(screen.queryByLabelText("Username")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /continue/i })).toBeEnabled();
  });
});
