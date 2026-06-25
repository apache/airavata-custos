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
import { StatusPill } from "@/features/core/audit/components/primitives/StatusPill";
import type { RowTone } from "@/features/core/audit/types";

const TONES: RowTone[] = ["ok", "error", "in-progress", "orphaned", "no-status"];

describe("StatusPill", () => {
  it("renders each tone with a default label", () => {
    for (const tone of TONES) {
      const { unmount } = render(<StatusPill tone={tone} />);
      expect(screen.getByRole("status")).toBeInTheDocument();
      unmount();
    }
  });

  it("dotOnly variant renders only the dot (no label text)", () => {
    render(<StatusPill tone="ok" dotOnly />);
    const node = screen.getByRole("status");
    expect(node).toHaveAttribute("aria-label", "Status: ok");
    expect(node.textContent ?? "").toBe("");
  });

  it("applies the pulsing class for in-progress", () => {
    const { container } = render(<StatusPill tone="in-progress" />);
    expect(container.querySelector(".custos-pulse-dot")).not.toBeNull();
  });

  it("does not apply the pulsing class for non-in-progress tones", () => {
    const { container } = render(<StatusPill tone="ok" />);
    expect(container.querySelector(".custos-pulse-dot")).toBeNull();
  });

  it("uses a custom label when provided", () => {
    render(<StatusPill tone="error" label="failed" />);
    expect(screen.getByText("failed")).toBeInTheDocument();
  });
});
