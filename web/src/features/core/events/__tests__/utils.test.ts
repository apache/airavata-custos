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
import { deliveryState, historyRuns } from "../utils";

function fixture(id: string) {
  const delivery = deliveryFixtures.find((d) => d.id === id);
  if (!delivery) throw new Error(`no fixture ${id}`);
  return delivery;
}

describe("historyRuns", () => {
  // Make sure an admin retry starts a new run, so the restarted attempt numbers are not mixed with the old ones.
  it("splits the history into runs at each admin retry", () => {
    const runs = historyRuns(fixture("dlv-failed-1").history);

    expect(runs.map((run) => run.map((step) => step.kind))).toEqual([
      ["attempt", "attempt"],
      ["retry", "attempt"],
    ]);
    expect(runs[1]?.[0]).toMatchObject({ kind: "retry", actorId: "user-admin", previousAttempts: 2 });
    expect(runs[1]?.[1]).toMatchObject({ kind: "attempt", attempt: 1, ok: false });
  });
});

describe("deliveryState", () => {
  // Make sure a pending delivery whose connector is not running is shown apart from one that is only queued.
  it("marks a pending delivery as not running when its connector is not loaded", () => {
    const running = new Set(["comanage-identity-provisioner", "slurm-association-mapper"]);

    expect(deliveryState(fixture("dlv-waiting-1"), running)).toBe("not-running");
    expect(deliveryState(fixture("dlv-retrying-1"), running)).toBe("retrying");
  });
});
