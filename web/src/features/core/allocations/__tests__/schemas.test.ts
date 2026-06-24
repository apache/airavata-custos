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

import { describe, expect, it } from "vitest";
import {
  allocationMembershipSchema,
  changeRequestEventSchema,
  changeRequestSchema,
  computeAllocationListSchema,
  computeAllocationSchema,
  createChangeRequestPayloadSchema,
  createMembershipPayloadSchema,
  updateChangeRequestPayloadSchema,
  updateMembershipPayloadSchema,
} from "../schemas";

const validAllocation = {
  id: "alloc-001",
  project_id: "project-001",
  name: "Test Alloc",
  status: "ACTIVE" as const,
  compute_cluster_id: "cluster-001",
  initial_su_amount: 100000,
  start_time: "2026-04-01T00:00:00.000Z",
  end_time: "2027-03-31T00:00:00.000Z",
};

const validMember = {
  id: "mem-001",
  compute_allocation_id: "alloc-001",
  user_id: "user-002",
  start_time: "2026-04-01T00:00:00.000Z",
  end_time: "2027-03-31T00:00:00.000Z",
  membership_status: "ACTIVE" as const,
};

const validChangeRequest = {
  id: "cr-001",
  compute_allocation_id: "alloc-001",
  requested_su_amount: 150000,
  requested_status: "ACTIVE" as const,
  reason: "Need more SUs",
  change_status: "PENDING" as const,
  requester_id: "user-pi-001",
  timestamp: "2026-06-01T10:00:00.000Z",
};

describe("computeAllocationSchema", () => {
  it("accepts a backend-shaped allocation", () => {
    expect(computeAllocationSchema.safeParse(validAllocation).success).toBe(true);
  });
  it("rejects an invalid status", () => {
    expect(
      computeAllocationSchema.safeParse({ ...validAllocation, status: "FOO" }).success,
    ).toBe(false);
  });
});

describe("computeAllocationListSchema", () => {
  it("accepts a valid envelope", () => {
    expect(
      computeAllocationListSchema.safeParse({ items: [validAllocation], total: 1 }).success,
    ).toBe(true);
  });
  it("rejects a negative total", () => {
    expect(computeAllocationListSchema.safeParse({ items: [], total: -1 }).success).toBe(false);
  });
});

describe("allocationMembershipSchema", () => {
  it("accepts a backend-shaped membership", () => {
    expect(allocationMembershipSchema.safeParse(validMember).success).toBe(true);
  });
  it("accepts role + display fields", () => {
    expect(
      allocationMembershipSchema.safeParse({
        ...validMember,
        role: "PI",
        display_name: "Test",
        email: "test@x",
      }).success,
    ).toBe(true);
  });
  it("rejects an unknown role", () => {
    expect(
      allocationMembershipSchema.safeParse({ ...validMember, role: "queen" }).success,
    ).toBe(false);
  });
});

describe("changeRequestSchema", () => {
  it("accepts a pending request without approver_id", () => {
    expect(changeRequestSchema.safeParse(validChangeRequest).success).toBe(true);
  });
  it("accepts an approved request with approver_id", () => {
    expect(
      changeRequestSchema.safeParse({
        ...validChangeRequest,
        change_status: "APPROVED",
        approver_id: "user-admin-001",
      }).success,
    ).toBe(true);
  });
  it("rejects an invalid change_status", () => {
    expect(
      changeRequestSchema.safeParse({ ...validChangeRequest, change_status: "MAYBE" }).success,
    ).toBe(false);
  });
});

describe("changeRequestEventSchema", () => {
  it("accepts a valid event", () => {
    expect(
      changeRequestEventSchema.safeParse({
        id: "evt-001",
        compute_allocation_change_request_id: "cr-001",
        event_type: "CREATED",
        timestamp: "2026-06-01T10:00:00.000Z",
      }).success,
    ).toBe(true);
  });
});

describe("createMembershipPayloadSchema", () => {
  it("accepts a complete payload", () => {
    expect(
      createMembershipPayloadSchema.safeParse({
        compute_allocation_id: "alloc-001",
        user_id: "user-002",
        start_time: "2026-04-01T00:00:00Z",
        end_time: "2027-03-31T00:00:00Z",
      }).success,
    ).toBe(true);
  });
  it("defaults membership_status to ACTIVE", () => {
    const parsed = createMembershipPayloadSchema.parse({
      compute_allocation_id: "alloc-001",
      user_id: "user-002",
      start_time: "x",
      end_time: "x",
    });
    expect(parsed.membership_status).toBe("ACTIVE");
  });
  it("rejects empty user_id", () => {
    expect(
      createMembershipPayloadSchema.safeParse({
        compute_allocation_id: "alloc-001",
        user_id: "",
        start_time: "x",
        end_time: "x",
      }).success,
    ).toBe(false);
  });
});

describe("updateMembershipPayloadSchema", () => {
  it("rejects empty patches", () => {
    expect(updateMembershipPayloadSchema.safeParse({}).success).toBe(false);
  });
  it("accepts a status-only patch", () => {
    expect(
      updateMembershipPayloadSchema.safeParse({ membership_status: "INACTIVE" }).success,
    ).toBe(true);
  });
});

describe("createChangeRequestPayloadSchema", () => {
  it("accepts a complete payload", () => {
    expect(
      createChangeRequestPayloadSchema.safeParse({
        compute_allocation_id: "alloc-001",
        requested_su_amount: 150000,
        requested_status: "ACTIVE",
        reason: "more sus",
        requester_id: "user-pi-001",
      }).success,
    ).toBe(true);
  });
  it("rejects negative SU amount", () => {
    expect(
      createChangeRequestPayloadSchema.safeParse({
        compute_allocation_id: "alloc-001",
        requested_su_amount: -1,
        requested_status: "ACTIVE",
        reason: "x",
        requester_id: "y",
      }).success,
    ).toBe(false);
  });
});

describe("updateChangeRequestPayloadSchema", () => {
  it("rejects empty patches", () => {
    expect(updateChangeRequestPayloadSchema.safeParse({}).success).toBe(false);
  });
  it("accepts a single-field patch", () => {
    expect(
      updateChangeRequestPayloadSchema.safeParse({ change_status: "APPROVED" }).success,
    ).toBe(true);
  });
});
