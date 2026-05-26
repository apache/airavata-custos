/**
 * Server-only signer integration settings.
 *
 * Importing this from any client component would leak the configured client
 * secret into the browser bundle. The /api/v1 proxy is the only intended
 * consumer; everything else should call the proxy.
 *
 * Each value falls back to a development default so an unconfigured
 * checkout still boots against a locally-running signer.
 */
export const API_BASE_URL =
  process.env.SIGNER_API_BASE_URL ?? "http://127.0.0.1:8084";

export const CLIENT_ID = process.env.SIGNER_CLIENT_ID ?? "tenant1:webapp";

export const CLIENT_SECRET = process.env.SIGNER_CLIENT_SECRET ?? "dev-secret";
