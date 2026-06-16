import { describe, expect, it } from "vitest";
import {
  linkUnmappedPayloadSchema,
  packetEventSchema,
  packetListResponseSchema,
  packetSchema,
  packetStatsSchema,
  replyListResponseSchema,
  resolvePacketPayloadSchema,
} from "../schemas";

describe("amie schemas", () => {
  it("packetSchema accepts a minimal NEW packet", () => {
    const parsed = packetSchema.parse({
      id: "pkt-x",
      amie_id: "1",
      type: "request_project_create",
      status: "NEW",
      source: "access",
      received_at: "2026-06-08T00:00:00Z",
      updated_at: "2026-06-08T00:00:00Z",
      retries: 0,
    });
    expect(parsed.status).toBe("NEW");
  });

  it("packetSchema accepts an optional linked_entity", () => {
    const parsed = packetSchema.parse({
      id: "pkt-x",
      amie_id: "1",
      type: "request_project_create",
      status: "PROCESSED",
      source: "access",
      received_at: "2026-06-08T00:00:00Z",
      updated_at: "2026-06-08T00:00:00Z",
      retries: 0,
      linked_entity: { type: "project", id: "prj-001", display_id: "BIO130001" },
    });
    expect(parsed.linked_entity?.type).toBe("project");
  });

  it("packetSchema rejects unknown status", () => {
    expect(() =>
      packetSchema.parse({
        id: "pkt-x",
        amie_id: "1",
        type: "x",
        status: "WAT",
        source: "access",
        received_at: "2026-06-08T00:00:00Z",
        updated_at: "2026-06-08T00:00:00Z",
        retries: 0,
      }),
    ).toThrow();
  });

  it("packetEventSchema accepts an optional trace_id", () => {
    const parsed = packetEventSchema.parse({
      id: "evt-1",
      packet_id: "pkt-1",
      event_type: "FAILED",
      actor: "amie-worker",
      status: "FAILED",
      timestamp: "2026-06-08T00:00:00Z",
      trace_id: "a3b1c92d3f4e5a6b7c8d9e0f12345678",
    });
    expect(parsed.trace_id).toBe("a3b1c92d3f4e5a6b7c8d9e0f12345678");
  });

  it("packetEventSchema rejects a malformed trace_id", () => {
    expect(() =>
      packetEventSchema.parse({
        id: "evt-1",
        packet_id: "pkt-1",
        event_type: "FAILED",
        actor: "amie-worker",
        status: "FAILED",
        timestamp: "2026-06-08T00:00:00Z",
        trace_id: "not-hex",
      }),
    ).toThrow();
  });

  it("packetListResponseSchema parses an empty list envelope", () => {
    const parsed = packetListResponseSchema.parse({
      packets: [],
      total: 0,
      limit: 20,
      offset: 0,
    });
    expect(parsed.packets).toEqual([]);
  });

  it("replyListResponseSchema parses a single reply", () => {
    const parsed = replyListResponseSchema.parse({
      replies: [
        {
          id: "r1",
          amie_id: "9",
          type: "inform_transaction_complete",
          status: "SENT",
          retries: 0,
          created_at: "2026-06-08T00:00:00Z",
        },
      ],
      total: 1,
      limit: 200,
      offset: 0,
    });
    expect(parsed.replies[0]?.status).toBe("SENT");
  });

  it("packetStatsSchema parses a byDay array", () => {
    const parsed = packetStatsSchema.parse({
      byDay: [
        { date: "2026-06-08", status: "PROCESSED", type: "request_project_create", count: 2 },
      ],
    });
    expect(parsed.byDay).toHaveLength(1);
  });

  it("resolvePacketPayloadSchema rejects a too-short reason", () => {
    expect(() => resolvePacketPayloadSchema.parse({ reason: "ok" })).toThrow();
  });

  it("linkUnmappedPayloadSchema rejects an empty entity_id", () => {
    expect(() => linkUnmappedPayloadSchema.parse({ entity_type: "project", entity_id: "" })).toThrow();
  });
});
