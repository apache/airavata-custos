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

import { describe, expect, it } from "vitest";
import type { Certificate } from "../schemas";
import {
  formatCertificateLifetime,
  getCertificateStatus,
  statusBadgeVariantFromCertificateStatus,
} from "../status";

const base: Certificate = {
  serial_number: 1,
  key_id: "k",
  principal: "p",
  public_key_fingerprint: "SHA256:pk",
  ca_fingerprint: "SHA256:ca",
  valid_after: 1_700_000_000,
  valid_before: 1_700_086_400,
  issued_at: 1_700_000_000,
  revoked: false,
};

const now = new Date(1_700_043_200 * 1000); // midway through the validity window

describe("getCertificateStatus", () => {
  it("returns revoked when revoked, regardless of validity", () => {
    expect(getCertificateStatus({ ...base, revoked: true }, now)).toBe("revoked");
  });

  it("returns expired when valid_before is in the past", () => {
    expect(getCertificateStatus({ ...base, valid_before: 1_600_000_000 }, now)).toBe("expired");
  });

  it("returns active within the validity window", () => {
    expect(getCertificateStatus(base, now)).toBe("active");
  });
});

describe("statusBadgeVariantFromCertificateStatus", () => {
  it("maps each status to a badge variant", () => {
    expect(statusBadgeVariantFromCertificateStatus("active")).toBe("active");
    expect(statusBadgeVariantFromCertificateStatus("expired")).toBe("expired");
    expect(statusBadgeVariantFromCertificateStatus("revoked")).toBe("deleted");
  });
});

describe("formatCertificateLifetime", () => {
  it("renders whole-hour lifetimes", () => {
    expect(formatCertificateLifetime({ ...base, valid_after: 0, valid_before: 7200 })).toBe(
      "2 hours",
    );
  });
});
