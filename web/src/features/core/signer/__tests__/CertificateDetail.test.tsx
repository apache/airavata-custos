// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0.

import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError } from "@/shared/api/client";
import type { Certificate } from "../schemas";

const baseCertificate: Certificate = {
  tenant_id: "tenant-1",
  client_id: "signer-client",
  serial_number: 42,
  key_id: "key-42",
  principal: "someone-else",
  user_email: "admin@example.org",
  public_key_fingerprint: "SHA256:pk",
  ca_fingerprint: "SHA256:ca",
  valid_after: 1_700_000_000,
  valid_before: 4_102_444_800,
  issued_at: 1_700_000_000,
  source_ip: "192.0.2.42",
  granted_extensions: ["permit-pty", "permit-user-rc"],
  force_command: "/usr/bin/id",
  revoked: false,
};

const state = vi.hoisted(() => ({
  canManage: true,
  query: {} as {
    data?: Certificate;
    isLoading: boolean;
    error: Error | null;
    refetch: ReturnType<typeof vi.fn>;
  },
  mutate: vi.fn(),
  isPending: false,
  toastSuccess: vi.fn(),
}));

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({
    can: (action: string, subject: string) =>
      action === "manage" && subject === "Signer" ? state.canManage : true,
  }),
}));

vi.mock("sonner", () => ({ toast: { success: state.toastSuccess } }));

vi.mock("../queries", () => ({
  useCertificate: () => state.query,
  useRevokeCertificate: () => ({ mutate: state.mutate, isPending: state.isPending }),
}));

import { CertificateDetail } from "../components/CertificateDetail";

beforeEach(() => {
  state.canManage = true;
  state.query = {
    data: { ...baseCertificate },
    isLoading: false,
    error: null,
    refetch: vi.fn(),
  };
  state.mutate.mockReset();
  state.toastSuccess.mockReset();
  state.isPending = false;
});

function submitRevoke(reason = "compromised") {
  fireEvent.click(screen.getByRole("button", { name: /^Revoke$/ }));
  fireEvent.change(screen.getByLabelText(/reason/i), { target: { value: reason } });
  fireEvent.click(screen.getByRole("button", { name: /^Confirm revoke$/ }));
}

describe("<CertificateDetail />", () => {
  it("renders complete issuance metadata", () => {
    render(<CertificateDetail serial="42" />);
    for (const value of [
      "tenant-1",
      "signer-client",
      "admin@example.org",
      "key-42",
      "SHA256:pk",
      "SHA256:ca",
      "192.0.2.42",
      "permit-pty, permit-user-rc",
      "/usr/bin/id",
    ]) {
      expect(screen.getByText(value)).toBeInTheDocument();
    }
  });

  it("renders a loading skeleton", () => {
    state.query = { ...state.query, data: undefined, isLoading: true };
    const { container } = render(<CertificateDetail serial="42" />);
    expect(container.querySelector(".animate-pulse")).toBeInTheDocument();
  });

  it("renders not-found handling for a 404", () => {
    state.query = {
      ...state.query,
      data: undefined,
      error: new ApiError(404, "/certificate/99", { error: "not_found" }),
    };
    render(<CertificateDetail serial="99" />);
    expect(screen.getByRole("heading", { name: /certificate not found/i })).toBeInTheDocument();
  });

  it("renders a retryable detail failure", () => {
    state.query = { ...state.query, data: undefined, error: new Error("signer unavailable") };
    render(<CertificateDetail serial="42" />);
    fireEvent.click(screen.getByRole("button", { name: /try again/i }));
    expect(state.query.refetch).toHaveBeenCalledOnce();
  });

  it("shows the Revoke button only for an active writer", () => {
    render(<CertificateDetail serial="42" />);
    expect(screen.getByRole("button", { name: /^Revoke$/ })).toBeInTheDocument();
  });

  it("hides Revoke without write privilege", () => {
    state.canManage = false;
    render(<CertificateDetail serial="42" />);
    expect(screen.queryByRole("button", { name: /^Revoke$/ })).not.toBeInTheDocument();
  });

  it.each([
    ["revoked", { revoked: true, revoked_at: 1_700_500_000, revocation_reason: "old" }],
    ["expired", { valid_before: 1 }],
    ["not yet valid", { valid_after: 4_102_444_800 }],
  ])("hides Revoke when the certificate is %s", (_label, overrides) => {
    state.query.data = { ...baseCertificate, ...overrides };
    render(<CertificateDetail serial="42" />);
    expect(screen.queryByRole("button", { name: /^Revoke$/ })).not.toBeInTheDocument();
  });

  it("renders authoritative revocation metadata", () => {
    state.query.data = {
      ...baseCertificate,
      revoked: true,
      revoked_at: 1_700_500_000,
      revocation_reason: "original reason",
      revoked_by: "admin-id",
    };
    render(<CertificateDetail serial="42" />);
    expect(screen.getByText("original reason")).toBeInTheDocument();
    expect(screen.getByText("admin-id")).toBeInTheDocument();
  });

  it.each([
    [403, "You no longer have permission to revoke certificates."],
    [409, "This certificate is no longer active and cannot be revoked."],
  ])("keeps the dialog open for API %s", (status, expected) => {
    state.mutate.mockImplementation((_variables, options) => {
      options.onError(new ApiError(status, "/revoke", { error: "failure" }));
    });
    render(<CertificateDetail serial="42" />);
    submitRevoke();
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText(expected)).toBeInTheDocument();
  });

  it("keeps the dialog open after a retryable network failure", () => {
    state.mutate.mockImplementation((_variables, options) => {
      options.onError(new Error("Network request failed"));
    });
    render(<CertificateDetail serial="42" />);
    submitRevoke();
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Network request failed")).toBeInTheDocument();
  });

  it("treats an already-revoked response as informational success", () => {
    state.mutate.mockImplementation((_variables, options) => {
      options.onSuccess({ already_revoked: true });
    });
    render(<CertificateDetail serial="42" />);
    submitRevoke();
    expect(state.toastSuccess).toHaveBeenCalledWith("Certificate was already revoked");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});
