import { cn } from "@/lib/utils";
import * as React from "react";

export type EmptyStateProps = {
  icon?: React.ReactNode;
  heading: string;
  description?: React.ReactNode;
  cta?: React.ReactNode;
  className?: string;
};

export function EmptyState({ icon, heading, description, cta, className }: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center justify-center gap-3 rounded-md border border-dashed bg-card px-6 py-12 text-center",
        className,
      )}
    >
      {icon ? (
        <div className="text-muted-foreground" aria-hidden="true">
          {icon}
        </div>
      ) : null}
      <h3 className="font-heading text-base font-medium text-foreground">{heading}</h3>
      {description ? <p className="max-w-md text-sm text-muted-foreground">{description}</p> : null}
      {cta ? <div className="mt-2">{cta}</div> : null}
    </div>
  );
}
