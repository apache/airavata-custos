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
import { PacketStatusBadge } from "../components/PacketStatusBadge";

describe("PacketStatusBadge", () => {
  it("shows a held packet as WAITING APPROVAL, without the underscore", () => {
    render(<PacketStatusBadge status="WAITING_APPROVAL" ageHours={0} />);
    expect(screen.getByText("WAITING APPROVAL")).toBeInTheDocument();
  });

  it("adds the loud '!' marker on aged FAILED packets", () => {
    const { container } = render(<PacketStatusBadge status="FAILED" ageHours={36} />);
    expect(container.textContent).toContain("FAILED");
    expect(container.textContent).toContain("!");
  });
});
