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

export const accessCheckTypeSchema = z.enum(["SIGN_IN", "JOB_SUBMISSION"]);
export type AccessCheckType = z.infer<typeof accessCheckTypeSchema>;

export const accessUIStateSchema = z.enum(["setting_up", "ok", "retrying", "stuck"]);
export type AccessUIState = z.infer<typeof accessUIStateSchema>;

export const accessCheckSchema = z.object({
  type: accessCheckTypeSchema,
  status: z.enum(["PENDING", "OK", "FAILING"]),
  ui_state: accessUIStateSchema,
  last_checked_at: z.string(),
  last_ok_at: z.string().nullish(),
  failing_since: z.string().nullish(),
});
export type AccessCheck = z.infer<typeof accessCheckSchema>;

export const accessCheckEventSchema = z.object({
  check_type: accessCheckTypeSchema.or(z.literal("")),
  event_type: z.enum(["STARTED", "ONLINE", "FAILED", "RECOVERED", "STUCK"]),
  occurred_at: z.string(),
});
export type AccessCheckEvent = z.infer<typeof accessCheckEventSchema>;

export const accessStatusSchema = z.object({
  checks: z.array(accessCheckSchema),
  events: z.array(accessCheckEventSchema),
  local_username: z.string().nullish(),
});
export type AccessStatus = z.infer<typeof accessStatusSchema>;

export const memberAccessStatusSchema = z.object({
  user_id: z.string(),
  display_name: z.string(),
  email: z.string(),
  local_username: z.string().nullish(),
  checks: z.array(accessCheckSchema),
});
export type MemberAccessStatus = z.infer<typeof memberAccessStatusSchema>;

export const allocationAccessMembersSchema = z.object({
  members: z.array(memberAccessStatusSchema),
});
export type AllocationAccessMembers = z.infer<typeof allocationAccessMembersSchema>;
