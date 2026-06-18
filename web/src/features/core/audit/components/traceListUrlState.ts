// URL <-> filter mapping for the trace list. Hardcoded whitelists keep
// parseFilters tolerant of pasted/manipulated URLs — stray values collapse to
// defaults instead of widening the TanStack cache key.

export type StatusFilter = "error" | "ok" | "in-progress" | "orphaned";
export type WindowPreset = "24h" | "7d" | "30d";

export type ListFilters = {
  status: StatusFilter[];
  source: string[];
  window: WindowPreset;
  q: string;
  page: number;
  pageSize: number;
  failingOver24h: boolean;
};

export const DEFAULT_FILTERS: ListFilters = {
  status: ["error"],
  source: [],
  window: "30d",
  q: "",
  page: 1,
  pageSize: 50,
  failingOver24h: false,
};

const VALID_STATUS: ReadonlyArray<StatusFilter> = ["error", "ok", "in-progress", "orphaned"];
const VALID_SOURCES: ReadonlyArray<string> = ["amie", "comanage", "slurm", "http", "core"];
const VALID_WINDOWS: ReadonlyArray<WindowPreset> = ["24h", "7d", "30d"];
const VALID_PAGE_SIZES: ReadonlyArray<number> = [25, 50, 100];

// API status codes: ok=0, error=1, cancelled=2, orphaned=3. `in-progress` is
// a UI-only filter (ended_at == null); never sent to the wire.
const STATUS_TO_API: Record<StatusFilter, number | null> = {
  ok: 0,
  error: 1,
  orphaned: 3,
  "in-progress": null,
};

type SearchParamsLike = {
  getAll: (key: string) => string[];
  get: (key: string) => string | null;
};

const DAY_MS = 24 * 60 * 60 * 1000;

export function parseFilters(params: SearchParamsLike): ListFilters {
  const rawStatus = params.getAll("status");
  const status = rawStatus.length
    ? (rawStatus.filter((s): s is StatusFilter =>
        VALID_STATUS.includes(s as StatusFilter),
      ) as StatusFilter[])
    : DEFAULT_FILTERS.status;
  const source = params.getAll("source").filter((s) => VALID_SOURCES.includes(s));
  const winRaw = params.get("window");
  const window: WindowPreset = VALID_WINDOWS.includes(winRaw as WindowPreset)
    ? (winRaw as WindowPreset)
    : DEFAULT_FILTERS.window;
  const q = (params.get("q") ?? "").trim();
  const pageRaw = Number.parseInt(params.get("page") ?? "", 10);
  const page = Number.isFinite(pageRaw) && pageRaw >= 1 ? pageRaw : DEFAULT_FILTERS.page;
  const pageSizeRaw = Number.parseInt(params.get("pageSize") ?? "", 10);
  const pageSize = VALID_PAGE_SIZES.includes(pageSizeRaw)
    ? pageSizeRaw
    : DEFAULT_FILTERS.pageSize;
  const failingOver24h = params.get("failingOver24h") === "1";
  return { status, source, window, q, page, pageSize, failingOver24h };
}

function arraysEqual<T>(a: ReadonlyArray<T>, b: ReadonlyArray<T>): boolean {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
  return true;
}

export function serializeFilters(filters: ListFilters): URLSearchParams {
  const params = new URLSearchParams();
  const sortedStatus = [...filters.status].sort();
  if (!arraysEqual(sortedStatus, [...DEFAULT_FILTERS.status].sort())) {
    for (const s of sortedStatus) params.append("status", s);
  }
  for (const s of [...filters.source].sort()) params.append("source", s);
  if (filters.window !== DEFAULT_FILTERS.window) params.set("window", filters.window);
  if (filters.q) params.set("q", filters.q);
  if (filters.page !== DEFAULT_FILTERS.page) params.set("page", String(filters.page));
  if (filters.pageSize !== DEFAULT_FILTERS.pageSize) {
    params.set("pageSize", String(filters.pageSize));
  }
  if (filters.failingOver24h) params.set("failingOver24h", "1");
  return params;
}

export function hasActiveFilters(filters: ListFilters): boolean {
  const statusChanged = !arraysEqual(
    [...filters.status].sort(),
    [...DEFAULT_FILTERS.status].sort(),
  );
  return (
    statusChanged ||
    filters.source.length > 0 ||
    filters.window !== DEFAULT_FILTERS.window ||
    filters.q.length > 0 ||
    filters.failingOver24h
  );
}

export function windowToFromTo(win: WindowPreset, now: number): { from: string; to: string } {
  const days = win === "24h" ? 1 : win === "7d" ? 7 : 30;
  const to = new Date(now).toISOString();
  const from = new Date(now - days * DAY_MS).toISOString();
  return { from, to };
}

// Map UI status filters to backend numeric codes; `in-progress` is filtered
// client-side and contributes no API status param.
export function statusFiltersToApi(status: StatusFilter[]): {
  apiStatus: number[];
  inProgressOnly: boolean;
} {
  const apiStatus: number[] = [];
  let hasInProgress = false;
  for (const s of status) {
    const code = STATUS_TO_API[s];
    if (code == null) {
      hasInProgress = true;
    } else {
      apiStatus.push(code);
    }
  }
  const inProgressOnly = hasInProgress && apiStatus.length === 0;
  return { apiStatus, inProgressOnly };
}

// Bounds shared by the 24h failure banner and the "Failing >24h" filter:
// traces that started between 30 days ago and 24 hours ago.
export function bannerBounds(now: number): { from: string; to: string } {
  return {
    from: new Date(now - 30 * 24 * 60 * 60 * 1000).toISOString(),
    to: new Date(now - 24 * 60 * 60 * 1000).toISOString(),
  };
}
