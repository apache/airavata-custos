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

import { z } from "zod";
import { zAppNameView } from "@/generated/core/zod.gen";
import { apiFetch } from "@/shared/api/client";
import {
  type Certificate,
  type CertificateList,
  type GeneratedCertificate,
  type RevokeResponse,
  certificateListSchema,
  certificateSchema,
  generatedCertificateSchema,
  revokeResponseSchema,
} from "./schemas";
import type { CertificateListParams } from "./types";

export async function listAppNames(): Promise<Record<string, string>> {
  const apps = z.array(zAppNameView).parse(await apiFetch("/apps/names"));
  return Object.fromEntries(apps.map((app) => [app.key, app.name]));
}

export async function getUserEmail(id: string): Promise<string> {
  const user = z.object({ email: z.string() }).parse(await apiFetch(`/users/${encodeURIComponent(id)}`));
  return user.email;
}

async function fetchCertificateList(
  path: string,
  params: CertificateListParams,
): Promise<CertificateList> {
  const search = new URLSearchParams();
  if (typeof params.limit === "number") search.set("limit", String(params.limit));
  if (params.cursor) search.set("cursor", params.cursor);
  const qs = search.toString();
  const raw = await apiFetch(`${path}${qs ? `?${qs}` : ""}`);
  return certificateListSchema.parse(raw);
}

export function listCertificates(params: CertificateListParams = {}): Promise<CertificateList> {
  return fetchCertificateList("/connectors/signer/certificates", params);
}

export function listMyCertificates(params: CertificateListParams = {}): Promise<CertificateList> {
  return fetchCertificateList("/connectors/signer/me/certificates", params);
}

export async function getMyCertificate(serial: number | string): Promise<Certificate> {
  const raw = await apiFetch(`/connectors/signer/me/certificates/${serial}`);
  return certificateSchema.parse(raw);
}

export async function generateCertificate(publicKey: string): Promise<GeneratedCertificate> {
  const raw = await apiFetch("/connectors/signer/me/certificates", {
    method: "POST",
    body: { public_key: publicKey },
  });
  return generatedCertificateSchema.parse(raw);
}

export async function revokeMyCertificate(serial: number, reason: string): Promise<RevokeResponse> {
  const raw = await apiFetch(`/connectors/signer/me/certificates/${serial}/revoke`, {
    method: "POST",
    body: { reason },
  });
  return revokeResponseSchema.parse(raw);
}

export async function getCertificate(serial: number | string): Promise<Certificate> {
  const raw = await apiFetch(`/connectors/signer/certificates/${serial}`);
  return certificateSchema.parse(raw);
}

// serial comes from the certificate; the backend keys revocation off the real
// serial_number (never a key id or timestamp).
export async function revokeCertificate(serial: number, reason: string): Promise<RevokeResponse> {
  const raw = await apiFetch(`/connectors/signer/certificates/${serial}/revoke`, {
    method: "POST",
    body: { reason },
  });
  return revokeResponseSchema.parse(raw);
}
