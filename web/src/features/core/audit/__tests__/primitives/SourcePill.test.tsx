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
import { SourcePill } from "@/features/core/audit/components/primitives/SourcePill";

describe("SourcePill", () => {
  it.each(["amie", "comanage", "slurm", "http", "core"])("renders the %s source", (src) => {
    render(<SourcePill source={src} />);
    expect(screen.getByText(src)).toBeInTheDocument();
  });

  it("falls back to muted styling for an unknown source without crashing", () => {
    render(<SourcePill source="future-connector" />);
    const node = screen.getByText("future-connector");
    expect(node).toBeInTheDocument();
    expect(node.className).toContain("bg-muted");
  });

  it("lowercases the displayed source name", () => {
    render(<SourcePill source="AMIE" />);
    expect(screen.getByText("amie")).toBeInTheDocument();
  });
});
