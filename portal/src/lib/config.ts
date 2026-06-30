// Public portal branding/config. Only NEXT_PUBLIC_* env vars are read here,
// so importing this module from a client component cannot leak server
// secrets. Each value falls back to a sensible default for unconfigured
// dev checkouts.
export const PORTAL_NAME =
  process.env.NEXT_PUBLIC_PORTAL_NAME ?? "Custos Portal";

export const ORG_NAME =
  process.env.NEXT_PUBLIC_ORG_NAME ?? "HPC Access Portal";

export const SUPPORT_EMAIL =
  process.env.NEXT_PUBLIC_SUPPORT_EMAIL ?? "support@example.org";

export const ALLOCATION_LABEL =
  process.env.NEXT_PUBLIC_ALLOCATION_LABEL ?? "Allocation";
