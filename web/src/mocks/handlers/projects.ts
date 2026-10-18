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
import { z } from "zod";
import {
  type ProjectMember,
  addProjectMemberPayloadSchema,
  createProjectPayloadSchema,
  projectMemberSchema,
  updateProjectMemberPayloadSchema,
  updateProjectStatusPayloadSchema,
} from "@/features/core/projects/schemas";
import projectsFixture from "@/features/core/projects/__fixtures__/projects.json";
import membersFixture from "@/features/core/projects/__fixtures__/members.json";
import { type ProjectResponse, zProjectResponse } from "@/generated/core/zod.gen";

const projects: ProjectResponse[] = z.array(zProjectResponse).parse(projectsFixture);
const membersByProject: Record<string, ProjectMember[]> = z
  .record(z.string(), z.array(projectMemberSchema))
  .parse(membersFixture);

let nextMemberSeq = 1000;

function ensureMembersBucket(projectId: string): ProjectMember[] {
  if (!membersByProject[projectId]) membersByProject[projectId] = [];
  return membersByProject[projectId];
}

function filterProjects(url: URL): ProjectResponse[] {
  const q = url.searchParams.get("q")?.toLowerCase() ?? "";
  const status = url.searchParams.get("status");
  const piId = url.searchParams.get("pi_id");
  return projects.filter((p) => {
    if (status && p.status !== status) return false;
    if (piId && p.project_pi_id !== piId) return false;
    if (q) {
      const hay = `${p.title} ${p.originated_id}`.toLowerCase();
      if (!hay.includes(q)) return false;
    }
    return true;
  });
}

export const projectsHandlers = [
  http.get("*/api/v1/projects", ({ request }) => {
    const url = new URL(request.url);
    const matched = filterProjects(url);
    const limit = Number(url.searchParams.get("limit") ?? matched.length);
    const offset = Number(url.searchParams.get("offset") ?? 0);
    const items = matched.slice(offset, offset + limit);
    return HttpResponse.json({ items, total: matched.length });
  }),

  http.get<{ id: string }>("*/api/v1/projects/:id", ({ params }) => {
    const id = String(params.id);
    const found = projects.find((p) => p.id === id);
    if (!found) return HttpResponse.json({ error: "project not found" }, { status: 404 });
    return HttpResponse.json(found);
  }),

  http.post("*/api/v1/projects", async ({ request }) => {
    const payload = createProjectPayloadSchema.parse(await request.json());
    const project: ProjectResponse = {
      id: `project-${Date.now()}`,
      originated_id: payload.originated_id ?? "",
      title: payload.title,
      origination: payload.origination,
      project_pi_id: payload.project_pi_id,
      status: "ACTIVE",
      created_time: new Date().toISOString(),
    };
    projects.unshift(project);
    return HttpResponse.json(project, { status: 201 });
  }),

  http.put<{ id: string }>("*/api/v1/projects/:id/status", async ({ params, request }) => {
    const id = String(params.id);
    const body = updateProjectStatusPayloadSchema.parse(await request.json());
    const existing = projects.find((p) => p.id === id);
    if (!existing) return HttpResponse.json({ error: "project not found" }, { status: 404 });
    existing.status = body.status;
    return HttpResponse.json(existing);
  }),

  http.get<{ id: string }>("*/api/v1/projects/:id/members", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(membersByProject[id] ?? []);
  }),

  http.post<{ id: string }>("*/api/v1/projects/:id/members", async ({ params, request }) => {
    const id = String(params.id);
    const body = addProjectMemberPayloadSchema.parse(await request.json());
    const bucket = ensureMembersBucket(id);
    const member: ProjectMember = {
      id: `pm-${nextMemberSeq++}`,
      project_id: id,
      user_id: body.user_id,
      email: `${body.user_id}@custos.local`,
      display_name: body.user_id,
      role: body.role,
      status: "ACTIVE",
      added_time: new Date().toISOString(),
      allocations: [],
    };
    bucket.push(member);
    return HttpResponse.json(member, { status: 201 });
  }),

  http.put<{ id: string; memberId: string }>("*/api/v1/projects/:id/members/:memberId", async ({ params, request }) => {
    const id = String(params.id);
    const memberId = String(params.memberId);
    const patch = updateProjectMemberPayloadSchema.parse(await request.json());
    const bucket = membersByProject[id] ?? [];
    const existing = bucket.find((m) => m.id === memberId);
    if (!existing) return HttpResponse.json({ error: "member not found" }, { status: 404 });
    if (patch.role) existing.role = patch.role;
    return HttpResponse.json(existing);
  }),

  http.delete<{ id: string; memberId: string }>("*/api/v1/projects/:id/members/:memberId", ({ params }) => {
    const id = String(params.id);
    const memberId = String(params.memberId);
    const bucket = membersByProject[id] ?? [];
    const idx = bucket.findIndex((m) => m.id === memberId);
    if (idx === -1) return HttpResponse.json({ error: "member not found" }, { status: 404 });
    bucket.splice(idx, 1);
    return new HttpResponse(null, { status: 204 });
  }),
];
