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

import { describe, expect, it } from "vitest";
import { deliveryFixtures } from "../__fixtures__/deliveries";
import { deliveryState, historyRuns, relativeTime } from "../utils";

function fixture(id: string) {
  const delivery = deliveryFixtures.find((d) => d.id === id);
  if (!delivery) throw new Error(`no fixture ${id}`);
  return delivery;
}

describe("historyRuns", () => {
  // Make sure an admin retry starts a new run and attempts keep counting, as the trace numbers its tries.
  it("splits the history into runs at each admin retry, numbering attempts continuously", () => {
    const runs = historyRuns(fixture("07000001-0000-4000-8000-000007000001").history);

    expect(runs.map((run) => run.map((step) => step.kind))).toEqual([
      ["attempt", "attempt"],
      ["retry", "attempt"],
    ]);
    expect(runs[1]?.[1]).toMatchObject({ kind: "attempt", attempt: 3, ok: false });
  });
});

describe("deliveryState", () => {
  // Make sure a pending delivery whose connector is not running is shown apart from one that is only queued.
  it("marks a pending delivery as not running when its connector is not loaded", () => {
    const running = new Set(["comanage-identity-provisioner", "slurm-association-mapper"]);

    expect(deliveryState(fixture("07000004-0000-4000-8000-000007000004"), running)).toBe(
      "not-running",
    );
    expect(deliveryState(fixture("07000003-0000-4000-8000-000007000003"), running)).toBe(
      "retrying",
    );
  });
});

describe("relativeTime", () => {
  // Make sure Events and the trace, which share this formatter, round a next try the same way.
  it("rounds down in both directions", () => {
    const now = Date.parse("2026-10-09T00:00:00Z");
    expect(relativeTime("2026-10-09T01:59:00Z", now)).toBe("in 1h");
    expect(relativeTime("2026-10-09T00:53:30Z", now)).toBe("in 53 min");
    expect(relativeTime("2026-10-08T23:57:00Z", now)).toBe("3 min ago");
  });
});
