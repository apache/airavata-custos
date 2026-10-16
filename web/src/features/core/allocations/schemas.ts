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

import type { ComputeAllocationListResponse } from "@/generated/core/zod.gen";
import {
  type ComputeAllocation,
  zAllocationMembershipResponse,
  zAllocationStatus,
  zComputeAllocationChangeRequest,
  zComputeAllocationChangeRequestInput,
  zComputeAllocationMembershipInput,
} from "@/generated/core/zod.gen";
import { z } from "zod";

export const changeRequestStatusSchema = z.enum(["PENDING", "APPROVED", "REJECTED"]);
export type ChangeRequestStatus = z.infer<typeof changeRequestStatusSchema>;

export type ComputeAllocationList = ComputeAllocationListResponse & {
  items: ComputeAllocation[];
};

export const allocationMembershipRoleSchema = z.enum([
  "PI",
  "CO_PI",
  "ALLOCATION_MANAGER",
  "MEMBER",
]);
export type AllocationMembershipRole = z.infer<typeof allocationMembershipRoleSchema>;

// The Go role is a string; the portal narrows it to the project roles.
export const allocationMembershipSchema = zAllocationMembershipResponse.extend({
  role: allocationMembershipRoleSchema,
});
export type AllocationMembership = z.infer<typeof allocationMembershipSchema>;

// ChangeStatus is a free-form string on the Go side; we narrow to the documented enum.
export const changeRequestSchema = zComputeAllocationChangeRequest.extend({
  change_status: changeRequestStatusSchema,
});
export type ChangeRequest = z.infer<typeof changeRequestSchema>;

// role is not on the Go membership; the backend ignores it.
export const createMembershipPayloadSchema = zComputeAllocationMembershipInput.extend({
  compute_allocation_id: z.string().min(1),
  user_id: z.string().min(1),
  start_time: z.string().min(1),
  end_time: z.string().min(1),
  membership_status: zAllocationStatus.default("ACTIVE"),
  role: allocationMembershipRoleSchema.optional(),
});
export type CreateMembershipPayload = z.input<typeof createMembershipPayloadSchema>;

export const updateMembershipPayloadSchema = zComputeAllocationMembershipInput
  .pick({ start_time: true, end_time: true, membership_status: true })
  .extend({ role: allocationMembershipRoleSchema.optional() })
  .refine((v) => Object.keys(v).length > 0, { message: "patch is empty" });
export type UpdateMembershipPayload = z.infer<typeof updateMembershipPayloadSchema>;

export const createChangeRequestPayloadSchema = zComputeAllocationChangeRequestInput.extend({
  compute_allocation_id: z.string().min(1),
  requested_su_amount: z.number().nonnegative(),
  requested_status: zAllocationStatus,
  reason: z.string().min(1),
  requester_id: z.string().min(1),
});
export type CreateChangeRequestPayload = z.infer<typeof createChangeRequestPayloadSchema>;

export const updateChangeRequestPayloadSchema = zComputeAllocationChangeRequestInput
  .pick({ approver_id: true, reason: true, requested_status: true })
  .extend({
    change_status: changeRequestStatusSchema.optional(),
    requested_su_amount: z.number().nonnegative().optional(),
  })
  .refine((v) => Object.keys(v).length > 0, { message: "patch is empty" });
export type UpdateChangeRequestPayload = z.infer<typeof updateChangeRequestPayloadSchema>;
