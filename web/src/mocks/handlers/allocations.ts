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

import { http, HttpResponse } from "msw";
import allocationsFixture from "@/features/core/allocations/__fixtures__/allocations.json";
import resourcesFixture from "@/features/core/allocations/__fixtures__/resources.json";
import diffsFixture from "@/features/core/allocations/__fixtures__/diffs.json";
import usageFixture from "@/features/core/allocations/__fixtures__/usage.json";
import membersFixture from "@/features/core/allocations/__fixtures__/members.json";
import changeRequestsFixture from "@/features/core/allocations/__fixtures__/change-requests.json";
import eventsFixture from "@/features/core/allocations/__fixtures__/events.json";
import type {
  AllocationDiff,
  AllocationMembership,
  AttachedResource,
  AllocationUsage,
  ChangeRequest,
  ChangeRequestEvent,
  ComputeAllocation,
  CreateChangeRequestPayload,
  CreateMembershipPayload,
  UpdateChangeRequestPayload,
  UpdateMembershipPayload,
} from "@/features/core/allocations/schemas";

const allocations: ComputeAllocation[] = (allocationsFixture as ComputeAllocation[]).map((a) => ({
  ...a,
}));
const resourcesByAlloc: Record<string, AttachedResource[]> = Object.fromEntries(
  Object.entries(resourcesFixture as Record<string, AttachedResource[]>).map(
    ([id, rows]) => [id, rows.map((r) => ({ ...r }))],
  ),
);
const usageByAlloc = usageFixture as Record<string, AllocationUsage[]>;
const diffsByAlloc = diffsFixture as Record<string, AllocationDiff[]>;
const membersByAlloc: Record<string, AllocationMembership[]> = Object.fromEntries(
  Object.entries(membersFixture as Record<string, AllocationMembership[]>).map(([id, rows]) => [
    id,
    rows.map((r) => ({ ...r })),
  ]),
);
const changeRequests: ChangeRequest[] = (changeRequestsFixture as ChangeRequest[]).map((cr) => ({
  ...cr,
}));
const eventsByCr: Record<string, ChangeRequestEvent[]> = Object.fromEntries(
  Object.entries(eventsFixture as Record<string, ChangeRequestEvent[]>).map(([id, rows]) => [
    id,
    rows.map((r) => ({ ...r })),
  ]),
);

let memberSeq = 1000;
let crSeq = 1000;
let eventSeq = 1000;

function ensureMembersBucket(allocId: string): AllocationMembership[] {
  if (!membersByAlloc[allocId]) membersByAlloc[allocId] = [];
  return membersByAlloc[allocId];
}

function ensureEventsBucket(crId: string): ChangeRequestEvent[] {
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

  http.get("*/api/v1/compute-allocations/:id", ({ params }) => {
    const id = String(params.id);
    const found = allocations.find((a) => a.id === id);
    if (!found) return HttpResponse.json({ error: "allocation not found" }, { status: 404 });
    return HttpResponse.json(found);
  }),

  http.get("*/api/v1/compute-allocations/:id/resources", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(resourcesByAlloc[id] ?? []);
  }),

  http.get("*/api/v1/compute-allocations/:id/usages", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(usageByAlloc[id] ?? []);
  }),

  http.get("*/api/v1/compute-allocations/:id/diffs", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(diffsByAlloc[id] ?? []);
  }),

  http.get("*/api/v1/compute-allocations/:id/diffs/latest", ({ params }) => {
    const id = String(params.id);
    const rows = [...(diffsByAlloc[id] ?? [])].sort(
      (a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime(),
    );
    if (rows.length === 0) return HttpResponse.json({ error: "no diffs" }, { status: 404 });
    return HttpResponse.json(rows[0]);
  }),

  http.get("*/api/v1/compute-allocations/:id/memberships", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(membersByAlloc[id] ?? []);
  }),

  http.post("*/api/v1/compute-allocation-memberships", async ({ request }) => {
    const body = (await request.json()) as CreateMembershipPayload;
    const bucket = ensureMembersBucket(body.compute_allocation_id);
    const member: AllocationMembership = {
      id: `mem-${memberSeq++}`,
      compute_allocation_id: body.compute_allocation_id,
      user_id: body.user_id,
      start_time: body.start_time,
      end_time: body.end_time,
      membership_status: body.membership_status ?? "ACTIVE",
      role: body.role,
      display_name: body.user_id,
      email: `${body.user_id}@custos.local`,
    };
    bucket.push(member);
    return HttpResponse.json(member, { status: 201 });
  }),

  http.put("*/api/v1/compute-allocation-memberships/:id", async ({ params, request }) => {
    const id = String(params.id);
    const patch = (await request.json()) as UpdateMembershipPayload;
    for (const bucket of Object.values(membersByAlloc)) {
      const existing = bucket.find((m) => m.id === id);
      if (existing) {
        if (patch.start_time) existing.start_time = patch.start_time;
        if (patch.end_time) existing.end_time = patch.end_time;
        if (patch.membership_status) existing.membership_status = patch.membership_status;
        if (patch.role) existing.role = patch.role;
        return HttpResponse.json(existing);
      }
    }
    return HttpResponse.json({ error: "membership not found" }, { status: 404 });
  }),

  http.delete("*/api/v1/compute-allocation-memberships/:id", ({ params }) => {
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

  http.get("*/api/v1/compute-allocations/:id/change-requests", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(changeRequests.filter((cr) => cr.compute_allocation_id === id));
  }),

  http.get("*/api/v1/users/:id/change-requests", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(changeRequests.filter((cr) => cr.requester_id === id));
  }),

  http.get("*/api/v1/compute-allocation-change-requests", ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get("status");
    const filtered = status ? changeRequests.filter((cr) => cr.change_status === status) : changeRequests;
    return HttpResponse.json(filtered);
  }),

  http.get("*/api/v1/compute-allocation-change-requests/:id", ({ params }) => {
    const id = String(params.id);
    const found = changeRequests.find((cr) => cr.id === id);
    if (!found) return HttpResponse.json({ error: "change request not found" }, { status: 404 });
    return HttpResponse.json(found);
  }),

  http.post("*/api/v1/compute-allocation-change-requests", async ({ request }) => {
    const body = (await request.json()) as CreateChangeRequestPayload;
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

  http.put("*/api/v1/compute-allocation-change-requests/:id", async ({ params, request }) => {
    const id = String(params.id);
    const patch = (await request.json()) as UpdateChangeRequestPayload;
    const existing = changeRequests.find((cr) => cr.id === id);
    if (!existing) return HttpResponse.json({ error: "change request not found" }, { status: 404 });
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
  }),

  http.delete("*/api/v1/compute-allocation-change-requests/:id", ({ params }) => {
    const id = String(params.id);
    const idx = changeRequests.findIndex((cr) => cr.id === id);
    if (idx === -1) return HttpResponse.json({ error: "change request not found" }, { status: 404 });
    changeRequests.splice(idx, 1);
    return new HttpResponse(null, { status: 204 });
  }),

  http.get("*/api/v1/compute-allocation-change-requests/:id/events", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(eventsByCr[id] ?? []);
  }),
];
