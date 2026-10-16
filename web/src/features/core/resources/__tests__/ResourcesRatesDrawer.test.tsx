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

import type { ComputeAllocationResourceRate as Rate } from "@/generated/core/types.gen";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

// Windows are anchored far in the past/future so rateStatus stays stable
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
  mutateAsync: vi.fn(async (_options: { body: Rate }) => ({})),
  isPending: false,
};

vi.mock("../queries", async (importOriginal) => {
  const { effectiveRate, rateStatus } = await importOriginal<typeof import("../queries")>();
  return {
    effectiveRate,
    rateStatus,
    useResourceRates: () => ({ data: rates, isLoading: false, error: null, refetch: vi.fn() }),
    useEffectiveRate: () => ({ data: effective, isLoading: false, error: null, refetch: vi.fn() }),
    useCreateResourceRate: () => mutation,
    useResource: () => ({ data: undefined, isLoading: false, error: null }),
    useResourceRate: () => ({ data: undefined, error: null }),
  };
});

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({ can: () => canManage }),
}));

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), dismiss: vi.fn() },
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
  mutation.mutateAsync.mockClear();
  vi.mocked(toast.success).mockClear();
});

describe("<ResourcesRatesDrawer />", () => {
  it("classifies each row and sorts start_time DESC", () => {
    effective = undefined; // keep "Active" unique to the table
    renderDrawer();
    const rows = screen.getAllByRole("row").slice(1); // drop header
    const badges = rows.map(
      (r) => within(r).getByText(/Active|Scheduled|Overridden|Expired/).textContent,
    );
    expect(badges).toEqual([
      expect.stringMatching(/^Scheduled from /),
      "Active",
      "Overridden",
      "Expired",
    ]);
  });

  it("hides Add rate without manage Allocation", () => {
    canManage = false;
    renderDrawer();
    expect(screen.queryByRole("button", { name: "+ Add rate" })).not.toBeInTheDocument();
  });

  it("posts a today rate with a now start_time, toasts and collapses", async () => {
    renderDrawer();
    openForm();
    fireEvent.change(screen.getByLabelText("Rate"), { target: { value: "0.07" } });
    fireEvent.click(screen.getByRole("button", { name: "Add rate" }));

    await waitFor(() => expect(mutation.mutateAsync).toHaveBeenCalledTimes(1));
    const payload = mutation.mutateAsync.mock.calls[0]?.[0].body;
    expect(payload?.compute_allocation_resource_id).toBe("res-1");
    expect(payload?.rate).toBe(0.07);
    // Today -> current wall-clock instant, not midnight.
    expect(payload?.start_time).not.toBe(`${todayUtc()}T00:00:00Z`);
    expect(payload?.start_time).toMatch(/T\d\d:\d\d/);
    await waitFor(() => expect(toast.success).toHaveBeenCalledWith("Rate added"));
    await waitFor(() =>
      expect(screen.queryByRole("form", { name: "Add rate form" })).not.toBeInTheDocument(),
    );
  });

  it("calls a past start immediate only when no later-starting rate stays in force", () => {
    renderDrawer();
    openForm();
    fireEvent.change(screen.getByLabelText("Effective date"), { target: { value: "2020-01-01" } });
    expect(screen.getByText(/becomes the active rate immediately/)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Effective date"), { target: { value: "2005-01-01" } });
    expect(screen.queryByText(/becomes the active rate immediately/)).not.toBeInTheDocument();
    expect(screen.getByText(/stays in force/)).toBeInTheDocument();
  });

  it("refuses an end on or before the start and an unparsable rate", async () => {
    renderDrawer();
    openForm();
    const ends = screen.getByLabelText("Ends") as HTMLInputElement;
    const start = (screen.getByLabelText("Effective date") as HTMLInputElement).value;
    fireEvent.change(ends, { target: { value: start } });
    expect(ends.validity.rangeUnderflow).toBe(true);
    fireEvent.change(ends, { target: { value: "2099-12-31" } });
    // jsdom has no badInput; a browser sets it for an entry it cannot parse, such as 1e309.
    const rate = screen.getByLabelText("Rate");
    fireEvent.change(rate, { target: { value: "5" } });
    Object.defineProperty(rate, "validity", { value: { badInput: true } });
    fireEvent.change(rate, { target: { value: "" } });
    fireEvent.click(screen.getByRole("button", { name: "Add rate" }));
    expect(await screen.findByText("Enter a valid rate.")).toBeInTheDocument();
    expect(mutation.mutateAsync).not.toHaveBeenCalled();
  });

  it("warns of the gap before a draft start only from now on, also for a first rate", () => {
    rates = [];
    effective = undefined;
    renderDrawer();
    openForm();
    fireEvent.change(screen.getByLabelText("Effective date"), { target: { value: "2005-01-01" } });
    expect(screen.queryByText(/^No rate is effective from/)).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Ends"), { target: { value: "2100-01-01" } });
    fireEvent.change(screen.getByLabelText("Effective date"), { target: { value: "2099-01-01" } });
    expect(
      screen.getByText(
        new RegExp(`^No rate is effective from .*${new Date().getUTCFullYear()} until .+2099\\.$`),
      ),
    ).toBeInTheDocument();
  });

  it("hides the preview and gap warnings while the end is not after the start", () => {
    rates = [];
    effective = undefined;
    renderDrawer();
    openForm();
    expect(screen.getByText(/^No rate is effective after/)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Ends"), { target: { value: "2005-01-01" } });
    expect(screen.queryByText(/never takes effect|immediately/)).not.toBeInTheDocument();
    expect(screen.queryByText(/^No rate is effective/)).not.toBeInTheDocument();
  });

  it("warns of a trailing gap only when no existing rate is in force at the draft end", () => {
    rates = [activeRate];
    const { rerender } = renderDrawer();
    openForm();
    expect(screen.queryByText(/^No rate is effective after/)).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Ends"), { target: { value: "2100-06-01" } });
    expect(screen.getByText(/^No rate is effective after/)).toBeInTheDocument();
    rates = [activeRate, rate("rate-later", 0.07, "2100-09-01T00:00:00Z", "2101-01-01T00:00:00Z")];
    rerender(
      <ResourcesRatesDrawer
        resourceId="res-1"
        resourceName="ClusterA CPU"
        onOpenChange={vi.fn()}
      />,
    );
    expect(screen.getByText(/^No rate is effective from .+ until .+2100\.$/)).toBeInTheDocument();
  });

  it("says the current rate comes back into force when it outlasts an immediate draft", () => {
    rates = [activeRate];
    renderDrawer();
    openForm();
    expect(
      screen.getByText(/The current rate \(0\.05\) comes back into force on/),
    ).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Ends"), { target: { value: "2100-06-01" } });
    expect(screen.getByText(/The current rate \(0\.05\) stays in history\./)).toBeInTheDocument();
  });

  it("names the current rate in a scheduled preview only while it lasts until the start", () => {
    const ending = rate("rate-ending", 0.05, "2010-01-01T00:00:00Z", "2050-01-01T00:00:00Z");
    rates = [ending];
    effective = ending;
    renderDrawer();
    openForm();
    fireEvent.change(screen.getByLabelText("Ends"), { target: { value: "2099-01-01" } });
    fireEvent.change(screen.getByLabelText("Effective date"), { target: { value: "2040-01-01" } });
    expect(
      screen.getByText(/The current rate \(0\.05\) stays active until then\./),
    ).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Effective date"), { target: { value: "2060-01-01" } });
    expect(screen.getByText(/^Scheduled to take effect on [^.]+\.$/)).toBeInTheDocument();
  });

  it("says the current rate resumes after a scheduled draft inside its window", () => {
    rates = [activeRate];
    renderDrawer();
    openForm();
    fireEvent.change(screen.getByLabelText("Ends"), { target: { value: "2099-01-01" } });
    fireEvent.change(screen.getByLabelText("Effective date"), { target: { value: "2098-01-01" } });
    expect(
      screen.getByText(
        /The current rate \(0\.05\) stays active until then and resumes on .+2099\./,
      ),
    ).toBeInTheDocument();
  });

  it("does not name the current rate when another rate is in force just before the draft takes over", () => {
    const ending = rate("rate-ending", 0.05, "2010-01-01T00:00:00Z", "2050-01-01T00:00:00Z");
    // Ties the draft's start with an earlier end, so the draft first wins in 2060.
    const tied = rate("rate-tied", 0.07, "2040-01-01T00:00:00Z", "2060-01-01T00:00:00Z");
    rates = [ending, tied];
    effective = ending;
    renderDrawer();
    openForm();
    fireEvent.change(screen.getByLabelText("Ends"), { target: { value: "2099-01-01" } });
    fireEvent.change(screen.getByLabelText("Effective date"), { target: { value: "2040-01-01" } });
    expect(screen.getByText(/^Scheduled to take effect on [^.]+\.$/)).toBeInTheDocument();
  });
});
