import { cn } from "@/lib/utils";
import type { AllocationStatus, ChangeRequestStatus } from "@/shared/api/domain";

export type StatusBadgeVariant =
  | "active"
  | "inactive"
  | "deleted"
  | "pending"
  | "approved"
  | "rejected"
  | "expired"
  | "warning";

// Flat fill + tinted text; bg-muted for neutral pills so they follow the theme.
const variantStyles: Record<StatusBadgeVariant, string> = {
  active: "bg-[color:var(--custos-green-50)] text-[color:var(--custos-green-700)]",
  inactive: "bg-muted text-muted-foreground",
  deleted: "bg-[color:var(--custos-red-50)] text-[color:var(--custos-red-700)]",
  pending: "bg-[color:var(--custos-blue-50)] text-[color:var(--custos-blue-700)]",
  approved: "bg-[color:var(--custos-green-50)] text-[color:var(--custos-green-700)]",
  rejected: "bg-[color:var(--custos-red-50)] text-[color:var(--custos-red-700)]",
  expired: "bg-muted text-muted-foreground",
  warning: "bg-[color:var(--custos-amber-50)] text-[color:var(--custos-amber-700)]",
};

const labels: Record<StatusBadgeVariant, string> = {
  active: "Active",
  inactive: "Inactive",
  deleted: "Deleted",
  pending: "Pending",
  approved: "Approved",
  rejected: "Rejected",
  expired: "Expired",
  warning: "Warning",
};

export type StatusBadgeProps = {
  variant: StatusBadgeVariant;
  label?: string;
  className?: string;
};

export function StatusBadge({ variant, label, className }: StatusBadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs font-medium",
        variantStyles[variant],
        className,
      )}
    >
      {variant === "active" ? (
        <span aria-hidden="true" className="h-2 w-2 rounded-full bg-current" />
      ) : null}
      {label ?? labels[variant]}
    </span>
  );
}

export function statusBadgeVariantFromAllocationStatus(
  status: AllocationStatus,
): StatusBadgeVariant {
  if (status === "ACTIVE") return "active";
  if (status === "INACTIVE") return "inactive";
  return "deleted";
}

export function statusBadgeVariantFromChangeRequest(
  status: ChangeRequestStatus,
): StatusBadgeVariant {
  if (status === "PENDING") return "pending";
  if (status === "APPROVED") return "approved";
  return "rejected";
}
