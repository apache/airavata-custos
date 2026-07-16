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

export const accessRequestStatusSchema = z.enum(["PENDING", "APPROVED", "DENIED"]);
export type AccessRequestStatus = z.infer<typeof accessRequestStatusSchema>;

export const accessRequestSchema = z.object({
  id: z.string(),
  oidc_sub: z.string(),
  email: z.string(),
  name: z.string(),
  institution: z.string(),
  event_code: z.string(),
  reason: z
    .string()
    .nullish()
    .transform((v) => v ?? ""),
  status: accessRequestStatusSchema,
  approver_id: z
    .string()
    .nullish()
    .transform((v) => v ?? ""),
  deny_reason: z
    .string()
    .nullish()
    .transform((v) => v ?? ""),
  expires_at: z.string().nullish(),
  created_user_id: z
    .string()
    .nullish()
    .transform((v) => v ?? ""),
  timestamp: z.string(),
});
export type AccessRequest = z.infer<typeof accessRequestSchema>;

export const accessRequestsSchema = z.array(accessRequestSchema);

// The resolve endpoint deliberately returns only the code and the
// allocation's display name — the caller is not a user yet.
export const accessEventResolveSchema = z.object({
  code: z.string(),
  name: z.string(),
});
export type AccessEventResolve = z.infer<typeof accessEventResolveSchema>;
