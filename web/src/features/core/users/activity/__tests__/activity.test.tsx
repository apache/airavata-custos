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

import { activityAnalytics, activityList } from "@/mocks/handlers/user-activity";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { getUserActivity } from "../api";
import { StatusComposition } from "../components/ActivityOverview";
import { ActivityPage } from "../components/ActivityPage";
import { ActivityStatus } from "../components/ActivityStatus";
import { ActivityTable } from "../components/ActivityTable";
import { DaysRangePicker } from "../components/DaysRangePicker";
import {
  activityActionLabel,
  activityStatus,
  formatLastLogin,
  utcCalendarDaysBetween,
} from "../lib";

const state = vi.hoisted(() => ({ allowed: true }));
vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({ can: () => state.allowed, cannot: () => !state.allowed }),
}));
vi.mock("next-auth/react", () => ({ useSession: () => ({ status: "authenticated" }) }));
vi.mock("recharts", async () => {
  const actual = await vi.importActual<typeof import("recharts")>("recharts");
  return {
    ...actual,
    ResponsiveContainer: ({ children }: { children: React.ReactNode }) => <div>{children}</div>,
  };
});
const fetcher = vi.fn();
beforeEach(() => {
  state.allowed = true;
  vi.stubGlobal("fetch", fetcher);
  fetcher.mockImplementation(async (input: string) => {
    const url = new URL(input, "http://localhost");
    const body = url.pathname.endsWith("analytics")
      ? activityAnalytics(
          Number(url.searchParams.get("window")),
          url.pathname.match(/users\/(activity-\d+)\//)?.[1],
        )
      : activityList(url);
    return new Response(JSON.stringify(body), { headers: { "content-type": "application/json" } });
  });
});
afterEach(() => {
  fetcher.mockReset();
  vi.unstubAllGlobals();
});
function dashboard() {
  return render(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <ActivityPage />
    </QueryClientProvider>,
  );
}

describe("activity dashboard", () => {
  it("renders API summary details and recomputes the comparison when the window changes", async () => {
    dashboard();
    await screen.findByText("1–10 of 225");
    const active = await screen.findByRole("button", { name: /^Active/ });
    expect(active).toHaveTextContent("-25 vs prior 30 days");
    expect(screen.getByRole("button", { name: /^Dormant/ })).toHaveTextContent("25 over 90 days");
    expect(screen.getByRole("button", { name: /^Never signed in/ })).toHaveTextContent(
      "oldest account created 324 days ago",
    );
    fireEvent.click(
      within(screen.getByRole("group", { name: "Activity window" })).getByRole("button", {
        name: "7 days",
      }),
    );
    await waitFor(() =>
      expect(screen.getByRole("button", { name: /^Active/ })).toHaveTextContent("vs prior 7 days"),
    );
  });
  it("does not fabricate missing comparison or creation data from a legacy API", async () => {
    fetcher.mockImplementation(async (input: string) => {
      const url = new URL(input, "http://localhost");
      const { prior_active_users, dormant_over_90_days, oldest_never_created_at, ...legacy } =
        activityAnalytics(30);
      return new Response(
        JSON.stringify(url.pathname.endsWith("analytics") ? legacy : activityList(url)),
        { headers: { "content-type": "application/json" } },
      );
    });
    dashboard();
    await screen.findByText("1–10 of 225");
    expect(screen.queryByText(/vs prior/)).not.toBeInTheDocument();
    expect(screen.queryByText(/oldest account/)).not.toBeInTheDocument();
  });
  it("does not fabricate null comparison or creation data", async () => {
    fetcher.mockImplementation(async (input: string) => {
      const url = new URL(input, "http://localhost");
      return new Response(
        JSON.stringify(
          url.pathname.endsWith("analytics")
            ? {
                ...activityAnalytics(30),
                prior_active_users: null,
                dormant_over_90_days: null,
                oldest_never_created_at: null,
              }
            : activityList(url),
        ),
        { headers: { "content-type": "application/json" } },
      );
    });
    dashboard();
    await screen.findByText("1–10 of 225");
    expect(screen.queryByText(/vs prior/)).not.toBeInTheDocument();
    expect(screen.queryByText(/over 90 days/)).not.toBeInTheDocument();
    expect(screen.queryByText(/oldest account/)).not.toBeInTheDocument();
  });
  it("hides a zero over-90-day subcount instead of showing 0 over 90 days", async () => {
    fetcher.mockImplementation(async (input: string) => {
      const url = new URL(input, "http://localhost");
      return new Response(
        JSON.stringify(
          url.pathname.endsWith("analytics")
            ? { ...activityAnalytics(30), dormant_over_90_days: 0 }
            : activityList(url),
        ),
        { headers: { "content-type": "application/json" } },
      );
    });
    dashboard();
    await screen.findByText("1–10 of 225");
    expect(screen.getByRole("button", { name: /^Dormant/ })).not.toHaveTextContent("over 90 days");
  });
  it("hides the over-90-day subcount when the selected window is 90 days or longer", async () => {
    dashboard();
    await screen.findByText("1–10 of 225");
    expect(screen.getByRole("button", { name: /^Dormant/ })).toHaveTextContent("25 over 90 days");
    fireEvent.click(
      within(screen.getByRole("group", { name: "Activity window" })).getByRole("button", {
        name: "90 days",
      }),
    );
    await waitFor(() =>
      expect(screen.getByRole("button", { name: /^Active/ })).toHaveTextContent("vs prior 90 days"),
    );
    expect(screen.getByRole("button", { name: /^Dormant/ })).not.toHaveTextContent("over 90 days");
  });
  it("opens Review access as an engagement audit, not an IAM panel", async () => {
    dashboard();
    await screen.findByText("1–10 of 225");
    fireEvent.change(screen.getByLabelText("Filter users by status"), {
      target: { value: "never" },
    });
    await screen.findByText("1–10 of 25");
    const reviewButtons = await screen.findAllByRole("button", { name: /Review access for/ });
    const reviewButton = reviewButtons[0];
    if (!reviewButton) throw new Error("Missing Review access action");
    fireEvent.click(reviewButton);
    expect(await screen.findByRole("dialog")).toHaveTextContent("Review access");
    expect(screen.getByRole("dialog")).toHaveTextContent(
      "Sign-in activity only; this does not change roles or cluster access.",
    );
    expect(await screen.findByText("Lifetime active days")).toBeInTheDocument();
  });
  it("renders status dots for dormant and never-signed-in pills", () => {
    const user = activityList(new URL("http://localhost")).items[0];
    if (!user) throw new Error("Missing fixture row");
    const { rerender, container } = render(
      <ActivityStatus
        user={{ ...user, last_login: "2026-08-01T12:00:00Z", inactive_days: 30 }}
        windowDays={30}
      />,
    );
    expect(container.querySelector("[aria-hidden='true']")).not.toBeNull();
    expect(screen.getByText("Dormant")).toBeInTheDocument();
    rerender(
      <ActivityStatus user={{ ...user, last_login: null, inactive_days: null }} windowDays={30} />,
    );
    expect(container.querySelector("[aria-hidden='true']")).not.toBeNull();
    expect(screen.getByText("Never signed in")).toBeInTheDocument();
  });
  it("omits the role suffix when role_names is missing or empty", () => {
    const user = activityList(new URL("http://localhost")).items[0];
    if (!user) throw new Error("Missing fixture row");
    const noop = () => undefined;
    const { rerender } = render(
      <ActivityTable
        rows={[{ ...user, email: "noroles@example.org", role_names: undefined }]}
        total={1}
        windowDays={30}
        page={1}
        pageSize={10}
        sort="last_login"
        direction="desc"
        onSort={noop}
        onPage={noop}
        onPageSize={noop}
        onSelect={noop}
      />,
    );
    expect(screen.getByText("noroles@example.org")).toBeInTheDocument();
    expect(screen.queryByText(/noroles@example.org ·/)).not.toBeInTheDocument();
    rerender(
      <ActivityTable
        rows={[{ ...user, email: "noroles@example.org", role_names: [] }]}
        total={1}
        windowDays={30}
        page={1}
        pageSize={10}
        sort="last_login"
        direction="desc"
        onSort={noop}
        onPage={noop}
        onPageSize={noop}
        onSelect={noop}
      />,
    );
    expect(screen.getByText("noroles@example.org")).toBeInTheDocument();
    expect(screen.queryByText(/noroles@example.org ·/)).not.toBeInTheDocument();
  });
  it("uses server totals and requests page 21 without a 200-user ceiling", async () => {
    dashboard();
    expect(await screen.findByText("1–10 of 225")).toBeInTheDocument();
    for (let page = 2; page <= 21; page++) {
      fireEvent.click(screen.getByRole("button", { name: "Next" }));
      await waitFor(() => expect(screen.getByRole("button", { name: "Next" })).not.toBeDisabled());
      await screen.findByText(`${(page - 1) * 10 + 1}–${page * 10} of 225`);
    }
    expect(fetcher.mock.calls.some(([url]) => String(url).includes("offset=200"))).toBe(true);
  });
  it("sends status, window and debounced search to the server and resets pagination", async () => {
    dashboard();
    await screen.findByText("1–10 of 225");
    fireEvent.click(screen.getByRole("button", { name: "Next" }));
    await screen.findByText("11–20 of 225");
    fireEvent.change(screen.getByLabelText("Filter users by status"), {
      target: { value: "dormant" },
    });
    await screen.findByText("1–10 of 75");
    fireEvent.click(
      within(screen.getByRole("group", { name: "Activity window" })).getByRole("button", {
        name: "7 days",
      }),
    );
    await screen.findByText("1–10 of 125");
    fireEvent.change(screen.getByLabelText("Search users"), {
      target: { value: "activity4@example.org" },
    });
    await screen.findByText("1–1 of 1");
    expect(screen.getByText("Activity User 004")).toBeInTheDocument();
    expect(
      fetcher.mock.calls.some(([url]) =>
        String(url).includes(
          "window=7&status=dormant&query=activity4%40example.org&limit=10&offset=0",
        ),
      ),
    ).toBe(true);
  });
  it("opens individual analytics and requests its independent range", async () => {
    dashboard();
    await screen.findByText("1–10 of 225");
    const button = screen.getAllByRole("button", { name: /View activity for/ })[0];
    if (!button) throw new Error("Missing activity action");
    fireEvent.click(button);
    expect(await screen.findByRole("dialog")).toBeInTheDocument();
    fireEvent.click(
      within(screen.getByRole("group", { name: "User analytics date range" })).getByRole("button", {
        name: "90 days",
      }),
    );
    await waitFor(() =>
      expect(
        fetcher.mock.calls.some(([url]) =>
          /users\/activity-\d+\/activity\/analytics\?window=90/.test(String(url)),
        ),
      ).toBe(true),
    );
  });
  it("denies direct navigation without making activity requests", () => {
    state.allowed = false;
    dashboard();
    expect(screen.getByText("Not permitted.")).toBeInTheDocument();
    expect(fetcher).not.toHaveBeenCalled();
  });
  it("keeps controls available after failure and retries", async () => {
    fetcher.mockResolvedValueOnce(new Response("{}", { status: 500 }));
    dashboard();
    await screen.findByText("We couldn't load the selected users.");
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    await screen.findByText("1–10 of 225");
  });
  it("keeps the current page on screen while the next page loads", async () => {
    dashboard();
    await screen.findByText("1–10 of 225");
    const firstRow = screen.getByText("Activity User 001");
    fetcher.mockImplementation(() => new Promise(() => {}));
    fireEvent.click(screen.getByRole("button", { name: "Next" }));
    // Swapping the table for a short skeleton collapses the page height and
    // yanks the trend chart above it; paging keeps the same columns, so the
    // rows stay put and only report that they are stale.
    expect(screen.getByRole("table")).toBeInTheDocument();
    expect(screen.queryByLabelText("Loading activity table")).not.toBeInTheDocument();
    expect(firstRow).toBeInTheDocument();
    expect(screen.getByRole("table").closest("[aria-busy]")).toHaveAttribute("aria-busy", "true");
    expect(screen.getByLabelText("Current page")).toHaveTextContent("Loading page 2 of 23");
    expect(screen.getByRole("button", { name: "Next" })).toBeDisabled();
    expect(
      screen.getByRole("button", { name: /View activity for Activity User 001/ }),
    ).toBeDisabled();
  });
  it("does not show previous-window rows while the next request is pending", async () => {
    dashboard();
    await screen.findByText("1–10 of 225");
    fetcher.mockImplementation(() => new Promise(() => {}));
    fireEvent.click(
      within(screen.getByRole("group", { name: "Activity window" })).getByRole("button", {
        name: "90 days",
      }),
    );
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.getByLabelText("Loading activity table")).toBeInTheDocument();
  });
  it("shows an empty search result", async () => {
    dashboard();
    await screen.findByText("1–10 of 225");
    fireEvent.change(screen.getByLabelText("Search users"), {
      target: { value: "no-such-person" },
    });
    await screen.findByText("No users match this view.");
  });
});
it("rejects an older backend that ignores the requested status/window", async () => {
  const result = activityList(new URL("http://localhost?window=30&status=all"));
  fetcher.mockResolvedValueOnce(
    new Response(JSON.stringify(result), { headers: { "content-type": "application/json" } }),
  );
  await expect(
    getUserActivity({
      window: 7,
      status: "active",
      query: "",
      limit: 10,
      offset: 0,
      sort: "name",
      direction: "asc",
    }),
  ).rejects.toThrow(/filters/);
});
it("rejects missing window counts instead of inventing zeros", async () => {
  const result = activityList(new URL("http://localhost?window=30&status=all"));
  const first = result.items[0];
  if (!first) throw new Error("Missing fixture row");
  const { window_login_count, ...oldRow } = first;
  fetcher.mockResolvedValueOnce(
    new Response(JSON.stringify({ ...result, items: [oldRow] }), {
      headers: { "content-type": "application/json" },
    }),
  );
  await expect(
    getUserActivity({
      window: 30,
      status: "all",
      query: "",
      limit: 10,
      offset: 0,
      sort: "name",
      direction: "asc",
    }),
  ).rejects.toThrow();
});
it("uses server-local calendar days at the exact inactivity boundary", () => {
  const user = activityList(new URL("http://localhost")).items[0];
  if (!user) throw new Error("Missing fixture row");
  expect(activityStatus({ ...user, inactive_days: 6 }, 7)).toBe("active");
  expect(formatLastLogin({ ...user, inactive_days: -1 })).toBe("Today");
  expect(activityStatus({ ...user, inactive_days: 7 }, 7)).toBe("dormant");
  expect(activityStatus({ ...user, last_login: null, inactive_days: null }, 7)).toBe("never");
  expect(activityActionLabel({ ...user, last_login: null, inactive_days: null }, 30)).toBe(
    "Review access",
  );
  expect(activityActionLabel({ ...user, inactive_days: 0 }, 30)).toBe("View");
});
it("counts oldest-account age in UTC calendar days, not elapsed hours", () => {
  expect(utcCalendarDaysBetween("2026-09-16T00:30:00Z", "2026-09-15T23:30:00Z")).toBe(1);
  expect(utcCalendarDaysBetween("2026-09-16T12:00:00Z", "2025-10-27T12:00:00Z")).toBe(324);
});
it("validates custom days without submitting invalid values", () => {
  const onChange = vi.fn();
  render(<DaysRangePicker value={30} onChange={onChange} label="Window" />);
  fireEvent.click(screen.getByRole("button", { name: "Custom" }));
  const input = screen.getByLabelText("Window custom days");
  for (const value of ["0", "366", "1.5", ""]) {
    fireEvent.change(input, { target: { value } });
    fireEvent.blur(input);
  }
  expect(onChange).not.toHaveBeenCalled();
  expect(screen.getByRole("alert")).toBeInTheDocument();
  fireEvent.change(input, { target: { value: "365" } });
  fireEvent.blur(input);
  expect(onChange).toHaveBeenCalledWith(365);
});
it("renders zero-user composition without NaN or Infinity", () => {
  const { container } = render(<StatusComposition counts={{ active: 0, dormant: 0, never: 0 }} />);
  expect(container.innerHTML).not.toMatch(/NaN|Infinity/);
});
