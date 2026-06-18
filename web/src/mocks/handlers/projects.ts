import { http, HttpResponse } from "msw";
import type {
  AddProjectMemberPayload,
  CreateProjectPayload,
  Project,
  ProjectMember,
  UpdateProjectMemberPayload,
  UpdateProjectStatusPayload,
} from "@/features/core/projects/schemas";
import projectsFixture from "@/features/core/projects/__fixtures__/projects.json";
import membersFixture from "@/features/core/projects/__fixtures__/members.json";

const projects: Project[] = (projectsFixture as Project[]).map((p) => ({ ...p }));
const membersByProject: Record<string, ProjectMember[]> = Object.fromEntries(
  Object.entries(membersFixture as Record<string, ProjectMember[]>).map(([id, rows]) => [
    id,
    rows.map((r) => ({ ...r })),
  ]),
);

let nextMemberSeq = 1000;

function ensureMembersBucket(projectId: string): ProjectMember[] {
  if (!membersByProject[projectId]) membersByProject[projectId] = [];
  return membersByProject[projectId];
}

function filterProjects(url: URL): Project[] {
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

  http.get("*/api/v1/projects/:id", ({ params }) => {
    const id = String(params.id);
    const found = projects.find((p) => p.id === id);
    if (!found) return HttpResponse.json({ error: "project not found" }, { status: 404 });
    return HttpResponse.json(found);
  }),

  http.post("*/api/v1/projects", async ({ request }) => {
    const payload = (await request.json()) as CreateProjectPayload;
    const project: Project = {
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

  http.put("*/api/v1/projects/:id/status", async ({ params, request }) => {
    const id = String(params.id);
    const body = (await request.json()) as UpdateProjectStatusPayload;
    const existing = projects.find((p) => p.id === id);
    if (!existing) return HttpResponse.json({ error: "project not found" }, { status: 404 });
    existing.status = body.status;
    return HttpResponse.json(existing);
  }),

  http.get("*/api/v1/projects/:id/members", ({ params }) => {
    const id = String(params.id);
    return HttpResponse.json(membersByProject[id] ?? []);
  }),

  http.post("*/api/v1/projects/:id/members", async ({ params, request }) => {
    const id = String(params.id);
    const body = (await request.json()) as AddProjectMemberPayload;
    const bucket = ensureMembersBucket(id);
    const member: ProjectMember = {
      id: `pm-${nextMemberSeq++}`,
      project_id: id,
      user_id: body.user_id,
      email: `${body.user_id}@custos.local`,
      display_name: body.user_id,
      role: body.role ?? "MEMBER",
      status: "ACTIVE",
      added_time: new Date().toISOString(),
      allocations: [],
    };
    bucket.push(member);
    return HttpResponse.json(member, { status: 201 });
  }),

  http.put("*/api/v1/projects/:id/members/:memberId", async ({ params, request }) => {
    const id = String(params.id);
    const memberId = String(params.memberId);
    const patch = (await request.json()) as UpdateProjectMemberPayload;
    const bucket = membersByProject[id] ?? [];
    const existing = bucket.find((m) => m.id === memberId);
    if (!existing) return HttpResponse.json({ error: "member not found" }, { status: 404 });
    if (patch.role) existing.role = patch.role;
    if (patch.status) existing.status = patch.status;
    return HttpResponse.json(existing);
  }),

  http.delete("*/api/v1/projects/:id/members/:memberId", ({ params }) => {
    const id = String(params.id);
    const memberId = String(params.memberId);
    const bucket = membersByProject[id] ?? [];
    const idx = bucket.findIndex((m) => m.id === memberId);
    if (idx === -1) return HttpResponse.json({ error: "member not found" }, { status: 404 });
    bucket.splice(idx, 1);
    return new HttpResponse(null, { status: 204 });
  }),
];
