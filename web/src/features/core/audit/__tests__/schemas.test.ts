import { describe, expect, it } from "vitest";
import failedFixture from "@/features/core/audit/__fixtures__/trace.amie.failed.json";
import successFixture from "@/features/core/audit/__fixtures__/trace.amie.success.json";
import httpFixture from "@/features/core/audit/__fixtures__/trace.http.json";
import inProgressFixture from "@/features/core/audit/__fixtures__/trace.in-progress.json";
import tracesListFixture from "@/features/core/audit/__fixtures__/traces.list.json";
import {
  spanSchema,
  traceDetailWireSchema,
  traceListWireSchema,
  traceSchema,
} from "@/features/core/audit/schemas";

describe("audit schemas — backend wire fixtures", () => {
  it("parses the list fixture (TraceSummary envelope)", () => {
    expect(() => traceListWireSchema.parse(tracesListFixture)).not.toThrow();
  });

  it("parses the AMIE success detail fixture (tree)", () => {
    expect(() => traceDetailWireSchema.parse(successFixture)).not.toThrow();
  });

  it("parses the AMIE failed detail fixture and surfaces the leaf error node", () => {
    const parsed = traceDetailWireSchema.parse(failedFixture);
    const errors = parsed.tree[0]?.children.filter((c) => c.status === "error") ?? [];
    expect(errors.length).toBeGreaterThan(0);
  });

  it("parses the HTTP detail fixture", () => {
    expect(() => traceDetailWireSchema.parse(httpFixture)).not.toThrow();
  });

  it("parses the in-progress fixture", () => {
    const parsed = traceDetailWireSchema.parse(inProgressFixture);
    expect(parsed.tree[0]?.status).toBe("in_progress");
  });
});

describe("audit schemas — UI null/optional tolerance", () => {
  const baseTrace = {
    trace_id: "a3b1c92d3f4e5a6b7c8d9e0f12345678",
    root_name: "amie.process_event:request_account_create",
    source: "amie" as const,
    status: 0,
    started_at: "2026-06-03T14:21:32.412Z",
    span_count: 1,
  };

  it("accepts explicit null root_event", () => {
    expect(() => traceSchema.parse({ ...baseTrace, root_event: null })).not.toThrow();
  });

  it("accepts omitted root_event", () => {
    expect(() => traceSchema.parse(baseTrace)).not.toThrow();
  });

  it("accepts explicit null ended_at and missing ended_at", () => {
    expect(() => traceSchema.parse({ ...baseTrace, ended_at: null })).not.toThrow();
    expect(() => traceSchema.parse(baseTrace)).not.toThrow();
  });

  const baseSpan = {
    span_id: "0000000000000001",
    name: "amie.process_event",
    kind: 0,
    status: 0,
    start_time: "2026-06-03T14:21:32.412Z",
  };

  it("accepts null/missing attributes on a span", () => {
    expect(() => spanSchema.parse({ ...baseSpan, attributes: null })).not.toThrow();
    expect(() => spanSchema.parse(baseSpan)).not.toThrow();
  });

  it("accepts null/missing end_time and status_message on a span", () => {
    expect(() =>
      spanSchema.parse({ ...baseSpan, end_time: null, status_message: null }),
    ).not.toThrow();
    expect(() => spanSchema.parse(baseSpan)).not.toThrow();
  });

  it("treats parent_span_id as optional but rejects explicit null (backend omits, never nulls)", () => {
    expect(() =>
      spanSchema.parse({ ...baseSpan, parent_span_id: "0000000000000002" }),
    ).not.toThrow();
    expect(() => spanSchema.parse(baseSpan)).not.toThrow();
    const result = spanSchema.safeParse({ ...baseSpan, parent_span_id: null });
    expect(result.success).toBe(false);
  });
});
