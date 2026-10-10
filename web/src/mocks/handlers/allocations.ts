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
import changeRequestsFixture from "@/features/core/allocations/__fixtures__/change-requests.json";
import diffsFixture from "@/features/core/allocations/__fixtures__/diffs.json";
import eventsFixture from "@/features/core/allocations/__fixtures__/events.json";
import membersFixture from "@/features/core/allocations/__fixtures__/members.json";
import resourcesFixture from "@/features/core/allocations/__fixtures__/resources.json";
import usageFixture from "@/features/core/allocations/__fixtures__/usage.json";
import type {
  AllocationMembershipResponse,
  ComputeAllocation,
  ComputeAllocationChangeRequest,
  DeleteComputeAllocationChangeRequestsByIdData,
  DeleteComputeAllocationMembershipsByIdData,
  GetComputeAllocationChangeRequestsByIdData,
  GetComputeAllocationChangeRequestsByIdEventsData,
  GetComputeAllocationsByIdChangeRequestsData,
  GetComputeAllocationsByIdData,
  GetComputeAllocationsByIdDiffsData,
  GetComputeAllocationsByIdDiffsLatestData,
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
  zComputeAllocationUsage,
} from "@/generated/core/zod.gen";
import { z } from "zod";
import { mockProjectRole, rejectUnlessCustos } from "./projects";
import { notFound, page } from "../paging";

const allocations: ComputeAllocation[] = z.array(zComputeAllocation).parse(allocationsFixture);
// Fixtures keyed by parent id.
const byParent = <T>(schema: z.ZodType<T>, fixture: unknown) =>
  z.record(z.string(), z.array(schema)).parse(fixture);
const resourcesByAlloc = byParent(zComputeAllocationResource, resourcesFixture);
const usageByAlloc = byParent(zComputeAllocationUsage, usageFixture);
const diffsByAlloc = byParent(zComputeAllocationDiff, diffsFixture);
const membersByAlloc = byParent(zAllocationMembershipResponse, membersFixture);
const changeRequests: ComputeAllocationChangeRequest[] = z
  .array(zComputeAllocationChangeRequest)
  .parse(changeRequestsFixture);
const eventsByCr = byParent(zComputeAllocationChangeRequestEvent, eventsFixture);

let memberSeq = 1000;
let crSeq = 1000;
let eventSeq = 1000;
let diffSeq = 1000;

function pushEvent(crId: string, eventType: string, description: string, timestamp: string) {
  eventsByCr[crId] ??= [];
  eventsByCr[crId].push({
    id: `evt-${eventSeq++}`,
    compute_allocation_change_request_id: crId,
    event_type: eventType,
    description,
    timestamp,
  });
}

function projectOf(allocId: string | undefined): string {
  return allocations.find((a) => a.id === allocId)?.project_id ?? "";
}

function filterAllocations(url: URL): ComputeAllocation[] {
  const projectId = url.searchParams.get("project_id");
  const status = url.searchParams.get("status");
  const q = url.searchParams.get("q")?.toLowerCase() ?? "";
  return allocations.filter((a) => {
    if (projectId && a.project_id !== projectId) return false;
    if (status && a.status !== status) return false;
    if (q && !(a.name ?? "").toLowerCase().includes(q)) return false;
    return true;
  });
}

export const allocationsHandlers = [
  http.get("*/api/v1/compute-allocations", ({ request }) => {
    const url = new URL(request.url);
    const { items, total } = page(url, filterAllocations(url));
    return HttpResponse.json({ items, total });
  }),

  http.get<GetComputeAllocationsByIdData["path"]>(
    "*/api/v1/compute-allocations/:id",
    ({ params }) => {
      const found = allocations.find((a) => a.id === params.id);
      return found ? HttpResponse.json(found) : notFound("allocation");
    },
  ),

  http.get<GetComputeAllocationsByIdResourcesData["path"]>(
    "*/api/v1/compute-allocations/:id/resources",
    ({ params }) => HttpResponse.json(resourcesByAlloc[params.id] ?? []),
  ),

  http.get<GetComputeAllocationsByIdUsagesData["path"]>(
    "*/api/v1/compute-allocations/:id/usages",
    ({ params }) => HttpResponse.json(usageByAlloc[params.id] ?? []),
  ),

  http.get<GetComputeAllocationsByIdDiffsData["path"]>(
    "*/api/v1/compute-allocations/:id/diffs",
    ({ params }) => HttpResponse.json(diffsByAlloc[params.id] ?? []),
  ),

  http.get<GetComputeAllocationsByIdDiffsLatestData["path"]>(
    "*/api/v1/compute-allocations/:id/diffs/latest",
    ({ params }) => {
      const [latest] = [...(diffsByAlloc[params.id] ?? [])].sort(
        (a, b) => Date.parse(b.timestamp ?? "") - Date.parse(a.timestamp ?? ""),
      );
      return latest ? HttpResponse.json(latest) : notFound("diff");
    },
  ),

  http.get<GetComputeAllocationsByIdMembershipsData["path"]>(
    "*/api/v1/compute-allocations/:id/memberships",
    ({ params }) => HttpResponse.json(membersByAlloc[params.id] ?? []),
  ),

  http.post<never, PostComputeAllocationMembershipsData["body"]>(
    "*/api/v1/compute-allocation-memberships",
    async ({ request }) => {
      const {
        compute_allocation_id: allocId = "",
        user_id: userId = "",
        ...body
      } = await request.json();
      const projectId = projectOf(allocId);
      const rejected = rejectUnlessCustos(projectId);
      if (rejected) return rejected;
      const member: AllocationMembershipResponse = {
        ...body,
        id: `mem-${memberSeq++}`,
        compute_allocation_id: allocId,
        user_id: userId,
        membership_status: body.membership_status ?? "ACTIVE",
        role: mockProjectRole(projectId, userId),
        display_name: userId,
        email: `${userId}@custos.local`,
      };
      membersByAlloc[allocId] ??= [];
      membersByAlloc[allocId].push(member);
      return HttpResponse.json(member, { status: 201 });
    },
  ),

  http.delete<DeleteComputeAllocationMembershipsByIdData["path"]>(
    "*/api/v1/compute-allocation-memberships/:id",
    ({ params }) => {
      for (const [allocId, bucket] of Object.entries(membersByAlloc)) {
        const idx = bucket.findIndex((m) => m.id === params.id);
        if (idx !== -1) {
          const rejected = rejectUnlessCustos(projectOf(allocId));
          if (rejected) return rejected;
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
      HttpResponse.json(changeRequests.filter((cr) => cr.compute_allocation_id === params.id)),
  ),

  http.get<GetUsersByIdChangeRequestsData["path"]>(
    "*/api/v1/users/:id/change-requests",
    ({ params }) => HttpResponse.json(changeRequests.filter((cr) => cr.requester_id === params.id)),
  ),

  http.get("*/api/v1/compute-allocation-change-requests", ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get("status");
    const limit = Math.min(Number(url.searchParams.get("limit") ?? 50), 200);
    return HttpResponse.json(
      changeRequests
        .filter((cr) => !status || cr.change_status === status)
        .sort((a, b) => Date.parse(b.timestamp ?? "") - Date.parse(a.timestamp ?? ""))
        .slice(0, limit),
    );
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
      const rejected = rejectUnlessCustos(projectOf(body.compute_allocation_id));
      if (rejected) return rejected;
      const id = `cr-${crSeq++}`;
      const timestamp = new Date().toISOString();
      const cr: ComputeAllocationChangeRequest = {
        ...body,
        id,
        change_status: "PENDING",
        timestamp,
      };
      changeRequests.unshift(cr);
      pushEvent(id, "CREATED", `Change request created by ${cr.requester_id}`, timestamp);
      return HttpResponse.json(cr, { status: 201 });
    },
  ),

  http.put<
    PutComputeAllocationChangeRequestsByIdData["path"],
    PutComputeAllocationChangeRequestsByIdData["body"]
  >("*/api/v1/compute-allocation-change-requests/:id", async ({ params, request }) => {
    const body = await request.json();
    const existing = changeRequests.find((cr) => cr.id === params.id);
    if (!existing) return notFound("change request");
    const rejected = rejectUnlessCustos(projectOf(existing.compute_allocation_id));
    if (rejected) return rejected;
    const previous = existing.change_status;
    Object.assign(existing, body, { id: params.id });
    const now = new Date().toISOString();
    if (existing.change_status === "APPROVED" || existing.change_status === "REJECTED") {
      const verb = existing.change_status === "APPROVED" ? "Approved" : "Rejected";
      pushEvent(
        params.id,
        existing.change_status,
        `${verb} by ${existing.approver_id ?? "unknown"}`,
        now,
      );
    }
    // Approval materialises the requested amount as a diff, as the backend does.
    if (existing.change_status === "APPROVED" && previous !== "APPROVED") {
      const allocId = existing.compute_allocation_id ?? "";
      diffsByAlloc[allocId] ??= [];
      diffsByAlloc[allocId].push({
        id: `diff-${diffSeq++}`,
        compute_allocation_id: allocId,
        diff_type: "CHANGE_REQUEST_APPROVED",
        new_su_amount: existing.requested_su_amount,
        status: existing.requested_status ?? "ACTIVE",
        timestamp: now,
        description: `Applied approved change request ${params.id}`,
      });
    }
    return HttpResponse.json(existing);
  }),

  http.delete<DeleteComputeAllocationChangeRequestsByIdData["path"]>(
    "*/api/v1/compute-allocation-change-requests/:id",
    ({ params }) => {
      const idx = changeRequests.findIndex((cr) => cr.id === params.id);
      if (idx === -1) return notFound("change request");
      const rejected = rejectUnlessCustos(projectOf(changeRequests[idx]?.compute_allocation_id));
      if (rejected) return rejected;
      changeRequests.splice(idx, 1);
      return new HttpResponse(null, { status: 204 });
    },
  ),

  http.get<GetComputeAllocationChangeRequestsByIdEventsData["path"]>(
    "*/api/v1/compute-allocation-change-requests/:id/events",
    ({ params }) => HttpResponse.json(eventsByCr[params.id] ?? []),
  ),
];
