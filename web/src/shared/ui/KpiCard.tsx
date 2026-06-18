"use client";

import { cn } from "@/lib/utils";
import { Sparkline } from "@/shared/charts/Sparkline";
import { Skeleton } from "@/shared/ui/skeleton";
import { ArrowDown, ArrowUp } from "lucide-react";
import type { ElementType, ReactNode } from "react";

// Card chrome matches StatCard so KPI strip + StatCardRow stay visually aligned.
const cardChrome = "rounded-lg border border-border bg-card p-5 shadow-sm";

export type KpiDeltaTone = "positive" | "negative" | "neutral";

export type KpiDelta = {
  value: number;
  unit?: string;
  direction: "up" | "down";
};

export type KpiCardProps = {
  icon?: ElementType;
  title: ReactNode;
  value: ReactNode;
  delta?: KpiDelta;
  deltaTone?: KpiDeltaTone;
  sparkline?: number[];
  onClick?: () => void;
  loading?: boolean;
  className?: string;
};

// "up" isn't always "good" — wait-time-down is positive, cost-up is negative.
// Tone drives color so the caller stays in charge of semantics.
const toneStyles: Record<KpiDeltaTone, string> = {
  positive: "bg-[color:var(--custos-green-50)] text-[color:var(--custos-green-700)]",
  negative: "bg-[color:var(--custos-red-50)] text-[color:var(--custos-red-700)]",
  neutral: "bg-muted text-muted-foreground",
};

function KpiCardTitle({ icon: Icon, title }: { icon?: ElementType; title: ReactNode }) {
  return (
    <div className="flex items-center gap-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
      {Icon ? <Icon className="h-4 w-4 stroke-[1.5]" /> : null}
      <span>{title}</span>
    </div>
  );
}

function DeltaPill({ delta, tone }: { delta: KpiDelta; tone: KpiDeltaTone }) {
  const Arrow = delta.direction === "up" ? ArrowUp : ArrowDown;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-0.5 rounded-full px-2 py-0.5 text-xs font-medium",
        toneStyles[tone],
      )}
      data-testid="kpi-delta"
    >
      <Arrow aria-hidden="true" className="h-3 w-3 stroke-[1.5]" />
      <span className="tabular-nums">
        {delta.value}
        {delta.unit ?? "%"}
      </span>
    </span>
  );
}

function sparklineLabelFor(title: ReactNode): string {
  // Per-card aria-label so each sparkline surfaces distinctly in the a11y
  // tree. Falls back to a generic label when the title isn't a plain string
  // (e.g. JSX badge composition).
  if (typeof title === "string" && title.trim().length > 0) {
    return `${title} trend sparkline`;
  }
  return "Trend sparkline";
}

function KpiCardBody({
  icon,
  title,
  value,
  delta,
  deltaTone = "neutral",
  sparkline,
}: Omit<KpiCardProps, "onClick" | "loading" | "className">) {
  return (
    <>
      <KpiCardTitle icon={icon} title={title} />
      <div className="mt-2 flex items-baseline gap-3">
        <span className="font-display text-3xl font-bold text-foreground tabular-nums">
          {value}
        </span>
        {delta ? <DeltaPill delta={delta} tone={deltaTone} /> : null}
      </div>
      {sparkline && sparkline.length > 0 ? (
        <div className="mt-3" data-testid="kpi-sparkline">
          <Sparkline
            data={sparkline.map((v) => ({ value: v }))}
            ariaLabel={sparklineLabelFor(title)}
            height={28}
          />
        </div>
      ) : null}
    </>
  );
}

function KpiCardSkeleton() {
  return (
    <div className="space-y-3" data-testid="kpi-skeleton">
      <Skeleton className="h-3 w-24" />
      <Skeleton className="h-8 w-32" />
      <Skeleton className="h-7 w-full" />
    </div>
  );
}

export function KpiCard({
  icon,
  title,
  value,
  delta,
  deltaTone,
  sparkline,
  onClick,
  loading,
  className,
}: KpiCardProps) {
  if (loading) {
    return (
      <div className={cn(cardChrome, className)} data-state="loading">
        <KpiCardSkeleton />
      </div>
    );
  }

  if (onClick) {
    return (
      <button
        type="button"
        onClick={onClick}
        className={cn(
          cardChrome,
          "w-full text-left transition-colors hover:bg-muted/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
          className,
        )}
      >
        <KpiCardBody
          icon={icon}
          title={title}
          value={value}
          delta={delta}
          deltaTone={deltaTone}
          sparkline={sparkline}
        />
      </button>
    );
  }

  return (
    <div className={cn(cardChrome, className)}>
      <KpiCardBody
        icon={icon}
        title={title}
        value={value}
        delta={delta}
        deltaTone={deltaTone}
        sparkline={sparkline}
      />
    </div>
  );
}
