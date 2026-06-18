"use client";

import { Check, Copy } from "lucide-react";
import * as React from "react";
import { cn } from "@/lib/utils";

export type CopyValueProps = {
  value: string;
  label?: string;
  explicit?: boolean;
  className?: string;
  children?: React.ReactNode;
};

const COPIED_RESET_MS = 1100;

export function CopyValue({ value, label, explicit = false, className, children }: CopyValueProps) {
  const [copied, setCopied] = React.useState(false);
  const timerRef = React.useRef<ReturnType<typeof setTimeout> | null>(null);

  React.useEffect(() => {
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
  }, []);

  const handleCopy = (event: React.MouseEvent) => {
    // Copying a trace id inside a list row should not also open the drawer.
    event.stopPropagation();
    event.preventDefault();
    void (async () => {
      try {
        await navigator.clipboard.writeText(value);
      } catch {
        // Best-effort: still flash the check so the user has feedback.
      }
      setCopied(true);
      if (timerRef.current) clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => setCopied(false), COPIED_RESET_MS);
    })();
  };

  const Icon = copied ? Check : Copy;
  const ariaLabel = `Copy ${label ?? value}`;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 font-mono text-xs text-foreground",
        className,
      )}
    >
      <span className="truncate">{children ?? value}</span>
      <button
        type="button"
        onClick={handleCopy}
        aria-label={ariaLabel}
        className={cn(
          "inline-flex h-5 w-5 shrink-0 items-center justify-center rounded-md text-muted-foreground",
          "hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
          copied ? "text-[color:var(--custos-green-700)]" : null,
          explicit ? "opacity-100" : "opacity-60",
        )}
      >
        <Icon className="h-3.5 w-3.5" aria-hidden="true" />
      </button>
    </span>
  );
}
