import { fireEvent, render, screen, within } from "@testing-library/react";
import * as React from "react";
import { describe, expect, it, vi } from "vitest";
import { DateRangePicker, type DateRangeValue } from "../DateRangePicker";

function Harness({
  initial,
  onChangeSpy,
}: {
  initial: DateRangeValue;
  onChangeSpy: (v: DateRangeValue) => void;
}) {
  const [value, setValue] = React.useState<DateRangeValue>(initial);
  const onChange = React.useCallback(
    (v: DateRangeValue) => {
      onChangeSpy(v);
      setValue(v);
    },
    [onChangeSpy],
  );
  return <DateRangePicker value={value} onChange={onChange} />;
}

describe("DateRangePicker", () => {
  const baseValue: DateRangeValue = {
    from: new Date("2026-04-22T00:00:00Z"),
    to: new Date("2026-05-22T00:00:00Z"),
    preset: "30d",
  };

  it("opens popover on trigger click and shows preset chips", () => {
    const spy = vi.fn();
    render(<Harness initial={baseValue} onChangeSpy={spy} />);
    fireEvent.click(screen.getByRole("button", { name: /Range:/ }));
    const presetGroup = screen.getByRole("group", { name: /Date range presets/i });
    expect(within(presetGroup).getByRole("button", { name: "24h" })).toBeInTheDocument();
    expect(within(presetGroup).getByRole("button", { name: "7d" })).toBeInTheDocument();
    expect(within(presetGroup).getByRole("button", { name: "30d" })).toBeInTheDocument();
    expect(within(presetGroup).getByRole("button", { name: "90d" })).toBeInTheDocument();
  });

  it("preset click fires onChange with from/to and preset", () => {
    const spy = vi.fn();
    render(<Harness initial={baseValue} onChangeSpy={spy} />);
    fireEvent.click(screen.getByRole("button", { name: /Range:/ }));
    const presetGroup = screen.getByRole("group", { name: /Date range presets/i });
    fireEvent.click(within(presetGroup).getByRole("button", { name: "7d" }));
    expect(spy).toHaveBeenCalledTimes(1);
    const arg = spy.mock.calls[0]?.[0] as DateRangeValue;
    expect(arg.preset).toBe("7d");
    expect(arg.from).toBeInstanceOf(Date);
    expect(arg.to).toBeInstanceOf(Date);
    expect(arg.from.getTime()).toBeLessThan(arg.to.getTime());
  });

  it("valid custom range fires onChange when Apply clicked", () => {
    const spy = vi.fn();
    render(<Harness initial={baseValue} onChangeSpy={spy} />);
    fireEvent.click(screen.getByRole("button", { name: /Range:/ }));
    fireEvent.click(screen.getByRole("button", { name: /Custom/i }));
    const fromInput = screen.getByLabelText("From") as HTMLInputElement;
    const toInput = screen.getByLabelText("To") as HTMLInputElement;
    fireEvent.change(fromInput, { target: { value: "2026-05-01" } });
    fireEvent.change(toInput, { target: { value: "2026-05-15" } });
    const apply = screen.getByRole("button", { name: "Apply" });
    expect(apply).not.toBeDisabled();
    fireEvent.click(apply);
    expect(spy).toHaveBeenCalledTimes(1);
    const arg = spy.mock.calls[0]?.[0] as DateRangeValue;
    expect(arg.preset).toBe("custom");
  });

  it("disables Apply when from > to", () => {
    const spy = vi.fn();
    render(<Harness initial={baseValue} onChangeSpy={spy} />);
    fireEvent.click(screen.getByRole("button", { name: /Range:/ }));
    fireEvent.click(screen.getByRole("button", { name: /Custom/i }));
    fireEvent.change(screen.getByLabelText("From"), { target: { value: "2026-05-15" } });
    fireEvent.change(screen.getByLabelText("To"), { target: { value: "2026-05-01" } });
    expect(screen.getByRole("button", { name: "Apply" })).toBeDisabled();
  });
});
