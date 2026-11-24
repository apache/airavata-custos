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

import * as React from "react";
import { toast } from "sonner";
import { useAbility } from "@/shared/casl/AbilityProvider";
import { ApiError } from "@/shared/api/client";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { useCertificate, useRevokeCertificate, useUserEmail } from "../queries";
import { userIdOf } from "../status";
import { CertificateView } from "./CertificateView";
import { RevokeCertificateDialog } from "./RevokeCertificateDialog";

export type CertificateDetailProps = {
  serial: string;
};

export function CertificateDetail({ serial }: CertificateDetailProps) {
  const ability = useAbility();
  const query = useCertificate(serial);
  const revoke = useRevokeCertificate();
  const revokerEmail = useUserEmail(userIdOf(query.data?.revoked_by));
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [revokeError, setRevokeError] = React.useState<string | null>(null);

  if (query.isLoading) return <CardSkeleton />;
  if (query.error) {
    if (query.error instanceof ApiError && query.error.status === 404) {
      return (
        <ErrorState heading="Certificate not found" message="No certificate has this serial." />
      );
    }
    return <ErrorState message={(query.error as Error).message} onRetry={() => query.refetch()} />;
  }
  const cert = query.data;
  if (!cert) return null;

  const openRevoke = () => {
    setRevokeError(null);
    setDialogOpen(true);
  };

  return (
    <>
      <CertificateView
        cert={cert}
        admin
        revokedBy={revokerEmail ?? cert.revoked_by ?? "unknown"}
        {...(ability.can("manage", "Signer") ? { onRevoke: openRevoke } : {})}
      />

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
                  res.already_revoked ? "Certificate was already revoked" : "Certificate revoked",
                );
              },
              onError: (err) => setRevokeError(revokeErrorMessage(err)),
            },
          );
        }}
      />
    </>
  );
}

export function revokeErrorMessage(error: unknown): string {
  if (error instanceof ApiError) {
    if (error.status === 403) return "You no longer have permission to revoke certificates.";
    if (error.status === 409) return "This certificate is no longer active and cannot be revoked.";
    if (error.body && typeof error.body === "object" && "message" in error.body) {
      const message = (error.body as { message?: unknown }).message;
      if (typeof message === "string" && message.length > 0) return message;
    }
  }
  return error instanceof Error ? error.message : "Revoke failed. Please try again.";
}
