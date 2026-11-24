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
import { traceListSchema } from "@/features/core/audit/schemas";
import type { Step } from "@/features/core/audit/types";
import { errorText, formatUntil, listStatusLabel, parseDetails } from "@/features/core/audit/utils";
import { describe, expect, it } from "vitest";

const list = traceListSchema.parse(tracesListFixture);
const byStatus = (status: string) => {
  const trace = list.traces.find((t) => t.status === status);
  if (!trace) throw new Error(`fixture has no ${status} trace`);
  return trace;
};

describe("listStatusLabel", () => {
  // Make sure the status word comes from the deliveries: a failed delivery is
  // Failed, a pending one with tries is Retrying with the count, and a trace
  // with every delivery done is Done.
  it("names the trace from its deliveries", () => {
    expect(listStatusLabel(byStatus("error"))).toBe("Failed");
    expect(listStatusLabel(byStatus("in_progress"))).toBe("Retrying 4 of 10");
    expect(listStatusLabel(byStatus("ok"))).toBe("Done");
  });

  // Make sure a pending delivery that has not run yet reads as Waiting, not
  // Retrying, since the connector may simply not be running.
  it("says Waiting before the first try", () => {
    const waiting = {
      ...byStatus("in_progress"),
      deliveries: { pending: 1, succeeded: 0, failed: 0, attempts: 0 },
    };
    expect(listStatusLabel(waiting)).toBe("Waiting");
  });
});

describe("parseDetails", () => {
  // Make sure the three shapes the backend writes all become key value pairs
  // or text: JSON, `key=value` pairs, and free text.
  it("reads JSON, key=value pairs, and free text", () => {
    expect(parseDetails('{"user_id":"u1","local_username":"jsmith"}')).toEqual({
      entries: [
        ["user_id", "u1"],
        ["local_username", "jsmith"],
      ],
    });
    expect(parseDetails("step=lookup err=401 Unauthorized: invalid API key")).toEqual({
      entries: [
        ["step", "lookup"],
        ["err", "401 Unauthorized: invalid API key"],
      ],
    });
    expect(parseDetails("DECODED")).toEqual({ text: "DECODED" });
    expect(parseDetails(undefined)).toBeNull();
  });
});

describe("errorText", () => {
  const step = (over: Partial<Step>): Step => ({
    id: "r1",
    span_id: "0123456789abcdef",
    source: "comanage",
    event_type: "ComanageProvisioningFailed",
    status: "error",
    created_at: "2026-10-09T00:00:00Z",
    ...over,
  });

  // Make sure the row shows the error the connector wrote, with the step it
  // failed at, and nothing for a row that did not fail.
  it("pulls the error out of a failed row's details", () => {
    expect(errorText(step({ description: "step=lookup err=401 Unauthorized" }))).toBe(
      "401 Unauthorized (step lookup)",
    );
    expect(errorText(step({ description: "dial tcp: connection refused" }))).toBe(
      "dial tcp: connection refused",
    );
    expect(errorText(step({ status: "ok", description: "fine" }))).toBeUndefined();
  });
});

describe("formatUntil", () => {
  // Make sure a retry whose time has passed reads "now", since the worker
  // picks it up on its next poll rather than at the exact second.
  it("says now once the time has passed", () => {
    const now = Date.parse("2026-10-09T00:00:00Z");
    expect(formatUntil("2026-10-09T00:04:00Z", now)).toBe("in 4m");
    expect(formatUntil("2026-10-08T23:59:00Z", now)).toBe("now");
  });
});
