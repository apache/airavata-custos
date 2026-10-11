/**
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

"use client";

import { cn } from "@/lib/utils";
import {
  replaceShallowSearchParams,
  useShallowSearchParams,
} from "@/shared/hooks/useShallowSearchParams";
import { EmptyState } from "@/shared/ui/EmptyState";
import { Button } from "@/shared/ui/button";
import * as React from "react";
import type { TraceDetail } from "../types";
import {
  type FlowModel,
  type FlowRow as FlowRowModel,
  buildFlow,
  containersOf,
  defaultExpanded,
  expandableKeys,
  flattenFlow,
  pathRails,
} from "../utils";
import { FlowRow } from "./FlowRow";
import { StepDetailPanel } from "./StepDetailPanel";

export type TraceFlowTabProps = {
  detail: TraceDetail;
};

const STEP_PARAM = "step";

function isNavigable(row: FlowRowModel): boolean {
  return row.kind !== "ghost" && row.kind !== "truncated";
}

function isSelectable(row: FlowRowModel): boolean {
  return row.kind === "step" || row.kind === "group" || row.kind === "hop";
}

// The row a `?span=` link means: the first row of that span.
function rowForSpan(model: FlowModel, spanId: string): string | null {
  return model.spanRows.get(spanId)?.[0]?.step.id ?? null;
}

export function TraceFlowTab({ detail }: TraceFlowTabProps) {
  const params = useShallowSearchParams();
  const model = React.useMemo(() => buildFlow(detail), [detail]);
  const [expanded, setExpanded] = React.useState<Set<string>>(() => defaultExpanded(model));
  const [failingOnly, setFailingOnly] = React.useState(false);
  const [focusedKey, setFocusedKey] = React.useState<string | null>(null);
  const treeRef = React.useRef<HTMLDivElement>(null);

  const stepParam = params.get(STEP_PARAM);
  const spanParam = params.get("span");
  const selectedKey = React.useMemo(() => {
    if (stepParam) return stepParam;
    if (spanParam) {
      const key = rowForSpan(model, spanParam);
      if (key) return key;
    }
    if (model.failingStepId) return model.failingStepId;
    const first = flattenFlow(model, new Set(expandableKeys(model))).find((r) => r.kind === "step");
    return first?.key ?? null;
  }, [stepParam, spanParam, model]);

  // A fresh model (new trace, or a refresh) resets what is open.
  const modelRef = React.useRef(model);
  React.useEffect(() => {
    if (modelRef.current === model) return;
    modelRef.current = model;
    setExpanded(defaultExpanded(model));
  }, [model]);

  // Whatever is selected must be in view, so its containers open.
  React.useEffect(() => {
    if (!selectedKey) return;
    const needed = containersOf(model, selectedKey);
    setExpanded((prev) => {
      if (needed.every((k) => prev.has(k))) return prev;
      const next = new Set(prev);
      for (const k of needed) next.add(k);
      return next;
    });
  }, [model, selectedKey]);

  const allRows = React.useMemo(() => flattenFlow(model, expanded), [model, expanded]);
  const rows = React.useMemo(() => {
    if (!failingOnly || !model.failingStepId) return allRows;
    return allRows.filter(
      (r) =>
        r.kind === "truncated" ||
        ("onPath" in r && r.onPath) ||
        (r.kind === "step" && r.failing) ||
        r.key === selectedKey,
    );
  }, [allRows, failingOnly, model.failingStepId, selectedKey]);
  const rails = React.useMemo(() => pathRails(rows), [rows]);
  const navigable = React.useMemo(() => rows.filter(isNavigable), [rows]);

  const select = React.useCallback((key: string) => {
    const next = new URLSearchParams(window.location.search);
    next.set(STEP_PARAM, key);
    next.delete("span");
    replaceShallowSearchParams(next);
    setFocusedKey(key);
  }, []);

  const toggle = React.useCallback((key: string) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  }, []);

  // Scroll the selected row into view when the selection comes from the URL
  // or the keyboard, not from a click that is already in view.
  const lastScrolled = React.useRef<string | null>(null);
  // biome-ignore lint/correctness/useExhaustiveDependencies: the row may only exist once `rows` has rendered it
  React.useEffect(() => {
    if (!selectedKey || lastScrolled.current === selectedKey) return;
    const el = treeRef.current?.querySelector<HTMLElement>(`[data-row-key="${selectedKey}"]`);
    if (!el) return;
    lastScrolled.current = selectedKey;
    el.scrollIntoView({ block: "center" });
  }, [selectedKey, rows]);

  const onKeyDown = (e: React.KeyboardEvent) => {
    const current = focusedKey ?? selectedKey;
    const index = navigable.findIndex((r) => r.key === current);
    const row = index >= 0 ? navigable[index] : undefined;
    const moveTo = (i: number) => {
      const target = navigable[Math.max(0, Math.min(navigable.length - 1, i))];
      if (target) setFocusedKey(target.key);
    };
    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        moveTo(index + 1);
        break;
      case "ArrowUp":
        e.preventDefault();
        moveTo(index - 1);
        break;
      case "Home":
        e.preventDefault();
        moveTo(0);
        break;
      case "End":
        e.preventDefault();
        moveTo(navigable.length - 1);
        break;
      case "ArrowRight":
        e.preventDefault();
        if (row?.kind === "group" && !row.expanded) toggle(row.key);
        else if (row?.kind === "fold") toggle(row.key);
        else moveTo(index + 1);
        break;
      case "ArrowLeft": {
        e.preventDefault();
        if (row?.kind === "group" && row.expanded) {
          toggle(row.key);
          break;
        }
        for (let i = index - 1; i >= 0 && row; i--) {
          const above = navigable[i];
          if (above && above.depth < row.depth) {
            setFocusedKey(above.key);
            break;
          }
        }
        break;
      }
      case "Enter":
      case " ":
        e.preventDefault();
        if (row && isSelectable(row)) select(row.key);
        else if (row?.kind === "fold") toggle(row.key);
        break;
      default:
        break;
    }
  };

  const expandAll = () => setExpanded(new Set(expandableKeys(model)));
  const collapseAll = () => setExpanded(new Set());

  const selectedRow = allRows.find((r) => r.key === selectedKey) ?? null;
  const focusTarget =
    focusedKey && navigable.some((r) => r.key === focusedKey) ? focusedKey : selectedKey;

  if (detail.tree.length === 0) {
    return (
      <EmptyState
        heading="No steps recorded"
        description="This trace has no audit rows yet."
        className="mt-6"
      />
    );
  }

  // Each step row shows its source pill only when it differs from its group.
  let groupSource = "";

  return (
    <div className="@container flex h-full min-h-0 flex-col">
      <div className="flex flex-wrap items-center gap-2 py-2 text-[12.5px]">
        <Button variant="ghost" size="sm" onClick={expandAll}>
          Expand all
        </Button>
        <Button variant="ghost" size="sm" onClick={collapseAll}>
          Collapse all
        </Button>
        <label className="ml-1 inline-flex items-center gap-1.5 text-foreground">
          <input
            type="checkbox"
            checked={failingOnly}
            disabled={!model.failingStepId}
            onChange={(e) => setFailingOnly(e.target.checked)}
            className="size-3.5 accent-[color:var(--brand)]"
          />
          Show failing path only
        </label>
      </div>

      <div className="flex min-h-0 flex-1 flex-col gap-3 @3xl:flex-row">
        <div
          ref={treeRef}
          role="tree"
          aria-label="Trace steps"
          data-testid="trace-flow"
          tabIndex={-1}
          onKeyDown={onKeyDown}
          className={cn(
            "min-w-0 overflow-auto rounded-[10px] border border-border bg-card py-1",
            "h-[45vh] shrink-0 @3xl:h-auto @3xl:min-h-0 @3xl:flex-1 @3xl:shrink",
          )}
        >
          {rows.map((row, i) => {
            if (row.kind === "group") groupSource = row.source;
            return (
              <FlowRow
                key={row.key}
                row={row}
                rails={rails[i] ?? []}
                groupSource={groupSource}
                startedAt={model.startedAt}
                selected={row.key === selectedKey}
                focused={row.key === focusTarget}
                onSelect={select}
                onToggle={toggle}
                onFocus={setFocusedKey}
              />
            );
          })}
        </div>
        <StepDetailPanel
          row={selectedRow}
          model={model}
          traceId={detail.trace_id}
          className="min-h-0 shrink-0 overflow-auto @3xl:w-[340px]"
        />
      </div>
    </div>
  );
}
