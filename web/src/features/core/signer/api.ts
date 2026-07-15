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

// Calls are namespaced under /signer so the proxy routes them to the signer
// service with the user's Bearer.
import { apiFetch } from "@/shared/api/client";
import {
  type Certificate,
  type CertificateList,
  type RevokeResponse,
  certificateListSchema,
  certificateSchema,
  revokeResponseSchema,
} from "./schemas";
import type { CertificateListParams } from "./types";

export async function listCertificates(
  params: CertificateListParams = {},
): Promise<CertificateList> {
  const search = new URLSearchParams();
  if (typeof params.limit === "number") search.set("limit", String(params.limit));
  if (typeof params.offset === "number") search.set("offset", String(params.offset));
  const qs = search.toString();
  const raw = await apiFetch(`/signer/certificates${qs ? `?${qs}` : ""}`);
  return certificateListSchema.parse(raw);
}

export async function getCertificate(serial: number | string): Promise<Certificate> {
  const raw = await apiFetch(`/signer/certificates/${serial}`);
  return certificateSchema.parse(raw);
}

// serial comes from the certificate; the backend keys revocation off the real
// serial_number (never a key id or timestamp).
export async function revokeCertificate(
  serial: number,
  reason: string,
): Promise<RevokeResponse> {
  const raw = await apiFetch(`/signer/certificates/${serial}/revoke`, {
    method: "POST",
    body: { reason },
  });
  return revokeResponseSchema.parse(raw);
}
