// Typed wrappers around the signer REST API. All calls are intentionally
// relative URLs so the Next /api/v1/[...path] proxy can apply server-side
// auth before forwarding to the signer backend.
import { apiFetch } from "../../lib/http";
import type {
  Certificate,
  CertificateListResponse,
  RevokeRequest,
  RevokeResponse,
  UserInfo,
} from "./types";

export function getUserInfo(): Promise<UserInfo> {
  return apiFetch<UserInfo>("/api/v1/userinfo");
}

export function listCertificates(params?: {
  limit?: number;
  offset?: number;
  username?: string;
}): Promise<CertificateListResponse> {
  const search = new URLSearchParams();

  if (params?.limit) search.set("limit", String(params.limit));
  if (params?.offset) search.set("offset", String(params.offset));
  if (params?.username) search.set("username", params.username);

  const query = search.toString();
  return apiFetch<CertificateListResponse>(
    `/api/v1/certificates${query ? `?${query}` : ""}`
  );
}

export function getCertificate(serial: string | number): Promise<Certificate> {
  return apiFetch<Certificate>(`/api/v1/certificates/${serial}`);
}

export function revokeCertificate(
  payload: RevokeRequest
): Promise<RevokeResponse> {
  return apiFetch<RevokeResponse>("/api/v1/revoke", {
    method: "POST",
    auth: "client",
    body: JSON.stringify(payload),
  });
}
