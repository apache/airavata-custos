// Certificate detail panel for /signer/certificates/[serial]. Rendered as
// a right-side overlay above a blurred snapshot of the list so users keep
// list context while inspecting or revoking a single certificate.
"use client";

import type React from "react";
import Link from "next/link";
import { Skeleton } from "@/components/ui/skeleton";
import {
  formatCertificateLifetime,
  formatCertificateStatus,
  formatRemainingTime,
  formatUnixTime,
  getCertificateStatus,
  getRemainingSeconds,
} from "./types";
import { useCertificate } from "./hooks";
import { RevokeDialog } from "./RevokeDialog";
import { CertificateListPage } from "./CertificateListPage";

export function CertificateDetailPage({ serial }: { serial: string }) {
  const { data: cert, loading, error, reload } = useCertificate(serial);

  if (loading && !cert) {
    return (
      <aside className="fixed inset-y-0 right-0 z-40 flex w-full max-w-[820px] flex-col gap-4 overflow-hidden rounded-l-2xl bg-white p-12 shadow-[-10px_0_30px_rgba(0,0,0,0.14)]">
        <Skeleton className="h-10 w-3/4" />
        <Skeleton className="h-6 w-1/2" />
        <Skeleton className="h-40 w-full" />
        <Skeleton className="h-40 w-full" />
      </aside>
    );
  }
  if (error) return <p className="text-red-600">{error}</p>;
  if (!cert) return <p>Certificate not found.</p>;

  const status = getCertificateStatus(cert);
  const isActive = status === "active";

  return (
    <section className="-mx-12 -my-8 min-h-[calc(100vh-5rem)]">
      <div className="pointer-events-none select-none px-12 py-8 blur-[3px]">
        <CertificateListPage />
      </div>
      <div className="fixed inset-0 z-30 bg-white/20 backdrop-blur-[2px]" />

      <aside className="fixed inset-y-0 right-0 z-40 flex w-full max-w-[820px] flex-col overflow-hidden rounded-l-2xl bg-white shadow-[-10px_0_30px_rgba(0,0,0,0.14)]">
        <header className="flex min-h-28 items-center justify-between border-b border-neutral-200 px-12">
          <h1 className="text-2xl font-semibold tracking-normal">
            Certificate {cert.serial_number}
          </h1>
          <div className="flex items-center gap-4">
            {/* Only show the countdown for active certs — once revoked or
                expired the remaining time is zero by definition and the
                status badge already conveys that. */}
            {isActive && (
              <div className="text-right">
                <div className="text-sm text-neutral-500">Remaining time</div>
                <div className="text-base font-bold tabular-nums">
                  {formatRemainingTime(getRemainingSeconds(cert))}
                </div>
              </div>
            )}
            <StatusBadge status={status} large />
          </div>
        </header>

        <div className="flex-1 overflow-y-auto px-12 py-8">
          <Section title="Main Info">
            <Detail label="Serial Number" value={String(cert.serial_number)} />
            <Detail label="Username" value={cert.principal} />
            <Detail label="Key ID" value={cert.key_id} />
            <Detail
              label="Public Key Fingerprint"
              value={cert.public_key_fingerprint}
            />
            <Detail label="CA Fingerprint" value={cert.ca_fingerprint} />
            <Detail
              label="Forced Command"
              value={cert.force_command || "N/A"}
            />
            <Detail label="Source IP" value={cert.source_ip || "N/A"} />
          </Section>

          <Section className="mt-10" title="Certificate Validity">
            <Detail label="Issued At" value={formatUnixTime(cert.issued_at)} />
            <Detail
              label="Valid until"
              value={formatUnixTime(cert.valid_before)}
            />
            <Detail label="Lifetime" value={formatCertificateLifetime(cert)} />
          </Section>

          {status === "revoked" && (
            <Section className="mt-10" title="Revocation">
              <Detail
                label="Revoked at"
                value={cert.revoked_at ? formatUnixTime(cert.revoked_at) : "N/A"}
              />
              <Detail
                label="Revocation reason"
                value={cert.revocation_reason || "N/A"}
              />
              <Detail label="Status" value={formatCertificateStatus(status)} />
            </Section>
          )}

          {status !== "revoked" && (
            <dl className="sr-only">
              <Detail label="Status" value={formatCertificateStatus(status)} />
            </dl>
          )}
        </div>

        <footer className="flex items-center justify-end gap-4 px-12 py-8">
          <Link
            className="inline-flex h-12 min-w-32 items-center justify-center rounded-lg border border-neutral-300 bg-white px-6 text-base font-semibold text-neutral-700 hover:bg-neutral-50"
            href="/signer/certificates"
          >
            Cancel
          </Link>

          {isActive && <RevokeDialog cert={cert} onRevoked={reload} />}
        </footer>
      </aside>
    </section>
  );
}

function Section({
  title,
  children,
  className = "",
}: {
  title: string;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={className}>
      <h2 className="text-lg font-semibold text-neutral-700">{title}</h2>
      <dl className="mt-2 space-y-3">{children}</dl>
    </section>
  );
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid min-h-12 grid-cols-[200px_minmax(0,1fr)] items-start rounded bg-neutral-50 px-4 py-3 text-sm">
      <dt className="flex items-center gap-3 text-neutral-500">
        <span className="h-1 w-1 rounded-full bg-neutral-500" />
        {label}
      </dt>
      <dd className="break-words text-neutral-950">{value}</dd>
    </div>
  );
}

function StatusBadge({
  status,
  large = false,
}: {
  status: "active" | "expired" | "revoked";
  large?: boolean;
}) {
  const sizeClasses = large ? "h-10 px-4 text-base" : "h-7 px-2.5 text-xs";
  const dotSize = large ? "h-2.5 w-2.5" : "h-2 w-2";

  if (status === "active") {
    return (
      <span
        className={`inline-flex items-center gap-2 rounded-full bg-emerald-50 font-medium text-emerald-700 ${sizeClasses}`}
      >
        <span className={`rounded-full bg-emerald-500 ${dotSize}`} />
        {formatCertificateStatus(status)}
      </span>
    );
  }

  if (status === "revoked") {
    return (
      <span
        className={`inline-flex items-center rounded-full bg-rose-50 font-medium text-rose-700 ${sizeClasses}`}
      >
        {formatCertificateStatus(status)}
      </span>
    );
  }

  return (
    <span
      className={`inline-flex items-center rounded-full bg-neutral-100 font-medium text-neutral-600 ${sizeClasses}`}
    >
      {formatCertificateStatus(status)}
    </span>
  );
}
