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

import { render, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { AgreementsList } from "../components/AgreementsList";
import type { AgreementListItem } from "../schemas";

const row = (version_label: string, status: AgreementListItem["status"]): AgreementListItem => ({
  id: `agr-${version_label}`,
  title: "Acceptable Use Policy",
  version_label,
  status,
  accepted_count: 4,
  pending_count: 2,
});

function pendingCell(version: string) {
  const link = screen.getByRole("link", { name: version });
  const cells = within(link.closest("tr") as HTMLElement).getAllByRole("cell");
  return cells[cells.length - 1];
}

describe("<AgreementsList />", () => {
  // Checks that a draft or removed row shows a dash for pending, and the published row its count.
  it("shows the pending count only for a published version", () => {
    render(
      <AgreementsList
        rows={[row("2.0", "DRAFT"), row("1.0", "PUBLISHED"), row("0.9", "REMOVED")]}
        isLoading={false}
        error={null}
      />,
    );
    expect(pendingCell("1.0")).toHaveTextContent("2");
    expect(pendingCell("2.0")).toHaveTextContent("—");
    expect(pendingCell("0.9")).toHaveTextContent("—");
  });
});
