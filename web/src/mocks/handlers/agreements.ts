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

import { HttpResponse, http } from "msw";
import type {
  AcceptanceReport,
  AgreementDetail,
  AgreementDraft,
} from "@/features/core/agreements/schemas";

const admin = { id: "user-admin", name: "Priya Nair", email: "priya.nair@example.edu" };
const user = (id: string, name: string) => ({ id, name, email: `${id.slice(2)}@example.edu` });

const BODY =
  "## Acceptable use\n\nUse the compute, storage and credentials only for your approved research.\n\nDo not share your account.";

const agreements: AgreementDetail[] = [
  {
    id: "agr-2",
    title: "Custos Acceptable Use Policy",
    version_label: "2.0",
    status: "DRAFT",
    summary: "Rules for using Custos compute, storage and credentials.",
    body_md: `${BODY}\n\nReport suspected misuse within 24 hours.`,
    change_note: "Adds a reporting rule for suspected misuse.",
    body_hash: "mock-hash-2",
    created_at: "2026-10-10T14:00:00Z",
    updated_at: "2026-10-10T14:00:00Z",
    created_by: admin,
    accepted_count: 0,
    pending_count: 0,
  },
  {
    id: "agr-1",
    title: "Custos Acceptable Use Policy",
    version_label: "1.0",
    status: "PUBLISHED",
    summary: "Rules for using Custos compute, storage and credentials.",
    body_md: BODY,
    change_note: "",
    body_hash: "mock-hash-1",
    created_at: "2026-03-02T09:00:00Z",
    updated_at: "2026-03-12T09:00:00Z",
    created_by: admin,
    published_at: "2026-03-12T09:00:00Z",
    published_by: admin,
    accepted_count: 3,
    pending_count: 2,
  },
  {
    id: "agr-0",
    title: "Custos Acceptable Use Policy",
    version_label: "0.9",
    status: "REMOVED",
    summary: "First draft of the rules.",
    body_md: BODY,
    change_note: "",
    body_hash: "mock-hash-0",
    created_at: "2025-08-20T09:00:00Z",
    updated_at: "2026-03-12T09:00:00Z",
    created_by: admin,
    published_at: "2025-09-04T09:00:00Z",
    published_by: admin,
    removed_at: "2026-03-12T09:00:00Z",
    removed_by: admin,
    accepted_count: 2,
    pending_count: 0,
  },
];

const reports: Record<string, AcceptanceReport> = {
  "agr-1": {
    accepted: [
      { user: user("u-jdoe", "John Doe"), accepted_at: "2026-10-17T15:42:00Z" },
      { user: user("u-tomas", "Tomas Herrera"), accepted_at: "2026-10-16T08:05:00Z" },
      { user: user("u-sofia", "Sofia Lindqvist"), accepted_at: "2026-10-09T11:30:00Z" },
    ],
    pending: [
      { user: user("u-marcus", "Marcus Chen"), first_sign_in_at: "2026-10-14T09:12:00Z" },
      { user: user("u-aisha", "Aisha Okafor"), first_sign_in_at: null },
    ],
  },
  "agr-0": {
    accepted: [
      { user: user("u-daniel", "Daniel Park"), accepted_at: "2025-09-28T17:18:00Z" },
      { user: user("u-nadia", "Nadia Rahman"), accepted_at: "2025-09-21T13:57:00Z" },
    ],
    pending: [],
  },
};

const find = (id: unknown) => agreements.find((a) => a.id === id);
const notFound = () => HttpResponse.json({ error: "agreement not found" }, { status: 404 });
const versionTaken = (label: string, exceptId?: string) =>
  agreements.some((a) => a.version_label === label && a.id !== exceptId);
const conflict = () =>
  HttpResponse.json({ error: "version label already exists" }, { status: 409 });

// Test seam: the non-httpOnly custos.test-gate cookie ("onboarding" or "consent")
// makes /me report that gate with the published version pending.
export function gateProfile() {
  const gate =
    typeof document === "undefined"
      ? undefined
      : document.cookie
          .split("; ")
          .find((c) => c.startsWith("custos.test-gate="))
          ?.slice("custos.test-gate=".length);
  const pending = gate
    ? agreements
        .filter((a) => a.status === "PUBLISHED")
        .map(({ created_by, published_by, removed_by, ...rest }) => rest)
    : [];
  return {
    consent: { pending },
    onboarding: {
      required: gate === "onboarding",
      username: gate === "onboarding" ? "custos-jdoe" : "",
      username_editable: gate === "onboarding",
    },
  };
}

const USERNAME = /^[a-z][a-z0-9_-]*$/;

export const agreementsHandlers = [
  http.get("*/api/v1/me/onboarding/username", ({ request }) => {
    const u = new URL(request.url).searchParams.get("u") ?? "";
    const valid = USERNAME.test(u);
    const available = valid && u !== "taken-user";
    return HttpResponse.json({ valid, available, suggestion: available ? "" : "taken-user2" });
  }),

  http.post("*/api/v1/me/onboarding", async ({ request }) => {
    const { username } = (await request.json()) as { username: string };
    if (username === "taken-user")
      return HttpResponse.json({ error: "username taken" }, { status: 409 });
    return HttpResponse.json({ id: "user-jdoe", username });
  }),

  http.get("*/api/v1/me/onboarding/status", () =>
    HttpResponse.json({
      steps: [
        { key: "registry", done_at: new Date(Date.now() - 9_000).toISOString() },
        {
          key: "cluster_account",
          name: "hpc",
          done_at: new Date(Date.now() - 8_000).toISOString(),
        },
        { key: "project_group", name: "CIS250123" },
      ],
    }),
  ),

  http.post("*/api/v1/me/agreements/:id/accept", ({ params }) =>
    HttpResponse.json({ agreement_id: params["id"], accepted_at: new Date().toISOString() }),
  ),

  http.get("*/api/v1/agreements", () =>
    HttpResponse.json(agreements.map(({ created_by, published_by, removed_by, ...rest }) => rest)),
  ),

  http.get("*/api/v1/agreements/:id", ({ params }) => {
    const found = find(params["id"]);
    return found ? HttpResponse.json(found) : notFound();
  }),

  http.get("*/api/v1/agreements/:id/acceptances", ({ params, request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get("status") ?? "accepted";
    const limit = Number(url.searchParams.get("limit") ?? 50);
    const offset = Number(url.searchParams.get("offset") ?? 0);
    const report = reports[String(params["id"])] ?? { accepted: [], pending: [] };
    return HttpResponse.json({
      status,
      accepted: status === "accepted" ? report.accepted.slice(offset, offset + limit) : [],
      pending: status === "pending" ? report.pending.slice(offset, offset + limit) : [],
    });
  }),

  http.post("*/api/v1/agreements", async ({ request }) => {
    const body = (await request.json()) as AgreementDraft;
    if (versionTaken(body.version_label)) return conflict();
    const created: AgreementDetail = {
      ...body,
      id: `agr-${Date.now()}`,
      status: "DRAFT",
      body_hash: `mock-hash-${Date.now()}`,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      created_by: admin,
      accepted_count: 0,
      pending_count: 0,
    };
    agreements.unshift(created);
    return HttpResponse.json(created, { status: 201 });
  }),

  http.put("*/api/v1/agreements/:id", async ({ params, request }) => {
    const found = find(params["id"]);
    if (!found) return notFound();
    const body = (await request.json()) as AgreementDraft;
    if (versionTaken(body.version_label, found.id)) return conflict();
    Object.assign(found, body);
    return HttpResponse.json(found);
  }),

  http.post("*/api/v1/agreements/:id/publish", ({ params }) => {
    const found = find(params["id"]);
    if (!found) return notFound();
    Object.assign(found, {
      status: "PUBLISHED",
      published_at: new Date().toISOString(),
      published_by: admin,
    });
    return HttpResponse.json(found);
  }),

  http.post("*/api/v1/agreements/:id/remove", ({ params }) => {
    const found = find(params["id"]);
    if (!found) return notFound();
    Object.assign(found, {
      status: "REMOVED",
      removed_at: new Date().toISOString(),
      removed_by: admin,
    });
    return HttpResponse.json(found);
  }),

  http.delete("*/api/v1/agreements/:id", ({ params }) => {
    const found = find(params["id"]);
    if (!found) return notFound();
    agreements.splice(agreements.indexOf(found), 1);
    return new HttpResponse(null, { status: 204 });
  }),
];
