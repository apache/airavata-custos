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

import { ApiError } from "@/shared/api/client";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const auditor = {
  id: "00000000-0000-4000-8000-0000000000a1",
  name: "Auditor",
  description: "",
  privileges: [],
  holderIds: [],
};
const rows = vi.hoisted(() => ({ data: [] as unknown[], refetch: vi.fn() }));
const row = vi.hoisted(() => ({ data: undefined as unknown, error: null as unknown }));

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({ can: () => false }),
}));

vi.mock("@/features/core/roles/queries", () => ({
  RoleSaveError: class extends Error {},
  useCreateRole: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateRole: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteRole: () => ({ mutateAsync: vi.fn(), isPending: false }),
  usePrivilegeCatalog: () => ({ data: [], isSuccess: false }),
  useRoleRows: () => ({ data: rows.data, isSuccess: true, refetch: rows.refetch }),
  useRoleRow: () => ({
    data: row.data,
    error: row.error,
    refetch: vi.fn(),
    dataUpdatedAt: row.error ? 0 : Number.MAX_SAFE_INTEGER,
    errorUpdatedAt: row.error ? Number.MAX_SAFE_INTEGER : 0,
  }),
}));

import { RolesGrid } from "../RolesGrid";

describe("<RolesGrid />", () => {
  it("says an edited role deleted elsewhere is gone and refreshes the grid", () => {
    rows.data = [auditor];
    row.data = auditor;
    const { rerender } = render(<RolesGrid />);
    fireEvent.click(screen.getByRole("button", { name: "Edit Auditor" }));
    expect(screen.getByLabelText("Name")).toHaveValue("Auditor");
    rows.data = [];
    row.error = new ApiError(
      404,
      "/roles/00000000-0000-4000-8000-0000000000a1",
      null,
      "role not found",
    );
    rerender(<RolesGrid />);
    expect(screen.getByText("This role no longer exists")).toBeVisible();
    expect(rows.refetch).toHaveBeenCalled();
  });
});
