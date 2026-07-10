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
import { StatusBadge, type StatusBadgeVariant } from "../StatusBadge";

const ALL_VARIANTS: StatusBadgeVariant[] = [
  "active",
  "inactive",
  "deleted",
  "pending",
  "approved",
  "rejected",
  "expired",
  "warning",
];

describe("StatusBadge", () => {
  it("renders a dot indicator on the active variant", () => {
    const { container } = render(<StatusBadge variant="active" />);
    const dot = container.querySelector("[aria-hidden='true']");
    expect(dot).not.toBeNull();
    expect(dot?.className).toContain("rounded-full");
  });

  it("does not render a dot on non-active variants", () => {
    for (const variant of ALL_VARIANTS) {
      if (variant === "active") continue;
      const { container, unmount } = render(<StatusBadge variant={variant} />);
      expect(container.querySelector("[aria-hidden='true']")).toBeNull();
      unmount();
    }
  });

  it("renders the default label per variant", () => {
    render(<StatusBadge variant="pending" />);
    expect(screen.getByText("Pending")).toBeInTheDocument();
  });

  it("uses bg-muted for inactive/expired and drops the ring", () => {
    const { container, rerender } = render(<StatusBadge variant="inactive" />);
    const span = container.firstElementChild as HTMLElement;
    expect(span.className).toContain("bg-muted");
    expect(span.className).not.toContain("ring-1");
    rerender(<StatusBadge variant="expired" />);
    const expired = container.firstElementChild as HTMLElement;
    expect(expired.className).toContain("bg-muted");
  });

  it("accepts a custom label override", () => {
    render(<StatusBadge variant="approved" label="OK" />);
    expect(screen.getByText("OK")).toBeInTheDocument();
  });
});
