import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { CopyValue } from "@/features/core/audit/components/primitives/CopyValue";

describe("CopyValue", () => {
  const writeText = vi.fn().mockResolvedValue(undefined);

  beforeEach(() => {
    writeText.mockClear();
    Object.assign(navigator, { clipboard: { writeText } });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("writes to the clipboard on click and stops propagation", () => {
    const onParentClick = vi.fn();
    render(
      <button type="button" onClick={onParentClick}>
        <CopyValue value="trace-abc" label="trace ID" />
      </button>,
    );
    const btn = screen.getByRole("button", { name: "Copy trace ID" });
    fireEvent.click(btn);
    expect(writeText).toHaveBeenCalledWith("trace-abc");
    expect(onParentClick).not.toHaveBeenCalled();
  });

  it("swaps to the check icon for 1.1s then reverts to copy", async () => {
    const { container } = render(<CopyValue value="trace-abc" />);
    const btn = screen.getByRole("button", { name: "Copy trace-abc" });
    fireEvent.click(btn);
    await waitFor(() => {
      expect(container.querySelector("svg.lucide-check")).not.toBeNull();
    });
    expect(container.querySelector("svg.lucide-copy")).toBeNull();

    await waitFor(
      () => {
        expect(container.querySelector("svg.lucide-check")).toBeNull();
      },
      { timeout: 1500 },
    );
    expect(container.querySelector("svg.lucide-copy")).not.toBeNull();
  });
});
