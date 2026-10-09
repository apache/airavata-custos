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

import { TraceRawTab } from "@/features/core/audit/components/TraceRawTab";
import type { TraceEvent, TraceSummary } from "@/generated/core/types.gen";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const writeText = vi.fn<Clipboard["writeText"]>().mockResolvedValue(undefined);

beforeEach(() => {
  writeText.mockClear();
  Object.assign(navigator, { clipboard: { writeText } });
});

const trace: TraceSummary = {
  trace_id: "a".repeat(32),
  root_operation: "amie.process_event:request_account_create",
  source: "amie",
  status: "ok",
  started_at: "2026-06-03T00:00:00.000Z",
  ended_at: "2026-06-03T00:00:01.000Z",
  event_count: 3,
  deliveries: { pending: 0, succeeded: 0, failed: 0, attempts: 0 },
};

const spans: TraceEvent[] = [
  {
    id: "1",
    span_id: "1".repeat(16),
    source: "amie",
    event_type: "amie.process_event",
    status: "ok",
    created_at: "2026-06-03T00:00:00.000Z",
  },
];

describe("TraceRawTab", () => {
  it("copies the stringified JSON to clipboard on click", async () => {
    render(<TraceRawTab trace={trace} spans={spans} />);
    fireEvent.click(screen.getByRole("button", { name: /copy trace JSON/i }));
    await waitFor(() => {
      expect(writeText).toHaveBeenCalledTimes(1);
    });
    const payload = writeText.mock.calls[0]?.[0] ?? "";
    expect(payload).toContain(`"trace_id": "${trace.trace_id}"`);
    expect(payload).toContain('"event_count":');
  });
});
