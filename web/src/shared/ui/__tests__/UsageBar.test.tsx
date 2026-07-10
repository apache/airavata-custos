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

import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { UsageBar } from "../UsageBar";

function inner(progressBar: HTMLElement): HTMLElement {
  const fill = progressBar.firstElementChild;
  if (!fill) throw new Error("UsageBar has no fill");
  return fill as HTMLElement;
}

describe("UsageBar thresholds", () => {
  it("uses the brand fill below 75%", () => {
    render(<UsageBar value={50} max={100} />);
    const bar = screen.getByRole("progressbar");
    expect(inner(bar).className).toContain("bg-brand");
  });

  it("uses the amber fill between 75% and 90%", () => {
    render(<UsageBar value={80} max={100} />);
    const bar = screen.getByRole("progressbar");
    expect(inner(bar).className).toContain("custos-amber-500");
  });

  it("uses the red fill at or above 90%", () => {
    render(<UsageBar value={95} max={100} />);
    const bar = screen.getByRole("progressbar");
    expect(inner(bar).className).toContain("custos-red-500");
  });

  it("flips to green when the bar is at 100%", () => {
    render(<UsageBar value={100} max={100} />);
    const bar = screen.getByRole("progressbar");
    expect(inner(bar).className).toContain("custos-green-500");
  });

  it("clamps when value exceeds max", () => {
    render(<UsageBar value={200} max={100} />);
    const bar = screen.getByRole("progressbar");
    expect(bar.getAttribute("aria-valuemax")).toBe("100");
    // Clamped to 100% → green.
    expect(inner(bar).className).toContain("custos-green-500");
  });

  it("treats a zero or negative max defensively", () => {
    render(<UsageBar value={10} max={0} />);
    const bar = screen.getByRole("progressbar");
    expect(bar).toBeInTheDocument();
  });
});
