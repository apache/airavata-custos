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
import { type SetupStep, zSetupStatusResponse } from "@/generated/core/zod.gen";
import { apiFetch } from "@/shared/api/client";
import {
  type AcceptanceReport,
  type AcceptanceStatus,
  type AgreementDetail,
  type AgreementDraft,
  type AgreementListItem,
  type UsernameCheck,
  acceptanceReportSchema,
  agreementDetailSchema,
  agreementListItemSchema,
  gateProfileSchema,
  usernameCheckSchema,
} from "./schemas";

export type AcceptanceParams = { status: AcceptanceStatus; limit: number; offset: number };

export async function listAgreements(): Promise<AgreementListItem[]> {
  return z.array(agreementListItemSchema).parse(await apiFetch("/agreements"));
}

export async function getAgreement(id: string): Promise<AgreementDetail> {
  return agreementDetailSchema.parse(await apiFetch(`/agreements/${encodeURIComponent(id)}`));
}

export async function listAcceptances(
  id: string,
  { status, limit, offset }: AcceptanceParams,
): Promise<AcceptanceReport> {
  const qs = new URLSearchParams({ status, limit: String(limit), offset: String(offset) });
  const raw = await apiFetch(`/agreements/${encodeURIComponent(id)}/acceptances?${qs}`);
  return acceptanceReportSchema.parse(raw);
}

export async function createAgreement(body: AgreementDraft): Promise<AgreementDetail> {
  return agreementDetailSchema.parse(await apiFetch("/agreements", { method: "POST", body }));
}

export async function updateAgreement({
  id,
  ...body
}: AgreementDraft & { id: string }): Promise<AgreementDetail> {
  const raw = await apiFetch(`/agreements/${encodeURIComponent(id)}`, { method: "PUT", body });
  return agreementDetailSchema.parse(raw);
}

export async function publishAgreement(id: string): Promise<AgreementDetail> {
  const raw = await apiFetch(`/agreements/${encodeURIComponent(id)}/publish`, { method: "POST" });
  return agreementDetailSchema.parse(raw);
}

export async function removeAgreement(id: string): Promise<AgreementDetail> {
  const raw = await apiFetch(`/agreements/${encodeURIComponent(id)}/remove`, { method: "POST" });
  return agreementDetailSchema.parse(raw);
}

export async function deleteAgreement(id: string): Promise<void> {
  await apiFetch(`/agreements/${encodeURIComponent(id)}`, { method: "DELETE" });
}

export async function getGateProfile() {
  const { consent, onboarding } = gateProfileSchema.parse(await apiFetch("/me"));
  return { pending: consent?.pending?.[0] ?? null, onboarding: onboarding ?? null };
}

export async function checkUsername(username: string): Promise<UsernameCheck> {
  const raw = await apiFetch(`/me/onboarding/username?u=${encodeURIComponent(username)}`);
  return usernameCheckSchema.parse(raw);
}

export async function getSetupSteps(): Promise<SetupStep[]> {
  return zSetupStatusResponse.parse(await apiFetch("/me/onboarding/status")).steps ?? [];
}

export async function completeOnboarding(body: { username: string; accepted: string[] }) {
  await apiFetch("/me/onboarding", { method: "POST", body });
}

export async function acceptAgreement(id: string) {
  await apiFetch(`/me/agreements/${encodeURIComponent(id)}/accept`, { method: "POST" });
}
