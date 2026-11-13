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

import { Check, Clock, Hourglass, UserSquare, X } from "lucide-react";
import type * as React from "react";
import { cn } from "@/lib/utils";
import { Button } from "@/shared/ui/button";
import { Card } from "@/shared/ui/card";
import { CopyValue } from "@/shared/ui/CopyValue";
import { MetaItem, MetaRow } from "@/shared/ui/MetaRow";
import { StatusBadge } from "@/shared/ui/StatusBadge";
import type { Certificate } from "../schemas";
import {
  CERTIFICATE_OPTIONS,
  CERTIFICATE_STATUS_LABELS,
  formatCertificateLifetime,
  formatUnixSeconds,
  formatValidUntil,
  getCertificateStatus,
  sshLoginCommand,
  statusBadgeVariantFromCertificateStatus,
} from "../status";

export type CertificateViewProps = {
  cert: Certificate;
  // Admins also see the subject, who asked and the CA.
  admin?: boolean;
  revokedBy: string;
  onRevoke?: () => void;
};

export function CertificateView({
  cert,
  admin = false,
  revokedBy,
  onRevoke,
}: CertificateViewProps) {
  const status = getCertificateStatus(cert);
  const granted = cert.granted_extensions ?? [];
  const otherGranted = granted.filter((name) => !CERTIFICATE_OPTIONS.some((o) => o.name === name));

  const command = cert.login_host
    ? sshLoginCommand(cert.principal, cert.login_host, cert.login_port ?? 22)
    : null;

  return (
    <section className="space-y-4">
      <header className="space-y-3">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <h1 className="font-display text-[28px] font-bold leading-tight">
            Certificate <span className="font-mono">{cert.serial_number}</span>
          </h1>
          {status === "active" && onRevoke ? (
            <Button variant="destructive" onClick={onRevoke}>
              Revoke
            </Button>
          ) : null}
        </div>
        <div className="flex flex-wrap items-center gap-3">
          <StatusBadge
            variant={statusBadgeVariantFromCertificateStatus(status)}
            label={CERTIFICATE_STATUS_LABELS[status]}
          />
          <MetaRow>
            <MetaItem
              icon={UserSquare}
              label="Logs in as"
              value={<span className="font-mono">{cert.principal}</span>}
            />
            <MetaItem
              icon={Clock}
              label={status === "active" ? "Expires" : "Valid until"}
              value={formatValidUntil(cert)}
            />
            <MetaItem icon={Hourglass} label="Lifetime" value={formatCertificateLifetime(cert)} />
          </MetaRow>
        </div>
      </header>

      {cert.revoked ? (
        <div className="space-y-1 rounded-lg border border-destructive p-4 text-sm">
          <p className="font-medium">
            Revoked
            {cert.revoked_at != null ? ` ${formatUnixSeconds(cert.revoked_at)}` : ""} by{" "}
            <span>{revokedBy}</span>
          </p>
          <p>
            Reason: <span>{cert.revocation_reason || "None given"}</span>
          </p>
          <p className="text-muted-foreground">
            Login nodes still accept this certificate until it expires.
          </p>
        </div>
      ) : status === "active" ? (
        <div className="space-y-2 rounded-lg border bg-card p-4 text-sm">
          {command ? <CopyValue value={command} label="ssh command" explicit wrap /> : null}
          <p className="text-muted-foreground">
            A revoked certificate is still accepted by login nodes until it expires.
          </p>
        </div>
      ) : (
        <p className="rounded-lg border bg-card p-4 text-sm text-muted-foreground">
          {status === "expired"
            ? `Expired ${formatUnixSeconds(cert.valid_before)}. It no longer logs in.`
            : `Works from ${formatUnixSeconds(cert.valid_after)}.`}
        </p>
      )}

      <div className="grid gap-4 lg:grid-cols-2">
        <InfoCard title="Validity">
          <dl className="divide-y">
            <Field label="Issued">{formatUnixSeconds(cert.issued_at)}</Field>
            <Field label="Valid until">{formatUnixSeconds(cert.valid_before)}</Field>
            {admin ? (
              <Field label="Valid after">{formatUnixSeconds(cert.valid_after)}</Field>
            ) : null}
          </dl>
        </InfoCard>

        <InfoCard title="What this certificate allows">
          <ul
            aria-label="Certificate options"
            className="grid grid-cols-1 gap-x-6 gap-y-1.5 border-b py-2.5 text-sm sm:grid-cols-2"
          >
            {CERTIFICATE_OPTIONS.map((option) => {
              const allowed = granted.includes(option.name);
              return (
                <li
                  key={option.name}
                  aria-label={`${allowed ? "Allowed" : "Blocked"}: ${option.label}`}
                  className={
                    allowed
                      ? "flex items-center gap-2"
                      : "flex items-center gap-2 text-muted-foreground"
                  }
                >
                  {allowed ? (
                    <Check className="h-4 w-4 text-custos-green-700" aria-hidden="true" />
                  ) : (
                    <X className="h-4 w-4" aria-hidden="true" />
                  )}
                  {option.label}
                </li>
              );
            })}
            {otherGranted.map((name) => (
              <li key={name} className="flex items-center gap-2 font-mono text-xs">
                <Check className="h-4 w-4 text-custos-green-700" aria-hidden="true" />
                {name}
              </li>
            ))}
          </ul>
          <dl className="divide-y">
            <Field label="Runs only">
              {cert.force_command ? (
                <span className="font-mono text-xs">{cert.force_command}</span>
              ) : (
                "Any command"
              )}
            </Field>
            <Field label="Log in from">
              {cert.source_address ? (
                <span className="font-mono text-xs">{cert.source_address}</span>
              ) : (
                "Any address"
              )}
            </Field>
          </dl>
        </InfoCard>

        <InfoCard title="Key and identity" className="lg:col-span-2">
          <dl className="divide-y">
            <Field label="Issued to">{cert.user_email}</Field>
            <Field label="Key fingerprint">
              <CopyValue
                value={cert.public_key_fingerprint}
                label="key fingerprint"
                explicit
                wrap
              />
            </Field>
            <Field label="Key ID">
              <CopyValue value={cert.key_id} label="key ID" explicit wrap />
            </Field>
            {admin ? (
              <>
                <Field label="Subject">
                  <span className="break-all font-mono text-xs">{cert.user_sub}</span>
                </Field>
                <Field label="Requested by">
                  <span className="font-mono text-xs">{cert.requested_by}</span>
                </Field>
                <Field label="CA fingerprint">
                  <CopyValue value={cert.ca_fingerprint} label="CA fingerprint" explicit wrap />
                </Field>
              </>
            ) : null}
          </dl>
        </InfoCard>
      </div>
    </section>
  );
}

function InfoCard({
  title,
  className,
  children,
}: {
  title: string;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <Card className={cn("gap-0 py-0", className)}>
      <h2 className="border-b px-5 py-3 text-sm font-semibold">{title}</h2>
      <div className="px-5 pb-1">{children}</div>
    </Card>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="grid grid-cols-1 gap-1 py-2.5 text-sm sm:grid-cols-[9rem_minmax(0,1fr)] sm:items-baseline sm:gap-4">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="min-w-0 text-foreground">{children}</dd>
    </div>
  );
}
