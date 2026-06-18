import { cn } from "@/lib/utils";
import type { PacketStatus, ReplyStatus } from "../types";

const packetStyles: Record<PacketStatus, string> = {
  NEW: "bg-[color:var(--custos-blue-50)] text-[color:var(--custos-blue-700)]",
  DECODED: "bg-[color:var(--custos-amber-50)] text-[color:var(--custos-amber-700)]",
  PROCESSED: "bg-[color:var(--custos-green-50)] text-[color:var(--custos-green-700)]",
  FAILED: "bg-[color:var(--custos-red-50)] text-[color:var(--custos-red-700)]",
};

const replyStyles: Record<ReplyStatus, string> = {
  PENDING: "bg-[color:var(--custos-amber-50)] text-[color:var(--custos-amber-700)]",
  SENT: "bg-[color:var(--custos-blue-50)] text-[color:var(--custos-blue-700)]",
  ACKED: "bg-[color:var(--custos-green-50)] text-[color:var(--custos-green-700)]",
  FAILED: "bg-[color:var(--custos-red-50)] text-[color:var(--custos-red-700)]",
};

export function PacketStatusBadge({
  status,
  ageHours,
  className,
}: {
  status: PacketStatus;
  ageHours?: number;
  className?: string;
}) {
  // Aged-FAILED packets stay louder via weight + trailing "!" — no ring needed.
  const loud = status === "FAILED" && ageHours !== undefined && ageHours > 24;
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium",
        packetStyles[status],
        loud && "font-semibold",
        className,
      )}
    >
      {status}
      {loud ? <span aria-hidden="true">!</span> : null}
    </span>
  );
}

export function ReplyStatusBadge({
  status,
  className,
}: {
  status: ReplyStatus;
  className?: string;
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
        replyStyles[status],
        className,
      )}
    >
      {status}
    </span>
  );
}
