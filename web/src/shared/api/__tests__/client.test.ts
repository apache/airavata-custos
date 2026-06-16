import { apiFetch } from "@/shared/api/client";
import { getLastTraceId, recordTraceId } from "@/shared/api/last-trace-id";
import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest";

const TRACE_ID = "a3b1c92d3f4e5a6b7c8d9e0f12345678";

const baseHandlers = [
  http.get("*/api/v1/healthz", () =>
    HttpResponse.json({ status: "ok" }, { headers: { "x-trace-id": TRACE_ID } }),
  ),
  http.get("*/api/v1/example", () => HttpResponse.json({ ok: true })),
];

const server = setupServer(...baseHandlers);

beforeAll(() => server.listen({ onUnhandledRequest: "error" }));
afterAll(() => server.close());
afterEach(() => server.resetHandlers(...baseHandlers));
beforeEach(() => recordTraceId(null));

describe("apiFetch", () => {
  it("prepends /api/v1 and returns the JSON body", async () => {
    const body = await apiFetch<{ status: string }>("/healthz");
    expect(body).toEqual({ status: "ok" });
  });

  it("captures X-Trace-Id into the singleton when the header is present", async () => {
    await apiFetch("/healthz");
    expect(getLastTraceId()).toBe(TRACE_ID);
  });

  it("leaves the singleton untouched when the header is absent", async () => {
    recordTraceId("previous-id");
    await apiFetch("/example");
    expect(getLastTraceId()).toBe("previous-id");
  });
});
