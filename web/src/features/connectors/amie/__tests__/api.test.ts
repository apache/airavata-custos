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

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError } from "@/shared/api/client";
import {
  getPacket,
  getPacketEvents,
  getPacketStats,
  linkUnmapped,
  listPackets,
  listReplies,
  listUnmapped,
  resolvePacket,
  retryPacket,
  retryReply,
} from "../api";

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

describe("amie api", () => {
  it("listPackets serializes filters into the query string", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ packets: [], total: 0, limit: 10, offset: 0 }));
    await listPackets({
      status: "FAILED",
      type: "request_project_create",
      source: "access",
      q: "BIO130001",
      from: "2026-05-01",
      to: "2026-05-22",
      limit: 10,
      offset: 20,
    });
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).toContain("/api/v1/connectors/amie/packets");
    expect(url).toContain("status=FAILED");
    expect(url).toContain("type=request_project_create");
    expect(url).toContain("source=access");
    expect(url).toContain("q=BIO130001");
    expect(url).toContain("limit=10");
    expect(url).toContain("offset=20");
  });

  it("listPackets strips empty params", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ packets: [], total: 0, limit: 0, offset: 0 }));
    await listPackets({ q: "" });
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).not.toContain("q=");
  });

  it("getPacket parses the packet shape", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse({
        id: "pkt-001",
        amie_id: "1234567",
        type: "request_project_create",
        status: "PROCESSED",
        source: "access",
        received_at: "2026-05-01T00:00:00Z",
        updated_at: "2026-05-01T00:01:00Z",
        retries: 0,
      }),
    );
    const packet = await getPacket("pkt-001");
    expect(packet.id).toBe("pkt-001");
    expect(packet.status).toBe("PROCESSED");
  });

  it("getPacket throws ApiError on 404", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ error: "not_found" }, { status: 404 }));
    await expect(getPacket("missing")).rejects.toBeInstanceOf(ApiError);
  });

  it("retryPacket posts and parses the envelope", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse({
        queued: true,
        packet: {
          id: "pkt-001",
          amie_id: "1234567",
          type: "request_project_create",
          status: "DECODED",
          source: "access",
          received_at: "2026-05-01T00:00:00Z",
          updated_at: "2026-05-01T00:01:00Z",
          retries: 1,
        },
      }),
    );
    const res = await retryPacket("pkt-001");
    expect(res.queued).toBe(true);
    expect(res.packet.retries).toBe(1);
    const init = fetchMock.mock.calls[0]?.[1] as RequestInit;
    expect(init.method).toBe("POST");
  });

  it("resolvePacket posts the reason", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse({
        id: "pkt-001",
        amie_id: "1234567",
        type: "request_project_create",
        status: "PROCESSED",
        source: "access",
        received_at: "2026-05-01T00:00:00Z",
        updated_at: "2026-05-01T00:01:00Z",
        retries: 0,
      }),
    );
    await resolvePacket("pkt-001", { reason: "Manual resolve" });
    const init = fetchMock.mock.calls[0]?.[1] as RequestInit;
    expect(init.method).toBe("POST");
    expect(init.body).toContain("Manual resolve");
  });

  it("listReplies validates the response envelope", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ replies: [], total: 0, limit: 50, offset: 0 }));
    const out = await listReplies({ status: "FAILED", limit: 50 });
    expect(out.replies).toEqual([]);
    const url = fetchMock.mock.calls[0]?.[0] as string;
    expect(url).toContain("/api/v1/connectors/amie/replies");
  });

  it("retryReply parses the queued envelope", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ queued: true }));
    const res = await retryReply("rply-001");
    expect(res.queued).toBe(true);
  });

  it("listUnmapped hits the unmapped endpoint", async () => {
    fetchMock.mockResolvedValueOnce(jsonResponse({ packets: [], total: 0, limit: 0, offset: 0 }));
    await listUnmapped({ limit: 25 });
    expect(fetchMock.mock.calls[0]?.[0]).toContain("/api/v1/connectors/amie/unmapped");
  });

  it("linkUnmapped posts the entity ref", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse({
        id: "pkt-001",
        amie_id: "1234567",
        type: "request_project_create",
        status: "PROCESSED",
        source: "access",
        received_at: "2026-05-01T00:00:00Z",
        updated_at: "2026-05-01T00:01:00Z",
        retries: 0,
      }),
    );
    await linkUnmapped("pkt-001", { entity_type: "project", entity_id: "BIO130001" });
    const init = fetchMock.mock.calls[0]?.[1] as RequestInit;
    expect(init.body).toContain("BIO130001");
    expect(init.body).toContain("project");
  });

  it("getPacketStats parses the byDay array", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse({
        byDay: [
          { date: "2026-05-20", status: "PROCESSED", type: "request_project_create", count: 3 },
        ],
      }),
    );
    const stats = await getPacketStats({ window: "7d" });
    expect(stats.byDay).toHaveLength(1);
    expect(stats.byDay[0]?.count).toBe(3);
  });

  it("getPacketEvents parses an array including trace_id", async () => {
    fetchMock.mockResolvedValueOnce(
      jsonResponse([
        {
          id: "evt-1",
          packet_id: "pkt-001",
          event_type: "FAILED",
          actor: "amie-worker",
          status: "FAILED",
          timestamp: "2026-05-01T00:00:00Z",
          trace_id: "a3b1c92d3f4e5a6b7c8d9e0f12345678",
        },
      ]),
    );
    const events = await getPacketEvents("pkt-001");
    expect(events).toHaveLength(1);
    expect(events[0]?.trace_id).toBe("a3b1c92d3f4e5a6b7c8d9e0f12345678");
  });
});
