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

import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { AccessStatus } from "../schemas";

const state: { data: AccessStatus | undefined } = { data: undefined };
vi.mock("../queries", () => ({
  useAccessStatus: () => ({ data: state.data, isLoading: false, error: null }),
}));

import { AllocationAccessBand } from "../components/AllocationAccessBand";

const now = new Date().toISOString();

function check(type: "SIGN_IN" | "JOB_SUBMISSION", ui_state: AccessStatus["checks"][0]["ui_state"]) {
  return {
    type,
    status: ui_state === "ok" ? ("OK" as const) : ui_state === "setting_up" ? ("PENDING" as const) : ("FAILING" as const),
    ui_state,
    last_checked_at: now,
    last_ok_at: ui_state === "ok" ? now : null,
    failing_since: ui_state === "retrying" || ui_state === "stuck" ? now : null,
  };
}

function renderBand(status: AccessStatus) {
  state.data = status;
  return render(<AllocationAccessBand allocationId="alloc-1" allocationName="pearc26-tutorial" />);
}

describe("<AllocationAccessBand />", () => {
  it("renders nothing before the first result", () => {
    state.data = undefined;
    const { container } = render(
      <AllocationAccessBand allocationId="alloc-1" allocationName="pearc26-tutorial" />,
    );
    expect(container).toBeEmptyDOMElement();
  });

  it("collapses to one line when everything is live, expanding on demand", () => {
    renderBand({
      checks: [check("SIGN_IN", "ok"), check("JOB_SUBMISSION", "ok")],
      events: [{ check_type: "SIGN_IN", event_type: "ONLINE", occurred_at: now }],
    });
    expect(screen.getByText("Your access is live")).toBeInTheDocument();
    expect(screen.queryByText("Cluster sign-in")).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "View checks" }));
    expect(screen.getByText("Cluster sign-in")).toBeInTheDocument();
    expect(screen.getByText("Job submission")).toBeInTheDocument();
    expect(screen.getByText("Cluster sign-in came online")).toBeInTheDocument();
  });

  it("expands automatically while access is still being set up", () => {
    renderBand({
      checks: [check("SIGN_IN", "ok"), check("JOB_SUBMISSION", "setting_up")],
      events: [],
    });
    expect(screen.getByText("Setting up your access")).toBeInTheDocument();
    expect(screen.getByText(/Job submission is still being set up/)).toBeInTheDocument();
    expect(screen.getByText("Ready")).toBeInTheDocument();
    // Both the band badge and the check chip carry the state label.
    expect(screen.getAllByText("Setting up")).toHaveLength(2);
  });

  it("reads a failing check as calm retrying, not an alarm", () => {
    renderBand({
      checks: [check("SIGN_IN", "ok"), check("JOB_SUBMISSION", "retrying")],
      events: [{ check_type: "JOB_SUBMISSION", event_type: "FAILED", occurred_at: now }],
    });
    expect(screen.getByText("Reconnecting your access")).toBeInTheDocument();
    expect(screen.getByText(/retrying automatically, and no action is needed/)).toBeInTheDocument();
    expect(screen.getByText("Retrying")).toBeInTheDocument();
    expect(screen.getByText("Job submission failed a check")).toBeInTheDocument();
  });

  it("escalates a stuck check with the support next step", () => {
    renderBand({
      checks: [check("SIGN_IN", "ok"), check("JOB_SUBMISSION", "stuck")],
      events: [{ check_type: "JOB_SUBMISSION", event_type: "STUCK", occurred_at: now }],
    });
    expect(screen.getByText("Your job submission access needs attention")).toBeInTheDocument();
    expect(screen.getByText(/Your sign-in still works/)).toBeInTheDocument();
    expect(screen.getByText(/Contact support if it doesn't clear/)).toBeInTheDocument();
    expect(screen.getByText("Not working")).toBeInTheDocument();
    expect(screen.getByText("Job submission still failing")).toBeInTheDocument();
  });
});
