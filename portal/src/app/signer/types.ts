export type CertificateStatus = "active" | "expired" | "revoked";

export type Certificate = {
  serial_number: number;
  client_id?: string;
  allocation?: string;
  key_id: string;
  principal: string;
  public_key_fingerprint: string;
  ca_fingerprint: string;
  valid_after: number;
  valid_before: number;
  issued_at: number;
  source_ip?: string;
  granted_extensions?: string[];
  force_command?: string | null;
  revoked: boolean;
  revoked_at?: number;
  revocation_reason?: string;
};

export type CertificateListResponse = {
  certificates: Certificate[];
  total: number;
  limit: number;
  offset: number;
};

export type UserInfo = {
  subject: string;
  issuer: string;
  email: string;
  principal: string;
  username?: string;
};

export type RevokeRequest = {
  serial_number?: number;
  key_id?: string;
  ca_fingerprint?: string;
  reason: string;
};

export type RevokeResponse = {
  success: boolean;
  message: string;
  revoked_count: number;
};

export function getCertificateStatus(cert: Certificate): CertificateStatus {
  const nowSeconds = Math.floor(Date.now() / 1000);

  // Revocation takes precedence over time-based validity.
  if (cert.revoked) return "revoked";
  if (cert.valid_before < nowSeconds) return "expired";
  return "active";
}

export function formatCertificateStatus(status: CertificateStatus): string {
  return status.charAt(0).toUpperCase() + status.slice(1);
}

export function formatUnixTime(seconds: number): string {
  return new Date(seconds * 1000).toLocaleString();
}

export function formatUnixDate(seconds: number): string {
  return new Date(seconds * 1000).toLocaleDateString();
}

export function formatUnixClock(seconds: number): string {
  return new Date(seconds * 1000)
    .toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    })
    .replace(/:/g, ".");
}

export function getRemainingSeconds(cert: Certificate): number {
  return Math.max(0, cert.valid_before - Math.floor(Date.now() / 1000));
}

export function formatRemainingTime(seconds: number): string {
  if (seconds <= 0) return "00:00";

  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remainingSeconds = seconds % 60;

  if (hours > 0) return `${hours}h ${String(minutes).padStart(2, "0")}m`;

  return `${String(minutes).padStart(2, "0")}:${String(
    remainingSeconds
  ).padStart(2, "0")}`;
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

export function getCertificateAllocation(cert: Certificate): string {
  return cert.allocation || cert.client_id || "Unassigned";
}
