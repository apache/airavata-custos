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
import type { ComputeAllocation } from "@/generated/core/zod.gen";
import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

const status = { mutate: vi.fn(), reset: vi.fn(), isPending: false, error: null };
vi.mock("@/shared/casl/permissions", () => ({
  usePermissions: () => ({ allocationsWrite: true }),
  requires: () => "",
}));
vi.mock("../queries", () => ({ useSetAllocationStatus: () => status }));

import { AllocationStatusButton } from "../components/AllocationStatusButton";

const allocation = (s: ComputeAllocation["status"]) =>
  ({ id: "a-1", name: "Alpha", status: s }) as ComputeAllocation;

describe("<AllocationStatusButton />", () => {
  beforeEach(() => status.mutate.mockClear());

  it("deactivates only after the confirm", () => {
    render(<AllocationStatusButton allocation={allocation("ACTIVE")} />);
    fireEvent.click(screen.getByRole("button", { name: "Deactivate" }));
    expect(status.mutate).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Deactivate allocation" }));
    expect(status.mutate).toHaveBeenCalledWith("INACTIVE", expect.anything());
  });

  it("reactivates at once", () => {
    render(<AllocationStatusButton allocation={allocation("INACTIVE")} />);
    fireEvent.click(screen.getByRole("button", { name: "Reactivate" }));
    expect(status.mutate).toHaveBeenCalledWith("ACTIVE", expect.anything());
  });
});
