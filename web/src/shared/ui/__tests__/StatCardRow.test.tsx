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

import { render } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { StatCardRow } from "../StatCardRow";

describe("StatCardRow", () => {
  it("defaults to a 3-col responsive grid (backward-compat with allocation header)", () => {
    const { container } = render(
      <StatCardRow>
        <div>A</div>
        <div>B</div>
        <div>C</div>
      </StatCardRow>,
    );
    const row = container.firstElementChild;
    expect(row?.className).toContain("md:grid-cols-3");
  });

  it("renders 5-col KPI strip when cols=5", () => {
    const { container } = render(
      <StatCardRow cols={5}>
        <div>A</div>
      </StatCardRow>,
    );
    const row = container.firstElementChild;
    expect(row?.className).toContain("sm:grid-cols-2");
    expect(row?.className).toContain("lg:grid-cols-5");
  });

  it("renders 4-col strip when cols=4", () => {
    const { container } = render(
      <StatCardRow cols={4}>
        <div>A</div>
      </StatCardRow>,
    );
    expect(container.firstElementChild?.className).toContain("lg:grid-cols-4");
  });

  it("passes through className", () => {
    const { container } = render(
      <StatCardRow cols={5} className="custom-marker">
        <div>A</div>
      </StatCardRow>,
    );
    expect(container.firstElementChild?.className).toContain("custom-marker");
  });
});
