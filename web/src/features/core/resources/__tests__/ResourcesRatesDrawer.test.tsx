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

import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ComputeAllocationResourceRate as Rate } from "@/generated/core/types.gen";

// Windows are anchored far in the past/future so classifyRate stays stable
// regardless of the wall clock at test time (assumes now is 2011..2098).
const expiredRate: Rate = rate(
  "rate-expired",
  0.03,
  "2000-01-01T00:00:00Z",
  "2001-01-01T00:00:00Z",
);
const supersededRate: Rate = rate(
  "rate-super",
  0.04,
  "2001-01-01T00:00:00Z",
  "2100-01-01T00:00:00Z",
);
const activeRate: Rate = rate("rate-active", 0.05, "2010-01-01T00:00:00Z", "2100-01-01T00:00:00Z");
const scheduledRate: Rate = rate(
  "rate-sched",
  0.06,
  "2099-01-01T00:00:00Z",
  "2100-01-01T00:00:00Z",
);

function rate(id: string, value: number, start: string, end: string): Rate {
  return {
    id,
    compute_allocation_resource_id: "res-1",
    rate: value,
    start_time: start,
    end_time: end,
  };
}

let rates: Rate[] = [];
let effective: Rate | undefined;
let canManage = true;
const mutation = {
  mutate: vi.fn((_options: { body: Rate }, callbacks?: { onSuccess?: () => void }) =>
    callbacks?.onSuccess?.(),
  ),
  isPending: false,
};

vi.mock("../queries", () => ({
  useResourceRates: () => ({ data: rates, isLoading: false, error: null, refetch: vi.fn() }),
  useEffectiveRate: () => ({ data: effective, isLoading: false, error: null, refetch: vi.fn() }),
  useCreateResourceRate: () => mutation,
  useResourceRate: () => ({ data: undefined, error: null }),
}));

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({ can: () => canManage }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn() },
}));

import { toast } from "sonner";
import { ResourcesRatesDrawer } from "../components/ResourcesRatesDrawer";

function renderDrawer() {
  return render(
    <ResourcesRatesDrawer resourceId="res-1" resourceName="ClusterA CPU" onOpenChange={vi.fn()} />,
  );
}

function openForm() {
  fireEvent.click(screen.getByRole("button", { name: "+ Add rate" }));
}

function todayUtc(): string {
  return new Date().toISOString().slice(0, 10);
}

beforeEach(() => {
  rates = [expiredRate, supersededRate, activeRate, scheduledRate];
  effective = activeRate;
  canManage = true;
  mutation.mutate.mockClear();
  vi.mocked(toast.success).mockClear();
});

describe("<ResourcesRatesDrawer />", () => {
  it("classifies each row and sorts start_time DESC", () => {
    effective = undefined; // keep ACTIVE unique to the table
    renderDrawer();
    const rows = screen.getAllByRole("row").slice(1); // drop header
    const badges = rows.map(
      (r) => within(r).getByText(/ACTIVE|SCHEDULED|SUPERSEDED|EXPIRED/).textContent,
    );
    expect(badges).toEqual(["SCHEDULED", "ACTIVE", "SUPERSEDED", "EXPIRED"]);
  });

  it("hides Add rate without manage Allocation and shows it with it", () => {
    canManage = false;
    const { rerender } = renderDrawer();
    expect(screen.queryByRole("button", { name: "+ Add rate" })).not.toBeInTheDocument();

    canManage = true;
    rerender(
      <ResourcesRatesDrawer
        resourceId="res-1"
        resourceName="ClusterA CPU"
        onOpenChange={vi.fn()}
      />,
    );
    expect(screen.getByRole("button", { name: "+ Add rate" })).toBeInTheDocument();
  });

  it("rejects end on or before start inline without posting", async () => {
    renderDrawer();
    openForm();
    fireEvent.change(screen.getByLabelText("Rate"), { target: { value: "0.07" } });
    fireEvent.change(screen.getByLabelText("Ends"), {
      target: { value: (screen.getByLabelText("Effective date") as HTMLInputElement).value },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add rate" }));
    expect(await screen.findByText(/after the start date/i)).toBeInTheDocument();
    expect(mutation.mutate).not.toHaveBeenCalled();
  });

  it("posts a today rate with a now start_time, toasts and collapses", async () => {
    renderDrawer();
    openForm();
    fireEvent.change(screen.getByLabelText("Rate"), { target: { value: "0.07" } });
    fireEvent.click(screen.getByRole("button", { name: "Add rate" }));

    await waitFor(() => expect(mutation.mutate).toHaveBeenCalledTimes(1));
    const payload = mutation.mutate.mock.calls[0]?.[0].body;
    expect(payload?.compute_allocation_resource_id).toBe("res-1");
    expect(payload?.rate).toBe(0.07);
    // Today -> current wall-clock instant, not midnight.
    expect(payload?.start_time).not.toBe(`${todayUtc()}T00:00:00Z`);
    expect(payload?.start_time).toMatch(/T\d\d:\d\d/);
    expect(toast.success).toHaveBeenCalledWith("Rate added");
    await waitFor(() =>
      expect(screen.queryByRole("form", { name: "Add rate form" })).not.toBeInTheDocument(),
    );
  });
});
