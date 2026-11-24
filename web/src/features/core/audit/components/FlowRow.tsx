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
  AlertTriangleIcon,
  ChevronRightIcon,
  CircleCheckIcon,
  CircleXIcon,
  PlayIcon,
  RadioIcon,
  RotateCcwIcon,
  SendIcon,
  UnlinkIcon,
} from "lucide-react";
import * as React from "react";
import type { FlowRow as FlowRowModel } from "../utils";
import {
  deliveryLabel,
  errorText,
  formatAbsoluteUtc,
  formatOffset,
  isCodeShaped,
  plainName,
  shortHex,
  usernameOf,
} from "../utils";
import { SourcePill } from "./primitives/SourcePill";
import { StatusPill } from "./primitives/StatusPill";

export const INDENT = 22;
const GUTTER = 16;

export type FlowRowProps = {
  row: FlowRowModel;
  rails: boolean[];
  groupSource: string;
  startedAt: string;
  selected: boolean;
  focused: boolean;
  onSelect(key: string): void;
  onToggle(key: string): void;
  onFocus(key: string): void;
};

function Rails({ rails, depth, onPath }: { rails: boolean[]; depth: number; onPath: boolean }) {
  return (
    <>
      {rails.map((red, i) => (
        <span
          // biome-ignore lint/suspicious/noArrayIndexKey: rails are positional
          key={i}
          aria-hidden="true"
          className={cn(
            "absolute top-0 bottom-0 w-0.5",
            red ? "bg-[color:var(--custos-red-500)]" : "bg-[color:var(--border-strong)]",
          )}
          style={{ left: i * INDENT + GUTTER / 2 - 1 }}
        />
      ))}
      {depth > 0 && (
        <span
          aria-hidden="true"
          className={cn(
            "absolute top-1/2 h-0.5 -translate-y-1/2",
            onPath ? "bg-[color:var(--custos-red-500)]" : "bg-[color:var(--border-strong)]",
          )}
          style={{ left: (depth - 1) * INDENT + GUTTER / 2, width: INDENT - GUTTER / 2 }}
        />
      )}
    </>
  );
}

function Marker({ children }: { children: React.ReactNode }) {
  return (
    <span
      aria-hidden="true"
      className="relative z-[1] flex size-4 shrink-0 items-center justify-center bg-card"
    >
      {children}
    </span>
  );
}

const rowBase =
  "relative flex w-full cursor-pointer items-center gap-2 pr-2 text-left outline-none select-none";
const rowFocus = "focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset";
const rowSelected =
  "bg-[color:var(--brand-tint)] shadow-[inset_0_0_0_1px_var(--brand)] before:absolute before:inset-y-0 before:left-0 before:w-[2.5px] before:bg-[color:var(--brand)]";

export function FlowRow(props: FlowRowProps) {
  const { row, rails, selected, focused, onSelect, onToggle, onFocus } = props;
  const ref = React.useRef<HTMLDivElement>(null);
  React.useEffect(() => {
    if (focused) ref.current?.focus({ preventScroll: true });
  }, [focused]);

  if (row.kind === "truncated") {
    return (
      <div className="mt-2 flex items-center gap-2 rounded-md bg-[color:var(--tone-warn-bg)] px-3 py-2 text-[12.5px] text-[color:var(--tone-warn-fg)]">
        <AlertTriangleIcon className="size-3.5 shrink-0" aria-hidden="true" />
        Showing the first 500 steps. Open Raw for all of them.
      </div>
    );
  }

  const indent = row.depth * INDENT;
  const onPath = "onPath" in row && row.onPath;

  if (row.kind === "ghost") {
    return (
      <div className="relative flex min-h-7 items-center gap-2 pr-2">
        <Rails rails={rails} depth={row.depth} onPath={false} />
        <span style={{ width: indent }} className="shrink-0" />
        <Marker>
          <span className="size-2 rounded-full ring-[1.5px] ring-inset ring-[color:var(--muted-foreground)]" />
        </Marker>
        <span className="truncate text-[12.5px] italic text-muted-foreground">{row.text}</span>
      </div>
    );
  }

  const common = {
    ref,
    tabIndex: focused ? 0 : -1,
    "data-row-key": row.key,
    onFocus: () => onFocus(row.key),
  };

  if (row.kind === "fold") {
    return (
      <div
        {...common}
        role="treeitem"
        aria-level={row.depth + 1}
        aria-expanded={false}
        aria-selected={false}
        onClick={() => onToggle(row.key)}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            onToggle(row.key);
          }
        }}
        className={cn(rowBase, rowFocus, "min-h-7 hover:bg-muted/60")}
      >
        <Rails rails={rails} depth={row.depth} onPath={onPath} />
        <span style={{ width: indent }} className="shrink-0" />
        <Marker>
          <ChevronRightIcon className="size-3.5 text-muted-foreground" />
        </Marker>
        <span className="text-[12.5px] font-medium text-muted-foreground">
          {row.count} more {row.count === 1 ? "try" : "tries"}
        </span>
      </div>
    );
  }

  if (row.kind === "group") {
    const Icon = row.delivery ? SendIcon : row.status === "muted" ? UnlinkIcon : PlayIcon;
    const isFailed = row.status === "failed" || row.status === "error";
    return (
      <div
        {...common}
        role="treeitem"
        aria-level={row.depth + 1}
        aria-expanded={row.expanded}
        aria-selected={selected}
        onClick={() => onSelect(row.key)}
        className={cn(
          rowBase,
          rowFocus,
          "sticky top-0 z-[2] min-h-9 bg-card hover:bg-muted/60",
          selected && rowSelected,
        )}
      >
        <Rails rails={rails} depth={row.depth} onPath={onPath} />
        <span style={{ width: indent }} className="shrink-0" />
        <button
          type="button"
          tabIndex={-1}
          aria-label={row.expanded ? "Collapse" : "Expand"}
          onClick={(e) => {
            e.stopPropagation();
            onToggle(row.key);
          }}
          className="relative z-[1] -ml-1 flex size-5 shrink-0 items-center justify-center rounded bg-card text-muted-foreground hover:text-foreground"
        >
          <ChevronRightIcon
            className={cn(
              "size-3.5 transition-transform duration-100",
              row.expanded && "rotate-90",
            )}
            aria-hidden="true"
          />
        </button>
        <Marker>
          <Icon
            className={cn(
              "size-3.5",
              isFailed
                ? "text-[color:var(--tone-error-fg)]"
                : row.status === "muted"
                  ? "text-muted-foreground"
                  : "text-foreground",
            )}
          />
        </Marker>
        <span
          className={cn(
            "min-w-0 truncate text-[13px] font-semibold text-foreground",
            row.status === "muted" && "font-medium text-muted-foreground",
            isCodeShaped(row.title) && "font-mono text-[12.5px]",
          )}
          title={row.title}
        >
          {row.delivery ? (
            <>
              <span className="font-medium text-muted-foreground">Delivery to </span>
              <span className="font-mono text-[12.5px]">{row.title}</span>
            </>
          ) : (
            row.title
          )}
        </span>
        {row.source && <SourcePill source={row.source} size="sm" />}
        {row.status !== "muted" && (
          <StatusPill
            status={row.status}
            label={row.delivery ? deliveryLabel(row.delivery) : undefined}
            size="sm"
          />
        )}
      </div>
    );
  }

  if (row.kind === "hop") {
    const done = row.deliveries.filter((d) => d.status === "SUCCEEDED").length;
    const failed = row.deliveries.filter((d) => d.status === "FAILED").length;
    const pending = row.deliveries.length - done - failed;
    const parts = [
      done > 0 && `${done} done`,
      failed > 0 && `${failed} failed`,
      pending > 0 && `${pending} pending`,
    ].filter(Boolean);
    return (
      <div
        {...common}
        role="treeitem"
        aria-level={row.depth + 1}
        aria-selected={selected}
        onClick={() => onSelect(row.key)}
        className={cn(rowBase, rowFocus, "min-h-8 hover:bg-muted/60", selected && rowSelected)}
      >
        <Rails rails={rails} depth={row.depth} onPath={onPath} />
        <span style={{ width: indent }} className="shrink-0" />
        <Marker>
          <RadioIcon className="size-3.5 text-[color:var(--tone-info-fg)]" />
        </Marker>
        <span className="min-w-0 truncate text-[13px] text-foreground">
          Published <span className="font-mono text-[12.5px]">{row.eventType}</span> to{" "}
          {row.deliveries.length} {row.deliveries.length === 1 ? "connector" : "connectors"}
        </span>
        <span className="ml-auto shrink-0 text-[12px] tabular-nums text-muted-foreground">
          {parts.join(", ")}
        </span>
      </div>
    );
  }

  const { step } = row;
  const isError = step.status === "error";
  const label = plainName(step.event_type) ?? step.event_type;
  const err = errorText(step);
  const username = usernameOf(step);
  const meta = err
    ? err
    : username
      ? username
      : step.entity_type && step.entity_id
        ? `${step.entity_type} ${shortHex(step.entity_id, 8)}`
        : "";
  const offset = Date.parse(step.created_at) - Date.parse(props.startedAt);
  let marker: React.ReactNode;
  if (row.try) {
    marker =
      row.try.outcome === "succeeded" ? (
        <CircleCheckIcon className="size-3.5 text-[color:var(--tone-ok-fg)]" />
      ) : row.try.outcome === "retried" ? (
        <RotateCcwIcon className="size-3.5 text-[color:var(--brand)]" />
      ) : (
        <CircleXIcon className="size-3.5 text-[color:var(--tone-error-fg)]" />
      );
  } else if (isError) {
    marker = <CircleXIcon className="size-3.5 text-[color:var(--tone-error-fg)]" />;
  } else {
    marker = <span className="size-2 rounded-full bg-[color:var(--custos-gray-400)]" />;
  }

  return (
    <div
      {...common}
      role="treeitem"
      aria-level={row.depth + 1}
      aria-selected={selected}
      aria-current={row.failing ? "true" : undefined}
      data-testid={row.failing ? "failing-step" : undefined}
      onClick={() => onSelect(row.key)}
      className={cn(
        rowBase,
        rowFocus,
        "min-h-8 hover:bg-muted/60",
        isError && !selected && "bg-[color:var(--tone-error-bg)]",
        row.failing &&
          !selected &&
          "before:absolute before:inset-y-0 before:left-0 before:w-[3px] before:bg-[color:var(--custos-red-500)]",
        selected && rowSelected,
      )}
    >
      <Rails rails={rails} depth={row.depth} onPath={onPath} />
      <span style={{ width: indent }} className="shrink-0" />
      <Marker>{marker}</Marker>
      <span
        className="w-[60px] shrink-0 text-right font-mono text-[11.5px] tabular-nums text-muted-foreground"
        title={formatAbsoluteUtc(step.created_at)}
      >
        {formatOffset(offset)}
      </span>
      {isError && <span className="sr-only">Failed:</span>}
      <span
        className={cn(
          "min-w-0 flex-1 truncate text-[13px] font-medium text-foreground",
          onPath && "font-semibold",
          isError && "text-[color:var(--tone-error-fg)]",
          isCodeShaped(label) && "font-mono text-[12.5px]",
        )}
        title={step.event_type}
      >
        {row.try ? `Try ${row.try.n} ${row.try.outcome}` : label}
      </span>
      {!row.try && step.source !== props.groupSource && (
        <SourcePill source={step.source} size="sm" />
      )}
      {meta && (
        <span
          className={cn(
            "hidden max-w-[45%] shrink-0 truncate text-[12px] @md:inline",
            err ? "font-mono text-[color:var(--tone-error-fg)]" : "font-mono text-muted-foreground",
          )}
          title={meta}
        >
          {meta}
        </span>
      )}
    </div>
  );
}
