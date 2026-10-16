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
import catalogFixture from "@/features/core/resources/__fixtures__/resources.json";
import type {
  AllocationMembershipResponse,
  ComputeAllocationChangeRequest,
  ComputeAllocationMembership,
  DeleteComputeAllocationChangeRequestsByIdData,
  DeleteComputeAllocationMembershipsByIdData,
  DeleteComputeAllocationUsagesByIdData,
  GetComputeAllocationChangeRequestsByIdData,
  GetComputeAllocationChangeRequestsByIdEventsData,
  GetComputeAllocationMembershipsByIdData,
  GetComputeAllocationUsagesByIdData,
  GetComputeAllocationsByIdChangeRequestsData,
  GetComputeAllocationsByIdData,
  GetComputeAllocationsByIdDiffsData,
  GetComputeAllocationsByIdMembershipsData,
  GetComputeAllocationsByIdResourcesData,
  GetComputeAllocationsByIdUsagesData,
  GetUsersByIdChangeRequestsData,
  PostComputeAllocationChangeRequestsData,
  PostComputeAllocationMembershipsData,
  PutComputeAllocationChangeRequestsByIdData,
} from "@/generated/core/types.gen";
import {
  zAllocationMembershipResponse,
  zComputeAllocation,
  zComputeAllocationChangeRequest,
  zComputeAllocationChangeRequestEvent,
  zComputeAllocationDiff,
  zComputeAllocationResource,
  zComputeAllocationResourceMapping,
  zComputeAllocationUsage,
} from "@/generated/core/zod.gen";
import { http, HttpResponse } from "msw";
import { z } from "zod";
import { listJson, notFound, page } from "../paging";
import { mockProjectRole } from "./projects";

const allocations = z.array(zComputeAllocation).parse(allocationsFixture);
// Fixtures keyed by parent id.
const byParent = <T>(schema: z.ZodType<T>, fixture: unknown) =>
  z.record(z.string(), z.array(schema)).parse(fixture);
const mappingsByAlloc = byParent(zComputeAllocationResourceMapping, resourcesFixture);
const catalog = z.array(zComputeAllocationResource).parse(catalogFixture);
// As master's join: the attached resources themselves, by name.
const attachedTo = (allocationId: string) =>
  catalog
    .filter((r) =>
      mappingsByAlloc[allocationId]?.some((m) => m.compute_allocation_resource_id === r.id),
    )
    .sort((a, b) => a.name.localeCompare(b.name));
const usageByAlloc = byParent(zComputeAllocationUsage, usageFixture);
const diffsByAlloc = byParent(zComputeAllocationDiff, diffsFixture);
const membersByAlloc = byParent(zAllocationMembershipResponse, membersFixture);
const changeRequests = z.array(zComputeAllocationChangeRequest).parse(changeRequestsFixture);
const eventsByCr = byParent(zComputeAllocationChangeRequestEvent, eventsFixture);

const byTime = (a: { timestamp: string }, b: { timestamp: string }) =>
  Date.parse(a.timestamp) - Date.parse(b.timestamp);
const badRequest = (error: string) => HttpResponse.json({ error }, { status: 400 });

function pushEvent(crId: string, eventType: string, description: string) {
  eventsByCr[crId] ??= [];
  eventsByCr[crId].push({
    id: crypto.randomUUID(),
    compute_allocation_change_request_id: crId,
    event_type: eventType,
    description,
    timestamp: new Date().toISOString(),
  });
}

// The stored membership, without the user and role the list joins.
const stored = ({ role, display_name, email, ...m }: AllocationMembershipResponse) =>
  m satisfies ComputeAllocationMembership;

function filterAllocations(url: URL) {
  const projectId = url.searchParams.get("project_id");
  const status = url.searchParams.get("status");
  const q = url.searchParams.get("q")?.toLowerCase() ?? "";
  return allocations.filter(
    (a) =>
      (!projectId || a.project_id === projectId) &&
      (!status || a.status === status) &&
      (!q || a.name.toLowerCase().includes(q)),
  );
}

export const allocationsHandlers = [
  http.get("*/api/v1/compute-allocations", ({ request }) => {
    const { items, total } = page(new URL(request.url), filterAllocations(new URL(request.url)));
    return HttpResponse.json({ items, total });
  }),

  http.get<GetComputeAllocationsByIdData["path"]>(
    "*/api/v1/compute-allocations/:id",
    ({ params }) => {
      const found = allocations.find((a) => a.id === params.id);
      return found ? HttpResponse.json(found) : notFound("compute allocation");
    },
  ),

  http.get<GetComputeAllocationsByIdResourcesData["path"]>(
    "*/api/v1/compute-allocations/:id/resources",
    ({ params }) => listJson(attachedTo(params.id)),
  ),

  http.get<GetComputeAllocationsByIdUsagesData["path"]>(
    "*/api/v1/compute-allocations/:id/usages",
    ({ params }) => listJson(usageByAlloc[params.id] ?? []),
  ),

  http.get<GetComputeAllocationUsagesByIdData["path"]>(
    "*/api/v1/compute-allocation-usages/:id",
    ({ params }) => {
      const usage = Object.values(usageByAlloc)
        .flat()
        .find((u) => u.id === params.id);
      return usage ? HttpResponse.json(usage) : notFound("usage");
    },
  ),

  http.delete<DeleteComputeAllocationUsagesByIdData["path"]>(
    "*/api/v1/compute-allocation-usages/:id",
    ({ params }) => {
      for (const bucket of Object.values(usageByAlloc)) {
        const idx = bucket.findIndex((u) => u.id === params.id);
        if (idx !== -1) {
          bucket.splice(idx, 1);
          return new HttpResponse(null, { status: 204 });
        }
      }
      return notFound("usage");
    },
  ),

  http.get<GetComputeAllocationsByIdDiffsData["path"]>(
    "*/api/v1/compute-allocations/:id/diffs",
    ({ params }) => listJson(diffsByAlloc[params.id] ?? []),
  ),

  http.get<GetComputeAllocationsByIdMembershipsData["path"]>(
    "*/api/v1/compute-allocations/:id/memberships",
    ({ params }) => HttpResponse.json(membersByAlloc[params.id] ?? []),
  ),

  http.post<never, PostComputeAllocationMembershipsData["body"]>(
    "*/api/v1/compute-allocation-memberships",
    async ({ request }) => {
      const body = await request.json();
      const allocId = body.compute_allocation_id ?? "";
      const alloc = allocations.find((a) => a.id === allocId);
      if (!alloc) return badRequest(`compute allocation "${allocId}" not found`);
      if (!body.user_id) return badRequest("user_id is required");
      const member = zAllocationMembershipResponse.parse({
        ...body,
        id: crypto.randomUUID(),
        membership_status: body.membership_status || "ACTIVE",
        role: mockProjectRole(alloc.project_id, body.user_id),
        display_name: body.user_id,
        email: `${body.user_id}@custos.local`,
      });
      membersByAlloc[allocId] ??= [];
      membersByAlloc[allocId].push(member);
      return HttpResponse.json(stored(member), { status: 201 });
    },
  ),

  http.get<GetComputeAllocationMembershipsByIdData["path"]>(
    "*/api/v1/compute-allocation-memberships/:id",
    ({ params }) => {
      const member = Object.values(membersByAlloc)
        .flat()
        .find((m) => m.id === params.id);
      return member ? HttpResponse.json(stored(member)) : notFound("membership");
    },
  ),

  http.get("*/api/v1/compute-allocation-memberships/:id/resource-overrides", () => listJson([])),

  http.delete<DeleteComputeAllocationMembershipsByIdData["path"]>(
    "*/api/v1/compute-allocation-memberships/:id",
    ({ params }) => {
      for (const bucket of Object.values(membersByAlloc)) {
        const idx = bucket.findIndex((m) => m.id === params.id);
        if (idx !== -1) {
          bucket.splice(idx, 1);
          return new HttpResponse(null, { status: 204 });
        }
      }
      return notFound("membership");
    },
  ),

  http.get<GetComputeAllocationsByIdChangeRequestsData["path"]>(
    "*/api/v1/compute-allocations/:id/change-requests",
    ({ params }) =>
      listJson(changeRequests.filter((cr) => cr.compute_allocation_id === params.id).sort(byTime)),
  ),

  http.get<GetUsersByIdChangeRequestsData["path"]>(
    "*/api/v1/users/:id/change-requests",
    ({ params }) =>
      listJson(changeRequests.filter((cr) => cr.requester_id === params.id).sort(byTime)),
  ),

  // The newest first; limit defaults to 50 and is capped at 200.
  http.get("*/api/v1/compute-allocation-change-requests", ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get("status");
    const limit = Math.min(Number(url.searchParams.get("limit")) || 50, 200);
    const rows = changeRequests
      .filter((cr) => !status || cr.change_status === status)
      .sort((a, b) => byTime(b, a));
    return HttpResponse.json(rows.slice(0, limit));
  }),

  http.get<GetComputeAllocationChangeRequestsByIdData["path"]>(
    "*/api/v1/compute-allocation-change-requests/:id",
    ({ params }) => {
      const found = changeRequests.find((cr) => cr.id === params.id);
      return found ? HttpResponse.json(found) : notFound("change request");
    },
  ),

  http.post<never, PostComputeAllocationChangeRequestsData["body"]>(
    "*/api/v1/compute-allocation-change-requests",
    async ({ request }) => {
      const body = await request.json();
      if (!body.compute_allocation_id) return badRequest("compute_allocation_id is required");
      if (!body.requester_id) return badRequest("requester_id is required");
      if (!allocations.some((a) => a.id === body.compute_allocation_id)) {
        return badRequest(`compute allocation "${body.compute_allocation_id}" not found`);
      }
      const cr = zComputeAllocationChangeRequest.parse({
        ...body,
        id: crypto.randomUUID(),
        change_status: body.change_status || "PENDING",
        timestamp: body.timestamp || new Date().toISOString(),
      });
      changeRequests.push(cr);
      pushEvent(cr.id, "CREATED", "Change request created");
      return HttpResponse.json(cr, { status: 201 });
    },
  ),

  // Empty fields keep the stored value, except approver_id; a status change is recorded as an
  // event of that status, and approval writes a diff.
  http.put<
    PutComputeAllocationChangeRequestsByIdData["path"],
    PutComputeAllocationChangeRequestsByIdData["body"]
  >("*/api/v1/compute-allocation-change-requests/:id", async ({ params, request }) => {
    const body = await request.json();
    const existing = changeRequests.find((cr) => cr.id === params.id);
    if (!existing) return notFound("change request");
    const previous = existing.change_status;
    Object.assign(existing, {
      compute_allocation_id: body.compute_allocation_id || existing.compute_allocation_id,
      requester_id: body.requester_id || existing.requester_id,
      requested_su_amount: body.requested_su_amount || existing.requested_su_amount,
      requested_status: body.requested_status || existing.requested_status,
      reason: body.reason || existing.reason,
      change_status: body.change_status || existing.change_status,
      approver_id: body.approver_id,
      timestamp: body.timestamp || existing.timestamp,
    });
    const next = existing.change_status;
    if (next === previous) pushEvent(params.id, "UPDATED", "Change request updated");
    else pushEvent(params.id, next, `Change request transitioned from ${previous} to ${next}`);
    const allocId = existing.compute_allocation_id;
    if (next === "APPROVED" && previous !== "APPROVED") {
      diffsByAlloc[allocId] = [
        ...(diffsByAlloc[allocId] ?? []),
        {
          id: crypto.randomUUID(),
          compute_allocation_id: allocId,
          diff_type: "CHANGE_REQUEST_APPROVED",
          new_su_amount: existing.requested_su_amount,
          status: existing.requested_status,
          description: `Applied approved change request ${params.id}`,
          timestamp: new Date().toISOString(),
        },
      ];
    }
    return HttpResponse.json(existing);
  }),

  http.delete<DeleteComputeAllocationChangeRequestsByIdData["path"]>(
    "*/api/v1/compute-allocation-change-requests/:id",
    ({ params }) => {
      const idx = changeRequests.findIndex((cr) => cr.id === params.id);
      if (idx === -1) return notFound("change request");
      pushEvent(params.id, "DELETED", "Change request deleted");
      changeRequests.splice(idx, 1);
      return new HttpResponse(null, { status: 204 });
    },
  ),

  http.get<GetComputeAllocationChangeRequestsByIdEventsData["path"]>(
    "*/api/v1/compute-allocation-change-requests/:id/events",
    ({ params }) => listJson(eventsByCr[params.id] ?? []),
  ),
];
