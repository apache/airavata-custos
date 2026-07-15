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

import { afterEach, describe, expect, it, vi } from "vitest";
import { getCertificate, listCertificates, revokeCertificate } from "../api";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

function mockResponse(status: number, body: unknown): Response {
  const text = typeof body === "string" ? body : JSON.stringify(body);
  return new Response(text, { status, headers: { "content-type": "application/json" } });
}

afterEach(() => {
  fetchMock.mockReset();
});

const certificate = {
  serial_number: 42,
  key_id: "k",
  principal: "dev-admin",
  public_key_fingerprint: "SHA256:pk",
  ca_fingerprint: "SHA256:ca",
  valid_after: 1_700_000_000,
  valid_before: 4_102_444_800,
  issued_at: 1_700_000_000,
  revoked: false,
};

describe("listCertificates", () => {
  it("round-trips pagination and returns the envelope", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(200, { certificates: [certificate], total: 1, limit: 20, offset: 0 }),
    );
    const result = await listCertificates({ limit: 20, offset: 0 });
    expect(result.total).toBe(1);
    expect(result.certificates).toHaveLength(1);
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).toContain("/api/v1/signer/certificates");
    expect(url).toContain("limit=20");
    expect(url).toContain("offset=0");
  });
});

describe("getCertificate", () => {
  it("validates the response", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, certificate));
    const out = await getCertificate(42);
    expect(out.serial_number).toBe(42);
    expect(fetchMock.mock.calls[0]?.[0]).toContain("/api/v1/signer/certificates/42");
  });
});

describe("revokeCertificate", () => {
  it("POSTs the reason to /certificates/{serial}/revoke and parses the response", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(200, {
        success: true,
        message: "Certificate revoked successfully",
        serial_number: 42,
        revoked: true,
        revoked_at: 1_700_500_000,
        reason: "compromised",
      }),
    );
    const out = await revokeCertificate(42, "compromised");
    expect(out.revoked).toBe(true);
    expect(out.reason).toBe("compromised");

    const [url, init] = fetchMock.mock.calls[0] ?? [];
    expect(url).toContain("/api/v1/signer/certificates/42/revoke");
    expect(init?.method).toBe("POST");
    const sent = JSON.parse(String(init?.body));
    expect(sent).toEqual({ reason: "compromised" });
  });

  it("rejects when the backend denies the revoke (403)", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(403, { error: "forbidden", message: "Requires the signer:certificates:write privilege" }),
    );
    await expect(revokeCertificate(42, "compromised")).rejects.toThrow();
  });
});
