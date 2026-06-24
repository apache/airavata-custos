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
import { Cpu } from "lucide-react";
import { describe, expect, it } from "vitest";
import { StatCard } from "../StatCard";
import { StatCardRow } from "../StatCardRow";

describe("StatCard", () => {
  it("renders text variant with title, value, and sub", () => {
    render(
      <StatCard variant="text" title="Remaining credits" value="324,241" sub="Updated 2m ago" />,
    );
    expect(screen.getByText("Remaining credits")).toBeInTheDocument();
    expect(screen.getByText("324,241")).toBeInTheDocument();
    expect(screen.getByText("Updated 2m ago")).toBeInTheDocument();
  });

  it("renders progress variant with usage bar and percent", () => {
    render(
      <StatCard variant="progress" icon={Cpu} title="GPU hours" value="6/6 hrs" percent={50} />,
    );
    expect(screen.getByText("GPU hours")).toBeInTheDocument();
    expect(screen.getByText("6/6 hrs")).toBeInTheDocument();
    expect(screen.getByText("50%")).toBeInTheDocument();
    expect(screen.getByRole("progressbar")).toBeInTheDocument();
  });

  it("keeps legacy sublabel prop working", () => {
    render(<StatCard title="Used" value="42" sublabel="Last 30 days" />);
    expect(screen.getByText("Last 30 days")).toBeInTheDocument();
  });

  it("StatCardRow wraps children in a 3-col responsive grid", () => {
    const { container } = render(
      <StatCardRow>
        <StatCard variant="text" title="A" value="1" />
        <StatCard variant="text" title="B" value="2" />
        <StatCard variant="text" title="C" value="3" />
      </StatCardRow>,
    );
    const row = container.firstElementChild;
    expect(row?.className).toContain("md:grid-cols-3");
  });
});
