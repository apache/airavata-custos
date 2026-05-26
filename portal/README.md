# Custos Portal

Next.js 15 portal for Custos SSH certificate management. Wraps the SSH
Certificate Signer service with an authenticated UI for browsing, inspecting,
and revoking certificates.

## Architecture at a glance

- **Next.js 15 App Router** with React 19, Tailwind CSS v4, and shadcn/ui
  primitives (Badge, Button, Dialog, Skeleton, Table).
- **NextAuth v5 (Auth.js)** drives sign-in. Configured to talk to a CILogon-
  compatible OIDC issuer (real CILogon, Keycloak, or any standards-compliant
  provider) and falls back to a local "Dev" credentials provider when no
  issuer is configured.
- **Server-side API proxy** at `/api/v1/[...path]` injects the session bearer
  on GETs and signer client credentials on POSTs so nothing sensitive is ever
  bundled into the browser.

```
Browser ── /api/v1/* ──▶ Next route ── adds auth ─▶ Signer service
   ▲                        │
   └── /api/auth/* ─────────┘
            (NextAuth handlers — sign-in, callback, session)
```

Key files:

| File | Role |
| --- | --- |
| [`auth.ts`](auth.ts) | NextAuth configuration; selects between OIDC and Dev providers. |
| [`src/app/api/auth/[...nextauth]/route.ts`](src/app/api/auth/[...nextauth]/route.ts) | Mounts NextAuth handlers under `/api/auth/`. |
| [`src/app/api/v1/[...path]/route.ts`](src/app/api/v1/[...path]/route.ts) | Server-side proxy that forwards browser calls to the signer with the right credentials. |
| [`src/app/signer/`](src/app/signer/) | Certificate list, detail, and revoke flows. |
| [`src/app/layout/PortalLayout.tsx`](src/app/layout/PortalLayout.tsx) | Sidebar/header chrome and the unauthenticated → `signIn()` redirect. |

## Local setup

The dev server binds to `http://localhost:5173` (see `package.json`).

```bash
cp .env.local.example .env.local
# Fill in OIDC_CLIENT_ID, OIDC_ISSUER_URL (and OIDC_CLIENT_SECRET if you
# have one — leave blank for public/PKCE-only clients).
# Generate NEXTAUTH_SECRET: openssl rand -base64 32
npm install
npm run dev
```

Then open <http://localhost:5173>.

## OIDC authentication

The portal targets CILogon by default but accepts any OIDC-compliant issuer
through env vars. Required values in `.env.local`:

| Variable | Purpose |
| --- | --- |
| `OIDC_CLIENT_ID` | Client identifier registered at the OIDC provider. |
| `OIDC_CLIENT_SECRET` | Client secret. Leave blank for public (PKCE-only) clients. |
| `OIDC_ISSUER_URL` | Issuer base URL — e.g. `https://cilogon.org` or a Keycloak realm. |
| `OIDC_SCOPE` (optional) | Override the requested scope set. Defaults to `openid profile email org.cilogon.userinfo` (CILogon's claim set). For Keycloak realms that don't advertise `org.cilogon.userinfo`, use `openid profile email`. |
| `NEXTAUTH_URL` | Public origin Auth.js uses to build the redirect URI. Must match the scheme/host/port reachable from the browser — `http://localhost:5173` for local dev. |
| `NEXTAUTH_SECRET` | Cookie/JWT encryption key. Generate with `openssl rand -base64 32`. |
| `AUTH_TRUST_HOST` | Set to `true` so Auth.js trusts the inbound Host header outside Vercel-style deployments. |

Register this callback URL with the OIDC client for local dev:

```
http://localhost:5173/api/auth/callback/cilogon
```

The `cilogon` segment comes from the provider id in [`auth.ts`](auth.ts);
changing it is a breaking config change for any deployment that has the URL
whitelisted.

## Dev OIDC fallback

When `OIDC_ISSUER_URL` is left blank, the portal registers a local
credentials-only **Dev** provider instead of an OIDC one. Pair it with the
signer's built-in dev OIDC mode (which disables real token validation and
returns a default identity) so contributors can run the full sign-in flow
without standing up a real provider. See
[`../extensions/SSH-Certificate-Signer/README.md`](../extensions/SSH-Certificate-Signer/README.md)
for the matching `DEV_MODE` / `dev_mode.enabled` configuration.

## Signer integration

The portal never talks to the signer directly from the browser. Set these in
`.env.local` (or leave them unset to use the dev defaults baked into
[`src/lib/serverConfig.ts`](src/lib/serverConfig.ts)):

| Variable | Default | Purpose |
| --- | --- | --- |
| `SIGNER_API_BASE_URL` | `http://127.0.0.1:8084` | Base URL the proxy forwards to. |
| `SIGNER_CLIENT_ID` | `tenant1:webapp` | Sent as `X-Client-Id` on `/revoke`. |
| `SIGNER_CLIENT_SECRET` | `dev-secret` | Sent as `X-Client-Secret` on `/revoke`. |

## Scripts

- `npm run dev` — Next dev server on <http://localhost:5173>
- `npm run build` — production build
- `npm run typecheck` — `tsc --noEmit`
- `npm run lint` — `next lint`
- `npm test` — Vitest unit tests
- `npm run test:e2e` — Playwright e2e (spawns its own dev server on port 3105)
