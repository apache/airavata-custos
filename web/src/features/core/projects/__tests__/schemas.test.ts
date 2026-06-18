import { describe, expect, it } from "vitest";
import {
  addProjectMemberPayloadSchema,
  createProjectPayloadSchema,
  projectListEnvelopeSchema,
  projectMemberSchema,
  projectSchema,
  updateProjectMemberPayloadSchema,
  updateProjectStatusPayloadSchema,
} from "../schemas";

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

describe("projectSchema", () => {
  it("accepts a backend-shaped project", () => {
    expect(projectSchema.safeParse(validProject).success).toBe(true);
  });
  it("rejects an invalid status", () => {
    expect(projectSchema.safeParse({ ...validProject, status: "FOO" }).success).toBe(false);
  });
});

describe("projectListEnvelopeSchema", () => {
  it("accepts a valid envelope", () => {
    expect(
      projectListEnvelopeSchema.safeParse({ items: [validProject], total: 1 }).success,
    ).toBe(true);
  });
  it("rejects a negative total", () => {
    expect(projectListEnvelopeSchema.safeParse({ items: [], total: -1 }).success).toBe(false);
  });
});

describe("projectMemberSchema", () => {
  it("accepts a valid member row", () => {
    expect(projectMemberSchema.safeParse(validMember).success).toBe(true);
  });
  it("rejects an invalid role", () => {
    expect(projectMemberSchema.safeParse({ ...validMember, role: "FOO" }).success).toBe(false);
  });
});

describe("createProjectPayloadSchema", () => {
  it("requires title 1+ chars", () => {
    expect(
      createProjectPayloadSchema.safeParse({
        title: "",
        origination: "ACCESS",
        project_pi_id: "user-pi-001",
      }).success,
    ).toBe(false);
  });
  it("accepts a minimal payload", () => {
    expect(
      createProjectPayloadSchema.safeParse({
        title: "New Project",
        origination: "ACCESS",
        project_pi_id: "user-pi-001",
      }).success,
    ).toBe(true);
  });
});

describe("updateProjectStatusPayloadSchema", () => {
  it("rejects an invalid status", () => {
    expect(updateProjectStatusPayloadSchema.safeParse({ status: "FOO" }).success).toBe(false);
  });
  it("accepts ACTIVE/INACTIVE/DELETED", () => {
    expect(updateProjectStatusPayloadSchema.safeParse({ status: "ACTIVE" }).success).toBe(true);
    expect(updateProjectStatusPayloadSchema.safeParse({ status: "DELETED" }).success).toBe(true);
  });
});

describe("addProjectMemberPayloadSchema", () => {
  it("defaults role to MEMBER", () => {
    const parsed = addProjectMemberPayloadSchema.parse({ user_id: "user-002" });
    expect(parsed.role).toBe("MEMBER");
  });
});

describe("updateProjectMemberPayloadSchema", () => {
  it("rejects an empty patch", () => {
    expect(updateProjectMemberPayloadSchema.safeParse({}).success).toBe(false);
  });
  it("accepts a role-only patch", () => {
    expect(updateProjectMemberPayloadSchema.safeParse({ role: "CO_PI" }).success).toBe(true);
  });
});
