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

import { Calendar, Fingerprint, UserSquare } from "lucide-react";
import * as React from "react";
import { toast } from "sonner";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { Button } from "@/shared/ui/button";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { MetaItem, MetaRow } from "@/shared/ui/MetaRow";
import { useCertificate, useRevokeCertificate } from "../queries";
import type { Certificate } from "../schemas";
import {
  CERTIFICATE_STATUS_LABELS,
  type CertificateStatus,
  formatCertificateLifetime,
  formatUnixSeconds,
  getCertificateStatus,
} from "../status";
import { RevokeCertificateDialog } from "./RevokeCertificateDialog";

export type CertificateDetailProps = {
  serial: string;
};

const STATUS_TONE: Record<CertificateStatus, "success" | "warning" | "danger"> = {
  active: "success",
  expired: "warning",
  revoked: "danger",
};

export function CertificateDetail({ serial }: CertificateDetailProps) {
  const ability = useAbility();
  const query = useCertificate(serial);
  const revoke = useRevokeCertificate();
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [revokeError, setRevokeError] = React.useState<string | null>(null);

  if (query.isLoading) return <CardSkeleton />;
  if (query.error) {
    return (
      <ErrorState message={(query.error as Error).message} onRetry={() => query.refetch()} />
    );
  }
  const cert = query.data;
  if (!cert) return null;

  const status = getCertificateStatus(cert);
  const canRevoke = ability.can("manage", "Signer") && !cert.revoked;

  return (
    <section className="space-y-6">
      <header className="space-y-4">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <h1 className="font-display text-[28px] font-bold leading-tight text-foreground">
            Certificate {cert.serial_number}
          </h1>
          {canRevoke ? (
            <Button
              variant="destructive"
              onClick={() => {
                setRevokeError(null);
                setDialogOpen(true);
              }}
            >
              Revoke
            </Button>
          ) : null}
        </div>
        <MetaRow>
          <MetaItem
            variant="status"
            tone={STATUS_TONE[status]}
            value={CERTIFICATE_STATUS_LABELS[status]}
          />
          <MetaItem icon={UserSquare} label="Principal" value={cert.principal} />
          <MetaItem icon={Calendar} label="Valid until" value={formatUnixSeconds(cert.valid_before)} />
          <MetaItem icon={Fingerprint} label="Lifetime" value={formatCertificateLifetime(cert)} />
        </MetaRow>
      </header>

      <CertificateMetadata cert={cert} />

      <RevokeCertificateDialog
        serialNumber={cert.serial_number}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        isPending={revoke.isPending}
        error={revokeError}
        onSubmit={(reason) => {
          setRevokeError(null);
          revoke.mutate(
            { serial: cert.serial_number, reason },
            {
              onSuccess: (res) => {
                setDialogOpen(false);
                toast.success(
                  res.already_revoked
                    ? "Certificate was already revoked"
                    : "Certificate revoked",
                );
              },
              onError: (err) =>
                setRevokeError(err instanceof Error ? err.message : "Revoke failed"),
            },
          );
        }}
      />
    </section>
  );
}

function CertificateMetadata({ cert }: { cert: Certificate }) {
  const items: Array<{ label: string; value: React.ReactNode }> = [
    { label: "Serial number", value: <span className="font-mono">{cert.serial_number}</span> },
    { label: "Key ID", value: <span className="break-all font-mono text-xs">{cert.key_id}</span> },
    { label: "Principal", value: cert.principal },
    {
      label: "Public key fingerprint",
      value: <span className="break-all font-mono text-xs">{cert.public_key_fingerprint}</span>,
    },
    {
      label: "CA fingerprint",
      value: <span className="break-all font-mono text-xs">{cert.ca_fingerprint}</span>,
    },
    { label: "Issued at", value: formatUnixSeconds(cert.issued_at) },
    { label: "Valid after", value: formatUnixSeconds(cert.valid_after) },
    { label: "Valid before", value: formatUnixSeconds(cert.valid_before) },
    { label: "Source IP", value: cert.source_ip || "—" },
    {
      label: "Granted extensions",
      value: cert.granted_extensions?.length ? cert.granted_extensions.join(", ") : "—",
    },
    { label: "Force command", value: cert.force_command || "—" },
  ];

  if (cert.revoked) {
    items.push(
      {
        label: "Revoked at",
        value: cert.revoked_at != null ? formatUnixSeconds(cert.revoked_at) : "—",
      },
      { label: "Revocation reason", value: cert.revocation_reason || "—" },
    );
  }

  return (
    <dl className="grid grid-cols-1 gap-x-8 gap-y-4 rounded-lg border bg-card p-5 sm:grid-cols-2">
      {items.map((item) => (
        <div key={item.label} className="space-y-1">
          <dt className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            {item.label}
          </dt>
          <dd className="text-sm text-foreground">{item.value}</dd>
        </div>
      ))}
    </dl>
  );
}
