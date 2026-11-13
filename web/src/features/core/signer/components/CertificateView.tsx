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

import { Check, X } from "lucide-react";
import type * as React from "react";
import { Button } from "@/shared/ui/button";
import { CopyValue } from "@/shared/ui/CopyValue";
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

  return (
    <section className="space-y-4">
      <header className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="font-display text-[28px] font-bold leading-tight">
            Certificate <span className="font-mono">{cert.serial_number}</span>
          </h1>
          <StatusBadge
            variant={statusBadgeVariantFromCertificateStatus(status)}
            label={CERTIFICATE_STATUS_LABELS[status]}
          />
          {status === "active" ? (
            <span className="text-sm text-muted-foreground">Expires {formatValidUntil(cert)}</span>
          ) : null}
        </div>
        {status === "active" && onRevoke ? (
          <Button variant="destructive" onClick={onRevoke}>
            Revoke
          </Button>
        ) : null}
      </header>

      {status === "active" ? (
        <p className="text-sm text-muted-foreground">
          A revoked certificate is still accepted by login nodes until it expires.
        </p>
      ) : null}

      {cert.revoked ? (
        <div className="space-y-1 rounded-lg border border-[color:var(--custos-red-700)] p-4 text-sm">
          <p className="font-medium">
            Revoked {cert.revoked_at != null ? formatUnixSeconds(cert.revoked_at) : ""} by{" "}
            <span>{revokedBy}</span>
          </p>
          <p>
            Reason: <span>{cert.revocation_reason || "None given"}</span>
          </p>
          <p className="text-muted-foreground">
            Login nodes still accept this certificate until it expires.
          </p>
        </div>
      ) : null}

      <Section title="Login">
        <Field label="Logs in as">
          <span className="font-mono">{cert.principal}</span>
        </Field>
        {cert.login_host ? (
          <Field label="Command">
            <CopyValue
              value={sshLoginCommand(cert.principal, cert.login_host, cert.login_port ?? 22)}
              label="ssh command"
              explicit
            />
          </Field>
        ) : null}
      </Section>

      <Section title="Validity">
        <Field label="Issued">{formatUnixSeconds(cert.issued_at)}</Field>
        <Field label="Valid until">{formatUnixSeconds(cert.valid_before)}</Field>
        <Field label="Lifetime">{formatCertificateLifetime(cert)}</Field>
        {admin ? <Field label="Valid after">{formatUnixSeconds(cert.valid_after)}</Field> : null}
      </Section>

      <Section title="What this certificate allows">
        <ul className="grid grid-cols-1 gap-2 text-sm sm:col-span-2 sm:grid-cols-2 lg:grid-cols-3">
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
                  <Check
                    className="h-4 w-4 text-[color:var(--custos-green-700)]"
                    aria-hidden="true"
                  />
                ) : (
                  <X className="h-4 w-4" aria-hidden="true" />
                )}
                {option.label}
              </li>
            );
          })}
          {otherGranted.map((name) => (
            <li key={name} className="flex items-center gap-2 font-mono text-xs">
              <Check className="h-4 w-4 text-[color:var(--custos-green-700)]" aria-hidden="true" />
              {name}
            </li>
          ))}
        </ul>
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
      </Section>

      <Section title="Key and identity">
        <Field label="Issued to">{cert.user_email}</Field>
        <Field label="Key fingerprint">
          <CopyValue value={cert.public_key_fingerprint} label="key fingerprint" explicit />
        </Field>
        <Field label="Key ID">
          <CopyValue value={cert.key_id} label="key ID" explicit />
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
              <CopyValue value={cert.ca_fingerprint} label="CA fingerprint" explicit />
            </Field>
          </>
        ) : null}
      </Section>
    </section>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="rounded-lg border bg-card p-5">
      <h2 className="mb-4 text-sm font-semibold">{title}</h2>
      <dl className="grid grid-cols-1 gap-x-8 gap-y-4 sm:grid-cols-2">{children}</dl>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="min-w-0 space-y-1">
      <dt className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{label}</dt>
      <dd className="text-sm text-foreground">{children}</dd>
    </div>
  );
}
