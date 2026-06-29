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
import { zAllocationStatus } from "@/generated/core/zod.gen";

export const allocationStatusSchema = zAllocationStatus;
export type AllocationStatus = z.infer<typeof allocationStatusSchema>;

export const changeRequestStatusSchema = z.enum(["PENDING", "APPROVED", "REJECTED"]);
export type ChangeRequestStatus = z.infer<typeof changeRequestStatusSchema>;

// Mirrors pkg/models/allocation.go#ComputeCluster. status is intentionally
// omitted until the Go model grows the column.
export const computeClusterSchema = z.object({
  id: z.string(),
  name: z.string(),
});
export type ComputeCluster = z.infer<typeof computeClusterSchema>;

// Mirrors pkg/models/allocation.go#ComputeAllocation 1:1.
export const computeAllocationSchema = z.object({
  id: z.string(),
  project_id: z.string(),
  name: z.string(),
  status: allocationStatusSchema,
  compute_cluster_id: z.string(),
  initial_su_amount: z.number().int(),
  start_time: z.string(),
  end_time: z.string(),
});
export type ComputeAllocation = z.infer<typeof computeAllocationSchema>;

// Portal-side envelope. The backend has no list endpoint today
// (only POST /compute-allocations + GET /compute-allocations/{id}); MSW serves
// the list and the real route is expected to mirror this shape.
export const computeAllocationListSchema = z.object({
  items: z.array(computeAllocationSchema),
  total: z.number().int().nonnegative(),
});
export type ComputeAllocationList = z.infer<typeof computeAllocationListSchema>;

export const computeAllocationResourceSchema = z.object({
  id: z.string(),
  name: z.string(),
  resource_type: z.string(),
  resource_amount: z.number().int(),
});
export type ComputeAllocationResource = z.infer<typeof computeAllocationResourceSchema>;

export const computeAllocationResourceRateSchema = z.object({
  id: z.string(),
  compute_allocation_resource_id: z.string(),
  rate: z.number(),
  start_time: z.string(),
  end_time: z.string(),
});
export type ComputeAllocationResourceRate = z.infer<typeof computeAllocationResourceRateSchema>;

export const allocationMembershipRoleSchema = z.enum([
  "PI",
  "CO_PI",
  "ALLOCATION_MANAGER",
  "MEMBER",
]);
export type AllocationMembershipRole = z.infer<typeof allocationMembershipRoleSchema>;

// Mirrors pkg/models/allocation.go#ComputeAllocationMembership.
export const allocationMembershipSchema = z.object({
  id: z.string(),
  compute_allocation_id: z.string(),
  user_id: z.string(),
  start_time: z.string(),
  end_time: z.string(),
  membership_status: allocationStatusSchema,
  role: allocationMembershipRoleSchema.optional(),
  display_name: z.string().optional(),
  email: z.string().optional(),
});
export type AllocationMembership = z.infer<typeof allocationMembershipSchema>;

export const allocationMembershipListSchema = z.array(allocationMembershipSchema);

// Mirrors pkg/models/allocation.go#ComputeAllocationChangeRequest. ChangeStatus
// is a free-form string on the Go side; we narrow to the documented enum.
export const changeRequestSchema = z.object({
  id: z.string(),
  compute_allocation_id: z.string(),
  requested_su_amount: z.number().int().nonnegative(),
  requested_status: allocationStatusSchema,
  reason: z.string(),
  change_status: changeRequestStatusSchema,
  requester_id: z.string(),
  approver_id: z.string().optional(),
  timestamp: z.string(),
});
export type ChangeRequest = z.infer<typeof changeRequestSchema>;

export const changeRequestListSchema = z.array(changeRequestSchema);

// Mirrors pkg/models/allocation.go#ComputeAllocationChangeRequestEvent.
export const changeRequestEventSchema = z.object({
  id: z.string(),
  compute_allocation_change_request_id: z.string(),
  event_type: z.string(),
  description: z.string().optional(),
  timestamp: z.string(),
});
export type ChangeRequestEvent = z.infer<typeof changeRequestEventSchema>;

export const changeRequestEventListSchema = z.array(changeRequestEventSchema);

export const allocationUsageSchema = z.object({
  id: z.string(),
  compute_allocation_id: z.string(),
  used_raw_amount: z.number().int(),
  used_su_amount: z.number().int(),
  last_updated: z.string(),
  user_id: z.string(),
  job_id: z.string(),
  compute_allocation_resource_id: z.string(),
});
export type AllocationUsage = z.infer<typeof allocationUsageSchema>;

export const allocationUsageListSchema = z.array(allocationUsageSchema);

export const createMembershipPayloadSchema = z.object({
  compute_allocation_id: z.string().min(1),
  user_id: z.string().min(1),
  start_time: z.string().min(1),
  end_time: z.string().min(1),
  membership_status: allocationStatusSchema.default("ACTIVE"),
  role: allocationMembershipRoleSchema.optional(),
});
export type CreateMembershipPayload = z.input<typeof createMembershipPayloadSchema>;

export const updateMembershipPayloadSchema = z
  .object({
    start_time: z.string().optional(),
    end_time: z.string().optional(),
    membership_status: allocationStatusSchema.optional(),
    role: allocationMembershipRoleSchema.optional(),
  })
  .refine((v) => Object.keys(v).length > 0, { message: "patch is empty" });
export type UpdateMembershipPayload = z.infer<typeof updateMembershipPayloadSchema>;

export const createChangeRequestPayloadSchema = z.object({
  compute_allocation_id: z.string().min(1),
  requested_su_amount: z.number().int().nonnegative(),
  requested_status: allocationStatusSchema,
  reason: z.string().min(1),
  requester_id: z.string().min(1),
});
export type CreateChangeRequestPayload = z.infer<typeof createChangeRequestPayloadSchema>;

export const updateChangeRequestPayloadSchema = z
  .object({
    change_status: changeRequestStatusSchema.optional(),
    approver_id: z.string().optional(),
    reason: z.string().optional(),
    requested_su_amount: z.number().int().nonnegative().optional(),
    requested_status: allocationStatusSchema.optional(),
  })
  .refine((v) => Object.keys(v).length > 0, { message: "patch is empty" });
export type UpdateChangeRequestPayload = z.infer<typeof updateChangeRequestPayloadSchema>;
