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
import { beforeEach, describe, expect, it, vi } from "vitest";
import { TraceTable } from "@/features/core/audit/components/TraceTable";
import type { Trace } from "@/features/core/audit/types";

const baseTrace = (over: Partial<Trace>): Trace => ({
  trace_id: "0123456789abcdef0123456789abcdef",
  root_name: "amie.process_event:request_account_create",
  source: "amie",
  status: 0,
  started_at: new Date(Date.now() - 60_000).toISOString(),
  ended_at: new Date().toISOString(),
  span_count: 4,
  root_event: null,
  ...over,
});

describe("TraceTable", () => {
  beforeEach(() => {
    Object.assign(navigator, { clipboard: { writeText: vi.fn().mockResolvedValue(undefined) } });
  });

  it("renders header and row columns", () => {
    render(
      <TraceTable
        traces={[baseTrace({ trace_id: "a".repeat(32) })]}
        total={1}
        page={1}
        pageSize={50}
        onView={() => {}}
        onPageChange={() => {}}
        onPageSizeChange={() => {}}
      />,
    );
    expect(screen.getByText("Started")).toBeInTheDocument();
    expect(screen.getByText("Trace ID")).toBeInTheDocument();
    expect(screen.getByText("Root action")).toBeInTheDocument();
    expect(screen.getByText("Source")).toBeInTheDocument();
    expect(screen.getByText("Spans")).toBeInTheDocument();
  });

  it("error rows include the red left rail", () => {
    const errored = baseTrace({
      trace_id: "b".repeat(32),
      status: 1,
      root_event: { error: "ComanageProvisioningFailed: 404" },
    });
    render(
      <TraceTable
        traces={[errored]}
        total={1}
        page={1}
        pageSize={50}
        onView={() => {}}
        onPageChange={() => {}}
        onPageSizeChange={() => {}}
      />,
    );
    const row = screen.getByTestId(`trace-row-${errored.trace_id}`);
    expect(row.getAttribute("data-tone")).toBe("error");
    expect(row.querySelector('[data-testid="error-rail"]')).not.toBeNull();
    expect(screen.getByText(/ComanageProvisioningFailed/)).toBeInTheDocument();
  });

  it("clicking the row triggers onView", () => {
    const onView = vi.fn();
    const t = baseTrace({ trace_id: "c".repeat(32) });
    render(
      <TraceTable
        traces={[t]}
        total={1}
        page={1}
        pageSize={50}
        onView={onView}
        onPageChange={() => {}}
        onPageSizeChange={() => {}}
      />,
    );
    fireEvent.click(screen.getByTestId(`trace-row-${t.trace_id}`));
    expect(onView).toHaveBeenCalledWith(t.trace_id);
  });

  it("clicking the trace ID copy button does NOT trigger onView", () => {
    const onView = vi.fn();
    const t = baseTrace({ trace_id: "d".repeat(32) });
    render(
      <TraceTable
        traces={[t]}
        total={1}
        page={1}
        pageSize={50}
        onView={onView}
        onPageChange={() => {}}
        onPageSizeChange={() => {}}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /copy trace ID/i }));
    expect(onView).not.toHaveBeenCalled();
  });

  it("disables Prev on page 1 and Next on the final page", () => {
    render(
      <TraceTable
        traces={[baseTrace({ trace_id: "e".repeat(32) })]}
        total={1}
        page={1}
        pageSize={50}
        onView={() => {}}
        onPageChange={() => {}}
        onPageSizeChange={() => {}}
      />,
    );
    expect(screen.getByRole("button", { name: /previous page/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /next page/i })).toBeDisabled();
  });

  it("enables Next when more pages exist", () => {
    render(
      <TraceTable
        traces={[baseTrace({ trace_id: "f".repeat(32) })]}
        total={120}
        page={1}
        pageSize={50}
        onView={() => {}}
        onPageChange={() => {}}
        onPageSizeChange={() => {}}
      />,
    );
    expect(screen.getByRole("button", { name: /next page/i })).not.toBeDisabled();
    expect(screen.getByRole("button", { name: /previous page/i })).toBeDisabled();
  });

  it("renders a loading skeleton when loading", () => {
    render(
      <TraceTable
        traces={[]}
        total={0}
        page={1}
        pageSize={50}
        loading
        onView={() => {}}
        onPageChange={() => {}}
        onPageSizeChange={() => {}}
      />,
    );
    expect(screen.getByTestId("trace-table-loading")).toBeInTheDocument();
  });

  it("shows the active-filters empty copy when hasActiveFilters", () => {
    render(
      <TraceTable
        traces={[]}
        total={0}
        page={1}
        pageSize={50}
        hasActiveFilters
        onView={() => {}}
        onPageChange={() => {}}
        onPageSizeChange={() => {}}
      />,
    );
    expect(screen.getByTestId("trace-table-empty").textContent).toMatch(/No traces match/i);
  });

  it("shows the unfiltered empty copy when no filters", () => {
    render(
      <TraceTable
        traces={[]}
        total={0}
        page={1}
        pageSize={50}
        onView={() => {}}
        onPageChange={() => {}}
        onPageSizeChange={() => {}}
      />,
    );
    expect(screen.getByTestId("trace-table-empty").textContent).toMatch(/No traces yet/i);
  });
});
