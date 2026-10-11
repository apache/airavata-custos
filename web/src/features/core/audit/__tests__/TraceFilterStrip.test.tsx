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

import { TraceFilterStrip } from "@/features/core/audit/components/TraceFilterStrip";
import {
  DEFAULT_FILTERS,
  type ListFilters,
} from "@/features/core/audit/components/traceListUrlState";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

function renderStrip(initial: ListFilters = DEFAULT_FILTERS) {
  const onChange = vi.fn<(next: ListFilters) => void>();
  let value: ListFilters = initial;
  const Wrap = () => (
    <TraceFilterStrip
      value={value}
      onChange={(next) => {
        value = next;
        onChange(next);
      }}
    />
  );
  const utils = render(<Wrap />);
  return { ...utils, onChange };
}

describe("TraceFilterStrip", () => {
  // Make sure no status is pre-selected, so an admin sees every trace until
  // they narrow the list.
  it("starts with every status off and the 7 day window on", () => {
    renderStrip();
    for (const name of ["Failed", "Retrying", "Waiting", "Done"]) {
      expect(screen.getByRole("button", { name, pressed: false })).toBeInTheDocument();
    }
    expect(screen.getByRole("button", { name: "7d", pressed: true })).toBeInTheDocument();
  });

  // Make sure changing any filter sends the list back to page 1, otherwise a
  // narrower result set can land on an empty page.
  it("resets to page 1 when a status is toggled", () => {
    const { onChange } = renderStrip({ ...DEFAULT_FILTERS, page: 3 });
    fireEvent.click(screen.getByRole("button", { name: "Failed" }));
    expect(onChange.mock.calls.at(-1)?.[0]).toMatchObject({ status: ["failed"], page: 1 });
  });

  // Make sure typing does not refetch on every keystroke.
  it("debounces the search", async () => {
    const { onChange } = renderStrip();
    fireEvent.change(screen.getByRole("searchbox", { name: "Search traces" }), {
      target: { value: "jsmith" },
    });
    expect(onChange).not.toHaveBeenCalled();
    await waitFor(() => expect(onChange).toHaveBeenCalled());
    expect(onChange.mock.calls.at(-1)?.[0]).toMatchObject({ q: "jsmith", page: 1 });
  });
});
