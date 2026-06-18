"use client";

import { AlertTriangle, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";
import type { RowTone } from "../types";
import type { VisibleRow } from "../utils";
import { isCodeShaped } from "../utils";
import { SourcePill } from "./primitives/SourcePill";
import { StatusPill } from "./primitives/StatusPill";

const INDENT_PX = 20;
const MAX_DEPTH = 5;

export type TraceTreeRowProps = {
  row: VisibleRow;
  tone: RowTone;
  source: string;
  isSelected: boolean;
  isOnErrorPath: boolean;
  isPreciseFailure: boolean;
  isExpanded: boolean;
  hasHiddenError: boolean;
  entityMeta?: string;
  onSelect: () => void;
  onToggle: () => void;
  rowRef: (el: HTMLDivElement | null) => void;
};

const ENTITY_KEYS: ReadonlyArray<string> = [
  "entity.user_id",
  "person.id",
  "amie.packet_id",
  "comanage.co_person_id",
  "slurm.cluster",
  "slurm.account",
  "allocation.id",
];

// First matching key wins.
export function pickEntityMeta(attributes: unknown): string | undefined {
  if (!attributes || typeof attributes !== "object") return undefined;
  const attrs = attributes as Record<string, unknown>;
  for (const k of ENTITY_KEYS) {
    const v = attrs[k];
    if (v == null) continue;
    if (typeof v === "string" || typeof v === "number" || typeof v === "boolean") {
      const short = k.split(".").pop() ?? k;
      return `${short}=${v}`;
    }
  }
  return undefined;
}

function statusCode(attributes: unknown): string | number | undefined {
  if (!attributes || typeof attributes !== "object") return undefined;
  const v = (attributes as Record<string, unknown>)["http.status_code"];
  if (typeof v === "string" || typeof v === "number") return v;
  return undefined;
}

export function TraceTreeRow({
  row,
  tone,
  source,
  isSelected,
  isOnErrorPath,
  isPreciseFailure,
  isExpanded,
  hasHiddenError,
  entityMeta,
  onSelect,
  onToggle,
  rowRef,
}: TraceTreeRowProps) {
  const { node, depth, hasChildren } = row;
  const span = node.span;
  const isError = tone === "error";
  const isRunning = span.running === true;
  const isNotRun = span.notRun === true;
  const isOrphan = span.orphan === true;
  const capped = Math.min(depth, MAX_DEPTH);
  const overCap = depth > MAX_DEPTH;
  const code = isCodeShaped(span.name);
  const display = overCap ? `…/${span.name}` : span.name;
  const errStatus = isError ? statusCode(span.attributes) : undefined;
  const resolvedEntityMeta = entityMeta ?? pickEntityMeta(span.attributes);
  const showEntityMeta = !isRunning && !isNotRun && Boolean(resolvedEntityMeta);

  // Color the precise failing leaf with a faint red wash. Ancestor error rows
  // stay calm — only the rail connects them.
  const rowBg = isSelected
    ? "var(--brand-tint)"
    : isPreciseFailure
      ? "var(--tone-error-bg)"
      : undefined;

  return (
    <div
      ref={rowRef}
      role="treeitem"
      aria-level={depth + 1}
      aria-selected={isSelected}
      aria-expanded={hasChildren ? isExpanded : undefined}
      data-testid={`trace-tree-row-${span.span_id}`}
      data-tone={tone}
      data-precise={isPreciseFailure ? "true" : undefined}
      data-on-error-path={isOnErrorPath ? "true" : undefined}
      onClick={onSelect}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onSelect();
        }
      }}
      tabIndex={-1}
      className={cn(
        "relative ml-1.5 mr-1 flex cursor-pointer items-center gap-2 rounded-md transition-colors",
        hasChildren ? "min-h-9" : "min-h-8",
        !isSelected && !isPreciseFailure ? "hover:bg-[color:var(--muted-2)]" : null,
      )}
      style={{
        paddingLeft: 0,
        paddingRight: 10,
        background: rowBg,
        boxShadow: isSelected ? "inset 0 0 0 1px var(--brand)" : undefined,
      }}
    >
      {isSelected && (
        <span
          aria-hidden="true"
          className="absolute top-1 bottom-1 left-0 rounded-sm"
          style={{ width: 2.5, background: "var(--brand)" }}
        />
      )}
      {!isSelected && isPreciseFailure && (
        <span
          aria-hidden="true"
          className="absolute top-1 bottom-1 left-0 rounded-sm"
          style={{ width: 3, background: "var(--custos-red-500)" }}
        />
      )}

      <span
        className="relative flex h-full shrink-0 items-center"
        style={{ width: 10 + capped * INDENT_PX }}
      >
        {hasChildren ? (
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onToggle();
            }}
            aria-label={isExpanded ? "Collapse" : "Expand"}
            className="absolute right-0.5 inline-flex h-4 w-4 items-center justify-center border-none bg-transparent p-0 text-muted-foreground hover:text-foreground"
          >
            <ChevronRight
              className="h-3.5 w-3.5 transition-transform duration-100"
              style={{ transform: isExpanded ? "rotate(90deg)" : undefined }}
              aria-hidden="true"
            />
          </button>
        ) : (
          <span aria-hidden="true" className="inline-block h-4 w-4" />
        )}
      </span>

      <StatusPill tone={tone} dotOnly />

      <span
        className={cn(
          "min-w-0 flex-shrink overflow-hidden text-ellipsis whitespace-nowrap text-foreground",
          code ? "font-mono text-[13px]" : "text-sm",
        )}
        style={{
          fontWeight: hasChildren || isOnErrorPath || isError ? 600 : 500,
        }}
      >
        {display}
      </span>

      {hasHiddenError && (
        <span
          title="Contains a failed span"
          className="inline-flex shrink-0 items-center text-[color:var(--banner-error-icon)]"
        >
          <AlertTriangle className="h-3 w-3" aria-hidden="true" />
        </span>
      )}
      {isOrphan && (
        <span className="shrink-0 text-[11px] font-medium text-muted-foreground">(orphan)</span>
      )}

      <span className="ml-0.5 shrink-0">
        <SourcePill source={source} />
      </span>

      <span
        className="ml-auto flex shrink-0 items-center gap-2 overflow-hidden pl-2 text-[12px] whitespace-nowrap text-muted-foreground"
        style={{ maxWidth: "46%" }}
      >
        {errStatus != null && (
          <span className="overflow-hidden font-mono text-ellipsis text-[color:var(--banner-error-fg)]">
            status={String(errStatus)}
          </span>
        )}
        {isRunning ? (
          <span className="overflow-hidden text-ellipsis italic">…still running</span>
        ) : isNotRun ? (
          <span className="overflow-hidden text-ellipsis">skipped (parent err)</span>
        ) : showEntityMeta ? (
          <span className="overflow-hidden font-mono text-ellipsis">{resolvedEntityMeta}</span>
        ) : null}
        {isPreciseFailure && (
          <AlertTriangle
            className="h-3 w-3 shrink-0 text-[color:var(--banner-error-icon)]"
            aria-hidden="true"
          />
        )}
      </span>
    </div>
  );
}
