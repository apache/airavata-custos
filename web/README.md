<!--
    Licensed to the Apache Software Foundation (ASF) under one
    or more contributor license agreements.  See the NOTICE file
    distributed with this work for additional information
    regarding copyright ownership.  The ASF licenses this file
    to you under the Apache License, Version 2.0 (the
    "License"); you may not use this file except in compliance
    with the License.  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing,
    software distributed under the License is distributed on an
    "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, either express or implied.  See the License for the
    specific language governing permissions and limitations
    under the License.
-->

# Custos Portal

The web portal for Apache Airavata Custos: allocation management, identity and admin tools for HPC sites. To run it
against a local server and OIDC provider, follow steps 6 and 7 of [INSTALL.md](../INSTALL.md).

## Stack

Next.js 16 (App Router) · TypeScript · Tailwind 4 + shadcn/ui · NextAuth v5 · TanStack Query · Zod · CASL · MSW ·
Playwright + Vitest · Biome · pnpm.

## Getting started

```bash
pnpm install
cp .env.example .env.local
pnpm dev
```

`.env.example` documents each variable. The portal does not start without the OIDC settings. The defaults point at
the bundled Keycloak from `dev-ops/compose`.

## Generated API clients

The portal builds its TypeScript types and Zod schemas from the OpenAPI specs in [`../api/`](../api/README.md) and
the connectors' `api/` directories. `pnpm gen:api` writes them to `src/generated/`, which is not committed.
Every script below that needs it runs it first.

## Scripts

| Command | Runs |
|---|---|
| `pnpm dev` | `gen:api`, then the development server on port 3000 |
| `pnpm build` | `gen:api`, then a production build |
| `pnpm start` | the production build |
| `pnpm lint` | `biome lint .` |
| `pnpm format` | `biome format --write .` |
| `pnpm typecheck` | `gen:api`, then `tsc --noEmit` |
| `pnpm test` | `gen:api`, then the Vitest unit tests |
| `pnpm test:watch` | `gen:api`, then Vitest in watch mode |
| `pnpm test:e2e` | Playwright tests in `tests/*.e2e.ts`, against a dev server with MSW fixtures |
| `pnpm test:e2e:signer` | the SSH certificate signer end-to-end test, with MSW off |
| `pnpm verify` | `lint`, `typecheck`, `test` and `build`. CI runs these and `test:e2e:signer` |
| `pnpm clean` | removes `.next`, `dist` and `coverage` |

The `tests/*.spec.ts` files run against a live portal instead, with
`pnpm exec playwright test -c playwright.live.config.ts`. They read the portal URL from `LIVE_PORTAL_URL`, which
defaults to `http://localhost:3001`.

## Layout

| Path | Contents |
|---|---|
| `src/app/` | App Router routes: `(auth)`, `(portal)` and the `api` route handlers |
| `src/features/` | feature modules, split into `core/` and `connectors/` |
| `src/shared/` | code shared across features: UI, layout, API client, auth, CASL abilities |
| `src/lib/` | environment parsing and small utilities |
| `src/mocks/` | MSW handlers for tests |
| `public/` | static files: brand assets, the signed-out landing page and the MSW worker |
| `tests/` | Playwright tests |
