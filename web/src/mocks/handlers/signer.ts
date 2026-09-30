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

import { http, HttpResponse } from "msw";
import certificatesFixture from "@/features/core/signer/__fixtures__/certificates.json";
import type { Certificate } from "@/features/core/signer/schemas";
import { effectivePrivileges } from "./privileges";

const certificates: Certificate[] = (certificatesFixture as Certificate[]).map((c) => ({ ...c }));

type SignerScenario =
  | "default"
  | "empty"
  | "paginated"
  | "already-revoked"
  | "inactive"
  | "forbidden"
  | "server-error";

function signerScenario(): SignerScenario {
  if (typeof document === "undefined") return "default";
  const match = document.cookie
    .split("; ")
    .find((cookie) => cookie.startsWith("custos.test-signer-scenario="));
  return (match
    ? decodeURIComponent(match.slice("custos.test-signer-scenario=".length))
    : "default") as SignerScenario;
}

function paginatedCertificates(): Certificate[] {
  const base = certificates[0];
  if (!base) return [];
  return Array.from({ length: 25 }, (_, index) => ({
    ...base,
    serial_number: 100 + index,
    key_id: `key-${100 + index}`,
    principal: `user-${index + 1}`,
  }));
}

export const signerHandlers = [
  http.get("*/api/v1/signer/admin/certificates", ({ request }) => {
    const url = new URL(request.url);
    const scenario = signerScenario();
    const source = scenario === "empty" ? [] : scenario === "paginated" ? paginatedCertificates() : certificates;
    const limit = Number(url.searchParams.get("limit") ?? source.length);
    const offset = Number(url.searchParams.get("offset") ?? 0);
    const items = source.slice(offset, offset + limit);
    return HttpResponse.json({
      certificates: items,
      total: source.length,
      limit,
      offset,
    });
  }),

  http.get("*/api/v1/signer/admin/certificates/:serial", ({ params }) => {
    const serial = Number(params.serial);
    const found = certificates.find((c) => c.serial_number === serial);
    if (!found) return HttpResponse.json({ error: "certificate not found" }, { status: 404 });
    return HttpResponse.json(found);
  }),

  http.post("*/api/v1/signer/admin/certificates/:serial/revoke", async ({ params, request }) => {
    const scenario = signerScenario();
    if (scenario === "forbidden") {
      return HttpResponse.json(
        { error: "insufficient_privilege", message: "Caller lacks required privilege" },
        { status: 403 },
      );
    }
    if (scenario === "inactive") {
      return HttpResponse.json(
        { error: "certificate_not_active", message: "Certificate is not active" },
        { status: 409 },
      );
    }
    if (scenario === "server-error") {
      return HttpResponse.json(
        { error: "authorization_unavailable", message: "Signer temporarily unavailable" },
        { status: 503 },
      );
    }
    if (!effectivePrivileges().includes("signer:certificates:write")) {
      return HttpResponse.json(
        { error: "insufficient_privilege", message: "Caller lacks required privilege" },
        { status: 403 },
      );
    }
    const serial = Number(params.serial);
    const cert = certificates.find((c) => c.serial_number === serial);
    if (!cert) return HttpResponse.json({ error: "certificate not found" }, { status: 404 });

    const body = (await request.json().catch(() => ({}))) as { reason?: string };
    const reason = (body.reason ?? "").trim();
    if (!reason || reason.length > 255) {
      return HttpResponse.json(
        { error: "invalid_reason", message: "Reason must be between 1 and 255 characters" },
        { status: 400 },
      );
    }

    const now = Math.floor(Date.now() / 1000);
    if (!cert.revoked && (cert.valid_after > now || cert.valid_before <= now)) {
      return HttpResponse.json(
        { error: "certificate_not_active", message: "Certificate is not active" },
        { status: 409 },
      );
    }

    // Idempotent: a repeat revoke returns the existing state without re-mutating.
    const racedRevocation = scenario === "already-revoked";
    if (racedRevocation) {
      cert.revoked = true;
      cert.revoked_at = 1_700_500_000;
      cert.revocation_reason = "Original backend reason";
      cert.revoked_by = "another-admin";
    }
    const alreadyRevoked = cert.revoked;
    if (!alreadyRevoked) {
      cert.revoked = true;
      cert.revoked_at = now;
      cert.revocation_reason = reason;
      cert.revoked_by = "test-admin";
    }

    return HttpResponse.json({
      success: true,
      message: alreadyRevoked
        ? "Certificate was already revoked"
        : "Certificate revoked successfully",
      serial_number: serial,
      revoked: true,
      revoked_at: cert.revoked_at ?? Math.floor(Date.now() / 1000),
      reason: cert.revocation_reason ?? reason,
      already_revoked: alreadyRevoked,
    });
  }),
];
