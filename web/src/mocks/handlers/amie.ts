import { http, HttpResponse } from "msw";
import eventsFixture from "@/features/connectors/amie/__fixtures__/events.json";
import packetsFixture from "@/features/connectors/amie/__fixtures__/packets.json";
import repliesFixture from "@/features/connectors/amie/__fixtures__/replies.json";
import statsFixture from "@/features/connectors/amie/__fixtures__/stats.json";
import type {
  Packet,
  PacketEvent,
  PacketListResponse,
  Reply,
  ReplyListResponse,
} from "@/features/connectors/amie/types";

const allPackets: Packet[] = ((packetsFixture as PacketListResponse).packets as Packet[]).map(
  (p) => ({ ...p }),
);
const eventsByPacket: Record<string, PacketEvent[]> = Object.fromEntries(
  Object.entries(eventsFixture as Record<string, PacketEvent[]>).map(([id, rows]) => [
    id,
    rows.map((r) => ({ ...r })),
  ]),
);
const allReplies: Reply[] = ((repliesFixture as ReplyListResponse).replies as Reply[]).map((r) => ({
  ...r,
}));

function filterPackets(url: URL): {
  rows: Packet[];
  total: number;
  limit: number;
  offset: number;
} {
  const status = url.searchParams.get("status");
  const type = url.searchParams.get("type");
  const source = url.searchParams.get("source");
  const q = url.searchParams.get("q")?.toLowerCase() ?? "";
  const limit = Number(url.searchParams.get("limit") ?? "100");
  const offset = Number(url.searchParams.get("offset") ?? "0");

  let filtered = allPackets;
  if (status) filtered = filtered.filter((p) => p.status === status);
  if (type) filtered = filtered.filter((p) => p.type === type);
  if (source) filtered = filtered.filter((p) => p.source === source);
  if (q) {
    filtered = filtered.filter((p) => {
      const hay = `${p.amie_id} ${p.id} ${p.linked_entity?.display_id ?? ""}`.toLowerCase();
      return hay.includes(q);
    });
  }
  const total = filtered.length;
  const sliced = filtered.slice(offset, offset + limit);
  return { rows: sliced, total, limit, offset };
}

export const amieHandlers = [
  http.get("/api/v1/connectors/amie/packets", ({ request }) => {
    const url = new URL(request.url);
    const { rows, total, limit, offset } = filterPackets(url);
    return HttpResponse.json({ packets: rows, total, limit, offset });
  }),

  http.get("/api/v1/connectors/amie/packets/:id", ({ params }) => {
    const id = String(params.id);
    const packet = allPackets.find((p) => p.id === id);
    if (!packet) return HttpResponse.json({ error: "not_found" }, { status: 404 });
    return HttpResponse.json(packet);
  }),

  http.get("/api/v1/connectors/amie/packets/:id/events", ({ params }) => {
    const id = String(params.id);
    const rows = eventsByPacket[id] ?? [];
    return HttpResponse.json(rows);
  }),

  http.post("/api/v1/connectors/amie/packets/:id/retry", ({ params }) => {
    const id = String(params.id);
    const packet = allPackets.find((p) => p.id === id);
    if (!packet) return HttpResponse.json({ error: "not_found" }, { status: 404 });
    packet.retries += 1;
    packet.status = "DECODED";
    packet.updated_at = new Date().toISOString();
    return HttpResponse.json({ queued: true, packet });
  }),

  http.post("/api/v1/connectors/amie/packets/:id/resolve", async ({ params, request }) => {
    const id = String(params.id);
    const packet = allPackets.find((p) => p.id === id);
    if (!packet) return HttpResponse.json({ error: "not_found" }, { status: 404 });
    const body = (await request.json().catch(() => ({}))) as { reason?: string };
    if (!body.reason || body.reason.trim().length < 3) {
      return HttpResponse.json({ error: "reason_required" }, { status: 400 });
    }
    packet.status = "PROCESSED";
    packet.processed_at = new Date().toISOString();
    packet.updated_at = packet.processed_at;
    return HttpResponse.json(packet);
  }),

  http.get("/api/v1/connectors/amie/replies", ({ request }) => {
    const url = new URL(request.url);
    const status = url.searchParams.get("status");
    const limit = Number(url.searchParams.get("limit") ?? "200");
    const offset = Number(url.searchParams.get("offset") ?? "0");
    let filtered = allReplies;
    if (status) filtered = filtered.filter((r) => r.status === status);
    return HttpResponse.json({
      replies: filtered.slice(offset, offset + limit),
      total: filtered.length,
      limit,
      offset,
    });
  }),

  http.post("/api/v1/connectors/amie/replies/:id/retry", ({ params }) => {
    const id = String(params.id);
    const reply = allReplies.find((r) => r.id === id);
    if (!reply) return HttpResponse.json({ error: "not_found" }, { status: 404 });
    reply.retries += 1;
    reply.status = "PENDING";
    return HttpResponse.json({ queued: true });
  }),

  http.get("/api/v1/connectors/amie/unmapped", ({ request }) => {
    const url = new URL(request.url);
    const limit = Number(url.searchParams.get("limit") ?? "100");
    const offset = Number(url.searchParams.get("offset") ?? "0");
    const filtered = allPackets.filter((p) => p.status === "DECODED" && !p.linked_entity);
    return HttpResponse.json({
      packets: filtered.slice(offset, offset + limit),
      total: filtered.length,
      limit,
      offset,
    });
  }),

  http.post("/api/v1/connectors/amie/unmapped/:id/link", async ({ params, request }) => {
    const id = String(params.id);
    const packet = allPackets.find((p) => p.id === id);
    if (!packet) return HttpResponse.json({ error: "not_found" }, { status: 404 });
    const body = (await request.json().catch(() => ({}))) as {
      entity_type?: "project" | "account" | "person" | "user_merge";
      entity_id?: string;
    };
    if (!body.entity_type || !body.entity_id) {
      return HttpResponse.json({ error: "entity_required" }, { status: 400 });
    }
    packet.linked_entity = {
      type: body.entity_type,
      id: body.entity_id,
      display_id: body.entity_id,
    };
    packet.status = "PROCESSED";
    packet.processed_at = new Date().toISOString();
    packet.updated_at = packet.processed_at;
    return HttpResponse.json(packet);
  }),

  http.get("/api/v1/connectors/amie/stats", () => {
    return HttpResponse.json(statsFixture);
  }),
];
