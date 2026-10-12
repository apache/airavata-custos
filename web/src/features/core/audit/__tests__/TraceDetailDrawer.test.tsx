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
import { ApiError } from "@/shared/api/client";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import * as React from "react";
import { describe, expect, it, vi } from "vitest";

const trace = vi.hoisted(() => ({ result: {} as Record<string, unknown> }));
vi.mock("../queries", () => ({ useTrace: () => trace.result }));
vi.mock("@/features/core/events/queries", () => ({
  useSubscriptions: () => ({ data: { items: [] } }),
}));
vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(window.location.search),
}));

import { TraceDetailDrawer } from "../components/TraceDetailDrawer";
import { traceDetailSchema } from "../schemas";

const detail = traceDetailSchema.parse(approveFailed);

function openAt(query: string) {
  window.history.replaceState(null, "", `/admin/traces?${query}`);
  render(<TraceDetailDrawer traceId={detail.trace_id} open onClose={vi.fn()} />);
}

describe("<TraceDetailDrawer />", () => {
  // Make sure ?step= is validated whichever tab is open, so a bogus step never lingers.
  it("drops a step naming no row while the Raw tab is open", async () => {
    trace.result = { data: detail, isLoading: false, error: null, refetch: vi.fn() };
    openAt(`trace=${detail.trace_id}&tab=raw&step=bogus`);
    await waitFor(() => expect(window.location.search).toBe(`?trace=${detail.trace_id}&tab=raw`));
  });

  // Make sure a trace that fails to load keeps no tab or step param it cannot show.
  it("drops the tab and step of a trace that fails to load", async () => {
    trace.result = {
      data: undefined,
      isLoading: false,
      error: new ApiError(404, "/audit/traces", {}),
      refetch: vi.fn(),
    };
    openAt(`trace=${detail.trace_id}&tab=raw&step=bogus`);
    await waitFor(() => expect(window.location.search).toBe(`?trace=${detail.trace_id}`));
  });

  // Make sure closing the drawer returns focus to the row's Open trace button, never the list or body.
  it("returns focus to the row that opened it", async () => {
    trace.result = { data: detail, isLoading: false, error: null, refetch: vi.fn() };
    window.history.replaceState(null, "", "/admin/traces");
    function List() {
      const [open, setOpen] = React.useState(false);
      return (
        <section aria-label="Traces">
          <div data-row data-drawer-trigger>
            <button type="button" onClick={() => setOpen(true)}>
              Open trace
            </button>
          </div>
          <TraceDetailDrawer traceId={detail.trace_id} open={open} onClose={() => setOpen(false)} />
        </section>
      );
    }
    render(<List />);
    const row = screen.getByRole("button", { name: "Open trace" });
    fireEvent.pointerDown(row);
    row.focus();
    fireEvent.click(row);
    await screen.findByTestId("trace-detail-drawer");
    fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });
    await waitFor(() => expect(row).toHaveFocus());
  });
});
