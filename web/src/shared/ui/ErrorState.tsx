import { cn } from "@/lib/utils";
import { Button } from "@/shared/ui/button";
import { AlertCircleIcon } from "lucide-react";
import * as React from "react";

export type ErrorStateProps = {
  heading?: string;
  message?: React.ReactNode;
  onRetry?: () => void;
  retryLabel?: string;
  className?: string;
};

export function ErrorState({
  heading = "Something went wrong",
  message,
  onRetry,
  retryLabel = "Try again",
  className,
}: ErrorStateProps) {
  return (
    <div
      role="alert"
      className={cn(
        // Soft-destructive tint; text stays neutral so contrast on the tint hits AA.
        "flex flex-col items-center justify-center gap-3 rounded-md border border-[color:var(--custos-red-200)] bg-[color:var(--custos-red-50)] px-6 py-10 text-center",
        className,
      )}
    >
      <AlertCircleIcon className="size-8 text-[color:var(--custos-red-600)]" aria-hidden="true" />
      <h3 className="font-heading text-base font-medium text-foreground">{heading}</h3>
      {message ? <p className="max-w-md text-sm text-muted-foreground">{message}</p> : null}
      {onRetry ? (
        <Button variant="outline" onClick={onRetry}>
          {retryLabel}
        </Button>
      ) : null}
    </div>
  );
}
