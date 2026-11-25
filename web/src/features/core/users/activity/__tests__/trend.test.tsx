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
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { LoginTrend } from "../components/ActivityOverview";
import { completeActivityTrend } from "../lib";
import type { UserActivityAnalytics } from "../schemas";

vi.mock("recharts", async () => {
  const actual = await vi.importActual<typeof import("recharts")>("recharts");
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: ReactNode }) => <div>{children}</div>,
    ComposedChart: ({ data }: { data: unknown }) => (
      <output data-testid="chart-data">{JSON.stringify(data)}</output>
    ),
  };
});
function analytics(overrides: Partial<UserActivityAnalytics> = {}): UserActivityAnalytics {
  return {
    generated_at: "2026-09-10T12:00:00Z",
    window_days: 10,
    total_users: 2,
    users_ever_logged_in: 2,
    active_users: 2,
    lifetime_login_count: 6,
    lifetime_active_days: 3,
    window_login_count: 6,
    window_active_days: 3,
    trend: [
      { date: "2026-09-10", login_count: 4, active_users: 2 },
      { date: "2026-09-01", login_count: 2, active_users: 1 },
    ],
    ...overrides,
  };
}
describe("completeActivityTrend", () => {
  it("sorts sparse data and retains every zero day between September 1 and 10", () => {
    const source = analytics();
    const before = JSON.stringify(source);
    const rows = completeActivityTrend(source);
    expect(rows).toHaveLength(10);
    expect(rows.map((row) => row.date)).toEqual(
      Array.from({ length: 10 }, (_, i) => `2026-09-${String(i + 1).padStart(2, "0")}`),
    );
    expect(rows.slice(1, 9).every((row) => row.login_count === 0 && row.active_users === 0)).toBe(
      true,
    );
    expect(rows.reduce((sum, row) => sum + row.login_count, 0)).toBe(6);
    expect(JSON.stringify(source)).toBe(before);
  });
  it("fills leading and trailing empty days even if the last sign-in was earlier", () => {
    const rows = completeActivityTrend(
      analytics({
        window_days: 3,
        trend: [{ date: "2026-09-09", login_count: 2, active_users: 1 }],
      }),
    );
    expect(rows.map((row) => [row.date, row.login_count])).toEqual([
      ["2026-09-08", 0],
      ["2026-09-09", 2],
      ["2026-09-10", 0],
    ]);
  });
  it("crosses leap day and month boundaries without fixed local-day durations", () => {
    const rows = completeActivityTrend(
      analytics({ generated_at: "2024-03-01T12:00:00Z", window_days: 3, trend: [] }),
    );
    expect(rows.map((row) => row.date)).toEqual(["2024-02-28", "2024-02-29", "2024-03-01"]);
  });
  it("uses the server instant across DST rather than the browser date", () => {
    const rows = completeActivityTrend(
      analytics({ generated_at: "2026-03-09T00:30:00+09:00", window_days: 3, trend: [] }),
    );
    expect(rows.map((row) => row.date)).toEqual(["2026-03-06", "2026-03-07", "2026-03-08"]);
  });
  it("preserves local-date boundary buckets on both sides of the UTC display window", () => {
    const rows = completeActivityTrend(
      analytics({
        window_days: 2,
        trend: [
          { date: "2026-09-08", login_count: 3, active_users: 1 },
          { date: "2026-09-11", login_count: 2, active_users: 1 },
        ],
      }),
    );
    expect(rows.map((row) => [row.date, row.login_count])).toEqual([
      ["2026-09-08", 3],
      ["2026-09-09", 0],
      ["2026-09-10", 0],
      ["2026-09-11", 2],
    ]);
  });
  it("supports the minimum and maximum custom windows across a year boundary", () => {
    expect(completeActivityTrend(analytics({ window_days: 1, trend: [] }))).toHaveLength(1);
    const rows = completeActivityTrend(
      analytics({ generated_at: "2026-01-01T12:00:00Z", window_days: 365, trend: [] }),
    );
    expect(rows).toHaveLength(365);
    expect(rows[0]?.date).toBe("2025-01-02");
    expect(rows.at(-1)?.date).toBe("2026-01-01");
  });
});
it("passes zero-filled points to the actual chart boundary", () => {
  render(<LoginTrend analytics={analytics()} />);
  const rows = JSON.parse(screen.getByTestId("chart-data").textContent ?? "[]");
  expect(rows).toHaveLength(10);
  expect(rows[4]).toEqual({ date: "2026-09-05", label: "09-05", loginSessions: 0, activeUsers: 0 });
});
it.each([{ trend: [] }, { trend: [{ date: "2026-09-10", login_count: 0, active_users: 0 }] }])(
  "shows an empty state for an all-zero trend",
  ({ trend }) => {
    render(<LoginTrend analytics={analytics({ trend })} />);
    expect(screen.getByText("No sign-ins in this period.")).toBeInTheDocument();
    expect(screen.queryByTestId("chart-data")).not.toBeInTheDocument();
    expect(screen.queryByText(/peak 0/)).not.toBeInTheDocument();
  },
);
