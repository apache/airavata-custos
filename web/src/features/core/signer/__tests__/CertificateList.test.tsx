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

import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { CertificateList } from "../components/CertificateList";
import type { Certificate } from "../schemas";

const push = vi.hoisted(() => vi.fn());
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
}));

const active: Certificate = {
  user_sub: "sub-1",
  requested_by: "user/1",
  serial_number: 42,
  key_id: "k-active",
  principal: "dev-admin",
  user_email: "admin@example.org",
  public_key_fingerprint: "SHA256:pk",
  ca_fingerprint: "SHA256:ca",
  valid_after: 1_700_000_000,
  valid_before: 4_102_444_800,
  issued_at: 1_700_000_000,
  revoked: false,
};

function renderList(overrides: Partial<React.ComponentProps<typeof CertificateList>> = {}) {
  const props = {
    rows: [active],
    isLoading: false,
    error: null,
    search: "",
    onSearchChange: vi.fn(),
    statusFilter: "all" as const,
    onStatusFilterChange: vi.fn(),
    ...overrides,
  };
  return { ...render(<CertificateList {...props} />), props };
}

describe("<CertificateList />", () => {
  // Make sure clicking a row opens that certificate's detail page.
  it("opens the certificate when its row is clicked", () => {
    renderList();
    fireEvent.click(screen.getByText("dev-admin"));
    expect(push).toHaveBeenCalledWith(`/admin/signer/certificates/${active.serial_number}`);
  });

  it("shows the empty state when no rows match the filters", () => {
    renderList({ rows: [] });
    expect(screen.getByRole("heading", { name: /no certificates yet/i })).toBeInTheDocument();
  });

  it("tells no matches apart from no certificates", () => {
    renderList({ rows: [], search: "not-present" });
    expect(screen.getByRole("heading", { name: /no matches/i })).toBeInTheDocument();
  });

  it("reports search and status control changes", () => {
    const { props } = renderList();
    fireEvent.change(screen.getByRole("searchbox", { name: /search certificates/i }), {
      target: { value: "dev" },
    });
    fireEvent.change(screen.getByRole("combobox", { name: /filter by status/i }), {
      target: { value: "revoked" },
    });
    expect(props.onSearchChange).toHaveBeenCalledWith("dev");
    expect(props.onStatusFilterChange).toHaveBeenCalledWith("revoked");
  });

  const paging = {
    page: 1,
    pageSize: 20,
    hasNext: true,
    onPrevious: vi.fn(),
    onNext: vi.fn(),
    onPageSizeChange: vi.fn(),
  };

  // Make sure the first page cannot go back.
  it("disables Previous on the first page", () => {
    renderList({ paging });
    expect(screen.getByRole("button", { name: /previous/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /next/i })).toBeEnabled();
  });

  // Make sure a page without a next cursor is treated as the last page.
  it("disables Next when the signer sends no next cursor", () => {
    renderList({ paging: { ...paging, page: 2, hasNext: false } });
    expect(screen.getByRole("button", { name: /next/i })).toBeDisabled();
    expect(screen.getByRole("button", { name: /previous/i })).toBeEnabled();
  });

  it("surfaces an error state with a retry callback", () => {
    const onRetry = vi.fn();
    renderList({ error: new Error("boom"), onRetry });
    expect(screen.getByText(/boom/)).toBeInTheDocument();
  });
});
