# CLAUDE.md

Context for AI assistants working in `airavata-custos/web/`. Human contributors should start with [README.md](./README.md).

## What this is

Web portal for **Apache Custos** — allocation management, identity, and admin tooling for HPC sites. Pairs with the Apache Custos backend; expects either a live backend (`CUSTOS_CORE_API_BASE_URL`) or the in-repo MSW mock layer.

This portal was ported from a working prototype. The prototype's ADRs and glossary remain the single source of truth — see References below.

## Hard rules in force (these override defaults)

1. **No `git commit` and no `git push` in `airavata-custos` from automation.** Only the human maintainer commits here. AI work stays in the working tree on the active branch.
2. **No prototype branding.** CSS tokens are `--custos-*`, env vars are `CUSTOS_*` / `NEXT_PUBLIC_CUSTOS_*`. Don't introduce prototype-era identifiers in code, comments, or specs.
3. **Code comments: precise, why-only, 2 lines max.** Default to no comments. No restate-what-the-code-does, no change-log comments, no client/SDK references.
4. **Keep it simple. No over-engineering.** Default to the simplest sufficient solution. Resist layers, hedges, "for safety" guards.

## Commands

```bash
pnpm dev            # Start dev server (localhost:3000)
pnpm build          # Production build
pnpm lint           # Biome lint
pnpm format         # Biome format (writes in place)
pnpm typecheck      # tsc --noEmit
pnpm test           # Vitest (unit, run once)
pnpm test:watch     # Vitest watch mode
pnpm test:e2e       # Playwright e2e tests
pnpm verify         # lint + typecheck + test + build (full gate)
```

Every change must leave `pnpm verify` green.

## Local development

Copy `.env.example` to `.env.local`. Defaults work without a backend once auth + MSW are wired (Phase 2+):

- `PORTAL_AUTH_MODE=dev` — credentials-based sign-in (Phase 3).
- `NEXT_PUBLIC_PORTAL_USE_MSW=true` — MSW intercepts `/api/v1/*` calls in the browser; no backend required.

To test against a real Custos backend, set `PORTAL_AUTH_MODE=oidc` and provide `OIDC_ISSUER_URL`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, plus `CUSTOS_CORE_API_BASE_URL`. The env schema in `src/lib/env.ts` fails fast at boot if anything required is missing.

## Architecture

### Next.js App Router layout

```
src/app/
  (auth)/sign-in/      — sign-in (credentials / OIDC)
  (portal)/            — authenticated portal shell
  api/
    auth/[...nextauth]/  — NextAuth handler (Phase 3)
    v1/[...path]/        — transparent proxy to Custos backend (Phase 2)
```

### Feature structure (lands in later phases)

Each feature in `src/features/{core,connectors}/<name>/` follows the same internal shape:

```
schemas.ts    — Zod schemas + inferred TypeScript types
types.ts      — non-schema types (query params, discriminated unions)
api.ts        — apiFetch calls; each validates response with Zod
queries.ts    — TanStack Query hooks + query key factories
components/   — React components scoped to this feature
__tests__/    — Vitest unit tests
```

Features must not import from each other. The one documented exception is the cross-feature trace deep-link primitive that lands with the audit feature.

### Shared layer (`src/shared/`)

- `api/client.ts` (Phase 2) — `apiFetch`: prepends `/api/v1`, attaches headers, throws `ApiError` on non-2xx, records the response `X-Trace-Id` on a singleton.
- `auth/auth.ts` (Phase 3) — NextAuth v5 config.
- `casl/abilities.ts` (Phase 3) — CASL `defineAbility()` driven by `/user/privileges`.
- `layout/nav.ts` — `portalNav` array; sidebar items declare an optional `ability` gate.
- `providers/Providers.tsx` — Root provider tree.
- `hooks/useShallowSearchParams.ts` — Drop-in for `useSearchParams()` whose writes don't trigger an RSC roundtrip.
- `ui/` — shadcn/ui components (style: `base-nova`). Add new components with `pnpm dlx shadcn@latest add <component>`.

### MSW (Mock Service Worker)

With `NEXT_PUBLIC_PORTAL_USE_MSW=true`, MSW boots in the browser via `src/mocks/browser.ts`. Handlers live in `src/mocks/handlers.ts` and aggregate per-feature handlers from later phases.

### Design tokens

CSS custom properties in `design-tokens/` are the source of truth for color, spacing, radius, and typography. `design-tokens/tokens.json` is the machine-readable version. A Vitest smoke test (`src/shared/__tests__/tokens.test.ts`) guards against regressions.

Light tokens are declared in `:root`; dark overrides live in `.dark`. Both selectors have equal specificity, so order in the source file matters — keep `.dark` overrides after the `:root` block they shadow.

## Code conventions

- **Biome** for lint + format. Config: `biome.json`. `noExplicitAny` is an error; `useImportType` is a warning (off inside `src/shared/ui/**`).
- **Zod** validation happens at the API boundary only.
- **Query keys** follow the factory pattern: each feature exports a `<feature>Keys` object with `all`, `list(params)`, `detail(id)` methods.
- **URL state** uses `useShallowSearchParams` for filter/drawer/tab state.
- **CASL** for permission gating.
- **Server-only code** (NextAuth, backend proxy, env validation) imports `"server-only"` to prevent accidental bundling into client code.

## Pitfalls

- **Don't gate UI with `session.role === "admin"`.** Use `<Can I="manage" a="Site">` or `useAbility()`.
- **Don't re-validate Zod schemas inside components.** Validation is at the API boundary (`api.ts`); components consume typed data and trust it.
- **Don't `router.replace` for filter/drawer/tab state.** Use `useShallowSearchParams`.
- **Don't import across `src/features/<name>/`.** The only sanctioned cross-feature import is the trace deep-link primitive.
- **Don't add `.dark` token overrides above the `:root` block they shadow.** Source order wins for equal specificity.
- **Don't write WHAT a piece of code does in a comment.** Comment only when WHY isn't obvious. Two-line max. No change-log comments.

## References (single source of truth)

The prototype repo holds the architecture decision records and glossary that govern this portal. Cite them from there rather than copying:

- ADR-0001 — CASL over role strings
- ADR-0002 — Zod at the boundary
- ADR-0003 — MSW browser-only
- ADR-0004 — Feature isolation
- ADR-0005 — Shallow URL state
- `docs/glossary.md` — domain vocabulary
- `docs/features/tracing.md` — tracing feature spec

- Prototype repo: maintainer-local sibling checkout (cite ADR paths from there)
- Portal scaffold spec: `docs/internal/portal/2026-06-10-portal-scaffold.md`
