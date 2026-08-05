// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
  active: "bg-[color:var(--tone-ok-bg)] text-[color:var(--tone-ok-fg)]",
  inactive: "bg-muted text-muted-foreground",
  deleted: "bg-[color:var(--tone-error-bg)] text-[color:var(--tone-error-fg)]",
  pending: "bg-[color:var(--tone-info-bg)] text-[color:var(--tone-info-fg)]",
  approved: "bg-[color:var(--tone-ok-bg)] text-[color:var(--tone-ok-fg)]",
  rejected: "bg-[color:var(--tone-error-bg)] text-[color:var(--tone-error-fg)]",
  expired: "bg-muted text-muted-foreground",
  warning: "bg-[color:var(--tone-warn-bg)] text-[color:var(--tone-warn-fg)]",
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
