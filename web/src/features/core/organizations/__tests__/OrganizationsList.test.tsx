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

import type { Organization } from "@/generated/core/types.gen";
import { Button } from "@/shared/ui/button";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { OrganizationsList } from "../components/OrganizationsList";

const org: Organization = {
  id: "15000001-0000-4000-8000-000015000001",
  name: "Georgia Institute of Technology",
  originated_id: "GATECH",
};

function renderList(overrides: Partial<React.ComponentProps<typeof OrganizationsList>> = {}) {
  const props = {
    rows: [org],
    isLoading: false,
    error: null,
    page: 1,
    pageSize: 50,
    total: 1,
    onPageChange: vi.fn(),
    ...overrides,
  };
  return { ...render(<OrganizationsList {...props} />), props };
}

describe("<OrganizationsList />", () => {
  it("renders a row with the organization fields", () => {
    renderList();
    expect(screen.getByText(org.name)).toBeInTheDocument();
    expect(screen.getByText(org.originated_id)).toBeInTheDocument();
    expect(screen.getByText(org.id)).toBeInTheDocument();
  });

  it("truncates a long name and originated ID and keeps each whole in the title", () => {
    const name = "N".repeat(255);
    const originated = "X".repeat(255);
    renderList({ rows: [{ ...org, name, originated_id: originated }] });
    for (const long of [name, originated]) {
      expect(screen.getByText(long)).toHaveAttribute("title", long);
      expect(screen.getByText(long)).toHaveClass("truncate");
    }
  });

  it("shows the empty state when there are no rows", () => {
    renderList({ rows: [], total: 0 });
    expect(screen.getByRole("heading", { name: /no organizations yet/i })).toBeInTheDocument();
  });

  it("offers the first page when the URL page is past the end", () => {
    const { props } = renderList({ rows: [], total: 3, page: 4 });
    expect(
      screen.getByRole("heading", { name: "No organizations on page 4" }),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "First page" }));
    expect(props.onPageChange).toHaveBeenCalledWith(1);
  });

  it("renders the create CTA only when one is provided", () => {
    const { rerender } = renderList();
    expect(screen.queryByRole("button", { name: /create organization/i })).not.toBeInTheDocument();
    rerender(
      <OrganizationsList
        rows={[org]}
        isLoading={false}
        error={null}
        page={1}
        pageSize={50}
        total={1}
        onPageChange={vi.fn()}
        headerCta={<Button>+ Create organization</Button>}
      />,
    );
    expect(screen.getByRole("button", { name: /create organization/i })).toBeInTheDocument();
  });

  it("surfaces an error state with a retry callback", () => {
    const onRetry = vi.fn();
    renderList({ error: new Error("boom"), onRetry });
    expect(screen.queryByText(/boom/)).toBeNull();
    screen.getByRole("button", { name: "Try again" }).click();
    expect(onRetry).toHaveBeenCalled();
  });
});
