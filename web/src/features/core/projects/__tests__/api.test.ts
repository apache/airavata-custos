import { afterEach, describe, expect, it, vi } from "vitest";
import {
  addProjectMember,
  createProject,
  getProject,
  listProjectMembers,
  listProjects,
  removeProjectMember,
  updateProjectMember,
  updateProjectStatus,
} from "../api";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

function mockResponse(status: number, body: unknown): Response {
  const text = typeof body === "string" ? body : JSON.stringify(body);
  return new Response(text, { status, headers: { "content-type": "application/json" } });
}

afterEach(() => {
  fetchMock.mockReset();
});

const validProject = {
  id: "project-001",
  originated_id: "BIO130000",
  title: "Test Project",
  origination: "ACCESS",
  project_pi_id: "user-pi-001",
  status: "ACTIVE" as const,
  created_time: "2026-04-01T00:00:00.000Z",
};

const validMember = {
  id: "pm-001",
  project_id: "project-001",
  user_id: "user-002",
  email: "x@custos.local",
  display_name: "X Y",
  role: "MEMBER" as const,
  status: "ACTIVE" as const,
  added_time: "2026-04-05T00:00:00.000Z",
};

describe("listProjects", () => {
  it("round-trips filters and returns the envelope", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, { items: [validProject], total: 1 }));
    const result = await listProjects({
      limit: 20,
      offset: 0,
      pi_id: "user-pi-001",
      status: "ACTIVE",
      q: "test",
    });
    expect(result.total).toBe(1);
    expect(result.items).toHaveLength(1);
    const calledUrl = fetchMock.mock.calls[0]?.[0] as string;
    expect(calledUrl).toContain("limit=20");
    expect(calledUrl).toContain("pi_id=user-pi-001");
    expect(calledUrl).toContain("status=ACTIVE");
    expect(calledUrl).toContain("q=test");
  });

  it("rejects on schema mismatch (total missing)", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, { items: [validProject] }));
    await expect(listProjects()).rejects.toThrow();
  });
});

describe("getProject", () => {
  it("validates the response", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, validProject));
    const out = await getProject("project-001");
    expect(out.id).toBe("project-001");
  });
});

describe("createProject", () => {
  it("POSTs the validated payload", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, validProject));
    const out = await createProject({
      title: "New Project",
      origination: "ACCESS",
      project_pi_id: "user-pi-001",
    });
    expect(out.id).toBe("project-001");
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe("POST");
  });
});

describe("updateProjectStatus", () => {
  it("PUTs to /projects/{id}/status", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, { ...validProject, status: "INACTIVE" }));
    const out = await updateProjectStatus("project-001", { status: "INACTIVE" });
    expect(out.status).toBe("INACTIVE");
    const init = fetchMock.mock.calls[0]?.[1];
    expect(init?.method).toBe("PUT");
    expect(fetchMock.mock.calls[0]?.[0]).toContain("/projects/project-001/status");
  });
});

describe("listProjectMembers", () => {
  it("returns [] when backend returns null", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, null));
    const out = await listProjectMembers("project-001");
    expect(out).toEqual([]);
  });

  it("validates each row", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, [validMember]));
    const out = await listProjectMembers("project-001");
    expect(out).toHaveLength(1);
    expect(out[0]?.role).toBe("MEMBER");
  });
});

describe("addProjectMember", () => {
  it("POSTs to /projects/{id}/members", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(201, validMember));
    const out = await addProjectMember("project-001", { user_id: "user-002" });
    expect(out.id).toBe("pm-001");
    expect(fetchMock.mock.calls[0]?.[0]).toContain("/projects/project-001/members");
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe("POST");
  });
});

describe("updateProjectMember", () => {
  it("PUTs the patch", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, { ...validMember, role: "CO_PI" }));
    const out = await updateProjectMember("project-001", "pm-001", { role: "CO_PI" });
    expect(out.role).toBe("CO_PI");
    expect(fetchMock.mock.calls[0]?.[0]).toContain("/projects/project-001/members/pm-001");
  });
});

describe("removeProjectMember", () => {
  it("DELETEs the member", async () => {
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 204 }));
    await removeProjectMember("project-001", "pm-001");
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe("DELETE");
  });
});
