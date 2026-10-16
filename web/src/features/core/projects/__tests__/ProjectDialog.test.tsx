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

import type { ProjectResponse } from "@/generated/core/types.gen";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

const update = { mutateAsync: vi.fn(async () => ({})), isPending: false };
vi.mock("../queries", () => ({
  useCreateProject: () => ({ mutateAsync: vi.fn(async () => ({})), isPending: false }),
  useUpdateProject: () => update,
  useProjectRecord: () => ({ refetch: vi.fn(), dataUpdatedAt: Number.MAX_SAFE_INTEGER }),
  useForgetProject: () => vi.fn(),
}));

vi.mock("@/shared/ui/sonner", () => ({
  dismissConfirm: vi.fn(),
  toastOnSuccess: (pending: Promise<unknown>) => pending,
}));

import { ProjectDialog } from "../components/ProjectDialog";

const project: ProjectResponse = {
  id: "p-1",
  originated_id: "",
  title: "Genomics",
  origination: "internal",
  project_pi_id: "u-old",
  status: "ACTIVE",
  created_time: "2026-01-01T00:00:00Z",
};

describe("<ProjectDialog />", () => {
  it("sends the full record, with the role the previous PI keeps on handover", () => {
    render(<ProjectDialog project={project} />);
    fireEvent.click(screen.getByRole("button", { name: "Edit project" }));
    expect(screen.queryByLabelText("Previous PI becomes")).toBeNull();
    fireEvent.change(screen.getByLabelText("PI user ID"), { target: { value: "u-new" } });
    fireEvent.change(screen.getByLabelText("Previous PI becomes"), {
      target: { value: "ALLOCATION_MANAGER" },
    });
    fireEvent.submit(screen.getByLabelText("Title").closest("form") as HTMLFormElement);
    expect(update.mutateAsync).toHaveBeenCalledWith({
      path: { id: "p-1" },
      body: { title: "Genomics", project_pi_id: "u-new", previous_pi_role: "ALLOCATION_MANAGER" },
    });
  });
});
