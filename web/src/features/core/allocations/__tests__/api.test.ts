import { afterEach, describe, expect, it, vi } from "vitest";
import {
  addMember,
  approveChangeRequest,
  getAllocation,
  getChangeRequest,
  listAllocationMembers,
  listAllocationResources,
  listAllocations,
  listChangeRequestEvents,
  listChangeRequests,
  rejectChangeRequest,
  removeMember,
  submitChangeRequest,
  updateMember,
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

const allocation = {
  id: "alloc-001",
  project_id: "project-001",
  name: "Test",
  status: "ACTIVE" as const,
  compute_cluster_id: "cluster-001",
  initial_su_amount: 100000,
  start_time: "2026-04-01T00:00:00Z",
  end_time: "2027-03-31T00:00:00Z",
};

const member = {
  id: "mem-001",
  compute_allocation_id: "alloc-001",
  user_id: "user-002",
  start_time: "2026-04-01T00:00:00Z",
  end_time: "2027-03-31T00:00:00Z",
  membership_status: "ACTIVE" as const,
};

const changeRequest = {
  id: "cr-001",
  compute_allocation_id: "alloc-001",
  requested_su_amount: 150000,
  requested_status: "ACTIVE" as const,
  reason: "more SUs",
  change_status: "PENDING" as const,
  requester_id: "user-pi-001",
  timestamp: "2026-06-01T10:00:00Z",
};

describe("listAllocations", () => {
  it("round-trips filters and returns the envelope", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, { items: [allocation], total: 1 }));
    const result = await listAllocations({
      limit: 20,
      offset: 0,
      project_id: "project-001",
      status: "ACTIVE",
      q: "test",
    });
    expect(result.total).toBe(1);
    expect(result.items).toHaveLength(1);
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).toContain("limit=20");
    expect(url).toContain("project_id=project-001");
    expect(url).toContain("status=ACTIVE");
    expect(url).toContain("q=test");
  });
});

describe("getAllocation", () => {
  it("validates the response", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, allocation));
    const out = await getAllocation("alloc-001");
    expect(out.id).toBe("alloc-001");
  });
});

describe("listAllocationResources", () => {
  it("returns [] when backend returns null", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, null));
    const out = await listAllocationResources("alloc-001");
    expect(out).toEqual([]);
  });
});

describe("listAllocationMembers", () => {
  it("validates rows", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, [member]));
    const out = await listAllocationMembers("alloc-001");
    expect(out).toHaveLength(1);
    expect(out[0]?.id).toBe("mem-001");
  });
});

describe("addMember", () => {
  it("POSTs to /compute-allocation-memberships", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(201, member));
    const out = await addMember({
      compute_allocation_id: "alloc-001",
      user_id: "user-002",
      start_time: "2026-04-01T00:00:00Z",
      end_time: "2027-03-31T00:00:00Z",
    });
    expect(out.id).toBe("mem-001");
    expect(fetchMock.mock.calls[0]?.[0]).toContain("/compute-allocation-memberships");
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe("POST");
  });
});

describe("updateMember", () => {
  it("PUTs to /compute-allocation-memberships/{id}", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, { ...member, membership_status: "INACTIVE" }));
    const out = await updateMember("mem-001", { membership_status: "INACTIVE" });
    expect(out.membership_status).toBe("INACTIVE");
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe("PUT");
  });
});

describe("removeMember", () => {
  it("DELETEs the membership", async () => {
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 204 }));
    await removeMember("mem-001");
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe("DELETE");
  });
});

describe("listChangeRequests", () => {
  it("scopes to allocation when allocation_id present", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, [changeRequest]));
    const out = await listChangeRequests({ allocation_id: "alloc-001" });
    expect(out).toHaveLength(1);
    expect(fetchMock.mock.calls[0]?.[0]).toContain("/compute-allocations/alloc-001/change-requests");
  });

  it("scopes to requester when requester_id present", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, [changeRequest]));
    await listChangeRequests({ requester_id: "user-pi-001" });
    expect(fetchMock.mock.calls[0]?.[0]).toContain("/users/user-pi-001/change-requests");
  });

  it("falls back to the generic list endpoint", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, [changeRequest]));
    await listChangeRequests({ status: "PENDING" });
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).toContain("/compute-allocation-change-requests");
    expect(url).toContain("status=PENDING");
  });
});

describe("getChangeRequest", () => {
  it("validates the response", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, changeRequest));
    const out = await getChangeRequest("cr-001");
    expect(out.id).toBe("cr-001");
  });
});

describe("submitChangeRequest", () => {
  it("POSTs the validated payload", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(201, changeRequest));
    const out = await submitChangeRequest({
      compute_allocation_id: "alloc-001",
      requested_su_amount: 150000,
      requested_status: "ACTIVE",
      reason: "more SUs",
      requester_id: "user-pi-001",
    });
    expect(out.id).toBe("cr-001");
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe("POST");
  });
});

describe("approveChangeRequest / rejectChangeRequest", () => {
  it("PUTs APPROVED for approve", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(200, { ...changeRequest, change_status: "APPROVED", approver_id: "u-admin" }),
    );
    const out = await approveChangeRequest("cr-001", "u-admin");
    expect(out.change_status).toBe("APPROVED");
    expect(fetchMock.mock.calls[0]?.[1]?.method).toBe("PUT");
  });
  it("PUTs REJECTED for reject", async () => {
    fetchMock.mockResolvedValueOnce(
      mockResponse(200, { ...changeRequest, change_status: "REJECTED", approver_id: "u-admin" }),
    );
    const out = await rejectChangeRequest("cr-001", "u-admin", "not eligible");
    expect(out.change_status).toBe("REJECTED");
  });
});

describe("listChangeRequestEvents", () => {
  it("returns [] when backend returns null", async () => {
    fetchMock.mockResolvedValueOnce(mockResponse(200, null));
    const out = await listChangeRequestEvents("cr-001");
    expect(out).toEqual([]);
  });
});
