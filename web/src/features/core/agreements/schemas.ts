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

import {
  zAcceptanceRow,
  zActorView,
  zOnboardingView,
  zPendingRow,
  zUserAgreement,
  zUserAgreementDetail,
  zUserAgreementListItem,
  type zUserAgreementRequest,
  zUsernameCheck,
} from "@/generated/core/zod.gen";
import { z } from "zod";

export type AgreementDraft = {
  [K in keyof z.infer<typeof zUserAgreementRequest>]-?: NonNullable<
    z.infer<typeof zUserAgreementRequest>[K]
  >;
};

const actorSchema = zActorView.required();
export type Actor = z.infer<typeof actorSchema>;

export const agreementListItemSchema = zUserAgreementListItem.required({
  id: true,
  title: true,
  version_label: true,
  status: true,
  accepted_count: true,
  pending_count: true,
});
export type AgreementListItem = z.infer<typeof agreementListItemSchema>;

// The backend sends null for an actor it cannot resolve.
export const agreementDetailSchema = zUserAgreementDetail
  .required({
    id: true,
    title: true,
    version_label: true,
    status: true,
    summary: true,
    body_md: true,
    change_note: true,
    created_at: true,
    accepted_count: true,
    pending_count: true,
  })
  .extend({
    created_by: actorSchema.nullish(),
    published_by: actorSchema.nullish(),
    removed_by: actorSchema.nullish(),
  });
export type AgreementDetail = z.infer<typeof agreementDetailSchema>;

export const acceptanceReportSchema = z.object({
  accepted: z.array(zAcceptanceRow.required().extend({ user: actorSchema })),
  pending: z.array(
    zPendingRow.extend({ user: actorSchema, first_sign_in_at: z.string().nullish() }),
  ),
});
export type AcceptanceReport = z.infer<typeof acceptanceReportSchema>;

export type AcceptanceStatus = "accepted" | "pending";

export const pendingAgreementSchema = zUserAgreement
  .pick({
    id: true,
    title: true,
    version_label: true,
    body_md: true,
    change_note: true,
    published_at: true,
  })
  .partial({ change_note: true });
export type PendingAgreement = z.infer<typeof pendingAgreementSchema>;

export const gateProfileSchema = z.object({
  consent: z.object({ pending: z.array(pendingAgreementSchema).nullish() }).nullish(),
  onboarding: zOnboardingView.nullish(),
});

export const usernameCheckSchema = zUsernameCheck.required({ valid: true, available: true });
export type UsernameCheck = z.infer<typeof usernameCheckSchema>;
