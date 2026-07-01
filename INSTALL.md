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

# Local Installation

End-to-end guide to running Apache Airavata Custos locally: database, OIDC, API server, and portal. For coding
conventions and contribution workflow, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Prerequisites

- Go **1.24+**
- Node **20+** and **pnpm** (for the portal)
- Docker and Docker Compose
- `git`

## Pick an OIDC provider

Custos is OIDC-generic and verifies bearer tokens against whatever issuer you point it at. For local development you
have two convenient options:

- **Bundled Keycloak** – the Compose stack ships a Keycloak service with a seeded realm and dev users. Zero external
  setup, fastest for feature work.
- **Any external OIDC provider** – register a client with your IdP whose redirect URI is
  `http://localhost:3000/api/auth/callback/oidc`, then plug the issuer, client_id, and client_secret into the env files.

The steps below cover both. Where they differ, they're labeled
**Bundled Keycloak** / **External OIDC**.

## 1. Start the database

From `dev-ops/compose/`:

```bash
docker compose up -d db
```

This starts MariaDB on `localhost:3306` and runs `dbinit/init-db.sh`, which
creates the `custos` database and the `admin` user (password `admin`).

**Bundled Keycloak only:** also bring up Keycloak. On first start it imports
the realm definition from
`dev-ops/compose/keycloak/import/custos-realm.json` (OIDC client + dev users inside Keycloak itself).

```bash
docker compose up -d keycloak
```

Keycloak will be available at `http://localhost:8081` once healthy (~60s).
Skip this if you're using an external OIDC provider.

Stop everything later with `docker compose down`, or `docker compose down -v`
to also wipe the DB volume.

## 2. Configure backend env

From the repo root:

```bash
cp .env.example .env
```

Edit `.env`:

- **Bundled Keycloak:** the defaults already point at the local realm, so no
  edits are needed. Bootstrap admin email stays `admin@custos.local` (matches
  the seeded realm user).
- **External OIDC:** register a client at your IdP with redirect
  `http://localhost:3000/api/auth/callback/oidc`, then:
    - Set `OIDC_ISSUER_URL` to the issuer URL.
    - Set `OIDC_AUDIENCE` to whatever your IdP puts in the access-token `aud`
      claim (typically your client_id).
    - Set `CUSTOS_BOOTSTRAP_ADMIN_EMAIL=<your email>`. A PENDING super_admin
      user is created on first boot; your first sign-in links your OIDC
      `sub` to it via email fallback.

Source it: `set -a && . ./.env && set +a`.

## 3. Run the API server

From the repo root:

```bash
go run ./cmd/server
```

On startup, the server opens the DB, runs the embedded migrations
(`internal/db/migrations/`), grants super_admin to
`CUSTOS_BOOTSTRAP_ADMIN_EMAIL` (creating a PENDING user if needed), wires the
event bus + service layer, and listens on the port from `config/custos.yaml`
(`core.api.port`, default `8080`).

Leave it running.

## 4. Apply the cluster seed

The migrations create the schema; the cluster seed inserts the default
`compute_clusters` row that connectors and portal features reference by ID.
Apply it once, in a second terminal, after the server has come up:

```bash
docker exec -i custos_db mariadb -uadmin -padmin custos < dev-ops/compose/seeds/default_cluster.sql
```

The seed uses `INSERT IGNORE`, so re-running is safe.

## 5. (Bundled Keycloak only) Seed dev users

For the bundled Keycloak path there is also a sample identity seed (org,
users, roles, role assignments) matched to the users in the imported realm:

```bash
docker exec -i custos_db mariadb -uadmin -padmin custos < dev-ops/compose/seeds/dev_users_and_roles.sql
```

Skip this for external OIDC. You don't need it, and the emails would collide
with the identity you actually sign in as.

## 6. Configure and run the portal

```bash
cd web
cp .env.example .env.local
pnpm install
```

Edit `web/.env.local`:

- Generate `NEXTAUTH_SECRET` with `openssl rand -base64 32`.
- **Bundled Keycloak:** defaults are correct.
- **External OIDC:** set `OIDC_ISSUER_URL`, `OIDC_CLIENT_ID`, and
  `OIDC_CLIENT_SECRET` to the values from the client you registered in step 2.

Start the portal:

```bash
pnpm dev
```

## 7. Sign in

Open <http://localhost:3000> and sign in:

- **Bundled Keycloak:** use one of the seeded users (e.g. `admin@custos.local`
  / `admin`; see the realm import for the full list).
- **External OIDC:** sign in with your provider. On first sign-in your OIDC
  `sub` is linked to the PENDING super_admin user created in step 3, and the
  user is promoted to ACTIVE.
