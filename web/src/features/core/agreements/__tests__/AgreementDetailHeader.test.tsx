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
import { AgreementDetailHeader } from "../components/AgreementDetailHeader";
import type { AgreementDetail } from "../schemas";

const agreement = (status: AgreementDetail["status"]): AgreementDetail => ({
  id: "agr-1",
  title: "Acceptable Use Policy",
  version_label: "1.0",
  status,
  summary: "",
  body_md: "",
  change_note: "",
  created_at: "2026-03-02T09:00:00Z",
  accepted_count: 0,
  pending_count: 0,
});

function renderHeader(status: AgreementDetail["status"]) {
  render(<AgreementDetailHeader agreement={agreement(status)} canManage onAction={vi.fn()} />);
}

describe("<AgreementDetailHeader />", () => {
  // Checks that a draft offers Publish and Delete, and not Remove, which only applies to a published version.
  it("offers publish and delete for a draft, not remove", () => {
    renderHeader("DRAFT");
    expect(screen.getByRole("button", { name: "Publish" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Delete draft" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Remove" })).not.toBeInTheDocument();
  });

  // Checks that a published version can only be removed.
  it("offers only remove for a published version", () => {
    renderHeader("PUBLISHED");
    expect(screen.getByRole("button", { name: "Remove" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Publish" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Delete draft" })).not.toBeInTheDocument();
  });
});
