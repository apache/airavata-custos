import { afterEach, describe, expect, it, vi } from "vitest";
import { assignUserRole, listUsers, removeUserRole } from "../api";

const fetchMock = vi.fn();
vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);

function mockResponse(status: number, body?: unknown): Response {
  return new Response(body === undefined ? null : JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" },
  });
}

afterEach(() => fetchMock.mockReset());

describe("users API", () => {
  it("sends limit and offset when listing users", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(200, {
        items: [{ id: "user-1", email: "user@example.org" }],
        total: 1,
      }),
    );
    const result = await listUsers({ limit: 25, offset: 50 });
    expect(result.total).toBe(1);
    const url = String(fetchMock.mock.calls[0]?.[0]);
    expect(url).toContain("limit=25");
    expect(url).toContain("offset=50");
  });

  it("posts the role id when assigning a role", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(201, { user_id: "user-1", role_id: "role-1" }));
    await assignUserRole("user-1", "role-1");
    expect(fetchMock.mock.calls[0]?.[1]).toMatchObject({
      method: "POST",
      body: JSON.stringify({ role_id: "role-1" }),
    });
  });

  it("uses the assignment resource when removing a role", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(204));
    await removeUserRole("user-1", "role-1");
    expect(fetchMock.mock.calls[0]?.[0]).toContain("/users/user-1/roles/role-1");
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe("DELETE");
  });
});
