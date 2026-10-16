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

let canReadAllocations = false;
vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({ can: () => canReadAllocations }),
}));
vi.mock("../../queries", () => ({
  useAnalyticsContexts: () => ({ data: [], isLoading: false, error: null }),
  useUsageSummary: () => ({ data: undefined, isLoading: false, error: null }),
}));
vi.mock("../../useAnalyticsSelection", () => ({
  useAnalyticsSelection: () => ({ project: undefined, allocation: undefined, select: vi.fn() }),
}));

import { AnalyticsPage } from "../AnalyticsPage";

describe("<AnalyticsPage />", () => {
  it("points an allocation reader without memberships at the allocation itself", () => {
    canReadAllocations = true;
    render(<AnalyticsPage />);
    expect(
      screen.getByText(
        "You are not a member of any allocation. Open an allocation to see its usage.",
      ),
    ).toBeInTheDocument();
  });
});
