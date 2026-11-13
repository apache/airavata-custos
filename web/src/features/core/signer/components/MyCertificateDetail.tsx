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
import { useCurrentUser } from "@/features/core/identity/queries";
import { ApiError } from "@/shared/api/client";
import { ErrorState } from "@/shared/ui/ErrorState";
import { CardSkeleton } from "@/shared/ui/Loading";
import { useMyCertificate, useRevokeMyCertificate } from "../queries";
import { userIdOf } from "../status";
import { revokeErrorMessage } from "./CertificateDetail";
import { CertificateView } from "./CertificateView";
import { loadErrorMessage } from "./MyCertificates";
import { RevokeCertificateDialog } from "./RevokeCertificateDialog";

export function MyCertificateDetail({ serial }: { serial: string }) {
  const query = useMyCertificate(serial);
  const { user } = useCurrentUser();
  const revoke = useRevokeMyCertificate();
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [revokeError, setRevokeError] = React.useState<string | null>(null);

  if (query.isLoading) return <CardSkeleton />;
  if (query.error) {
    if (query.error instanceof ApiError && query.error.status === 404) {
      return <ErrorState heading="Certificate not found" message="You have no certificate with this serial." />;
    }
    return <ErrorState message={loadErrorMessage(query.error)} onRetry={() => query.refetch()} />;
  }
  const cert = query.data;
  if (!cert) return null;

  return (
    <>
      <CertificateView
        cert={cert}
        revokedBy={userIdOf(cert.revoked_by) === user?.id ? "you" : "your site"}
        onRevoke={() => {
          setRevokeError(null);
          setDialogOpen(true);
        }}
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
              onSuccess: () => {
                setDialogOpen(false);
                toast.success("Certificate revoked");
              },
              onError: (err) => setRevokeError(revokeErrorMessage(err)),
            },
          );
        }}
      />
    </>
  );
}
