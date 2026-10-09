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

import { TraceTable } from "@/features/core/audit/components/TraceTable";
import type { TraceSummary } from "@/generated/core/types.gen";
import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const trace: TraceSummary = {
  trace_id: "c".repeat(32),
  root_operation: "amie.process_event:request_account_create",
  source: "amie",
  status: "ok",
  started_at: new Date(Date.now() - 60_000).toISOString(),
  ended_at: new Date().toISOString(),
  event_count: 4,
  deliveries: { pending: 0, succeeded: 0, failed: 0, attempts: 0 },
};

function renderTable(onView: (id: string) => void) {
  render(
    <TraceTable
      traces={[trace]}
      total={1}
      page={1}
      pageSize={50}
      onView={onView}
      onPageChange={() => {}}
      onPageSizeChange={() => {}}
    />,
  );
}

describe("TraceTable", () => {
  beforeEach(() => {
    Object.assign(navigator, { clipboard: { writeText: vi.fn().mockResolvedValue(undefined) } });
  });

  it("clicking the row triggers onView", () => {
    const onView = vi.fn();
    renderTable(onView);
    fireEvent.click(screen.getByTestId(`trace-row-${trace.trace_id}`));
    expect(onView).toHaveBeenCalledWith(trace.trace_id);
  });

  it("clicking the trace ID copy button does NOT trigger onView", () => {
    const onView = vi.fn();
    renderTable(onView);
    fireEvent.click(screen.getByRole("button", { name: /copy trace ID/i }));
    expect(onView).not.toHaveBeenCalled();
  });
});
