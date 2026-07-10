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
import { CopyValue } from "@/features/core/audit/components/primitives/CopyValue";

describe("CopyValue", () => {
  const writeText = vi.fn().mockResolvedValue(undefined);

  beforeEach(() => {
    writeText.mockClear();
    Object.assign(navigator, { clipboard: { writeText } });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("writes to the clipboard on click and stops propagation", () => {
    const onParentClick = vi.fn();
    render(
      <button type="button" onClick={onParentClick}>
        <CopyValue value="trace-abc" label="trace ID" />
      </button>,
    );
    const btn = screen.getByRole("button", { name: "Copy trace ID" });
    fireEvent.click(btn);
    expect(writeText).toHaveBeenCalledWith("trace-abc");
    expect(onParentClick).not.toHaveBeenCalled();
  });

  it("swaps to the check icon for 1.1s then reverts to copy", async () => {
    const { container } = render(<CopyValue value="trace-abc" />);
    const btn = screen.getByRole("button", { name: "Copy trace-abc" });
    fireEvent.click(btn);
    await waitFor(() => {
      expect(container.querySelector("svg.lucide-check")).not.toBeNull();
    });
    expect(container.querySelector("svg.lucide-copy")).toBeNull();

    await waitFor(
      () => {
        expect(container.querySelector("svg.lucide-check")).toBeNull();
      },
      { timeout: 1500 },
    );
    expect(container.querySelector("svg.lucide-copy")).not.toBeNull();
  });
});
