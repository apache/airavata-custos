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

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { TraceRawTab } from "@/features/core/audit/components/TraceRawTab";
import type { Span, Trace } from "@/features/core/audit/types";

const writeText = vi.fn().mockResolvedValue(undefined);

beforeEach(() => {
  writeText.mockClear();
  Object.assign(navigator, { clipboard: { writeText } });
});

afterEach(() => {
  vi.useRealTimers();
});

const trace: Trace = {
  trace_id: "a".repeat(32),
  root_name: "amie.process_event:request_account_create",
  source: "amie",
  status: 0,
  started_at: "2026-06-03T00:00:00.000Z",
  ended_at: "2026-06-03T00:00:01.000Z",
  span_count: 3,
  root_event: null,
};

const spans: Span[] = [
  {
    span_id: "1".repeat(16),
    name: "amie.process_event",
    kind: 1,
    status: 0,
    start_time: "2026-06-03T00:00:00.000Z",
    end_time: "2026-06-03T00:00:00.500Z",
    attributes: null,
  },
  {
    span_id: "2".repeat(16),
    parent_span_id: "1".repeat(16),
    name: "comanage.ensure_co_person",
    kind: 0,
    status: 0,
    start_time: "2026-06-03T00:00:00.100Z",
    end_time: "2026-06-03T00:00:00.250Z",
    attributes: null,
  },
  {
    span_id: "3".repeat(16),
    parent_span_id: "1".repeat(16),
    name: "slurm.create_account",
    kind: 0,
    status: 0,
    start_time: "2026-06-03T00:00:00.300Z",
    end_time: "2026-06-03T00:00:00.450Z",
    attributes: null,
  },
];

describe("TraceRawTab", () => {
  it("renders the Copy JSON button and span count label", () => {
    render(<TraceRawTab trace={trace} spans={spans} />);
    expect(screen.getByRole("button", { name: /copy trace JSON/i })).toBeInTheDocument();
    expect(screen.getByText(/3 spans/)).toBeInTheDocument();
  });

  it("emits highlighted tokens for keys (syntax-key) and strings (syntax-str)", () => {
    const { container } = render(<TraceRawTab trace={trace} spans={spans} />);
    const pre = container.querySelector('[data-testid="trace-raw-json"]');
    expect(pre).not.toBeNull();
    const keys = pre?.querySelectorAll('[data-token="key"]');
    expect(keys?.length ?? 0).toBeGreaterThan(0);
    const keyToken = Array.from(keys ?? []).find((el) =>
      (el.getAttribute("style") ?? "").includes("--syntax-key"),
    );
    expect(keyToken).toBeTruthy();
    const strings = pre?.querySelectorAll('[data-token="str"]');
    expect(strings?.length ?? 0).toBeGreaterThan(0);
  });

  it("copies the stringified JSON to clipboard on click", async () => {
    render(<TraceRawTab trace={trace} spans={spans} />);
    fireEvent.click(screen.getByRole("button", { name: /copy trace JSON/i }));
    await waitFor(() => {
      expect(writeText).toHaveBeenCalledTimes(1);
    });
    const payload = writeText.mock.calls[0]?.[0] as string;
    expect(payload).toContain(`"trace_id": "${trace.trace_id}"`);
    expect(payload).toContain('"span_count":');
  });
});
