import { cn } from "@/lib/utils";
import {
  StatusBadge,
  type StatusBadgeProps,
  type StatusBadgeVariant,
} from "@/shared/ui/StatusBadge";
import type { ElementType, ReactNode } from "react";

export function MetaRow({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  // Not a <dl> — items aren't true term/definition pairs (status pill has no
  // term, value can be a node, etc.). A labeled group reads correctly for AT.
  return (
    <div
      role="group"
      aria-label="Metadata"
      className={cn(
        "flex flex-wrap items-center gap-x-6 gap-y-2 text-sm text-foreground",
        className,
      )}
    >
      {children}
    </div>
  );
}

type MetaTone = "success" | "warning" | "danger";

type MetaItemBase = {
  className?: string;
};

type DefaultMetaItem = MetaItemBase & {
  variant?: "default";
  icon?: ElementType;
  label?: string;
  value: ReactNode;
  tone?: never;
};

type StatusMetaItem = MetaItemBase & {
  variant: "status";
  icon?: never;
  label?: never;
  value: ReactNode;
  tone?: MetaTone;
};

export type MetaItemProps = DefaultMetaItem | StatusMetaItem;

const toneToVariant: Record<MetaTone, StatusBadgeVariant> = {
  success: "active",
  warning: "warning",
  danger: "rejected",
};

export function MetaItem(props: MetaItemProps) {
  if (props.variant === "status") {
    const variant: StatusBadgeProps["variant"] = props.tone
      ? toneToVariant[props.tone]
      : "active";
    return <StatusBadge variant={variant} label={String(props.value)} className={props.className} />;
  }

  const { icon: Icon, label, value, className } = props;
  return (
    <div className={cn("flex items-center gap-2", className)}>
      {Icon ? <Icon className="h-4 w-4 stroke-[1.5] text-muted-foreground" /> : null}
      {label ? <span className="text-muted-foreground">{label} :</span> : null}
      <span className="font-medium text-foreground">{value}</span>
    </div>
  );
}
