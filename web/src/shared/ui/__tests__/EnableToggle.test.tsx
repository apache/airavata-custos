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
import { describe, expect, it, vi } from "vitest";
import { EnableToggle } from "../EnableToggle";

const baseProps = {
  resource: { kind: "cluster" as const, id: "cluster-001", label: "Bridges-2" },
  impact: { activeAllocations: 47, activeUsers: 213, inflightJobs: 5 },
};

describe("EnableToggle", () => {
  it("renders enabled state with green dot and Enabled label", () => {
    render(<EnableToggle {...baseProps} enabled onConfirm={vi.fn()} />);
    const trigger = screen.getByRole("switch");
    expect(trigger).toHaveAttribute("aria-checked", "true");
    expect(trigger).toHaveTextContent("Enabled");
    expect(trigger.getAttribute("data-state")).toBe("enabled");
  });

  it("renders disabled state with Disabled label", () => {
    render(<EnableToggle {...baseProps} enabled={false} onConfirm={vi.fn()} />);
    const trigger = screen.getByRole("switch");
    expect(trigger).toHaveAttribute("aria-checked", "false");
    expect(trigger).toHaveTextContent("Disabled");
    expect(trigger.getAttribute("data-state")).toBe("disabled");
  });

  it("click opens the confirmation dialog with impact numbers", () => {
    render(<EnableToggle {...baseProps} enabled onConfirm={vi.fn()} />);
    fireEvent.click(screen.getByRole("switch"));
    expect(screen.getByText(/Disable Bridges-2\?/)).toBeInTheDocument();
    expect(screen.getByText("47")).toBeInTheDocument();
    expect(screen.getByText("213")).toBeInTheDocument();
    expect(screen.getByText("5")).toBeInTheDocument();
  });

  it("confirm fires onConfirm with the toggled value", async () => {
    const onConfirm = vi.fn().mockResolvedValue(undefined);
    render(<EnableToggle {...baseProps} enabled onConfirm={onConfirm} />);
    fireEvent.click(screen.getByRole("switch"));
    fireEvent.click(screen.getByRole("button", { name: "Disable cluster" }));
    await waitFor(() => expect(onConfirm).toHaveBeenCalledWith(false));
  });

  it("renders ErrorState inside the dialog when onConfirm rejects", async () => {
    const onConfirm = vi.fn().mockRejectedValue(new Error("backend down"));
    render(<EnableToggle {...baseProps} enabled onConfirm={onConfirm} />);
    fireEvent.click(screen.getByRole("switch"));
    fireEvent.click(screen.getByRole("button", { name: "Disable cluster" }));
    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent("backend down");
    });
  });

  it("disabled prop renders the switch non-interactive and announces read-only", () => {
    const onConfirm = vi.fn();
    render(<EnableToggle {...baseProps} enabled disabled onConfirm={onConfirm} />);
    const trigger = screen.getByRole("switch");
    expect(trigger).toBeDisabled();
    expect(trigger.getAttribute("aria-label")).toMatch(/read-only/i);
    fireEvent.click(trigger);
    expect(onConfirm).not.toHaveBeenCalled();
  });
});
