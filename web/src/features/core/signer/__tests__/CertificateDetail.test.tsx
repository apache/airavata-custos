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
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Certificate } from "../schemas";

// Presentation gating only — backend enforces the real authorization.
const state = vi.hoisted(() => ({
  canManage: true,
  cert: {
    serial_number: 42,
    key_id: "k",
    principal: "someone-else",
    public_key_fingerprint: "SHA256:pk",
    ca_fingerprint: "SHA256:ca",
    valid_after: 1_700_000_000,
    valid_before: 4_102_444_800,
    issued_at: 1_700_000_000,
    revoked: false,
  } as Certificate,
}));

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({
    can: (action: string, subject: string) =>
      action === "manage" && subject === "Signer" ? state.canManage : true,
  }),
}));

vi.mock("../queries", () => ({
  useCertificate: () => ({ data: state.cert, isLoading: false, error: null, refetch: vi.fn() }),
  useRevokeCertificate: () => ({ mutate: vi.fn(), isPending: false }),
}));

import { CertificateDetail } from "../components/CertificateDetail";

beforeEach(() => {
  state.canManage = true;
  state.cert = { ...state.cert, revoked: false };
});

describe("<CertificateDetail /> revoke gating", () => {
  it("shows the Revoke button for an admin with the signer write privilege", () => {
    state.canManage = true;
    render(<CertificateDetail serial="42" />);
    expect(screen.getByRole("button", { name: /^Revoke$/ })).toBeInTheDocument();
  });

  it("hides the Revoke button for a user without the privilege (ownership is irrelevant)", () => {
    state.canManage = false;
    render(<CertificateDetail serial="42" />);
    expect(screen.queryByRole("button", { name: /^Revoke$/ })).not.toBeInTheDocument();
  });

  it("hides the Revoke button when the certificate is already revoked", () => {
    state.canManage = true;
    state.cert = { ...state.cert, revoked: true, revoked_at: 1_700_500_000, revocation_reason: "old" };
    render(<CertificateDetail serial="42" />);
    expect(screen.queryByRole("button", { name: /^Revoke$/ })).not.toBeInTheDocument();
  });
});
