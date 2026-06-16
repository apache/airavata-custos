"use client";

import { Popover as PopoverPrimitive } from "@base-ui/react/popover";
import { Calendar } from "lucide-react";
import * as React from "react";

import { cn } from "@/lib/utils";
import { Button } from "@/shared/ui/button";

const DAY_MS = 24 * 60 * 60 * 1000;

export type DateRangePreset = "24h" | "7d" | "30d" | "90d" | "custom";

export type DateRangeValue = {
  from: Date;
  to: Date;
  preset?: DateRangePreset;
};

export type DateRangePickerProps = {
  value: DateRangeValue;
  onChange: (v: DateRangeValue) => void;
  presets?: Array<{ label: string; days: number; preset: Exclude<DateRangePreset, "custom"> }>;
  className?: string;
};

const DEFAULT_PRESETS: NonNullable<DateRangePickerProps["presets"]> = [
  { label: "24h", days: 1, preset: "24h" },
  { label: "7d", days: 7, preset: "7d" },
  { label: "30d", days: 30, preset: "30d" },
  { label: "90d", days: 90, preset: "90d" },
];

function formatTrigger(value: DateRangeValue, presets: DateRangePickerProps["presets"]): string {
  const matched = (presets ?? DEFAULT_PRESETS).find((p) => p.preset === value.preset);
  if (matched) return `Last ${matched.label}`;
  if (value.preset === "custom" || !value.preset) {
    const fmt = new Intl.DateTimeFormat(undefined, {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
    return `${fmt.format(value.from)} – ${fmt.format(value.to)}`;
  }
  return "Pick a range";
}

function toDateInputValue(d: Date): string {
  // `<input type="date">` expects yyyy-MM-dd in the local timezone.
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

function fromDateInputValue(s: string): Date | null {
  if (!s) return null;
  const ms = Date.parse(`${s}T00:00:00`);
  if (Number.isNaN(ms)) return null;
  return new Date(ms);
}

export function DateRangePicker({
  value,
  onChange,
  presets = DEFAULT_PRESETS,
  className,
}: DateRangePickerProps) {
  const [open, setOpen] = React.useState(false);
  const [showCustom, setShowCustom] = React.useState(value.preset === "custom");
  const [customFrom, setCustomFrom] = React.useState<string>(toDateInputValue(value.from));
  const [customTo, setCustomTo] = React.useState<string>(toDateInputValue(value.to));

  // Reset custom drafts when popover opens so the user starts from the current value.
  React.useEffect(() => {
    if (open) {
      setShowCustom(value.preset === "custom");
      setCustomFrom(toDateInputValue(value.from));
      setCustomTo(toDateInputValue(value.to));
    }
  }, [open, value.preset, value.from, value.to]);

  const handlePreset = (preset: Exclude<DateRangePreset, "custom">, days: number) => {
    const to = new Date();
    const from = new Date(to.getTime() - days * DAY_MS);
    onChange({ from, to, preset });
    setOpen(false);
  };

  const customFromDate = fromDateInputValue(customFrom);
  const customToDate = fromDateInputValue(customTo);
  const canApplyCustom =
    customFromDate !== null && customToDate !== null && customFromDate.getTime() <= customToDate.getTime();

  const handleApplyCustom = () => {
    if (!canApplyCustom || !customFromDate || !customToDate) return;
    onChange({ from: customFromDate, to: customToDate, preset: "custom" });
    setOpen(false);
  };

  return (
    <PopoverPrimitive.Root open={open} onOpenChange={setOpen}>
      <PopoverPrimitive.Trigger
        render={
          <Button variant="outline" size="sm" className={className}>
            <Calendar aria-hidden="true" className="h-3.5 w-3.5 stroke-[1.5]" />
            <span>Range: {formatTrigger(value, presets)}</span>
          </Button>
        }
      />
      <PopoverPrimitive.Portal>
        <PopoverPrimitive.Positioner sideOffset={4} className="z-50 outline-none">
          <PopoverPrimitive.Popup
            className={cn(
              "z-50 w-72 rounded-lg bg-popover p-3 text-popover-foreground shadow-md ring-1 ring-foreground/10",
              "data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0",
            )}
          >
            <div role="group" aria-label="Date range presets" className="flex flex-wrap gap-2">
              {presets.map((p) => (
                <Button
                  key={p.preset}
                  variant={value.preset === p.preset ? "default" : "outline"}
                  size="sm"
                  onClick={() => handlePreset(p.preset, p.days)}
                >
                  {p.label}
                </Button>
              ))}
              <Button
                variant={showCustom ? "default" : "outline"}
                size="sm"
                onClick={() => setShowCustom((v) => !v)}
              >
                Custom…
              </Button>
            </div>
            {showCustom ? (
              <div className="mt-3 space-y-2">
                <label className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
                  <span>From</span>
                  <input
                    type="date"
                    value={customFrom}
                    onChange={(e) => setCustomFrom(e.target.value)}
                    className="rounded-md border border-border bg-background px-2 py-1 text-sm text-foreground"
                  />
                </label>
                <label className="flex items-center justify-between gap-2 text-xs text-muted-foreground">
                  <span>To</span>
                  <input
                    type="date"
                    value={customTo}
                    onChange={(e) => setCustomTo(e.target.value)}
                    className="rounded-md border border-border bg-background px-2 py-1 text-sm text-foreground"
                  />
                </label>
                <div className="flex justify-end pt-1">
                  <Button
                    variant="default"
                    size="sm"
                    onClick={handleApplyCustom}
                    disabled={!canApplyCustom}
                  >
                    Apply
                  </Button>
                </div>
              </div>
            ) : null}
          </PopoverPrimitive.Popup>
        </PopoverPrimitive.Positioner>
      </PopoverPrimitive.Portal>
    </PopoverPrimitive.Root>
  );
}
