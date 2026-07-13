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
import { describe, expect, it } from "vitest";
import { IdentitiesCard } from "../IdentitiesCard";
import type { UserIdentity } from "../../schemas";

describe("IdentitiesCard", () => {
  it("renders an empty state when no identities are linked", () => {
    render(<IdentitiesCard identities={[]} />);
    expect(screen.getByText("No identities linked to this account.")).toBeInTheDocument();
  });

  it("renders a row per identity with source, external id, and type badge", () => {
    const identities: UserIdentity[] = [
      {
        id: "i1",
        source: "cilogon",
        external_id: "elena@access-ci.org",
        oidc_sub: "http://cilogon.org/serverA/users/1",
        created_at: "2026-03-12T09:00:00Z",
      },
      { id: "i2", source: "comanage", external_id: "CO-4821", oidc_sub: "" },
    ];
    render(<IdentitiesCard identities={identities} />);
    expect(screen.getByText("cilogon")).toBeInTheDocument();
    expect(screen.getByText("elena@access-ci.org")).toBeInTheDocument();
    expect(screen.getByText("OIDC")).toBeInTheDocument();
    expect(screen.getByText("comanage")).toBeInTheDocument();
    expect(screen.getByText("Registry")).toBeInTheDocument();
  });
});
