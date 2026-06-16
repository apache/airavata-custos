"use client";

import {
  AlertTriangle,
  ArrowDown,
  ArrowRight,
  ArrowUp,
  ChevronsDownUp,
  ChevronsUpDown,
} from "lucide-react";
import * as React from "react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { Button } from "@/shared/ui/button";
import type { Span, Trace, UISpan } from "../types";
import {
  buildTree,
  detectErrorPath,
  enrichSpan,
  flattenTree,
  rowTone,
  subtreeHasError,
  type TreeNode,
  type VisibleRow,
} from "../utils";
import { TraceSpanDetailPanel } from "./TraceSpanDetailPanel";
import { TraceTreeRow } from "./TraceTreeRow";

export type TraceTreeTabProps = {
  trace: Trace;
  spans: Span[];
  onSwitchToTab: (tab: "overview" | "raw" | "linked") => void;
};

const SPAN_PARAM = "span";
const SCROLL_PAD_PX = 40;

function readSpanSource(span: Span, traceSource: string): string {
  const attrs = span.attributes;
  if (attrs && typeof attrs === "object" && !Array.isArray(attrs)) {
    const s = (attrs as Record<string, unknown>).source;
    if (typeof s === "string" && s) return s;
  }
  return traceSource;
}

export function TraceTreeTab({ trace, spans, onSwitchToTab }: TraceTreeTabProps) {
  const traceSource = String(trace.source);

  const enrichedSpans = React.useMemo<UISpan[]>(() => {
    const byId = new Map<string, Span>(spans.map((s) => [s.span_id, s]));
    return spans.map((s) =>
      enrichSpan(s, byId, s.parent_span_id ? byId.get(s.parent_span_id)?.status === 1 : false),
    );
  }, [spans]);

  const { roots, byId } = React.useMemo(() => buildTree(enrichedSpans), [enrichedSpans]);
  const { pathSet, errorLeafIds } = React.useMemo(
    () => detectErrorPath(enrichedSpans),
    [enrichedSpans],
  );
  const pathSetRef = React.useRef(pathSet);
  pathSetRef.current = pathSet;
  const errorLeafSet = React.useMemo(() => new Set(errorLeafIds), [errorLeafIds]);

  const structuralIds = React.useMemo(() => {
    const out: string[] = [];
    for (const node of byId.values()) if (node.children.length > 0) out.push(node.span.span_id);
    return out;
  }, [byId]);

  const [expanded, setExpanded] = React.useState<Set<string>>(() => new Set(structuralIds));
  // Re-seed when the trace changes — stale expanded-set can carry over span IDs
  // from the previous trace.
  const traceIdRef = React.useRef(trace.trace_id);
  if (traceIdRef.current !== trace.trace_id) {
    traceIdRef.current = trace.trace_id;
    setExpanded(new Set(structuralIds));
  }

  const [errorsOnly, setErrorsOnly] = React.useState(false);
  const [errorCursor, setErrorCursor] = React.useState(0);

  const searchParams = useShallowSearchParams();
  const urlSelectedSpan = searchParams.get(SPAN_PARAM);
  const fallbackSelected = errorLeafIds[0] ?? roots[0]?.span.span_id ?? null;
  const selectedSpanId =
    urlSelectedSpan && byId.has(urlSelectedSpan) ? urlSelectedSpan : fallbackSelected;
  const selectedNode: TreeNode | null = selectedSpanId
    ? byId.get(selectedSpanId) ?? null
    : null;

  const writeSelected = React.useCallback((spanId: string) => {
    const params = new URLSearchParams(window.location.search);
    params.set(SPAN_PARAM, spanId);
    replaceShallowSearchParams(params);
  }, []);

  const visible = React.useMemo<VisibleRow[]>(
    () => flattenTree(roots, expanded, errorsOnly, pathSet),
    [roots, expanded, errorsOnly, pathSet],
  );

  // Order error leaves by visible-row index (unfiltered flat list) so `n`/`p`
  // matches the eye's top-to-bottom scan even in errors-only mode.
  const errorLeavesInOrder = React.useMemo(() => {
    const all = flattenTree(roots, expanded, false, pathSet);
    return all
      .filter((v) => errorLeafSet.has(v.node.span.span_id))
      .map((v) => v.node.span.span_id);
  }, [roots, expanded, pathSet, errorLeafSet]);

  const containerRef = React.useRef<HTMLDivElement | null>(null);
  const rowRefs = React.useRef<Map<string, HTMLDivElement>>(new Map());
  const setRowRef = React.useCallback((id: string, el: HTMLDivElement | null) => {
    if (el) rowRefs.current.set(id, el);
    else rowRefs.current.delete(id);
  }, []);

  const [rail, setRail] = React.useState<{ top: number; height: number } | null>(null);

  const recomputeRail = React.useCallback(() => {
    if (errorLeafIds.length === 0) {
      setRail(null);
      return;
    }
    const ps = pathSetRef.current;
    const inPath = visible.filter((v) => ps.has(v.node.span.span_id));
    const firstRow = inPath[0];
    const lastRow = inPath[inPath.length - 1];
    if (!firstRow || !lastRow) {
      setRail(null);
      return;
    }
    const firstId = firstRow.node.span.span_id;
    const lastId = lastRow.node.span.span_id;
    const first = rowRefs.current.get(firstId);
    const last = rowRefs.current.get(lastId);
    const container = containerRef.current;
    if (!first || !last || !container) {
      setRail(null);
      return;
    }
    const cRect = container.getBoundingClientRect();
    const cTop = cRect.top + container.scrollTop;
    const top = first.getBoundingClientRect().top + container.scrollTop - cTop;
    const bottom = last.getBoundingClientRect().bottom + container.scrollTop - cTop;
    setRail({ top, height: Math.max(0, bottom - top) });
  }, [errorLeafIds.length, visible]);

  React.useLayoutEffect(() => {
    recomputeRail();
  }, [recomputeRail]);

  React.useEffect(() => {
    const container = containerRef.current;
    if (!container) return;
    let raf = 0;
    const schedule = () => {
      if (raf) cancelAnimationFrame(raf);
      raf = requestAnimationFrame(() => {
        raf = 0;
        recomputeRail();
      });
    };
    const RO: typeof ResizeObserver | undefined =
      typeof ResizeObserver === "undefined" ? undefined : ResizeObserver;
    const ro = RO ? new RO(schedule) : null;
    ro?.observe(container);
    window.addEventListener("resize", schedule);
    return () => {
      if (raf) cancelAnimationFrame(raf);
      ro?.disconnect();
      window.removeEventListener("resize", schedule);
    };
  }, [recomputeRail]);

  const scrollRowIntoView = React.useCallback((spanId: string) => {
    const container = containerRef.current;
    const el = rowRefs.current.get(spanId);
    if (!container || !el) return;
    // Manual scroll math — scrollIntoView shifts the drawer and breaks the
    // rail's measured geometry.
    container.scrollTop = el.offsetTop - SCROLL_PAD_PX;
  }, []);

  const selectSpan = React.useCallback(
    (id: string) => {
      writeSelected(id);
    },
    [writeSelected],
  );

  const toggleRow = React.useCallback((id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const expandAll = React.useCallback(() => {
    setExpanded(new Set(structuralIds));
  }, [structuralIds]);

  const collapseAll = React.useCallback(() => {
    setExpanded(new Set());
  }, []);

  const gotoError = React.useCallback(
    (delta: number, cursorOverride?: number) => {
      if (errorLeavesInOrder.length === 0) return;
      const base = cursorOverride ?? errorCursor;
      const next = (base + delta + errorLeavesInOrder.length) % errorLeavesInOrder.length;
      setErrorCursor(next);
      const id = errorLeavesInOrder[next];
      if (!id) return;
      selectSpan(id);
      requestAnimationFrame(() => scrollRowIntoView(id));
    },
    [errorLeavesInOrder, errorCursor, selectSpan, scrollRowIntoView],
  );

  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    const ids = visible.map((v) => v.node.span.span_id);
    const curIdx = selectedSpanId ? ids.indexOf(selectedSpanId) : -1;
    const cur = curIdx >= 0 ? visible[curIdx] : undefined;

    if (e.key === "ArrowDown") {
      e.preventDefault();
      if (ids.length === 0) return;
      const next = Math.min(ids.length - 1, Math.max(0, curIdx + 1));
      const id = ids[next];
      if (id) selectSpan(id);
      return;
    }
    if (e.key === "ArrowUp") {
      e.preventDefault();
      if (ids.length === 0) return;
      const next = Math.max(0, curIdx <= 0 ? 0 : curIdx - 1);
      const id = ids[next];
      if (id) selectSpan(id);
      return;
    }
    if (e.key === "ArrowRight") {
      e.preventDefault();
      if (!cur) return;
      const id = cur.node.span.span_id;
      if (cur.hasChildren && !expanded.has(id)) {
        toggleRow(id);
      } else if (cur.hasChildren) {
        const firstChild = cur.node.children[0];
        if (firstChild) selectSpan(firstChild.span.span_id);
      }
      return;
    }
    if (e.key === "ArrowLeft") {
      e.preventDefault();
      if (!cur) return;
      const id = cur.node.span.span_id;
      if (cur.hasChildren && expanded.has(id)) {
        toggleRow(id);
      } else if (cur.node.parent) {
        selectSpan(cur.node.parent.span.span_id);
      }
      return;
    }
    if (e.key === "n") {
      e.preventDefault();
      gotoError(1);
      return;
    }
    if (e.key === "p") {
      e.preventDefault();
      gotoError(-1);
      return;
    }
    if ((e.metaKey || e.ctrlKey) && (e.key === "c" || e.key === "C")) {
      if (selectedSpanId) {
        void (async () => {
          try {
            await navigator.clipboard.writeText(selectedSpanId);
            toast.success("Span ID copied");
          } catch {
            toast.error("Could not copy to clipboard");
          }
        })();
      }
    }
  };

  const hasRunning = enrichedSpans.some((s) => s.running);
  const hasErrors = errorLeafIds.length > 0;
  const statusSummary = (() => {
    const n = spans.length;
    if (errorsOnly && hasErrors) {
      return (
        <span className="tabular-nums text-muted-foreground">
          {visible.length} of {n} rows
        </span>
      );
    }
    if (hasRunning) {
      return (
        <span className="tabular-nums text-muted-foreground">
          {n} spans · <span className="text-[color:var(--tone-warn-fg)]">running</span>
        </span>
      );
    }
    if (hasErrors) {
      return <span className="tabular-nums text-muted-foreground">{n} spans</span>;
    }
    return (
      <span className="tabular-nums text-muted-foreground">
        {n} spans · <span className="text-[color:var(--tone-ok-fg)]">all ok</span>
      </span>
    );
  })();

  return (
    <div className="flex h-full flex-col" data-testid="trace-tree-tab">
      <div className="mb-3 flex flex-wrap items-center gap-2.5">
        <Button variant="outline" size="sm" onClick={expandAll} aria-label="Expand all">
          <ChevronsUpDown className="h-3.5 w-3.5" aria-hidden="true" />
          <span>Expand all</span>
        </Button>
        <Button variant="outline" size="sm" onClick={collapseAll} aria-label="Collapse all">
          <ChevronsDownUp className="h-3.5 w-3.5" aria-hidden="true" />
          <span>Collapse all</span>
        </Button>
        <label className="ml-1 inline-flex cursor-pointer items-center gap-1.5 text-[12.5px] font-medium text-muted-foreground">
          <input
            type="checkbox"
            checked={errorsOnly}
            onChange={(e) => setErrorsOnly(e.target.checked)}
            data-testid="errors-only-checkbox"
            className="h-3.5 w-3.5 accent-[color:var(--brand)]"
          />
          Errors only
        </label>
        <span className="ml-auto text-[12.5px]">{statusSummary}</span>
      </div>

      {hasErrors && (
        <TraceErrorChip
          count={errorLeafIds.length}
          cursor={errorCursor}
          onPrev={() => gotoError(-1)}
          onNext={() => gotoError(1)}
          onJump={() => gotoError(0)}
        />
      )}

      <div className="flex min-h-0 flex-1 gap-6">
        <div
          ref={containerRef}
          role="tree"
          aria-label="Trace span tree"
          // biome-ignore lint/a11y/noNoninteractiveTabindex: role="tree" must be focusable for keyboard nav.
          tabIndex={0}
          onKeyDown={handleKeyDown}
          data-testid="trace-tree"
          className={cn(
            "relative flex-1 overflow-auto rounded-[10px] border border-[color:var(--border)] bg-[color:var(--card)] py-1.5 outline-none",
            "focus-visible:ring-2 focus-visible:ring-ring",
          )}
        >
          {rail && (
            <div
              aria-hidden="true"
              data-testid="trace-error-rail"
              className="absolute z-[2] rounded-sm"
              style={{
                left: 10,
                top: rail.top,
                height: rail.height,
                width: 3,
                background: "var(--custos-red-500)",
                pointerEvents: "none",
              }}
            />
          )}
          {visible.length === 0 ? (
            <div className="px-6 py-8 text-center text-sm text-muted-foreground">
              No spans match the current filter.
            </div>
          ) : (
            visible.map((row) => {
              const span = row.node.span;
              const tone = rowTone(span);
              const isErrorLeaf = errorLeafSet.has(span.span_id);
              const isOnPath = pathSet.has(span.span_id);
              const isExpanded = expanded.has(span.span_id);
              const hidden = row.hasChildren && !isExpanded && subtreeHasError(row.node);
              return (
                <TraceTreeRow
                  key={span.span_id}
                  row={row}
                  tone={tone}
                  source={readSpanSource(span, traceSource)}
                  isSelected={selectedSpanId === span.span_id}
                  isOnErrorPath={isOnPath}
                  isPreciseFailure={isErrorLeaf}
                  isExpanded={isExpanded}
                  hasHiddenError={hidden}
                  onSelect={() => selectSpan(span.span_id)}
                  onToggle={() => toggleRow(span.span_id)}
                  rowRef={(el) => setRowRef(span.span_id, el)}
                />
              );
            })
          )}
        </div>

        <TraceSpanDetailPanel
          span={selectedNode ? selectedNode.span : null}
          trace={trace}
          source={selectedNode ? readSpanSource(selectedNode.span, traceSource) : traceSource}
          onOpenInRaw={() => onSwitchToTab("raw")}
        />
      </div>
    </div>
  );
}

type TraceErrorChipProps = {
  count: number;
  cursor: number;
  onPrev: () => void;
  onNext: () => void;
  onJump: () => void;
};

function TraceErrorChip({ count, cursor, onPrev, onNext, onJump }: TraceErrorChipProps) {
  return (
    <div
      data-testid="trace-error-chip"
      className="mb-3 flex items-center gap-2.5 rounded-[10px] border px-3 py-2.5 text-[13px] font-semibold"
      style={{
        background: "var(--banner-error-bg)",
        borderColor: "var(--banner-error-border)",
        color: "var(--banner-error-fg)",
      }}
    >
      <AlertTriangle className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
      <span>{count === 1 ? "1 error" : `${count} errors`}</span>
      {count > 1 && (
        <>
          <span aria-hidden="true" className="opacity-50">
            ·
          </span>
          <span className="tabular-nums">
            {cursor + 1} of {count}
          </span>
          <div className="ml-0.5 inline-flex gap-1">
            <button
              type="button"
              onClick={onPrev}
              aria-label="Previous error"
              data-testid="trace-error-chip-prev"
              className="inline-flex h-[22px] w-[22px] items-center justify-center rounded-md border bg-[color:var(--card)] text-[color:var(--banner-error-fg)]"
              style={{ borderColor: "var(--banner-error-border)" }}
            >
              <ArrowUp className="h-3 w-3" aria-hidden="true" />
            </button>
            <button
              type="button"
              onClick={onNext}
              aria-label="Next error"
              data-testid="trace-error-chip-next"
              className="inline-flex h-[22px] w-[22px] items-center justify-center rounded-md border bg-[color:var(--card)] text-[color:var(--banner-error-fg)]"
              style={{ borderColor: "var(--banner-error-border)" }}
            >
              <ArrowDown className="h-3 w-3" aria-hidden="true" />
            </button>
          </div>
        </>
      )}
      <button
        type="button"
        onClick={onJump}
        data-testid="trace-error-chip-jump"
        className="ml-auto inline-flex items-center gap-1 border-none bg-transparent text-[12.5px] font-semibold text-[color:var(--banner-error-fg)] hover:underline"
      >
        jump to row <ArrowRight className="h-3 w-3" aria-hidden="true" />
      </button>
    </div>
  );
}
