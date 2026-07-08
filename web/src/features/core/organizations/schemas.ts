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
import { zOrganization } from "@/generated/core/zod.gen";

// The generated schema marks every field optional; the backend always returns
// all three, so tighten them here where the UI reads them.
export const organizationSchema = zOrganization.required({
  id: true,
  name: true,
  originated_id: true,
});
export type Organization = z.infer<typeof organizationSchema>;

// Portal-shaped paginated envelope. Mirrors the projects list contract.
export const organizationListEnvelopeSchema = z.object({
  items: z.array(organizationSchema),
  total: z.number().int().nonnegative(),
});
export type OrganizationListEnvelope = z.infer<typeof organizationListEnvelopeSchema>;

export const createOrganizationPayloadSchema = z.object({
  name: z.string().min(1).max(200),
  originated_id: z.string().optional(),
});
export type CreateOrganizationPayload = z.infer<typeof createOrganizationPayloadSchema>;
