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
import { slurmHandlers } from "@/mocks/handlers/slurm";
import type { Conflict } from "../api";
import { SlurmActions } from "../components/SlurmActions";

vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => true }) }));

const server = setupServer(...slurmHandlers);
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const conflict: Conflict = {
  key: "qos:debug",
  name: "debug",
  label: "QoS",
  kind: "qos_definition",
  custos: { name: "debug", priority: 50, max_tres_per_user: "", flags: [] },
  slurm: { name: "debug", priority: 80, max_tres_per_user: "", flags: [] },
  choices: ["custos", "slurm", "manual"],
};

describe("SlurmActions", () => {
  it("posts the choice with the SLURM value it saw", async () => {
    let posted: unknown;
    server.use(
      http.post("*/api/v1/connectors/slurm/clusters/c1/conflicts/resolve", async ({ request }) => {
        posted = await request.json();
        return new HttpResponse(null, { status: 204 });
      }),
    );
    render(
      <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
      >
        <SlurmActions clusterId="c1" scope="cluster" conflicts={[conflict]} />
      </QueryClientProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Actions" }));
    fireEvent.click(await screen.findByText("Use SLURM"));
    await waitFor(() =>
      expect(posted).toEqual({
        key: "qos:debug",
        choice: "slurm",
        seen_slurm: conflict.slurm,
      }),
    );
  });
});
