import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { TraceFilterStrip } from "@/features/core/audit/components/TraceFilterStrip";
import {
  DEFAULT_FILTERS,
  type ListFilters,
} from "@/features/core/audit/components/traceListUrlState";

function renderStrip(initial: ListFilters = DEFAULT_FILTERS) {
  const onChange = vi.fn<(next: ListFilters) => void>();
  let value: ListFilters = initial;
  const Wrap = () => (
    <TraceFilterStrip
      value={value}
      onChange={(next) => {
        value = next;
        onChange(next);
      }}
    />
  );
  const utils = render(<Wrap />);
  return { ...utils, onChange, getValue: () => value };
}

describe("TraceFilterStrip", () => {
  it("renders the default state with error pre-pressed", () => {
    renderStrip();
    const errorPill = screen.getByRole("button", { name: /error/i, pressed: true });
    expect(errorPill).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^ok$/i, pressed: false })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "30d", pressed: true })).toBeInTheDocument();
  });

  it("toggles a status pill and resets to page 1", () => {
    const { onChange } = renderStrip({ ...DEFAULT_FILTERS, page: 3 });
    fireEvent.click(screen.getByRole("button", { name: /^ok$/i }));
    expect(onChange).toHaveBeenCalled();
    const last = onChange.mock.calls.at(-1)?.[0];
    expect(last?.status).toContain("ok");
    expect(last?.status).toContain("error");
    expect(last?.page).toBe(1);
  });

  it("clicking a window pill replaces selection (radio behavior)", () => {
    const { onChange } = renderStrip();
    fireEvent.click(screen.getByRole("button", { name: "24h" }));
    const last = onChange.mock.calls.at(-1)?.[0];
    expect(last?.window).toBe("24h");
  });

  it("debounces the search input and emits page=1", async () => {
    const { onChange } = renderStrip();
    const input = screen.getByRole("searchbox", { name: /search traces/i });
    fireEvent.change(input, { target: { value: "alice" } });
    await waitFor(
      () => {
        const call = onChange.mock.calls.find((c) => c[0].q === "alice");
        expect(call).toBeTruthy();
      },
      { timeout: 1000 },
    );
    const last = onChange.mock.calls.find((c) => c[0].q === "alice")?.[0];
    expect(last?.page).toBe(1);
  });

  it("toggles source pills", () => {
    const { onChange } = renderStrip();
    fireEvent.click(screen.getByRole("button", { name: "amie" }));
    const last = onChange.mock.calls.at(-1)?.[0];
    expect(last?.source).toEqual(["amie"]);
  });

  it("renders the Failing >24h chip when filter is on, and clears it on click", () => {
    const { onChange } = renderStrip({ ...DEFAULT_FILTERS, failingOver24h: true, page: 4 });
    const chip = screen.getByTestId("failing-over-24h-chip");
    expect(chip).toBeInTheDocument();
    fireEvent.click(chip);
    const last = onChange.mock.calls.at(-1)?.[0];
    expect(last?.failingOver24h).toBe(false);
    expect(last?.page).toBe(1);
  });

  it("does not render the Failing >24h chip when filter is off", () => {
    renderStrip();
    expect(screen.queryByTestId("failing-over-24h-chip")).toBeNull();
  });
});
