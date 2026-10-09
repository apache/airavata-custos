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

const mutate = vi.hoisted(() => vi.fn());

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => defineAbilitiesFor(["core:users:write"]),
}));

vi.mock("../../queries", () => ({
  useUpdateMyName: () => ({ mutate, isPending: false }),
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
  it("shows name, email, ACTIVE badge, and read-only fields", () => {
    render(<ProfileCard user={user} />);
    expect(screen.getByText("Elena Vasquez")).toBeInTheDocument();
    expect(screen.getByText("elena.vasquez@sample.example.edu")).toBeInTheDocument();
    expect(screen.getByText("Active")).toBeInTheDocument();
    expect(screen.getByText("elena.vasquez")).toBeInTheDocument();
  });

  it("expands the inline editor and saves the name fields", () => {
    render(<ProfileCard user={user} />);
    fireEvent.click(screen.getByRole("button", { name: /edit name/i }));
    fireEvent.change(screen.getByLabelText(/last name/i), { target: { value: "Vasquez-Ng" } });
    fireEvent.click(screen.getByRole("button", { name: /^save$/i }));

    expect(mutate).toHaveBeenCalledWith(
      {
        path: { id: "u1" },
        body: { first_name: "Elena", middle_name: "", last_name: "Vasquez-Ng" },
      },
      expect.anything(),
    );
  });
});
