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
import { afterAll, afterEach, beforeAll, describe, expect, it } from "vitest";
import { ProfileCard } from "../ProfileCard";
import type { UserProfile } from "../../schemas";

const user: UserProfile = {
  id: "u1",
  email: "elena.vasquez@sample.example.edu",
  first_name: "Elena",
  middle_name: "",
  last_name: "Vasquez",
  organization_id: "org-1",
  status: "ACTIVE",
  type: "CLUSTER_LOCAL",
};

let putBody: unknown;
const server = setupServer(
  http.put("*/api/v1/users/:id", async ({ request }) => {
    putBody = await request.json();
    return HttpResponse.json({ ...user, first_name: "Elena", last_name: "Vasquez-Ng" });
  }),
);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());
afterEach(() => server.resetHandlers());

function renderCard() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <ProfileCard user={user} />
    </QueryClientProvider>,
  );
}

describe("ProfileCard", () => {
  it("shows name, email, ACTIVE badge, and read-only fields", () => {
    renderCard();
    expect(screen.getByText("Elena Vasquez")).toBeInTheDocument();
    expect(screen.getByText("elena.vasquez@sample.example.edu")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("elena.vasquez")).toBeInTheDocument();
  });

  it("expands the inline editor and PUTs the name fields on save", async () => {
    renderCard();
    fireEvent.click(screen.getByRole("button", { name: /edit name/i }));

    fireEvent.change(screen.getByLabelText(/last name/i), {
      target: { value: "Vasquez-Ng" },
    });
    fireEvent.click(screen.getByRole("button", { name: /^save$/i }));

    await waitFor(() =>
      expect(putBody).toEqual({
        first_name: "Elena",
        middle_name: "",
        last_name: "Vasquez-Ng",
      }),
    );
  });
});
