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

"use client";

import { AlertTriangle, ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { type RowTone, type VisibleRow, isCodeShaped } from "../utils";
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
  onSelect: () => void;
  onToggle: () => void;
  rowRef: (el: HTMLDivElement | null) => void;
};

export function TraceTreeRow({
  row,
  tone,
  source,
  isSelected,
  isOnErrorPath,
  isPreciseFailure,
  isExpanded,
  hasHiddenError,
  onSelect,
  onToggle,
  rowRef,
}: TraceTreeRowProps) {
  const { node, depth, hasChildren } = row;
  const span = node.span;
  const isError = tone === "error";
  const capped = Math.min(depth, MAX_DEPTH);
  const overCap = depth > MAX_DEPTH;
  const name = span.event_type;
  const code = isCodeShaped(name);
  const display = overCap ? `…/${name}` : name;
  const entityMeta = span.entity_id ? `${span.entity_type}=${span.entity_id}` : undefined;

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

      <span className="ml-0.5 shrink-0">
        <SourcePill source={source} />
      </span>

      <span
        className="ml-auto flex shrink-0 items-center gap-2 overflow-hidden pl-2 text-[12px] whitespace-nowrap text-muted-foreground"
        style={{ maxWidth: "46%" }}
      >
        {entityMeta && (
          <span className="overflow-hidden font-mono text-ellipsis">{entityMeta}</span>
        )}
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
