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

const user = { id: "u-1", first_name: "Elena", last_name: "Vasquez", status: "ACTIVE" };
vi.mock("../../queries", () => ({
  useMe: () => ({ isPending: false, isError: false, data: { user, privileges: [], roles: [] } }),
  useMyIdentities: () => ({ data: undefined }),
  useMyAccess: () => ({ data: undefined }),
}));
vi.mock("@/features/core/organizations/queries", () => ({
  useOrganization: () => ({ data: undefined }),
}));
vi.mock("../ProfileCard", () => ({ ProfileCard: () => null }));
vi.mock("../AppearanceCard", () => ({ AppearanceCard: () => null }));

import { SettingsPage } from "../SettingsPage";

describe("<SettingsPage />", () => {
  it("leaves linked identities out of the summary when the identities card is not shown", () => {
    render(<SettingsPage />);
    expect(screen.getByText("Your profile, access, and appearance.")).toBeInTheDocument();
  });
});
