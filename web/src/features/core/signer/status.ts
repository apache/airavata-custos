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

import type { StatusBadgeVariant } from "@/shared/ui/StatusBadge";
import type { Certificate } from "./schemas";

export type CertificateStatus = "active" | "expired" | "revoked";

export const CERTIFICATE_STATUS_LABELS: Record<CertificateStatus, string> = {
  active: "Active",
  expired: "Expired",
  revoked: "Revoked",
};

// Revocation takes precedence over time-based validity.
export function getCertificateStatus(cert: Certificate, now: Date = new Date()): CertificateStatus {
  if (cert.revoked) return "revoked";
  const nowSeconds = Math.floor(now.getTime() / 1000);
  if (cert.valid_before < nowSeconds) return "expired";
  return "active";
}

export function statusBadgeVariantFromCertificateStatus(
  status: CertificateStatus,
): StatusBadgeVariant {
  if (status === "active") return "active";
  if (status === "expired") return "expired";
  return "deleted";
}

export function formatUnixSeconds(seconds: number): string {
  return new Date(seconds * 1000).toLocaleString();
}

export function formatUnixDate(seconds: number): string {
  return new Date(seconds * 1000).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

export function formatCertificateLifetime(cert: Certificate): string {
  const seconds = Math.max(0, cert.valid_before - cert.valid_after);
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  if (hours > 0 && minutes === 0) return `${hours} hour${hours === 1 ? "" : "s"}`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  if (minutes > 0) return `${minutes} minute${minutes === 1 ? "" : "s"}`;
  return `${seconds} seconds`;
}
