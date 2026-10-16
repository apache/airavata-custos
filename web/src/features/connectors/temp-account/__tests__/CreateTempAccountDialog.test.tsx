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

const create = vi.hoisted(() => ({ isPending: false, mutateAsync: vi.fn(async () => ({})) }));
const organizations = vi.hoisted(() =>
  vi.fn(() => ({
    data: {
      items: [
        { id: "system", name: "System" },
        { id: "o-1", name: "Org" },
      ],
    },
  })),
);
const ability = vi.hoisted(() => ({ canReadOrganization: false }));
vi.mock("../queries", () => ({ useCreateTempAccount: () => create }));
vi.mock("@/features/core/organizations/queries", () => ({ useOrganizations: organizations }));
vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({ can: () => ability.canReadOrganization }),
}));

import { CreateTempAccountDialog } from "../components/CreateTempAccountDialog";

describe("<CreateTempAccountDialog />", () => {
  it("asks for the organization id without Organization read", () => {
    ability.canReadOrganization = false;
    render(<CreateTempAccountDialog onCreated={() => {}} />);
    fireEvent.click(screen.getByRole("button", { name: "+ Create temporary account" }));
    expect(organizations).toHaveBeenCalledWith({ limit: 200 }, false);
    fireEvent.change(screen.getByLabelText("First name"), { target: { value: "Tia" } });
    fireEvent.change(screen.getByLabelText("Email"), { target: { value: "tia@example.org" } });
    fireEvent.change(screen.getByLabelText("Organization ID"), { target: { value: "o-1" } });
    fireEvent.submit(screen.getByLabelText("Email").closest("form") as HTMLFormElement);
    expect(create.mutateAsync).toHaveBeenCalledWith({
      first_name: "Tia",
      last_name: "",
      email: "tia@example.org",
      organization_id: "o-1",
    });
  });

  it("picks the organization from the listed ones with Organization read", () => {
    ability.canReadOrganization = true;
    render(<CreateTempAccountDialog onCreated={() => {}} />);
    fireEvent.click(screen.getByRole("button", { name: "+ Create temporary account" }));
    expect(screen.getAllByRole("option").map((o) => o.textContent)).toEqual(["System", "Org"]);
    expect(screen.queryByLabelText("Organization ID")).toBeNull();
  });
});
