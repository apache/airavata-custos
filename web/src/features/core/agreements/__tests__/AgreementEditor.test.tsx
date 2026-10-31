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
import { AgreementEditor } from "../components/AgreementEditor";

vi.mock("next/navigation", () => ({ useRouter: () => ({ push: vi.fn() }) }));

const server = setupServer(
  http.get("*/api/v1/agreements", () => HttpResponse.json([])),
  http.post("*/api/v1/agreements", () =>
    HttpResponse.json({ error: "version label already exists" }, { status: 409 }),
  ),
);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());
afterEach(() => server.resetHandlers());

async function renderEditor() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <AgreementEditor />
    </QueryClientProvider>,
  );
  await screen.findByRole("heading", { name: "New version" });
}

const type = (label: RegExp, value: string) =>
  fireEvent.change(screen.getByLabelText(label), { target: { value } });

describe("<AgreementEditor />", () => {
  // Checks that Save draft needs a title, a version and a body.
  it("keeps Save draft disabled until title, version and body are filled", async () => {
    await renderEditor();
    const save = screen.getByRole("button", { name: "Save draft" });
    type(/title/i, "Acceptable Use Policy");
    type(/version label/i, "2.0");
    expect(save).toBeDisabled();
    type(/body/i, "Use resources only for approved research.");
    expect(save).toBeEnabled();
  });

  // Checks that a 409 shows under the version field.
  it("shows the duplicate-version message on a 409", async () => {
    await renderEditor();
    type(/title/i, "Acceptable Use Policy");
    type(/version label/i, "1.0");
    type(/body/i, "Use resources only for approved research.");
    fireEvent.click(screen.getByRole("button", { name: "Save draft" }));
    expect(await screen.findByText("Version 1.0 already exists.")).toBeInTheDocument();
  });
});
