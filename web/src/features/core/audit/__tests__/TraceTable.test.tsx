/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import tracesListFixture from "@/features/core/audit/__fixtures__/traces.list.json";
import { TraceTable, type TraceTableProps } from "@/features/core/audit/components/TraceTable";
import { traceListSchema } from "@/features/core/audit/schemas";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const list = traceListSchema.parse(tracesListFixture);
const byStatus = (status: string) => {
  const trace = list.traces.find((t) => t.status === status);
  if (!trace) throw new Error(`fixture has no ${status} trace`);
  return trace;
};
const failed = byStatus("error");
const retrying = byStatus("in_progress");

function renderTable(over: Partial<TraceTableProps> = {}) {
  const onView = vi.fn();
  render(
    <TraceTable
      traces={[failed, retrying]}
      total={2}
      page={1}
      pageSize={50}
      windowLabel="7 days"
      onView={onView}
      onPageChange={() => {}}
      onPageSizeChange={() => {}}
      onClearFilters={() => {}}
      {...over}
    />,
  );
  return { onView };
}

describe("TraceTable", () => {
  beforeEach(() => {
    Object.assign(navigator, { clipboard: { writeText: vi.fn().mockResolvedValue(undefined) } });
  });

  // Make sure a row says in words what is wrong: the status word with the try
  // count, and the delivery count in red when one failed.
  it("shows the status word and delivery count per row", () => {
    renderTable();
    expect(screen.getByText("Failed")).toBeInTheDocument();
    expect(screen.getByText("Retrying 4 of 10")).toBeInTheDocument();
    const failedRow = screen.getByTestId(`trace-row-${failed.trace_id}`);
    const retryingRow = screen.getByTestId(`trace-row-${retrying.trace_id}`);
    expect(within(failedRow).getByText("1 of 2 done")).toBeInTheDocument();
    expect(failedRow.querySelector('[data-testid="error-rail"]')).not.toBeNull();
    expect(retryingRow.querySelector('[data-testid="error-rail"]')).toBeNull();
  });

  // Make sure copying the trace id from a row does not also open the drawer.
  it("opens on row click but not on the copy button", () => {
    const { onView } = renderTable();
    const failedRow = screen.getByTestId(`trace-row-${failed.trace_id}`);
    fireEvent.click(within(failedRow).getByRole("button", { name: "Copy trace ID" }));
    expect(onView).not.toHaveBeenCalled();
    fireEvent.click(screen.getByTestId(`trace-row-${failed.trace_id}`));
    expect(onView).toHaveBeenCalledWith(failed.trace_id);
  });

  // Make sure an empty list says whether filters hid the rows, with a way
  // out, or whether there is simply nothing in the window.
  it("tells filtered and unfiltered empty states apart", () => {
    const onClearFilters = vi.fn();
    const { unmount } = render(
      <TraceTable
        traces={[]}
        total={0}
        page={1}
        pageSize={50}
        windowLabel="7 days"
        hasActiveFilters
        onView={() => {}}
        onPageChange={() => {}}
        onPageSizeChange={() => {}}
        onClearFilters={onClearFilters}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Clear filters" }));
    expect(onClearFilters).toHaveBeenCalled();
    unmount();
    renderTable({ traces: [], total: 0 });
    expect(screen.getByText("No traces in the last 7 days.")).toBeInTheDocument();
  });
});
