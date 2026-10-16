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

import type { User } from "@/generated/core/types.gen";
import { defineAbilitiesFor } from "@/shared/casl/abilities";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ProfileCard } from "../ProfileCard";

const mutateAsync = vi.hoisted(() => vi.fn(async () => ({})));
const me = vi.hoisted(() => ({ user: {} as Record<string, unknown> }));

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => defineAbilitiesFor(["core:users:write"]),
}));

vi.mock("../../queries", () => ({
  useUpdateMyName: () => ({ mutateAsync, isPending: false }),
  useMe: () => ({ data: me, refetch: vi.fn(), dataUpdatedAt: Number.MAX_SAFE_INTEGER }),
}));

const user: User = {
  id: "u1",
  email: "elena.vasquez@sample.example.edu",
  first_name: "Elena",
  middle_name: "",
  last_name: "Vasquez",
  organization_id: "org-1",
  status: "ACTIVE",
  type: "CLUSTER_LOCAL",
};

describe("ProfileCard", () => {
  it("shows name, email, ACTIVE badge, and read-only fields with the organization name", () => {
    render(<ProfileCard user={user} organizationName="Sample University" />);
    expect(screen.getByText("Elena Vasquez")).toBeInTheDocument();
    expect(screen.getByText("elena.vasquez@sample.example.edu")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("elena.vasquez")).toBeInTheDocument();
    expect(screen.getByText("Sample University")).toBeInTheDocument();
  });

  // Another tab renamed the user; typing the old value back must still save it.
  it("prefills from the fresh read and saves the whole name", () => {
    me.user = { ...user, last_name: "Vasquez-Ng" };
    render(<ProfileCard user={user} organizationName="Sample University" />);
    fireEvent.click(screen.getByRole("button", { name: /edit name/i }));
    expect(screen.getByLabelText(/last name/i)).toHaveValue("Vasquez-Ng");
    fireEvent.change(screen.getByLabelText(/last name/i), { target: { value: "Vasquez" } });
    fireEvent.click(screen.getByRole("button", { name: /^save$/i }));

    expect(mutateAsync).toHaveBeenCalledWith({
      path: { id: "u1" },
      body: { first_name: "Elena", middle_name: "", last_name: "Vasquez" },
    });
  });

  it("mentions linked identities in the lock note only when their card is shown", () => {
    const { rerender } = render(<ProfileCard user={user} organizationName="Sample University" />);
    fireEvent.click(screen.getByRole("button", { name: /edit name/i }));
    expect(screen.getByText("Email is managed by your identity provider.")).toBeInTheDocument();
    rerender(<ProfileCard user={user} organizationName="Sample University" showsIdentities />);
    expect(screen.getByText(/Email and linked identities/)).toBeInTheDocument();
  });

  it("saves a middle name for a user without first and last names", () => {
    me.user = { ...user, first_name: "", last_name: "" };
    render(<ProfileCard user={me.user as User} organizationName="Sample University" />);
    fireEvent.click(screen.getByRole("button", { name: /edit name/i }));
    fireEvent.change(screen.getByLabelText(/middle name/i), { target: { value: "Q" } });
    fireEvent.click(screen.getByRole("button", { name: /^save$/i }));

    expect(mutateAsync).toHaveBeenCalledWith({
      path: { id: "u1" },
      body: { first_name: "", middle_name: "Q", last_name: "" },
    });
  });
});
