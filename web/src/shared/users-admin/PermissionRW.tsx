import { cn } from "@/lib/utils";

function Cell({
  label,
  active,
  activeClass,
  onClick,
}: {
  label: string;
  active: boolean;
  activeClass: string;
  onClick?: () => void;
}) {
  const title = label === "read" ? "Read" : "Write";
  const className = cn(
    "inline-flex h-6 items-center justify-center rounded px-2 text-xs font-medium",
    active ? activeClass : "bg-[color:var(--custos-gray-100)] text-[color:var(--custos-gray-400)]",
    onClick && "cursor-pointer transition-transform hover:scale-105",
  );
  if (!onClick) {
    return (
      <span title={title} className={className}>
        {label}
      </span>
    );
  }
  return (
    <button
      type="button"
      title={title}
      aria-pressed={active}
      aria-label={`${title}${active ? ", granted" : ", not granted"}`}
      onClick={onClick}
      className={className}
    >
      {label}
    </button>
  );
}

export function PermissionRW({
  read,
  write,
  onToggleRead,
  onToggleWrite,
}: {
  read: boolean;
  write: boolean;
  onToggleRead?: () => void;
  onToggleWrite?: () => void;
}) {
  return (
    <div className="flex items-center gap-1">
      <Cell
        label="read"
        active={read}
        activeClass="bg-[color:var(--custos-blue-50)] text-[color:var(--custos-blue-700)]"
        onClick={onToggleRead}
      />
      <Cell
        label="write"
        active={write}
        activeClass="bg-[color:var(--custos-green-50)] text-[color:var(--custos-green-700)]"
        onClick={onToggleWrite}
      />
    </div>
  );
}
