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
import { CertificateList } from "../components/CertificateList";
import type { Certificate } from "../schemas";

const active: Certificate = {
  serial_number: 42,
  key_id: "k-active",
  principal: "dev-admin",
  public_key_fingerprint: "SHA256:pk",
  ca_fingerprint: "SHA256:ca",
  valid_after: 1_700_000_000,
  valid_before: 4_102_444_800,
  issued_at: 1_700_000_000,
  revoked: false,
};

const revoked: Certificate = {
  ...active,
  serial_number: 44,
  key_id: "k-revoked",
  revoked: true,
  revoked_at: 1_700_500_000,
  revocation_reason: "Key compromised",
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
  it("renders the serial as a link to the detail page", () => {
    renderList();
    const link = screen.getByRole("link", { name: String(active.serial_number) });
    expect(link).toHaveAttribute("href", `/signer/certificates/${active.serial_number}`);
  });

  it("shows the empty state when no rows match the filters", () => {
    renderList({ rows: [] });
    expect(screen.getByRole("heading", { name: /no certificates yet/i })).toBeInTheDocument();
  });

  it("filters client-side by status", () => {
    renderList({ rows: [active, revoked], statusFilter: "revoked" });
    expect(
      screen.queryByRole("link", { name: String(active.serial_number) }),
    ).not.toBeInTheDocument();
    expect(screen.getByRole("link", { name: String(revoked.serial_number) })).toBeInTheDocument();
  });

  it("surfaces an error state with a retry callback", () => {
    const onRetry = vi.fn();
    renderList({ error: new Error("boom"), onRetry });
    expect(screen.getByText(/boom/)).toBeInTheDocument();
  });
});
