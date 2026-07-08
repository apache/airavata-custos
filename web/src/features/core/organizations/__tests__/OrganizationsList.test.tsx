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
import { Button } from "@/shared/ui/button";
import { OrganizationsList } from "../components/OrganizationsList";
import type { Organization } from "../schemas";

const org: Organization = {
  id: "org-gatech",
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

  it("shows the empty state when there are no rows", () => {
    renderList({ rows: [], total: 0 });
    expect(screen.getByRole("heading", { name: /no organizations yet/i })).toBeInTheDocument();
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
    expect(screen.getByText(/boom/)).toBeInTheDocument();
  });
});
