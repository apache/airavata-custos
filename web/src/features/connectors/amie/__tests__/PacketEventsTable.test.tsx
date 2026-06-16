import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { PacketEvent } from "../types";

const can = vi.fn<(action: string, subject: string) => boolean>();

vi.mock("@/shared/casl/AbilityProvider", () => ({
  useAbility: () => ({ can }),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  usePathname: () => "/admin/amie/packets",
  useSearchParams: () => new URLSearchParams(),
}));

import { PacketEventsTable } from "../components/PacketEventsTable";

const EVENTS: PacketEvent[] = [
  {
    id: "evt-1",
    packet_id: "pkt-1",
    event_type: "RECEIVED",
    actor: "amie-worker",
    status: "SUCCEEDED",
    timestamp: "2026-06-08T12:00:00Z",
  },
  {
    id: "evt-2",
    packet_id: "pkt-1",
    event_type: "FAILED",
    actor: "amie-worker",
    status: "FAILED",
    timestamp: "2026-06-08T12:00:10Z",
    message: "cluster not reachable",
    trace_id: "a3b1c92d3f4e5a6b7c8d9e0f12345678",
  },
];

describe("PacketEventsTable", () => {
  it("renders a spinner while loading", () => {
    render(<PacketEventsTable events={[]} isLoading={true} />);
    expect(screen.getByLabelText(/Loading/i)).toBeInTheDocument();
  });

  it("renders an empty-state when there are no events", () => {
    render(<PacketEventsTable events={[]} isLoading={false} />);
    expect(screen.getByText(/No events recorded/i)).toBeInTheDocument();
  });

  it("renders one row per event with its event_type", () => {
    can.mockReturnValue(true);
    render(<PacketEventsTable events={EVENTS} isLoading={false} />);
    expect(screen.getByText("RECEIVED")).toBeInTheDocument();
    expect(screen.getByText("FAILED")).toBeInTheDocument();
  });

  it("surfaces ViewTraceLink only on events that carry a trace_id", () => {
    can.mockReturnValue(true);
    render(<PacketEventsTable events={EVENTS} isLoading={false} />);
    const links = screen.getAllByRole("link", { name: /View trace/i });
    expect(links).toHaveLength(1);
  });

  it("omits ViewTraceLink when the ability denies read Trace", () => {
    can.mockReturnValue(false);
    render(<PacketEventsTable events={EVENTS} isLoading={false} />);
    expect(screen.queryByRole("link", { name: /View trace/i })).not.toBeInTheDocument();
  });
});
