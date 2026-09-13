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

# SSH Certificate Signer web zone

This directory contains the signer-owned Next.js application. It uses the
`/signer` base path and is exposed through the Custos portal as a Next.js
multi-zone application.

## Development

Start the signer service on port 8084, then run the zone on port 3001:

```bash
cp .env.example .env.local
pnpm install
pnpm dev --port 3001
```

In the portal's `.env.local`, register the signer manifest:

```dotenv
CUSTOS_EXTENSION_MANIFEST_URLS=http://localhost:8084/.well-known/custos-extension.json
```

Start the portal on port 3000. Requests under `/signer/*` are rewritten to
the zone while remaining on the portal origin. The portal and zone must share
`NEXTAUTH_SECRET`; the session cookie is scoped to `/`.

The zone's browser code calls `/signer/api/v1/*`. That server-side BFF decodes
the shared session cookie and forwards the selected bearer token to the signer
API, so tokens are never exposed to client components.

## Verification

```bash
pnpm lint
pnpm typecheck
pnpm test
pnpm build
pnpm test:e2e
```

The Playwright suite starts a deterministic manifest endpoint, the signer
zone, and the portal to verify the complete multi-zone flow.

On Windows, Playwright may print a passing result and then remain at
`Terminating the WebServer` while shutting down the pnpm-launched development
servers. This is a local process-teardown issue, not a stalled test. The Linux
CI job is the authoritative end-to-end result.
