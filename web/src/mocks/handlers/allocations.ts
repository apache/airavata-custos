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

import allocationsFixture from "@/features/core/allocations/__fixtures__/allocations.json";
import changeRequestsFixture from "@/features/core/allocations/__fixtures__/change-requests.json";
import diffsFixture from "@/features/core/allocations/__fixtures__/diffs.json";
import eventsFixture from "@/features/core/allocations/__fixtures__/events.json";
import membersFixture from "@/features/core/allocations/__fixtures__/members.json";
import resourcesFixture from "@/features/core/allocations/__fixtures__/resources.json";
import usageFixture from "@/features/core/allocations/__fixtures__/usage.json";
import {
  type AllocationMembership,
  type ChangeRequest,
  allocationMembershipSchema,
  changeRequestSchema,
  createChangeRequestPayloadSchema,
  createMembershipPayloadSchema,
  updateChangeRequestPayloadSchema,
} from "@/features/core/allocations/schemas";
import {
  type ComputeAllocation,
  type ComputeAllocationChangeRequestEvent,
  zComputeAllocation,
  zComputeAllocationChangeRequestEvent,
  zComputeAllocationDiff,
  zComputeAllocationResource,
  zComputeAllocationUsage,
} from "@/generated/core/zod.gen";
import { http, HttpResponse } from "msw";
import { z } from "zod";

const byId = <T extends z.ZodType>(row: T) => z.record(z.string(), z.array(row));

const allocations = z.array(zComputeAllocation).parse(allocationsFixture);
const resourcesByAlloc = byId(zComputeAllocationResource).parse(resourcesFixture);
const usageByAlloc = byId(zComputeAllocationUsage).parse(usageFixture);
const diffsByAlloc = byId(zComputeAllocationDiff).parse(diffsFixture);
const membersByAlloc = byId(allocationMembershipSchema).parse(membersFixture);
const changeRequests = z.array(changeRequestSchema).parse(changeRequestsFixture);
const eventsByCr = byId(zComputeAllocationChangeRequestEvent).parse(eventsFixture);

let memberSeq = 1000;
let crSeq = 1000;
let eventSeq = 1000;

function ensureMembersBucket(allocId: string): AllocationMembership[] {
  if (!membersByAlloc[allocId]) membersByAlloc[allocId] = [];
  return membersByAlloc[allocId];
}

function ensureEventsBucket(crId: string): ComputeAllocationChangeRequestEvent[] {
  if (!eventsByCr[crId]) eventsByCr[crId] = [];
  return eventsByCr[crId];
}

function filterAllocations(url: URL): ComputeAllocation[] {
  const projectId = url.searchParams.get("project_id");
  const status = url.searchParams.get("status");
  const q = url.searchParams.get("q")?.toLowerCase() ?? "";
  return allocations.filter((a) => {
    if (projectId && a.project_id !== projectId) return false;
    if (status && a.status !== status) return false;
    if (q && !a.name.toLowerCase().includes(q) && !a.id.toLowerCase().includes(q)) return false;
    return true;
  });
}

export const allocationsHandlers = [
  http.get("*/api/v1/compute-allocations", ({ request }) => {
    const url = new URL(request.url);
    const matched = filterAllocations(url);
    const limit = Number(url.searchParams.get("limit") ?? matched.length);
    const offset = Number(url.searchParams.get("offset") ?? 0);
    const items = matched.slice(offset, offset + limit);
    return HttpResponse.json({ items, total: matched.length });
  }),

  http.get<{ id: string }>("*/api/v1/compute-allocations/:id", ({ params }) => {
    const id = String(params.id);
    const found = allocations.find((a) => a.id === id);
    if (!found) return HttpResponse.json({ error: "allocation not found" }, { status: 404 });
    return HttpResponse.json(found);
  }),

  http.get<{ id: string }>("*/api/v1/compute-allocations/:id/resources", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(resourcesByAlloc[id] ?? []);
  }),

  http.get<{ id: string }>("*/api/v1/compute-allocations/:id/usages", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(usageByAlloc[id] ?? []);
  }),

  http.get<{ id: string }>("*/api/v1/compute-allocations/:id/diffs", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(diffsByAlloc[id] ?? []);
  }),

  http.get<{ id: string }>("*/api/v1/compute-allocations/:id/diffs/latest", ({ params }) => {
    const id = String(params.id);
    const rows = [...(diffsByAlloc[id] ?? [])].sort(
      (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime(),
    );
    if (rows.length === 0) return HttpResponse.json({ error: "no diffs" }, { status: 404 });
    return HttpResponse.json(rows[0]);
  }),

  http.get<{ id: string }>("*/api/v1/compute-allocations/:id/memberships", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(membersByAlloc[id] ?? []);
  }),

  http.post("*/api/v1/compute-allocation-memberships", async ({ request }) => {
    const body = createMembershipPayloadSchema.parse(await request.json());
    const bucket = ensureMembersBucket(body.compute_allocation_id);
    const member: AllocationMembership = {
      id: `mem-${memberSeq++}`,
      compute_allocation_id: body.compute_allocation_id,
      user_id: body.user_id,
      start_time: body.start_time,
      end_time: body.end_time,
      membership_status: body.membership_status,
      role: "MEMBER",
      display_name: body.user_id,
      email: `${body.user_id}@custos.local`,
    };
    bucket.push(member);
    return HttpResponse.json(member, { status: 201 });
  }),

  http.post<{ id: string }>(
    "*/api/v1/compute-allocation-memberships/:id/deactivate",
    ({ params }) => {
      const member = Object.values(membersByAlloc)
        .flat()
        .find((m) => m.id === String(params.id));
      if (!member) return HttpResponse.json({ error: "membership not found" }, { status: 404 });
      member.membership_status = "INACTIVE";
      return new HttpResponse(null, { status: 204 });
    },
  ),

  http.delete<{ id: string }>("*/api/v1/compute-allocation-memberships/:id", ({ params }) => {
    const id = String(params.id);
    for (const bucket of Object.values(membersByAlloc)) {
      const idx = bucket.findIndex((m) => m.id === id);
      if (idx !== -1) {
        bucket.splice(idx, 1);
        return new HttpResponse(null, { status: 204 });
      }
    }
    return HttpResponse.json({ error: "membership not found" }, { status: 404 });
  }),

  http.get<{ id: string }>("*/api/v1/compute-allocations/:id/change-requests", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(changeRequests.filter((cr) => cr.compute_allocation_id === id));
  }),

  http.get<{ id: string }>("*/api/v1/users/:id/change-requests", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(changeRequests.filter((cr) => cr.requester_id === id));
  }),

  http.get("*/api/v1/compute-allocation-change-requests", ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get("status");
    const filtered = status
      ? changeRequests.filter((cr) => cr.change_status === status)
      : changeRequests;
    return HttpResponse.json(filtered);
  }),

  http.get<{ id: string }>("*/api/v1/compute-allocation-change-requests/:id", ({ params }) => {
    const id = String(params.id);
    const found = changeRequests.find((cr) => cr.id === id);
    if (!found) return HttpResponse.json({ error: "change request not found" }, { status: 404 });
    return HttpResponse.json(found);
  }),

  http.post("*/api/v1/compute-allocation-change-requests", async ({ request }) => {
    const body = createChangeRequestPayloadSchema.parse(await request.json());
    const cr: ChangeRequest = {
      id: `cr-${crSeq++}`,
      compute_allocation_id: body.compute_allocation_id,
      requested_su_amount: body.requested_su_amount,
      requested_status: body.requested_status,
      reason: body.reason,
      change_status: "PENDING",
      requester_id: body.requester_id,
      timestamp: new Date().toISOString(),
    };
    changeRequests.unshift(cr);
    ensureEventsBucket(cr.id).push({
      id: `evt-${eventSeq++}`,
      compute_allocation_change_request_id: cr.id,
      event_type: "CREATED",
      description: `Change request created by ${cr.requester_id}`,
      timestamp: cr.timestamp,
    });
    return HttpResponse.json(cr, { status: 201 });
  }),

  http.put<{ id: string }>(
    "*/api/v1/compute-allocation-change-requests/:id",
    async ({ params, request }) => {
      const id = String(params.id);
      const patch = updateChangeRequestPayloadSchema.parse(await request.json());
      const existing = changeRequests.find((cr) => cr.id === id);
      if (!existing)
        return HttpResponse.json({ error: "change request not found" }, { status: 404 });
      if (patch.change_status) existing.change_status = patch.change_status;
      if (patch.approver_id) existing.approver_id = patch.approver_id;
      if (patch.reason) existing.reason = patch.reason;
      if (typeof patch.requested_su_amount === "number")
        existing.requested_su_amount = patch.requested_su_amount;
      if (patch.requested_status) existing.requested_status = patch.requested_status;
      if (patch.change_status === "APPROVED" || patch.change_status === "REJECTED") {
        ensureEventsBucket(existing.id).push({
          id: `evt-${eventSeq++}`,
          compute_allocation_change_request_id: existing.id,
          event_type: patch.change_status,
          description: `${patch.change_status === "APPROVED" ? "Approved" : "Rejected"} by ${patch.approver_id ?? "unknown"}`,
          timestamp: new Date().toISOString(),
        });
      }
      return HttpResponse.json(existing);
    },
  ),

  http.delete<{ id: string }>("*/api/v1/compute-allocation-change-requests/:id", ({ params }) => {
    const id = String(params.id);
    const idx = changeRequests.findIndex((cr) => cr.id === id);
    if (idx === -1)
      return HttpResponse.json({ error: "change request not found" }, { status: 404 });
    changeRequests.splice(idx, 1);
    return new HttpResponse(null, { status: 204 });
  }),

  http.get<{ id: string }>(
    "*/api/v1/compute-allocation-change-requests/:id/events",
    ({ params }) => {
      const id = String(params.id);
      return HttpResponse.json(eventsByCr[id] ?? []);
    },
  ),
];
