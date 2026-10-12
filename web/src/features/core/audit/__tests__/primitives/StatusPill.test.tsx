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
  // Make sure only a pill that changes in place is a live region, so a list does not announce every row.
  it("is a live region only when asked", () => {
    const { rerender } = render(
      <StatusPill status="retrying" label="Waiting, 6 of 10 left this run" />,
    );
    expect(screen.queryByRole("status")).toBeNull();
    expect(screen.getByText("Waiting, 6 of 10 left this run")).toBeInTheDocument();
    rerender(<StatusPill status="retrying" label="Waiting, 6 of 10 left this run" live />);
    expect(screen.getByRole("status")).toHaveAttribute(
      "aria-label",
      "Status: Waiting, 6 of 10 left this run",
    );
  });
});
