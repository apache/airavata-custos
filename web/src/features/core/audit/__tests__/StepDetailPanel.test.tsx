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

import approveFailed from "@/features/core/audit/__fixtures__/trace.approve.failed.json";
import { StepDetailPanel } from "@/features/core/audit/components/StepDetailPanel";
import { traceDetailSchema } from "@/features/core/audit/schemas";
import { type FlowRow, buildFlow, expandableKeys, flattenFlow } from "@/features/core/audit/utils";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const subscriptions = vi.hoisted(() => ({ loaded: true }));
vi.mock("@/features/core/events/queries", () => ({
  useSubscriptions: () => ({
    data: {
      items: [
        {
          subscriber: "amie-processor",
          event_type: "ComputeClusterUserApproved",
          loaded: subscriptions.loaded,
        },
      ],
    },
  }),
}));

const detail = traceDetailSchema.parse(approveFailed);
const model = buildFlow(detail);
const rows = flattenFlow(model, new Set(expandableKeys(model)));

describe("<StepDetailPanel />", () => {
  // Make sure a try's panel is titled as its row reads.
  it("titles a try row's panel Try N with its outcome", () => {
    const row = rows.find((r) => r.kind === "step" && r.try?.outcome === "failed");
    if (row?.kind !== "step" || !row.try) throw new Error("no failed try row");
    render(<StepDetailPanel row={row} model={model} traceId={detail.trace_id} />);
    expect(screen.getByText(`Try ${row.try.n} failed`)).toBeInTheDocument();
  });

  // Make sure a delivery with no try yet names whether its connector is running.
  it("keys the no-try-yet copy on the connector's subscription", () => {
    const row: FlowRow = {
      kind: "group",
      key: "d",
      depth: 0,
      title: "amie-processor",
      source: "amie",
      status: "waiting",
      expanded: true,
      onPath: false,
      tries: 0,
      delivery: {
        id: "dlv-1",
        subscriber: "amie-processor",
        event_type: "ComputeClusterUserApproved",
        status: "PENDING",
        attempts: 0,
        next_run_at: "2026-10-09T00:00:00Z",
        span_id: "00f067aa0ba902b7",
      },
    };
    const { unmount } = render(
      <StepDetailPanel row={row} model={model} traceId={detail.trace_id} />,
    );
    expect(screen.getByText(/The connector is running/)).toBeInTheDocument();
    unmount();
    subscriptions.loaded = false;
    render(<StepDetailPanel row={row} model={model} traceId={detail.trace_id} />);
    expect(screen.getByText(/The connector is not running on this server/)).toBeInTheDocument();
  });
});
