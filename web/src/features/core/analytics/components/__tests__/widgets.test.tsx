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

import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type {
  AnalyticsAllocation,
  AnalyticsContext,
  UsageMember,
  UsageResource,
} from "../../schemas";
import { ContextSwitcher } from "../ContextSwitcher";
import { HeroTiles } from "../HeroTiles";
import { MemberBreakdown } from "../MemberBreakdown";
import { ResourceBreakdown } from "../ResourceBreakdown";

const NOW = new Date("2026-07-14T12:00:00Z");

function alloc(overrides: Partial<AnalyticsAllocation> = {}): AnalyticsAllocation {
  return {
    id: "a1",
    name: "Alloc One",
    status: "ACTIVE",
    initial_su_amount: 1000,
    used_su_amount: 200,
    start_time: "2026-06-14T12:00:00Z",
    end_time: "2026-08-14T12:00:00Z",
    ...overrides,
  };
}

function ctx(role: AnalyticsContext["role"], allocations: AnalyticsAllocation[]): AnalyticsContext {
  return {
    project_id: "p1",
    project_name: "Project One",
    role,
    allocations,
  };
}

function resource(id: string, used: number, usedByCaller: number): UsageResource {
  return {
    resource_id: id,
    name: id.toUpperCase(),
    resource_type: "GPU_HOURS",
    used,
    cap: null,
    used_native: used,
    native_unit: "GPU-hours",
    used_by_caller: usedByCaller,
  };
}

describe("HeroTiles", () => {
  it("renders the tiles and a critical pill when credits are nearly gone", () => {
    render(
      <HeroTiles
        allocation={alloc({ initial_su_amount: 1000, used_su_amount: 970 })}
        callerUsed={30}
        now={NOW}
      />,
    );
    expect(screen.getByText("Credits left")).toBeInTheDocument();
    expect(screen.getByText("Time remaining")).toBeInTheDocument();
    // 3% left is <= the red threshold.
    expect(screen.getByText("Critical")).toBeInTheDocument();
  });

  it("shows the project total used with the caller's own share on the subline", () => {
    render(
      <HeroTiles
        allocation={alloc({ initial_su_amount: 1000, used_su_amount: 800 })}
        callerUsed={120}
        now={NOW}
      />,
    );
    // The tile value is the project total; "your share" is the subline.
    expect(screen.getByText("your share: 120")).toBeInTheDocument();
  });
});

describe("ContextSwitcher", () => {
  it("renders a static label when there is only one allocation", () => {
    render(
      <ContextSwitcher
        contexts={[ctx("MEMBER", [alloc()])]}
        selectedProjectId="p1"
        selectedAllocationId="a1"
        onSelect={vi.fn()}
      />,
    );
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    // Allocation-first label, plus the project name alongside.
    expect(screen.getByText(/Alloc One/)).toBeInTheDocument();
    expect(screen.getByText("Project One")).toBeInTheDocument();
  });

  it("opens a menu and selects an allocation when there are several", () => {
    const onSelect = vi.fn();
    render(
      <ContextSwitcher
        contexts={[
          ctx("PI", [alloc({ id: "a1", name: "First" }), alloc({ id: "a2", name: "Second" })]),
        ]}
        selectedProjectId="p1"
        selectedAllocationId="a1"
        onSelect={onSelect}
      />,
    );
    fireEvent.click(screen.getByRole("button"));
    fireEvent.click(screen.getByText("Second"));
    expect(onSelect).toHaveBeenCalledWith("p1", "a2");
  });
});

describe("resource and member breakdowns", () => {
  const summary = {
    total: 1000,
    used: 300,
    daily: [],
    by_resource: [resource("gpu", 200, 40), resource("cpu", 100, 90)] as UsageResource[],
    by_member: null,
  };

  it("resource view shares are against the allocation budget, with the remainder", () => {
    render(<ResourceBreakdown summary={summary} />);
    // gpu is 200 of the 1000 budget, not 67% of the 300 consumed.
    expect(screen.getByText(/20% · 200/)).toBeInTheDocument();
    expect(screen.getByText(/10% · 100/)).toBeInTheDocument();
    // The unused remainder is a legend row of its own.
    expect(screen.getByText("Available")).toBeInTheDocument();
    expect(screen.getByText(/70% · 700/)).toBeInTheDocument();
  });

  it("flips to an over-budget center and drops the remainder when usage exceeds the budget", () => {
    const overrun = {
      ...summary,
      total: 100,
      by_resource: [resource("gpu", 150, 0), resource("cpu", 50, 0)] as UsageResource[],
    };
    render(<ResourceBreakdown summary={overrun} />);
    // Shares scale to the 200 consumed; the center shows how far over.
    expect(screen.getByText(/75% · 150/)).toBeInTheDocument();
    expect(screen.getByText("-100")).toBeInTheDocument();
    expect(screen.queryByText("Available")).not.toBeInTheDocument();
  });

  it("renders a full available ring when nothing is used yet", () => {
    const untouched = { ...summary, used: 0, by_resource: [] as UsageResource[] };
    render(<ResourceBreakdown summary={untouched} />);
    expect(screen.getByText("Available")).toBeInTheDocument();
    expect(screen.getByText(/of 1K available/)).toBeInTheDocument();
  });

  it("member breakdown lists every member and folds the tail into one ring arc", () => {
    const many: UsageMember[] = [
      { user_id: "u1", name: "A A", used: 600 },
      { user_id: "u2", name: "B B", used: 500 },
      { user_id: "u3", name: "C C", used: 400 },
      { user_id: "u4", name: "D D", used: 300 },
      { user_id: "u5", name: "E E", used: 200 },
      { user_id: "u6", name: "F F", used: 100 },
      { user_id: "u7", name: "G G", used: 50 },
    ];
    const { container } = render(<MemberBreakdown members={many} total={4300} />);
    // Every member appears in the legend, including the folded tail.
    expect(screen.getByText("A A")).toBeInTheDocument();
    expect(screen.getByText("F F")).toBeInTheDocument();
    expect(screen.getByText("G G")).toBeInTheDocument();
    // The tail shares a single ring arc.
    expect(container.querySelector('[data-slice="__others"]')).toBeInTheDocument();
    // 2150 of the 4300 budget remains.
    expect(screen.getByText(/50% · 2.15K/)).toBeInTheDocument();
  });

  it("hovering a sector swaps the donut center to that slice", () => {
    const { container } = render(<ResourceBreakdown summary={summary} />);
    const arc = container.querySelector('[data-slice="gpu"]') as Element;
    fireEvent.mouseEnter(arc);
    // Center shows the hovered slice's share of the budget.
    expect(screen.getByText(/20% · of 1K/)).toBeInTheDocument();
    fireEvent.mouseLeave(arc);
    expect(screen.getByText(/of 1K available/)).toBeInTheDocument();
  });
});
