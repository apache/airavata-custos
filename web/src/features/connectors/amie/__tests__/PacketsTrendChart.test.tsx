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
import { PacketsTrendChart } from "../components/PacketsTrendChart";

describe("PacketsTrendChart", () => {
  it("renders an empty-state when there are no buckets", () => {
    render(<PacketsTrendChart buckets={[]} />);
    expect(screen.getByText(/No packet activity/i)).toBeInTheDocument();
  });

  it("renders the chart container and legend when buckets exist", () => {
    render(
      <PacketsTrendChart
        buckets={[
          { date: "2026-05-20", status: "PROCESSED", type: "request_project_create", count: 5 },
          { date: "2026-05-20", status: "FAILED", type: "request_account_create", count: 1 },
          { date: "2026-05-21", status: "PROCESSED", type: "request_project_create", count: 7 },
        ]}
      />,
    );
    expect(screen.getByLabelText(/AMIE packets per day/i)).toBeInTheDocument();
    expect(screen.getByText("PROCESSED")).toBeInTheDocument();
    expect(screen.getByText("FAILED")).toBeInTheDocument();
  });
});
