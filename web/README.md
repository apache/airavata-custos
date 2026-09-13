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

Web portal for **Apache Custos** — allocation management, identity, and admin tooling for HPC sites.

## Stack

Next.js 15 (App Router) · TypeScript · Tailwind 4 + shadcn/ui · NextAuth v5 · TanStack Query · React Hook Form + Zod · CASL · MSW · Playwright + Vitest · Biome · pnpm.

## Getting started

```bash
pnpm install
cp .env.example .env.local
pnpm dev
```

## Extension zones

The portal discovers independently deployed UI extensions from the
comma-separated `CUSTOS_EXTENSION_MANIFEST_URLS` setting. Every configured
manifest is mandatory: an unavailable or invalid manifest fails configuration
instead of silently dropping navigation or rewrites.

For the SSH Certificate Signer development setup, start its Go service on
`:8084`, its signer-owned Next.js app on `:3001`, and configure:

```dotenv
CUSTOS_EXTENSION_MANIFEST_URLS=http://localhost:8084/.well-known/custos-extension.json
```

The portal then serves the signer zone at `/signer/*`. Cross-zone navigation
uses a normal HTML link so the destination zone loads its own Next.js runtime.

See [CLAUDE.md](./CLAUDE.md) for commands, env modes, and architecture overview.
