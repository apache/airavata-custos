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

vi.mock("next-auth/react", () => ({
  useSession: () => ({
    status: "authenticated",
    data: { user: { name: "IdP Name", email: "pat@x.org" } },
  }),
}));

vi.mock("@/features/core/identity/queries", () => ({
  useMe: () => ({
    data: {
      user: {
        id: "u-1",
        first_name: "Saved",
        middle_name: "Q",
        last_name: "Name",
        email: "pat@account.org",
      },
    },
  }),
}));

import { UserPill } from "../UserPill";

describe("<UserPill />", () => {
  it("shows the account's full name and email, not the IdP's sign-in ones", () => {
    render(<UserPill />);
    expect(screen.getByText("Saved Q Name")).toBeInTheDocument();
    expect(screen.queryByText("IdP Name")).not.toBeInTheDocument();
    expect(screen.getByText("pat@account.org")).toBeInTheDocument();
    expect(screen.queryByText("pat@x.org")).not.toBeInTheDocument();
  });
});
