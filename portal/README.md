# Custos Portal

Next.js 15 portal for Custos SSH certificate management.

## Setup

```bash
cp .env.local.example .env.local
# Fill in OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, OIDC_ISSUER_URL.
# Generate NEXTAUTH_SECRET: openssl rand -base64 32
npm install
npm run dev
```

## Dev OIDC mode

The signer service supports a built-in dev OIDC mode that disables real token
validation and returns a default identity, so contributors can run the full
sign-in flow locally without standing up a real OIDC provider (Keycloak,
CILogon, etc.). See
[../extensions/SSH-Certificate-Signer/README.md](../extensions/SSH-Certificate-Signer/README.md)
for `DEV_MODE` / `dev_mode.enabled` configuration.

## Scripts

- `npm run dev` — Next dev server on `127.0.0.1:5173`
- `npm run build` — production build
- `npm run typecheck` — `tsc --noEmit`
- `npm run lint` — `next lint`
- `npm test` — vitest unit tests
- `npm run test:e2e` — playwright e2e (requires a running portal + signer)
