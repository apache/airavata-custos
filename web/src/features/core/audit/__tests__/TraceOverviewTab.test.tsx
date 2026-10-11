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
import { TraceOverviewTab } from "@/features/core/audit/components/TraceOverviewTab";
import type { Span, Trace } from "@/features/core/audit/types";

beforeEach(() => {
  Object.assign(navigator, { clipboard: { writeText: vi.fn().mockResolvedValue(undefined) } });
});

function buildTrace(over: Partial<Trace> = {}): Trace {
  return {
    trace_id: "0".repeat(32),
    root_name: "amie.process_event:request_account_create",
    source: "amie",
    status: 0,
    started_at: "2026-06-03T00:00:00.000Z",
    ended_at: "2026-06-03T00:00:01.000Z",
    span_count: 2,
    root_event: null,
    ...over,
  };
}

function buildSpan(over: Partial<Span> & { span_id: string }): Span {
  return {
    name: "amie.process_event",
    kind: 1,
    status: 0,
    start_time: "2026-06-03T00:00:00.000Z",
    end_time: "2026-06-03T00:00:01.000Z",
    attributes: null,
    ...over,
  };
}

describe("TraceOverviewTab", () => {
  it("renders every facts row for an ok amie trace", () => {
    render(<TraceOverviewTab trace={buildTrace()} spans={[buildSpan({ span_id: "1".repeat(16) })]} />);
    expect(screen.getByText("Trace ID")).toBeInTheDocument();
    expect(screen.getByText("Source")).toBeInTheDocument();
    expect(screen.getByText("Root action")).toBeInTheDocument();
    expect(screen.getByText("Status")).toBeInTheDocument();
    expect(screen.getByText("Started")).toBeInTheDocument();
    expect(screen.getByText("Ended")).toBeInTheDocument();
    expect(screen.getByText("Duration")).toBeInTheDocument();
    expect(screen.getByText("Span count")).toBeInTheDocument();
    expect(screen.getByText(/No retry attempts yet/)).toBeInTheDocument();
  });

  it("renders an error status pill for a failed trace", () => {
    render(
      <TraceOverviewTab
        trace={buildTrace({ status: 1 })}
        spans={[buildSpan({ span_id: "1".repeat(16), status: 1 })]}
      />,
    );
    const statusPills = screen.getAllByText(/error/i);
    expect(statusPills.length).toBeGreaterThan(0);
  });

  it("renders the still-running label when ended_at is null", () => {
    render(
      <TraceOverviewTab
        trace={buildTrace({ ended_at: null })}
        spans={[buildSpan({ span_id: "1".repeat(16), end_time: null })]}
      />,
    );
    expect(screen.getByText("still running")).toBeInTheDocument();
  });

  it("shows the empty-root-entity copy when no entity attrs on the root span", () => {
    const t = buildTrace();
    const s = [
      buildSpan({
        span_id: "f".repeat(16),
        attributes: { "irrelevant.key": "x" },
      }),
    ];
    render(<TraceOverviewTab trace={t} spans={s} />);
    expect(screen.getByText(/No root entity attributes captured/)).toBeInTheDocument();
  });
});
