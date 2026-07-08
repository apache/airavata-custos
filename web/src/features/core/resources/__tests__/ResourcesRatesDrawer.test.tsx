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

import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Rate } from "../schemas";

// Windows are anchored far in the past/future so the ACTIVE/EXPIRED derivation
// stays stable regardless of the wall clock at test time.
const activeRate: Rate = {
  id: "rate-current",
  compute_allocation_resource_id: "res-1",
  rate: 0.05,
  start_time: "2000-01-01T00:00:00Z",
  end_time: "2100-01-01T00:00:00Z",
};
const expiredRate: Rate = {
  id: "rate-old",
  compute_allocation_resource_id: "res-1",
  rate: 0.04,
  start_time: "2000-01-01T00:00:00Z",
  end_time: "2001-01-01T00:00:00Z",
};

let rates: Rate[] = [];
let effective: Rate | undefined;

vi.mock("../queries", () => ({
  useResourceRates: () => ({ data: rates, isLoading: false, error: null, refetch: vi.fn() }),
  useEffectiveRate: () => ({ data: effective, isLoading: false, error: null, refetch: vi.fn() }),
}));

import { ResourcesRatesDrawer } from "../components/ResourcesRatesDrawer";

beforeEach(() => {
  rates = [activeRate, expiredRate];
  effective = activeRate;
});

describe("<ResourcesRatesDrawer />", () => {
  it("derives ACTIVE and EXPIRED badges per rate window", () => {
    render(
      <ResourcesRatesDrawer resourceId="res-1" resourceName="Anvil CPU" onOpenChange={vi.fn()} />,
    );
    // One badge in the effective card plus one per active history row.
    expect(screen.getAllByText("ACTIVE").length).toBeGreaterThanOrEqual(2);
    expect(screen.getByText("EXPIRED")).toBeInTheDocument();
  });

  it("titles the drawer with the resource name and shows the effective rate", () => {
    render(
      <ResourcesRatesDrawer resourceId="res-1" resourceName="Anvil CPU" onOpenChange={vi.fn()} />,
    );
    expect(screen.getByText("Rates: Anvil CPU")).toBeInTheDocument();
    expect(screen.getByText("Effective rate")).toBeInTheDocument();
  });

  it("shows the empty state when no rates are recorded", () => {
    rates = [];
    effective = undefined;
    render(
      <ResourcesRatesDrawer resourceId="res-1" resourceName="Anvil CPU" onOpenChange={vi.fn()} />,
    );
    expect(
      screen.getByRole("heading", { name: /no rates recorded for this resource/i }),
    ).toBeInTheDocument();
  });
});
