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
import { StateMachineBreadcrumb } from "../components/PacketDetailDrawer";

describe("<StateMachineBreadcrumb />", () => {
  // Make sure the trail ends at the reached state, with no time master never recorded for it.
  it("ends at the state the packet actually reached", () => {
    render(
      <StateMachineBreadcrumb
        packet={{
          id: "p-1",
          amie_id: "1",
          type: "request_account_create",
          status: "REFUSED",
          source: "access",
          received_at: "2026-06-08T14:00:00Z",
          decoded_at: "2026-06-08T14:00:01Z",
          updated_at: "2026-06-08T15:00:00Z",
          retries: 0,
        }}
      />,
    );
    expect(
      screen.getAllByRole("listitem").map((li) => li.querySelector("span")?.textContent),
    ).toEqual(["New", "Decoded", "Refused"]);
    expect(screen.getAllByRole("listitem")[2]?.querySelector("time")).toBeNull();
  });
});
