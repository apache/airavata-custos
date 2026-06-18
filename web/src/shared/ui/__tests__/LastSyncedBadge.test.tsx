import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { LastSyncedBadge } from "../LastSyncedBadge";

const NOW = new Date("2026-05-22T12:00:00Z");

describe("LastSyncedBadge", () => {
  it("shows abbreviated relative-time copy for a minutes-old sync", () => {
    const syncedAt = new Date(NOW.getTime() - 12 * 60 * 1000);
    render(<LastSyncedBadge syncedAt={syncedAt} now={NOW} />);
    expect(screen.getByText(/Synced 12m ago/)).toBeInTheDocument();
  });

  it("abbreviates hours and days too", () => {
    const hours = new Date(NOW.getTime() - 3 * 60 * 60 * 1000);
    const { rerender } = render(<LastSyncedBadge syncedAt={hours} now={NOW} />);
    expect(screen.getByText(/Synced 3h ago/)).toBeInTheDocument();
    const days = new Date(NOW.getTime() - 2 * 24 * 60 * 60 * 1000);
    rerender(<LastSyncedBadge syncedAt={days} now={NOW} />);
    expect(screen.getByText(/Synced 2d ago/)).toBeInTheDocument();
  });

  it("uses fresh tint when sync is under 1 hour old", () => {
    const syncedAt = new Date(NOW.getTime() - 30 * 60 * 1000);
    const { container } = render(<LastSyncedBadge syncedAt={syncedAt} now={NOW} />);
    expect(container.firstElementChild?.getAttribute("data-tint")).toBe("fresh");
  });

  it("flips to stale tint between 1h and 24h and applies the amber color token", () => {
    const syncedAt = new Date(NOW.getTime() - 3 * 60 * 60 * 1000);
    const { container } = render(<LastSyncedBadge syncedAt={syncedAt} now={NOW} />);
    const badge = container.firstElementChild;
    expect(badge?.getAttribute("data-tint")).toBe("stale");
    // Amber tint is sourced from `--custos-amber-*` semantic tokens; assert
    // the class is present so a stylesheet refactor that drops it fails here.
    expect(badge?.className ?? "").toMatch(/var\(--custos-amber-50\)/);
    expect(badge?.className ?? "").toMatch(/var\(--custos-amber-700\)/);
  });

  it("does not stay fresh once the sync crosses the 1h boundary", () => {
    const just = new Date(NOW.getTime() - (60 * 60 * 1000 + 1));
    const { container } = render(<LastSyncedBadge syncedAt={just} now={NOW} />);
    expect(container.firstElementChild?.getAttribute("data-tint")).toBe("stale");
  });

  it("flips to expired tint after 24h", () => {
    const syncedAt = new Date(NOW.getTime() - 48 * 60 * 60 * 1000);
    const { container } = render(<LastSyncedBadge syncedAt={syncedAt} now={NOW} />);
    expect(container.firstElementChild?.getAttribute("data-tint")).toBe("expired");
  });

  it("fires onRefetch when the refresh icon is clicked", () => {
    const onRefetch = vi.fn();
    render(
      <LastSyncedBadge syncedAt={new Date(NOW.getTime() - 1000)} now={NOW} onRefetch={onRefetch} />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Refetch data/i }));
    expect(onRefetch).toHaveBeenCalledTimes(1);
  });
});
