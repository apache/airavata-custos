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

This guide covers running Apache Airavata Custos locally for development and testing. For coding conventions, build commands, and contribution workflow, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Prerequisites

- Go **1.24+**
- Docker and Docker Compose (for the local MariaDB and supporting services)
- `git`

## Running locally with a database

The server requires a MariaDB/MySQL database. The easiest way to get one is via the bundled Docker Compose stack.

### 1. Start MariaDB

```bash
cd dev-ops/compose
docker compose up -d db
```

This starts MariaDB on `localhost:3306` and runs `dbinit/init-db.sh`, which creates the `custos` database and the `admin` user (password `admin`).

To stop it later:

```bash
docker compose down
```

Use `docker compose down -v` if you also want to wipe the database volume.

### 2. Run the API server

From the repository root:

```bash
export DATABASE_DSN='admin:admin@tcp(localhost:3306)/custos?parseTime=true&charset=utf8mb4&multiStatements=true'
go run ./cmd/server
```

On startup, the server:

1. Opens the database connection (`internal/db`).
2. Runs the embedded migrations (`internal/db/migrations/`).
3. Wires the event bus, service layer, and HTTP router.
4. Listens on the port from `config/custos.yaml` (`core.api.port`, default `8080`).

### 3. Seed dev users (optional)

The schema is created by the server's migrations on first run. After that, apply
the dev seed to populate sample identity data (org, users, roles, privileges):

```bash
docker exec -i custos_db mariadb -uadmin -padmin custos \
  < dev-ops/compose/seeds/dev_users_and_roles.sql
```

The seed uses `INSERT IGNORE`, so re-running is safe. See the seed file for
exactly what it inserts.

### Environment variables

See [.env.example](.env.example) for the template. Copy it to `.env`, fill in
the required values (`DATABASE_DSN`, `OIDC_ISSUER_URL`, `OIDC_AUDIENCE`), and
source it before running the server.

### Optional services

The Compose file also defines Adminer, Prometheus, Grafana, and Vault. Start them as needed, for example:

```bash
docker compose up -d adminer       # DB UI at http://localhost:18080
```
