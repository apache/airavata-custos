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

const certificates: Certificate[] = (certificatesFixture as Certificate[]).map((c) => ({ ...c }));

export const signerHandlers = [
  http.get("*/api/v1/signer/certificates", ({ request }) => {
    const url = new URL(request.url);
    const limit = Number(url.searchParams.get("limit") ?? certificates.length);
    const offset = Number(url.searchParams.get("offset") ?? 0);
    const items = certificates.slice(offset, offset + limit);
    return HttpResponse.json({
      certificates: items,
      total: certificates.length,
      limit,
      offset,
    });
  }),

  http.get("*/api/v1/signer/certificates/:serial", ({ params }) => {
    const serial = Number(params.serial);
    const found = certificates.find((c) => c.serial_number === serial);
    if (!found) return HttpResponse.json({ error: "certificate not found" }, { status: 404 });
    return HttpResponse.json(found);
  }),

  http.post("*/api/v1/signer/certificates/:serial/revoke", async ({ params, request }) => {
    const serial = Number(params.serial);
    const cert = certificates.find((c) => c.serial_number === serial);
    if (!cert) return HttpResponse.json({ error: "certificate not found" }, { status: 404 });

    const body = (await request.json().catch(() => ({}))) as { reason?: string };
    const reason = (body.reason ?? "").trim();

    // Idempotent: a repeat revoke returns the existing state without re-mutating.
    const alreadyRevoked = cert.revoked;
    if (!alreadyRevoked) {
      cert.revoked = true;
      cert.revoked_at = Math.floor(Date.now() / 1000);
      cert.revocation_reason = reason;
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
