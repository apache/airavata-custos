/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { StatusPill } from "@/features/core/audit/components/primitives/StatusPill";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

describe("StatusPill", () => {
  // Make sure the status is announced in words, never by color alone.
  it("names the status for screen readers", () => {
    render(<StatusPill status="retrying" label="Retrying 4 of 10" />);
    expect(screen.getByRole("status")).toHaveAttribute("aria-label", "Status: Retrying 4 of 10");
    expect(screen.getByText("Retrying 4 of 10")).toBeInTheDocument();
  });
});
