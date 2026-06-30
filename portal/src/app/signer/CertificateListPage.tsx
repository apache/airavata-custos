// Certificate list view at /signer/certificates. Owns the filter controls
// (status, allocation, time range, username) and the paginated table.
// Server-side filtering is limited to username today; the rest filter
// client-side within the current page window.
"use client";

import type React from "react";
import { useMemo, useState } from "react";
import Link from "next/link";
import { ChevronLeft, ChevronRight, RefreshCw } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { Certificate, CertificateStatus } from "./types";
import {
  formatCertificateStatus,
  formatRemainingTime,
  formatUnixClock,
  formatUnixDate,
  getCertificateAllocation,
  getCertificateStatus,
  getRemainingSeconds,
} from "./types";
import { useCertificates } from "./hooks";

// Client-side time-range filter over cert.issued_at. Same scope caveat as the
// status and allocation filters: only narrows within the current page window.
// TODO: push these filters server-side once the signer API supports them.
const timeRangeOptions = [
  { label: "Last 24 hrs", seconds: 24 * 60 * 60 },
  { label: "Last 7 days", seconds: 7 * 24 * 60 * 60 },
  { label: "Last 30 days", seconds: 30 * 24 * 60 * 60 },
  { label: "All time", seconds: 0 },
];

export function CertificateListPage() {
  const [statusFilter, setStatusFilter] = useState<CertificateStatus | "all">(
    "all"
  );
  const [allocationFilter, setAllocationFilter] = useState("all");
  const [timeRange, setTimeRange] = useState(timeRangeOptions[3].label);
  const [usernameDraft, setUsernameDraft] = useState("");
  const [usernameFilter, setUsernameFilter] = useState("");
  const [page, setPage] = useState(1);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const { data, loading, error, reload } = useCertificates(
    rowsPerPage,
    (page - 1) * rowsPerPage,
    usernameFilter
  );

  const certificates = useMemo(() => data?.certificates ?? [], [data]);
  const allocationOptions = useMemo(
    () => Array.from(new Set(certificates.map(getCertificateAllocation))).sort(),
    [certificates]
  );
  const timeRangeSeconds =
    timeRangeOptions.find((option) => option.label === timeRange)?.seconds ?? 0;
  const nowSeconds = Math.floor(Date.now() / 1000);
  const visibleCertificates = certificates.filter((cert) => {
    if (statusFilter !== "all" && getCertificateStatus(cert) !== statusFilter) {
      return false;
    }

    if (
      allocationFilter !== "all" &&
      getCertificateAllocation(cert) !== allocationFilter
    ) {
      return false;
    }

    if (timeRangeSeconds > 0 && cert.issued_at < nowSeconds - timeRangeSeconds) {
      return false;
    }

    return true;
  });
  const totalRows = data?.total ?? visibleCertificates.length;
  const totalPages = Math.max(1, Math.ceil(totalRows / rowsPerPage));
  const pageWindow = buildPageWindow(page, totalPages);

  function updateRowsPerPage(value: number) {
    setRowsPerPage(value);
    setPage(1);
  }

  if (loading && !data) return <CertificateListSkeleton />;
  if (error) return <p className="text-red-600">{error}</p>;

  return (
    <section className="mx-auto max-w-[1440px] space-y-6">
      <div className="max-w-4xl">
        <h1 className="text-[30px] font-semibold leading-tight tracking-normal">
          SSH Certificates
        </h1>
        <p className="mt-4 max-w-3xl text-lg leading-snug text-neutral-900">
          Manage and monitor your certificates, view details, track validity
          periods, and revoke them to ensure safe and controlled authentication.
        </p>
      </div>

      <div className="h-px bg-neutral-300" />

      <div className="flex flex-wrap items-center gap-x-5 gap-y-3">
        <FilterSelect
          label="Status"
          value={statusFilter}
          onChange={(value) =>
            setStatusFilter(value as CertificateStatus | "all")
          }
          options={[
            { label: "All", value: "all" },
            { label: "Active", value: "active" },
            { label: "Expired", value: "expired" },
            { label: "Revoked", value: "revoked" },
          ]}
        />

        <FilterSelect
          label="Allocation"
          className="min-w-[330px]"
          value={allocationFilter}
          onChange={setAllocationFilter}
          options={[
            { label: "All", value: "all" },
            ...allocationOptions.map((allocation) => ({
              label: allocation,
              value: allocation,
            })),
          ]}
        />

        <FilterSelect
          label="Time range"
          value={timeRange}
          onChange={setTimeRange}
          options={timeRangeOptions.map((option) => ({
            label: option.label,
            value: option.label,
          }))}
        />

        <form
          className="flex items-center gap-2"
          onSubmit={(event) => {
            event.preventDefault();
            setPage(1);
            setUsernameFilter(usernameDraft.trim());
          }}
        >
          <label
            className="text-sm text-neutral-500"
            htmlFor="username-filter"
          >
            Username :
          </label>
          <input
            id="username-filter"
            className="h-10 w-40 rounded-lg border border-neutral-300 bg-white px-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
            placeholder="Filter by username"
            value={usernameDraft}
            onChange={(event) => setUsernameDraft(event.target.value)}
          />
          <button
            className="h-10 rounded-lg border border-neutral-300 bg-white px-3 text-sm font-medium hover:bg-neutral-50"
            type="submit"
          >
            Apply
          </button>
        </form>

        <button
          className="ml-auto inline-flex h-10 items-center gap-2 text-sm font-medium text-blue-700 hover:text-blue-900"
          onClick={reload}
          type="button"
        >
          <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          Refresh list
        </button>
      </div>

      <div className="overflow-hidden rounded-lg bg-white shadow-[0_8px_24px_rgba(0,0,0,0.12)]">
        <Table className="min-w-[1000px] text-left">
          <TableHeader className="bg-[#f1f1f1] text-neutral-950">
            <TableRow>
              <TableHead className="px-5 py-4 text-sm font-bold">Serial</TableHead>
              <TableHead className="px-5 py-4 text-sm font-bold">Username</TableHead>
              <TableHead className="px-5 py-4 text-sm font-bold">Allocation</TableHead>
              <TableHead className="px-5 py-4 text-sm font-bold">Issued Date</TableHead>
              <TableHead className="px-5 py-4 text-sm font-bold">Issued Time</TableHead>
              <TableHead className="px-5 py-4 text-sm font-bold">Status</TableHead>
              <TableHead className="px-5 py-4 text-right text-sm font-bold"> </TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {visibleCertificates.map((cert) => (
              <CertificateRow key={cert.serial_number} cert={cert} />
            ))}
          </TableBody>
        </Table>
      </div>

      {visibleCertificates.length === 0 && (
        <div className="rounded-lg border border-dashed bg-white p-6 text-center text-sm text-neutral-500">
          No SSH certificates found.
        </div>
      )}

      <div className="flex items-center justify-end gap-4 text-sm text-neutral-500">
        <label className="flex items-center gap-2">
          Rows per page :
          <select
            className="h-10 rounded-lg border border-neutral-300 bg-white px-3 pr-8 text-neutral-950"
            value={rowsPerPage}
            onChange={(event) => updateRowsPerPage(Number(event.target.value))}
          >
            <option value={10}>10</option>
            <option value={20}>20</option>
            <option value={50}>50</option>
          </select>
        </label>
        <div className="h-10 w-px bg-neutral-200" />
        <button
          aria-label="Previous page"
          className="flex h-10 w-10 items-center justify-center rounded-md border border-neutral-200 text-neutral-400 disabled:opacity-50"
          disabled={page === 1}
          onClick={() => setPage((current) => Math.max(1, current - 1))}
          type="button"
        >
          <ChevronLeft className="h-5 w-5" />
        </button>
        {pageWindow.map((pageNumber) => (
          <button
            key={pageNumber}
            className={`flex h-10 min-w-10 items-center justify-center rounded-md px-3 ${
              page === pageNumber
                ? "bg-black text-white"
                : "bg-neutral-100 text-neutral-950"
            }`}
            onClick={() => setPage(pageNumber)}
            type="button"
          >
            {pageNumber}
          </button>
        ))}
        <button
          aria-label="Next page"
          className="flex h-10 w-10 items-center justify-center rounded-md border border-neutral-300 text-neutral-950 disabled:opacity-50"
          disabled={page >= totalPages}
          onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
          type="button"
        >
          <ChevronRight className="h-5 w-5" />
        </button>
      </div>
    </section>
  );
}

function buildPageWindow(current: number, total: number): number[] {
  const windowSize = 3;
  const start = Math.max(1, Math.min(current - 1, total - windowSize + 1));
  const end = Math.min(total, start + windowSize - 1);
  const pages: number[] = [];

  for (let i = start; i <= end; i += 1) pages.push(i);

  return pages;
}

function CertificateRow({ cert }: { cert: Certificate }) {
  const status = getCertificateStatus(cert);
  const remaining = formatRemainingTime(getRemainingSeconds(cert));

  return (
    <TableRow className="border-t border-neutral-200 hover:bg-transparent">
      <TableCell className="px-5 py-3">{cert.serial_number}</TableCell>
      <TableCell className="px-5 py-3">{cert.principal}</TableCell>
      <TableCell className="max-w-[360px] truncate px-5 py-3">
        {getCertificateAllocation(cert)}
      </TableCell>
      <TableCell className="px-5 py-3">{formatUnixDate(cert.issued_at)}</TableCell>
      <TableCell className="px-5 py-3">{formatUnixClock(cert.issued_at)}</TableCell>
      <TableCell className="px-5 py-3">
        <div className="flex items-center gap-3">
          <StatusBadge status={status} />
          {status === "active" && (
            <span className="tabular-nums text-neutral-600">{remaining}</span>
          )}
        </div>
      </TableCell>
      <TableCell className="px-5 py-3 text-right">
        <Link
          href={`/signer/certificates/${cert.serial_number}`}
          className="inline-flex h-7 min-w-16 items-center justify-center rounded-md border border-neutral-300 px-3 text-xs font-semibold text-neutral-600 hover:bg-neutral-50"
        >
          More
        </Link>
      </TableCell>
    </TableRow>
  );
}

function FilterSelect({
  label,
  value,
  onChange,
  options,
  className = "min-w-[170px]",
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  options: Array<{ label: string; value: string }>;
  className?: string;
}) {
  return (
    <label className="flex items-center gap-2 text-sm text-neutral-500">
      {label} :
      <select
        className={`h-10 rounded-lg border border-neutral-300 bg-white px-3 pr-8 text-neutral-950 outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100 ${className}`}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

function StatusBadge({ status }: { status: CertificateStatus }) {
  if (status === "active") {
    return (
      <Badge className="h-7 gap-1.5 rounded-full bg-emerald-50 px-2.5 text-emerald-700">
        <span className="h-2 w-2 rounded-full bg-emerald-500" />
        {formatCertificateStatus(status)}
      </Badge>
    );
  }

  if (status === "revoked") {
    return (
      <Badge className="h-7 rounded-full bg-rose-50 px-2.5 text-rose-700">
        {formatCertificateStatus(status)}
      </Badge>
    );
  }

  return (
    <Badge className="h-7 rounded-full bg-neutral-100 px-2.5 text-neutral-600">
      {formatCertificateStatus(status)}
    </Badge>
  );
}

function CertificateListSkeleton() {
  return (
    <section className="mx-auto max-w-[1440px] space-y-6">
      <Skeleton className="h-10 w-72" />
      <Skeleton className="h-6 w-full max-w-3xl" />
      <div className="h-px bg-neutral-300" />
      <div className="space-y-2">
        {Array.from({ length: 6 }).map((_, i) => (
          <Skeleton key={i} className="h-12 w-full" />
        ))}
      </div>
    </section>
  );
}
