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
import { describe, expect, it, vi } from "vitest";
import type { AccessRequest } from "../schemas";

vi.mock("@/features/core/allocations/queries", () => ({
  useAllocation: (id: string | undefined) => ({
    data: id ? { id, name: "pearc26-tutorial", initial_su_amount: 25000 } : undefined,
  }),
}));

vi.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(""),
}));

import { AccessRequestDrawer } from "../components/AccessRequestDrawer";

const base: AccessRequest = {
  id: "areq-1",
  oidc_sub: "sub-1",
  email: "amara.osei@example.edu",
  name: "Amara Osei",
  institution: "Example University",
  event_code: "PEARC26",
  reason: "Hands-on tutorial attendee.",
  status: "PENDING",
  approver_id: "",
  deny_reason: "",
  expires_at: null,
  created_user_id: "",
  timestamp: "2026-07-10T14:05:00Z",
  decided_at: null,
  allocation_id: "alloc-1",
};

describe("<AccessRequestDrawer />", () => {
  it("renders request fields without a decision block while pending", () => {
    render(<AccessRequestDrawer request={base} onClose={vi.fn()} />);
    expect(screen.getByText("Example University")).toBeInTheDocument();
    expect(screen.getByText("Hands-on tutorial attendee.")).toBeInTheDocument();
    expect(screen.queryByText("Decision")).not.toBeInTheDocument();
  });

  it("renders the decision block with allocation and expiry for approved", () => {
    render(
      <AccessRequestDrawer
        request={{
          ...base,
          status: "APPROVED",
          approver_id: "user-admin-0001",
          decided_at: "2026-07-11T08:00:00Z",
          expires_at: "2026-08-09T00:00:00Z",
        }}
        onClose={vi.fn()}
      />,
    );
    expect(screen.getByText("Decision")).toBeInTheDocument();
    expect(screen.getByText("Approved")).toBeInTheDocument();
    expect(screen.getByText("user-adm…")).toBeInTheDocument();
    expect(screen.getByText(/pearc26-tutorial/)).toBeInTheDocument();
    const credits = new Intl.NumberFormat(undefined, { maximumFractionDigits: 0 }).format(25000);
    expect(screen.getByText(new RegExp(`${credits} credits`))).toBeInTheDocument();
    expect(screen.getByText("Allocation ends")).toBeInTheDocument();
  });

  it("renders the deny reason verbatim for denied", () => {
    render(
      <AccessRequestDrawer
        request={{
          ...base,
          status: "DENIED",
          approver_id: "user-admin-0001",
          decided_at: "2026-07-11T08:00:00Z",
          deny_reason: "Not a tutorial attendee",
        }}
        onClose={vi.fn()}
      />,
    );
    expect(screen.getByText("Denied")).toBeInTheDocument();
    expect(screen.getByText("Not a tutorial attendee")).toBeInTheDocument();
    expect(screen.queryByText("Granted allocation")).not.toBeInTheDocument();
  });
});
