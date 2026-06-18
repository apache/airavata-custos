import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import auditEventsFixture from "@/features/core/audit/__fixtures__/audit-events.json";
import failedFixture from "@/features/core/audit/__fixtures__/trace.amie.failed.json";
import tracesListFixture from "@/features/core/audit/__fixtures__/traces.list.json";
import {
  RetryApiError,
  getAuditEventsForTrace,
  getTrace,
  listAuditSources,
  listTraces,
  retryTrace,
} from "@/features/core/audit/api";
import { ApiError } from "@/shared/api/client";

const fetchMock = vi.fn();

beforeEach(() => {
  vi.stubGlobal("fetch", fetchMock);
});

afterEach(() => {
  vi.unstubAllGlobals();
  fetchMock.mockReset();
});

function jsonResponse(body: unknown, init: ResponseInit = {}): Response {
  return new Response(JSON.stringify(body), {
    status: init.status ?? 200,
    headers: { "content-type": "application/json", ...(init.headers ?? {}) },
  });
}

const traceId = "a3b1c92d3f4e5a6b7c8d9e0f12345678";

describe("audit api", () => {
  it("listTraces serializes filters as repeated params on /audit/traces", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(tracesListFixture));
    await listTraces({
      source: ["amie", "http"],
      status: [0, 1],
      from: "2026-05-01T00:00:00Z",
      to: "2026-06-01T00:00:00Z",
      q: "alice",
      limit: 50,
      offset: 0,
    });
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).toMatch(/\/audit\/traces\?/);
    expect(url).toContain("source=amie");
    expect(url).toContain("source=http");
    expect(url).toContain("status=ok");
    expect(url).toContain("status=error");
    expect(url).toContain("q=alice");
    expect(url).toContain("limit=50");
  });

  it("listTraces sorts multi-value source filters for cache-key stability", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(tracesListFixture));
    await listTraces({ source: ["http", "amie"], status: [1, 0] });
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url.indexOf("source=amie")).toBeLessThan(url.indexOf("source=http"));
  });

  it("listTraces adapts backend fields onto the UI shape", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(tracesListFixture));
    const out = await listTraces();
    expect(out.traces.length).toBeGreaterThan(0);
    const first = out.traces[0];
    expect(first?.root_name).toBe("amie.process_event:request_account_create");
    expect(first?.span_count).toBe(14);
    expect(first?.status).toBe(1);
  });

  it("listTraces marks zero-time ended_at as null (in-progress override)", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(tracesListFixture));
    const out = await listTraces();
    const inProgress = out.traces.find((t) => t.ended_at == null);
    expect(inProgress).toBeDefined();
  });

  it("listTraces rejects payloads that fail schema validation", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse({ traces: [{ trace_id: "not-hex" }], total: 0, limit: 0, offset: 0 }),
    );
    await expect(listTraces()).rejects.toBeInstanceOf(Error);
  });

  it("getTrace hits the detail route and flattens the tree", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse(failedFixture));
    const detail = await getTrace(failedFixture.trace_id);
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).toContain(`/audit/traces/${failedFixture.trace_id}`);
    expect(detail.spans.length).toBeGreaterThan(1);
    expect(detail.trace.trace_id).toBe(failedFixture.trace_id);
  });

  it("listAuditSources hits /audit/sources and returns the array", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse({ sources: ["amie", "core"] }),
    );
    const out = await listAuditSources();
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).toContain("/audit/sources");
    expect(out).toEqual(["amie", "core"]);
  });

  it("retryTrace POSTs to /audit/traces/{id}/retry and resolves on 202", async () => {
    fetchMock.mockResolvedValueOnce(new Response(null, { status: 202 }));
    await retryTrace(traceId);
    const init = fetchMock.mock.calls[0]?.[1] as RequestInit;
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(init.method).toBe("POST");
    expect(url).toContain(`/audit/traces/${traceId}/retry`);
  });

  it("retryTrace throws a typed RetryApiError on 409", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse({ error: "source not registered" }, { status: 409 }),
    );
    const err = await retryTrace(traceId).catch((e) => e);
    expect(err).toBeInstanceOf(RetryApiError);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as RetryApiError).status).toBe(409);
    expect((err as RetryApiError).body.error).toBe("source not registered");
  });

  it("retryTrace falls back to bare ApiError when the body is not the envelope", async () => {
    fetchMock.mockResolvedValueOnce(new Response("<html>500</html>", { status: 500 }));
    const err = await retryTrace(traceId).catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err).not.toBeInstanceOf(RetryApiError);
  });

  it("getAuditEventsForTrace passes trace_id and span_id query params", async () => {
    const bucket = (auditEventsFixture as Record<string, { events: unknown[] }>)[traceId];
    fetchMock.mockResolvedValueOnce(jsonResponse({ events: bucket?.events ?? [] }));
    await getAuditEventsForTrace(traceId, "1000000000000006");
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).toContain("/audit/events");
    expect(url).toContain(`trace_id=${traceId}`);
    expect(url).toContain("span_id=1000000000000006");
  });
});
