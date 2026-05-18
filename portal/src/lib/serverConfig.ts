// Signer integration settings stay server-side so bearer and client credentials
// are never bundled into the browser application.
export const API_BASE_URL =
  process.env.SIGNER_API_BASE_URL ?? "http://127.0.0.1:8084";

export const DEV_BEARER_TOKEN =
  process.env.SIGNER_DEV_BEARER_TOKEN ?? "dev-token";

export const CLIENT_ID = process.env.SIGNER_CLIENT_ID ?? "tenant1:webapp";

export const CLIENT_SECRET = process.env.SIGNER_CLIENT_SECRET ?? "dev-secret";
