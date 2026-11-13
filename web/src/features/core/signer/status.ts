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

export type CertificateStatus = "active" | "not-yet-valid" | "expired" | "revoked";

export function userIdOf(actor: string | undefined): string | null {
  return actor?.startsWith("user/") ? actor.slice("user/".length) : null;
}

export const CERTIFICATE_STATUS_LABELS: Record<CertificateStatus, string> = {
  active: "Active",
  "not-yet-valid": "Not yet valid",
  expired: "Expired",
  revoked: "Revoked",
};

// Revocation takes precedence over time-based validity.
export function getCertificateStatus(cert: Certificate, now: Date = new Date()): CertificateStatus {
  if (cert.revoked) return "revoked";
  const nowSeconds = Math.floor(now.getTime() / 1000);
  if (cert.valid_after > nowSeconds) return "not-yet-valid";
  if (cert.valid_before <= nowSeconds) return "expired";
  return "active";
}

export function statusBadgeVariantFromCertificateStatus(
  status: CertificateStatus,
): StatusBadgeVariant {
  if (status === "active") return "active";
  if (status === "expired") return "expired";
  if (status === "not-yet-valid") return "pending";
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

// Time left on an active certificate, otherwise its end date. Certificates last at most a day.
export function formatValidUntil(cert: Certificate, now: Date = new Date()): string {
  if (getCertificateStatus(cert, now) !== "active") return formatUnixDate(cert.valid_before);
  const left = cert.valid_before - Math.floor(now.getTime() / 1000);
  if (left < 60) return "in less than a minute";
  if (left < 3600) return `in ${Math.floor(left / 60)} min`;
  return `in ${Math.floor(left / 3600)}h ${Math.floor((left % 3600) / 60)}m`;
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
