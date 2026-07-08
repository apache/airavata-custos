import { BookUser, Fingerprint, Globe, KeyRound, type LucideIcon } from "lucide-react";

// Display metadata for the `source` field on UserIdentity (GET
// /users/{id}/user-identities) — "the source's native identifier" per the
// API doc, e.g. "access", "cilogon", "orcid", "nairr".
export const IDENTITY_SOURCE_LABELS: Record<string, string> = {
  access: "ACCESS",
  cilogon: "CILogon",
  orcid: "ORCID",
  nairr: "NAIRR",
};

const IDENTITY_SOURCE_ICONS: Record<string, LucideIcon> = {
  access: Globe,
  cilogon: KeyRound,
  orcid: BookUser,
  nairr: Fingerprint,
};

export function identitySourceLabel(source: string | undefined): string {
  if (!source) return "Unknown";
  return IDENTITY_SOURCE_LABELS[source] ?? source;
}

export function identitySourceIcon(source: string | undefined): LucideIcon {
  return (source ? IDENTITY_SOURCE_ICONS[source] : undefined) ?? Fingerprint;
}
