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
import { http, HttpResponse } from "msw";
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
  if (existing && (role !== "MEMBER" || existing.allocations?.length)) existing.role = role;
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

// The bare Project master's writes answer with, without the joined PI fields.
const bare = ({ project_pi_display_name, project_pi_email, ...project }: ProjectResponse) =>
  project;

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
    const id = crypto.randomUUID();
    const project: ProjectResponse = {
      id,
      originated_id: "",
      title,
      origination: "internal",
      project_pi_id,
      status: "ACTIVE",
      created_time: new Date().toISOString(),
    };
    projects.unshift(project);
    setTag(id, project_pi_id, "PI");
    return HttpResponse.json(project, { status: 201 });
  }),

  http.put<PutProjectsByIdStatusData["path"], PutProjectsByIdStatusData["body"]>(
    "*/api/v1/projects/:id/status",
    async ({ params, request }) => {
      const existing = projects.find((p) => p.id === params.id);
      if (!existing) return notFound("project");
      existing.status = (await request.json()).status ?? existing.status;
      return HttpResponse.json(bare(existing));
    },
  ),

  http.put<PutProjectsByIdData["path"], PutProjectsByIdData["body"]>(
    "*/api/v1/projects/:id",
    async ({ params, request }) => {
      const { title, project_pi_id, previous_pi_role = "" } = await request.json();
      const existing = projects.find((p) => p.id === params.id);
      if (!existing) return notFound("project");
      const previousPi = existing.project_pi_id;
      if (project_pi_id && project_pi_id !== previousPi) {
        if (!["CO_PI", "ALLOCATION_MANAGER", "MEMBER"].includes(previous_pi_role)) {
          return HttpResponse.json(
            {
              error: "invalid input: previous_pi_role must be CO_PI, ALLOCATION_MANAGER or MEMBER",
            },
            { status: 400 },
          );
        }
        setTag(params.id, previousPi, previous_pi_role);
        setTag(params.id, project_pi_id, "PI");
        existing.project_pi_id = project_pi_id;
      }
      // Blank fields keep the stored value.
      existing.title = title || existing.title;
      return HttpResponse.json(bare(existing));
    },
  ),

  http.delete<DeleteProjectsByIdData["path"]>("*/api/v1/projects/:id", ({ params }) => {
    const idx = projects.findIndex((p) => p.id === params.id);
    if (idx === -1) return notFound("project");
    if (allocationsFixture.some((a) => a.project_id === params.id)) {
      return HttpResponse.json({ error: "project still has allocations" }, { status: 409 });
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
      const { role } = await request.json();
      if (!role) return HttpResponse.json({ error: "role is required" }, { status: 400 });
      if (mockProjectRole(id, userId) === "PI") return new HttpResponse(null, { status: 204 });
      if (role === "PI") {
        return HttpResponse.json({ error: "project PI cannot be changed" }, { status: 409 });
      }
      setTag(id, userId, role);
      return new HttpResponse(null, { status: 204 });
    },
  ),
];
