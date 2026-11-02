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

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook } from "@testing-library/react";
import { createElement, type ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

const api = vi.hoisted(() => ({ revokeCertificate: vi.fn() }));

vi.mock("../api", () => ({
  getCertificate: vi.fn(),
  listCertificates: vi.fn(),
  revokeCertificate: api.revokeCertificate,
}));

import { signerKeys, useRevokeCertificate } from "../queries";

describe("signerKeys", () => {
  it("namespaces under 'signer-certificates'", () => {
    expect(signerKeys.all).toEqual(["signer-certificates"]);
  });

  it("list key carries the params", () => {
    const params = { limit: 20, offset: 0 };
    expect(signerKeys.list(params)).toEqual(["signer-certificates", "list", params]);
  });

  it("detail key stringifies the serial", () => {
    expect(signerKeys.detail(42)).toEqual(["signer-certificates", "detail", "42"]);
    expect(signerKeys.detail("42")).toEqual(["signer-certificates", "detail", "42"]);
  });
});

describe("useRevokeCertificate", () => {
  it("invalidates authoritative list and detail state after success", async () => {
    api.revokeCertificate.mockResolvedValueOnce({
      success: true,
      message: "Certificate revoked successfully",
      serial_number: 42,
      revoked: true,
      revoked_at: 1_700_500_000,
      reason: "compromised",
      already_revoked: false,
    });
    const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
    const invalidate = vi.spyOn(client, "invalidateQueries");
    const wrapper = ({ children }: { children: ReactNode }) =>
      createElement(QueryClientProvider, { client }, children);
    const { result } = renderHook(() => useRevokeCertificate(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ serial: 42, reason: "compromised" });
    });

    expect(invalidate).toHaveBeenCalledWith({ queryKey: signerKeys.all });
    expect(invalidate).toHaveBeenCalledWith({ queryKey: signerKeys.detail(42) });
  });
});
