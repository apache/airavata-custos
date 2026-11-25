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
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import { clustersHandlers } from "@/mocks/handlers/clusters";
import { slurmConflicts, slurmHandlers } from "@/mocks/handlers/slurm";
import { ClusterQoSTab } from "../components/ClusterQoSTab";

vi.mock("@/shared/casl/AbilityProvider", () => ({ useAbility: () => ({ can: () => true }) }));

const server = setupServer(...clustersHandlers, ...slurmHandlers);
beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe("<ClusterQoSTab />", () => {
  it("marks a QoS only SLURM has", async () => {
    server.use(
      slurmConflicts([
        {
          key: "qos:openfold",
          name: "openfold",
          label: "QoS",
          kind: "qos_definition",
          custos: null,
          slurm: { name: "openfold", priority: 100, max_tres_per_user: "", flags: [] },
          choices: ["custos", "slurm"],
        },
      ]),
    );
    render(
      <QueryClientProvider
        client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
      >
        <ClusterQoSTab clusterId="c1" />
      </QueryClientProvider>,
    );
    expect(await screen.findByText("only on SLURM")).toBeInTheDocument();
    expect(screen.getByText("openfold")).toBeInTheDocument();
  });
});
