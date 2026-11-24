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

const state = vi.hoisted(() => ({ canRead: true }));

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({
    cannot: (action: string, subject: string) =>
      action === "read" && subject === "Signer" ? !state.canRead : false,
  }),
}));

import { SignerPermissionGate } from "../PermissionGate";

describe("<SignerPermissionGate />", () => {
  it("renders signer administration for a reader", () => {
    state.canRead = true;
    render(
      <SignerPermissionGate>
        <div>Signer administration</div>
      </SignerPermissionGate>,
    );
    expect(screen.getByText("Signer administration")).toBeInTheDocument();
  });

  it("fails closed without signer read privilege", () => {
    state.canRead = false;
    render(
      <SignerPermissionGate>
        <div>Signer administration</div>
      </SignerPermissionGate>,
    );
    expect(screen.queryByText("Signer administration")).not.toBeInTheDocument();
    expect(screen.getByText("Not permitted.")).toBeInTheDocument();
  });
});
