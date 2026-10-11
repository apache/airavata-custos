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

import sourcesFixture from "@/features/core/audit/__fixtures__/sources.json";
import approveDone from "@/features/core/audit/__fixtures__/trace.approve.done.json";
import approveFailed from "@/features/core/audit/__fixtures__/trace.approve.failed.json";
import approveRetrying from "@/features/core/audit/__fixtures__/trace.approve.retrying.json";
import createdDone from "@/features/core/audit/__fixtures__/trace.created.done.json";
import deniedDone from "@/features/core/audit/__fixtures__/trace.denied.done.json";
import packetFailed from "@/features/core/audit/__fixtures__/trace.packet.failed.json";
import tracesListFixture from "@/features/core/audit/__fixtures__/traces.list.json";
import {
  stepSchema,
  traceDetailSchema,
  traceListSchema,
  traceSourceListSchema,
} from "@/features/core/audit/schemas";
import { describe, expect, it } from "vitest";

// The fixtures are real responses captured from the backend, so these parses
// are the contract between the two.
describe("audit schemas", () => {
  it("parse every captured response", () => {
    expect(() => traceListSchema.parse(tracesListFixture)).not.toThrow();
    expect(() => traceSourceListSchema.parse(sourcesFixture)).not.toThrow();
    for (const detail of [
      approveFailed,
      approveRetrying,
      approveDone,
      deniedDone,
      packetFailed,
      createdDone,
    ]) {
      expect(() => traceDetailSchema.parse(detail)).not.toThrow();
    }
  });

  // Make sure a root row with no parent is accepted while an explicit null is
  // not, since the backend omits the field and never sends null.
  it("treat parent_span_id as omitted, never null", () => {
    const row = { ...approveFailed.tree[0], children: undefined };
    expect(stepSchema.safeParse({ ...row, parent_span_id: undefined }).success).toBe(true);
    expect(stepSchema.safeParse({ ...row, parent_span_id: null }).success).toBe(false);
  });
});
