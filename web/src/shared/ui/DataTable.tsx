"use client";

import { cn } from "@/lib/utils";
import { Button } from "@/shared/ui/button";
import { ChevronDown, ChevronsUpDown, ChevronUp } from "lucide-react";
import * as React from "react";

export type DataTableSortValue = string | number | Date | null | undefined;

export type DataTableColumn<T> = {
  key: string;
  header: React.ReactNode;
  cell: (row: T, index: number) => React.ReactNode;
  width?: string;
  align?: "left" | "right" | "center";
  // Set true when the cell already renders its own interactive control (a
  // checkbox, a button, a link). The DataTable will then skip wrapping the cell
  // in its keyboard-activation <button>, which would otherwise nest controls.
  interactive?: boolean;
  // Mark the column header as a sort toggle. Requires `sortValue`. The toggle
  // cycles none → asc → desc → none.
  sortable?: boolean;
  sortValue?: (row: T) => DataTableSortValue;
};

type SortState = { key: string; direction: "asc" | "desc" } | null;

export type DataTablePagination = {
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange?: (size: number) => void;
  pageSizeOptions?: number[];
  // Hard-disable Next beyond `page >= totalPages` — used when the backend
  // caps cumulative offset (e.g. tracing's 1M offset ceiling).
  nextDisabled?: boolean;
};

export type DataTableProps<T> = {
  columns: Array<DataTableColumn<T>>;
  rows: T[];
  rowKey: (row: T, index: number) => string;
  onRowClick?: (row: T) => void;
  rowClassName?: (row: T) => string | undefined;
  caption?: string;
  empty?: React.ReactNode;
  pagination?: DataTablePagination;
  className?: string;
};

function compareSortable(
  a: DataTableSortValue,
  b: DataTableSortValue,
  direction: "asc" | "desc",
): number {
  // Nullish values always sort last regardless of direction so a missing
  // entry doesn't bubble to the top under desc.
  if (a == null && b == null) return 0;
  if (a == null) return 1;
  if (b == null) return -1;
  let cmp = 0;
  if (a instanceof Date && b instanceof Date) {
    cmp = a.getTime() - b.getTime();
  } else if (typeof a === "number" && typeof b === "number") {
    cmp = a - b;
  } else {
    cmp = String(a).localeCompare(String(b), undefined, { numeric: true, sensitivity: "base" });
  }
  return direction === "asc" ? cmp : -cmp;
}

export function DataTable<T>({
  columns,
  rows,
  rowKey,
  onRowClick,
  rowClassName,
  caption,
  empty,
  pagination,
  className,
}: DataTableProps<T>) {
  const [sortState, setSortState] = React.useState<SortState>(null);

  const sortedRows = React.useMemo(() => {
    if (!sortState) return rows;
    const col = columns.find((c) => c.key === sortState.key);
    if (!col?.sortValue) return rows;
    const sortValue = col.sortValue;
    return [...rows]
      .map((row, i) => ({ row, i }))
      .sort((a, b) => {
        const cmp = compareSortable(sortValue(a.row), sortValue(b.row), sortState.direction);
        // Preserve original order for equal keys so re-sorts feel stable.
        return cmp !== 0 ? cmp : a.i - b.i;
      })
      .map((entry) => entry.row);
  }, [rows, sortState, columns]);

  const cycleSort = (key: string) => {
    setSortState((prev) => {
      if (!prev || prev.key !== key) return { key, direction: "asc" };
      if (prev.direction === "asc") return { key, direction: "desc" };
      return null;
    });
  };

  return (
    <div
      className={cn(
        "overflow-hidden rounded-lg border border-border bg-card",
        className,
      )}
    >
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          {caption ? <caption className="sr-only">{caption}</caption> : null}
          <thead className="bg-muted text-xs font-medium uppercase tracking-wide text-muted-foreground">
            <tr>
              {columns.map((col) => {
                const isSorted = sortState?.key === col.key;
                const ariaSort: React.AriaAttributes["aria-sort"] = isSorted
                  ? sortState.direction === "asc"
                    ? "ascending"
                    : "descending"
                  : col.sortable
                    ? "none"
                    : undefined;
                return (
                  <th
                    key={col.key}
                    scope="col"
                    style={col.width ? { width: col.width } : undefined}
                    aria-sort={ariaSort}
                    className={cn(
                      "px-4 py-3 font-medium",
                      col.align === "right" && "text-right",
                      col.align === "center" && "text-center",
                    )}
                  >
                    {col.sortable && col.sortValue ? (
                      <button
                        type="button"
                        onClick={() => cycleSort(col.key)}
                        className={cn(
                          "inline-flex items-center gap-1 rounded-sm text-xs font-medium uppercase tracking-wide text-muted-foreground transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring",
                          col.align === "right" && "flex-row-reverse",
                          col.align === "center" && "justify-center",
                        )}
                      >
                        <span>{col.header}</span>
                        {isSorted ? (
                          sortState.direction === "asc" ? (
                            <ChevronUp className="h-3.5 w-3.5 text-foreground" aria-hidden />
                          ) : (
                            <ChevronDown className="h-3.5 w-3.5 text-foreground" aria-hidden />
                          )
                        ) : (
                          <ChevronsUpDown className="h-3.5 w-3.5 opacity-40" aria-hidden />
                        )}
                      </button>
                    ) : (
                      col.header
                    )}
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {sortedRows.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="px-4 py-12 text-center">
                  {empty ?? (
                    <span className="text-sm text-muted-foreground">
                      No data.
                    </span>
                  )}
                </td>
              </tr>
            ) : (
              sortedRows.map((row, index) => (
                <tr
                  key={rowKey(row, index)}
                  className={cn(
                    "border-t border-border/60 bg-card",
                    onRowClick && "cursor-pointer",
                    rowClassName?.(row),
                  )}
                  onClick={
                    onRowClick
                      ? (e) => {
                          // Ignore clicks on interactive children (buttons,
                          // checkboxes, links) — they have their own handlers.
                          // Also covers portal-rendered content (dropdown
                          // menus, popovers): the clicked DOM node's
                          // ancestors won't include the <tr>, but React's
                          // synthetic events still bubble through the
                          // *component* tree, so without this the row click
                          // fires alongside the menu item's own click.
                          const target = e.target as HTMLElement;
                          if (
                            target.closest(
                              "button, a, input, select, textarea, label, [role='menu'], [role='menuitem'], [role='menuitemcheckbox'], [role='menuitemradio'], [role='dialog'], [role='listbox'], [role='option']",
                            )
                          )
                            return;
                          onRowClick(row);
                        }
                      : undefined
                  }
                  onKeyDown={
                    onRowClick
                      ? (e) => {
                          if (
                            e.key === "Enter" &&
                            e.target === e.currentTarget
                          ) {
                            onRowClick(row);
                          }
                        }
                      : undefined
                  }
                >
                  {(() => {
                    const firstWrappable = onRowClick
                      ? columns.findIndex((c) => !c.interactive)
                      : -1;
                    return columns.map((col, ci) => {
                      const wrap = ci === firstWrappable;
                      const content = col.cell(row, index);
                      return (
                        <td
                          key={col.key}
                          className={cn(
                            "px-4 py-3 align-top text-sm text-foreground",
                            col.align === "right" && "text-right",
                            col.align === "center" && "text-center",
                          )}
                        >
                          {wrap ? (
                            // A real button avoids the nested-interactive a11y warning
                            // that role="button" on the <tr> triggered.
                            <button
                              type="button"
                              className="-mx-1 inline-flex items-center rounded-sm bg-transparent px-1 text-left focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                              onClick={(e) => {
                                e.stopPropagation();
                                onRowClick?.(row);
                              }}
                            >
                              {content}
                            </button>
                          ) : (
                            content
                          )}
                        </td>
                      );
                    });
                  })()}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      {pagination ? <DataTablePager pagination={pagination} /> : null}
    </div>
  );
}

function DataTablePager({ pagination }: { pagination: DataTablePagination }) {
  const {
    page,
    pageSize,
    total,
    onPageChange,
    onPageSizeChange,
    pageSizeOptions = [10, 20, 50, 100],
    nextDisabled,
  } = pagination;
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const start = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const end = Math.min(total, page * pageSize);
  return (
    <div className="flex items-center justify-between border-t border-border/60 bg-card px-4 py-3 text-xs text-muted-foreground">
      <span>
        Showing {start}–{end} of {total}
      </span>
      <div className="flex items-center gap-4">
        {onPageSizeChange && (
          <div className="flex items-center gap-2">
            <span>Rows per page</span>
            <div className="relative">
              <select
                value={pageSize}
                onChange={(e) => {
                  onPageSizeChange(Number(e.target.value));
                  onPageChange(1);
                }}
                aria-label="Rows per page"
                className="h-8 appearance-none rounded-md border border-border bg-background pl-2 pr-8 text-xs"
              >
                {pageSizeOptions.map((n) => (
                  <option key={n} value={n}>
                    {n}
                  </option>
                ))}
              </select>
              <ChevronDown className="pointer-events-none absolute right-2 top-1/2 h-3 w-3 -translate-y-1/2 text-muted-foreground" />
            </div>
          </div>
        )}
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(Math.max(1, page - 1))}
            disabled={page <= 1}
          >
            Previous
          </Button>
          {/* Active page is solid primary; sibling prev/next are outlined. */}
          <span
            aria-current="page"
            className="inline-flex h-8 min-w-8 items-center justify-center rounded-md bg-primary px-3 text-[0.8rem] font-semibold text-primary-foreground"
          >
            {page}
          </span>
          <span className="text-muted-foreground">of {totalPages}</span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(Math.min(totalPages, page + 1))}
            disabled={page >= totalPages || Boolean(nextDisabled)}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
