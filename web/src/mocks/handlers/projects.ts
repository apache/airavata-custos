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
import membersFixture from "@/features/core/projects/__fixtures__/members.json";
import projectsFixture from "@/features/core/projects/__fixtures__/projects.json";
import type {
  DeleteProjectsByIdData,
  GetProjectsByIdData,
  GetProjectsByIdMembersData,
  PostProjectsData,
  ProjectMemberResponse,
  ProjectResponse,
  PutProjectsByIdData,
  PutProjectsByIdMembersByUserIdData,
  PutProjectsByIdStatusData,
} from "@/generated/core/types.gen";
import { zProjectMemberResponse, zProjectResponse } from "@/generated/core/zod.gen";
import { z } from "zod";
import { notFound, page } from "../paging";

const projects: ProjectResponse[] = z.array(zProjectResponse).parse(projectsFixture);
const membersByProject: Record<string, ProjectMemberResponse[]> = z
  .record(z.string(), z.array(zProjectMemberResponse))
  .parse(membersFixture);

// Mirrors the backend: MEMBER drops the tag, keeping rows still backed by allocations.
function setTag(projectId: string, userId: string, role: string) {
  membersByProject[projectId] ??= [];
  const bucket = membersByProject[projectId];
  const existing = bucket.find((m) => m.user_id === userId);
  if (existing && (role !== "MEMBER" || existing.allocations.length)) existing.role = role;
  else if (existing) bucket.splice(bucket.indexOf(existing), 1);
  else if (role !== "MEMBER") {
    bucket.push({
      id: userId,
      project_id: projectId,
      user_id: userId,
      added_time: new Date().toISOString(),
      email: `${userId}@custos.local`,
      display_name: userId,
      role,
      status: "ACTIVE",
      allocations: [],
    });
  }
}

export function mockProjectRole(projectId: string, userId: string): string {
  return membersByProject[projectId]?.find((m) => m.user_id === userId)?.role ?? "MEMBER";
}

export function rejectUnlessCustos(projectId: string | undefined) {
  if (projects.find((p) => p.id === projectId)?.origination === "internal") return undefined;
  return HttpResponse.json({ error: "project is managed externally" }, { status: 409 });
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
    const { items, total } = page(url, filterProjects(url));
    return HttpResponse.json({ items, total });
  }),

  http.get<GetProjectsByIdData["path"]>("*/api/v1/projects/:id", ({ params }) => {
    const found = projects.find((p) => p.id === params.id);
    return found ? HttpResponse.json(found) : notFound("project");
  }),

  http.post<never, PostProjectsData["body"]>("*/api/v1/projects", async ({ request }) => {
    const { title = "", project_pi_id = "" } = await request.json();
    const project: ProjectResponse = {
      id: `project-${Date.now()}`,
      originated_id: "",
      title,
      origination: "internal",
      project_pi_id,
      status: "ACTIVE",
      created_time: new Date().toISOString(),
    };
    projects.unshift(project);
    setTag(project.id, project_pi_id, "PI");
    return HttpResponse.json(project, { status: 201 });
  }),

  http.put<PutProjectsByIdStatusData["path"], PutProjectsByIdStatusData["body"]>(
    "*/api/v1/projects/:id/status",
    async ({ params, request }) => {
      const existing = projects.find((p) => p.id === params.id);
      if (!existing) return notFound("project");
      const rejected = rejectUnlessCustos(params.id);
      if (rejected) return rejected;
      existing.status = (await request.json()).status ?? existing.status;
      return HttpResponse.json(existing);
    },
  ),

  http.put<PutProjectsByIdData["path"], PutProjectsByIdData["body"]>(
    "*/api/v1/projects/:id",
    async ({ params, request }) => {
      const body = await request.json();
      const existing = projects.find((p) => p.id === params.id);
      if (!existing) return notFound("project");
      const rejected = rejectUnlessCustos(params.id);
      if (rejected) return rejected;
      if (body.project_pi_id && body.project_pi_id !== existing.project_pi_id) {
        setTag(params.id, existing.project_pi_id, "MEMBER");
        setTag(params.id, body.project_pi_id, "PI");
      }
      Object.assign(existing, body);
      return HttpResponse.json(existing);
    },
  ),

  http.delete<DeleteProjectsByIdData["path"]>("*/api/v1/projects/:id", ({ params }) => {
    const idx = projects.findIndex((p) => p.id === params.id);
    if (idx === -1) return notFound("project");
    const rejected = rejectUnlessCustos(params.id);
    if (rejected) return rejected;
    if (allocationsFixture.some((a) => a.project_id === params.id)) {
      return HttpResponse.json({ error: "project still has allocations" }, { status: 400 });
    }
    projects.splice(idx, 1);
    return new HttpResponse(null, { status: 204 });
  }),

  http.get<GetProjectsByIdMembersData["path"]>("*/api/v1/projects/:id/members", ({ params }) =>
    HttpResponse.json(membersByProject[params.id] ?? []),
  ),

  http.put<PutProjectsByIdMembersByUserIdData["path"], PutProjectsByIdMembersByUserIdData["body"]>(
    "*/api/v1/projects/:id/members/:userId",
    async ({ params: { id, userId }, request }) => {
      const { role = "MEMBER" } = await request.json();
      const rejected = rejectUnlessCustos(id);
      if (rejected) return rejected;
      if (role === "PI" || mockProjectRole(id, userId) === "PI") {
        return HttpResponse.json({ error: "project PI cannot be changed" }, { status: 409 });
      }
      setTag(id, userId, role);
      return new HttpResponse(null, { status: 204 });
    },
  ),
];
