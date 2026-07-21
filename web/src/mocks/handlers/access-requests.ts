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

import type { AccessRequest } from "@/features/core/access-requests/schemas";
import { http, HttpResponse } from "msw";

const EVENT_NAMES: Record<string, string> = {
  PEARC26: "pearc26-tutorial",
};

const EVENT_ALLOCATIONS: Record<string, string> = {
  PEARC26: "alloc-pearc26-001",
};

function seedRequests(): AccessRequest[] {
  return [
    {
      id: "areq-001",
      oidc_sub: "sub-queue-001",
      email: "amara.osei@example.edu",
      name: "Amara Osei",
      institution: "Example University",
      event_code: "PEARC26",
      reason: "Hands-on tutorial attendee.",
      status: "PENDING",
      approver_id: "",
      deny_reason: "",
      expires_at: null,
      created_user_id: "",
      timestamp: "2026-07-10T14:05:00Z",
      decided_at: null,
    },
    {
      id: "areq-002",
      oidc_sub: "sub-queue-002",
      email: "li.wei@example.org",
      name: "Li Wei",
      institution: "Example Institute",
      event_code: "PEARC26",
      reason: "",
      status: "PENDING",
      approver_id: "",
      deny_reason: "",
      expires_at: null,
      created_user_id: "",
      timestamp: "2026-07-11T09:30:00Z",
      decided_at: null,
    },
    {
      id: "areq-003",
      oidc_sub: "sub-queue-003",
      email: "marcus.rivera@utexas.edu",
      name: "Marcus Rivera",
      institution: "University of Texas",
      event_code: "PEARC26",
      reason: "Tutorial follow-up experiments.",
      status: "APPROVED",
      approver_id: "user-admin-001",
      deny_reason: "",
      expires_at: "2026-08-09T00:00:00Z",
      created_user_id: "u2",
      timestamp: "2026-07-09T11:00:00Z",
      decided_at: "2026-07-10T08:15:00Z",
    },
  ];
}

let requests = seedRequests();
// The signed-in mock caller's own request; null = no request yet.
let mine: AccessRequest | null = null;

export function resetAccessRequests() {
  requests = seedRequests();
  mine = null;
}

const notFound = () => HttpResponse.json({ error: "not found" }, { status: 404 });

// Test seam: e2e seeds the caller's own denied request via a non-httpOnly
// cookie, mirroring the privileges handler.
function deniedMineFromCookie(): AccessRequest | null {
  if (typeof document === "undefined") return null;
  if (!document.cookie.split("; ").includes("custos.test-my-access-request=DENIED")) return null;
  return {
    id: "areq-mine-denied",
    oidc_sub: "sub-mock-caller",
    email: "caller@example.edu",
    name: "Mock Caller",
    institution: "Example University",
    event_code: "PEARC26",
    reason: "",
    status: "DENIED",
    approver_id: "user-admin-001",
    deny_reason: "Could not verify event registration for this attendee.",
    expires_at: null,
    created_user_id: "",
    timestamp: "2026-07-10T14:05:00Z",
    decided_at: "2026-07-11T09:00:00Z",
  };
}

// The privileged list carries decision context; /me stays the bare model.
function bare(row: AccessRequest): Omit<AccessRequest, "decided_at" | "allocation_id"> {
  const { decided_at: _decidedAt, allocation_id: _allocationId, ...rest } = row;
  return rest;
}

export const accessRequestsHandlers = [
  http.get("*/api/v1/access-requests/me", () => {
    const row = mine ?? deniedMineFromCookie();
    return row ? HttpResponse.json(bare(row)) : notFound();
  }),

  http.get("*/api/v1/access-requests/events/:code", ({ params }) => {
    const code = String(params.code);
    const name = EVENT_NAMES[code];
    return name ? HttpResponse.json({ code, name }) : notFound();
  }),

  http.get("*/api/v1/access-requests/username", ({ request }) => {
    const url = new URL(request.url);
    if (!EVENT_NAMES[url.searchParams.get("event_code") ?? ""]) return notFound();
    const username = url.searchParams.get("username") ?? "";
    const valid = username === "" || /^[a-z][a-z0-9_-]{0,31}$/.test(username);
    // "taken" stands in for a claimed login so the red-cross path is testable.
    const available = valid && username !== "taken";
    return HttpResponse.json({ suggestion: "nexus-mockcaller", valid, available });
  }),

  http.post("*/api/v1/access-requests", async ({ request }) => {
    if (mine?.status === "PENDING") {
      return HttpResponse.json({ error: "a pending request already exists" }, { status: 409 });
    }
    const body = (await request.json()) as {
      institution?: string;
      event_code?: string;
      reason?: string;
    };
    if (!body.institution || !body.event_code) {
      return HttpResponse.json(
        { error: "institution and event_code are required" },
        { status: 400 },
      );
    }
    if (!EVENT_NAMES[body.event_code]) return notFound();
    mine = {
      id: `areq-${String(requests.length + 1).padStart(3, "0")}`,
      oidc_sub: "sub-mock-caller",
      email: "caller@example.edu",
      name: "Mock Caller",
      institution: body.institution,
      event_code: body.event_code,
      reason: body.reason ?? "",
      status: "PENDING",
      approver_id: "",
      deny_reason: "",
      expires_at: null,
      created_user_id: "",
      timestamp: new Date().toISOString(),
    };
    requests = [mine, ...requests];
    return HttpResponse.json(mine, { status: 201 });
  }),

  http.get("*/api/v1/access-requests", ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get("status");
    const event = url.searchParams.get("event");
    let rows = requests;
    if (status) rows = rows.filter((r) => r.status === status);
    if (event) rows = rows.filter((r) => r.event_code === event);
    return HttpResponse.json(
      rows.map((r) => ({ ...r, allocation_id: EVENT_ALLOCATIONS[r.event_code] ?? null })),
    );
  }),

  http.put("*/api/v1/access-requests/:id", async ({ params, request }) => {
    const row = requests.find((r) => r.id === String(params.id));
    if (!row) return notFound();
    if (row.status !== "PENDING") {
      return HttpResponse.json({ error: "request already decided" }, { status: 400 });
    }
    const body = (await request.json()) as {
      status?: string;
      expires_at?: string;
      deny_reason?: string;
    };
    if (body.status !== "APPROVED" && body.status !== "DENIED") {
      return HttpResponse.json({ error: "status must be APPROVED or DENIED" }, { status: 400 });
    }
    row.status = body.status;
    row.approver_id = "user-admin-001";
    row.decided_at = new Date().toISOString();
    if (body.status === "APPROVED") {
      row.expires_at =
        body.expires_at ?? new Date(Date.now() + 30 * 24 * 3600 * 1000).toISOString();
    } else {
      row.deny_reason = body.deny_reason ?? "";
    }
    return HttpResponse.json(row);
  }),
];
