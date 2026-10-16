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

import type { UserManagementRow } from "@/features/core/users/queries";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const save = vi.hoisted(() => vi.fn(async () => ({})));
vi.mock("@/features/core/users/queries", () => ({
  useDeleteIdentity: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useSaveIdentity: () => ({ mutateAsync: save, isPending: false }),
  userIdentities: () => ({}),
}));
vi.mock("@tanstack/react-query", async (orig) => ({
  ...(await orig<typeof import("@tanstack/react-query")>()),
  useQuery: () => ({
    refetch: vi.fn(),
    dataUpdatedAt: Number.MAX_SAFE_INTEGER,
    data: user.identities,
  }),
}));
vi.mock("@/shared/ui/sonner", async (orig) => ({
  ...(await orig<typeof import("@/shared/ui/sonner")>()),
  dismissConfirm: vi.fn(),
}));

import { IdentitiesSection } from "../IdentitiesSection";

const user: UserManagementRow = {
  id: "00000000-0000-4000-8000-000000000001",
  email: "u1@example.org",
  first_name: "U",
  last_name: "One",
  organization_id: "00000000-0000-4000-8000-000000000002",
  status: "ACTIVE",
  type: "VIRTUAL",
  roles: [],
  identities: [
    {
      id: "00000000-0000-4000-8000-000000000011",
      user_id: "00000000-0000-4000-8000-000000000001",
      source: "oidc",
      external_id: "sub-1",
      oidc_sub: "sub-1",
      created_at: "2026-01-01T00:00:00Z",
    },
    {
      id: "00000000-0000-4000-8000-000000000012",
      user_id: "00000000-0000-4000-8000-000000000001",
      source: "orcid",
      external_id: "0000-0001",
      metadata: '{"verified":true}',
      created_at: "2026-01-01T00:00:00Z",
    },
    {
      id: "00000000-0000-4000-8000-000000000013",
      user_id: "00000000-0000-4000-8000-000000000001",
      source: "cilogon",
      external_id: "c-1",
      oidc_sub: "sub-2",
      created_at: "2026-01-01T00:00:00Z",
    },
    {
      id: "00000000-0000-4000-8000-000000000014",
      user_id: "00000000-0000-4000-8000-000000000001",
      source: "github",
      external_id: "gh-1",
      created_at: "2026-01-01T00:00:00Z",
    },
  ],
  rolesLoading: false,
  identitiesLoading: false,
  rolesError: false,
  identitiesError: false,
};

describe("<IdentitiesSection />", () => {
  it("offers edit and remove on every identity to a users:write holder", () => {
    render(<IdentitiesSection user={user} canWrite />);
    for (const name of [
      "OIDC identity sub-1",
      "ORCID identity 0000-0001",
      "CILogon identity c-1",
    ]) {
      expect(screen.getByRole("button", { name: `Edit ${name}` })).toBeInTheDocument();
      expect(screen.getByRole("button", { name: `Remove ${name}` })).toBeInTheDocument();
    }
  });

  it("saves an edit as the whole identity with the form's fields", async () => {
    render(<IdentitiesSection user={user} canWrite />);
    fireEvent.click(screen.getByRole("button", { name: "Edit ORCID identity 0000-0001" }));
    fireEvent.change(screen.getByLabelText(/External ID/), { target: { value: "0000-0002" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() =>
      expect(save).toHaveBeenCalledWith({
        id: "00000000-0000-4000-8000-000000000012",
        user_id: "00000000-0000-4000-8000-000000000001",
        source: "orcid",
        external_id: "0000-0002",
        email: "",
        metadata: '{"verified":true}',
        created_at: "2026-01-01T00:00:00Z",
      }),
    );
  });

  it("keeps an unlisted source and sends nothing on an untouched save", () => {
    save.mockClear();
    render(<IdentitiesSection user={user} canWrite />);
    fireEvent.click(screen.getByRole("button", { name: "Edit github identity gh-1" }));
    expect(screen.getByLabelText("Source")).toHaveValue("github");
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(save).not.toHaveBeenCalled();
  });
});
