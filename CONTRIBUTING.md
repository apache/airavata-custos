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

# Contributing

Thanks for your interest in contributing to Apache Airavata Custos. This document covers the repository layout, build commands, coding conventions, and the pull request workflow.

For instructions on running Custos locally against a database, see [INSTALL.md](INSTALL.md). For configuration keys,
see [CONFIG.md](CONFIG.md).

## Prerequisites

- Go **1.27+**
- `git`
- Docker and Docker Compose (integration tests and the local stack)
- Node **24+** and **pnpm 12+** (the portal under `web/`)

## Repository layout

- `cmd/server/` — main HTTP API entry point
- `internal/` — server, store, and database wiring (not importable outside the module)
- `pkg/` — public packages (models, service, events)
- `connectors/` — protocol/site-specific connectors (ACCESS, SLURM, …)
- `extensions/` — out-of-process extensions (PAM module, SSH cert signer, …)
- `dev-ops/compose/` — local Docker Compose stack (PostgreSQL, Keycloak, Adminer, Prometheus, Grafana, Vault)

The [architecture doc](docs/architecture.md#repository-layout) describes the layout.
[docs/contributing/writing-a-connector.md](docs/contributing/writing-a-connector.md) covers adding a connector.

## Build and check

From the repository root:

```bash
gofmt -l .        # must print nothing
go build ./...
go vet ./...
go test ./...
```

All four should be clean before opening a pull request. CI runs the same checks
([.github/workflows/README.md](.github/workflows/README.md)).

The Makefile wraps the common tasks:

| Target | Runs |
|---|---|
| `make gen-api` | `go generate ./...`, which regenerates the OpenAPI specs, the service mock and the event graph |
| `make build` | `gen-api`, then builds the `custos` binary and runs `go build ./...` |
| `make verify-no-drift` | `gen-api`, then fails on spec drift ([api/README.md](api/README.md)) |
| `make test` | `go test ./...` |
| `make verify` | `verify-no-drift`, `go vet ./...` and `make test` |
| `make integration-test-amie` | the AMIE connector integration suite in its own stack |

After changing a handler's swag annotations, run `make gen-api` and commit the regenerated spec.
[api/README.md](api/README.md) explains how the specs are produced.

## Database migrations

Migrations live in `internal/db/migrations/` and are embedded into the binary via `//go:embed`. They are applied automatically on server start.

To add a new migration, create a matching pair following the existing six-digit numbering:

```
internal/db/migrations/000009_short_name.up.sql
internal/db/migrations/000009_short_name.down.sql
```

Every SQL file must carry the Apache 2.0 license header (see existing files for the exact comment block).

A connector keeps its migrations next to its own code. The connector guide covers them.

## Coding conventions

- Every new Go or SQL file must include the Apache 2.0 license header used throughout the repo.
- Service-layer errors use the sentinel errors in `pkg/service` (`ErrInvalidInput`, `ErrNotFound`, `ErrAlreadyExists`); HTTP handlers translate them to 400/404/409 respectively with `common.WriteServiceError` ([pkg/README.md](pkg/README.md#errors)).
- Database work happens inside `s.inTx(ctx, func(*sql.Tx) error { … })`. Do not nest transactions.
- HTTP routes register on `identity.Router` ([architecture](docs/architecture.md#identity-and-authentication)), with `r.PathValue` for path parameters.
- Publish domain events through `pkg/events.Bus` from the service layer; never from handlers or stores.
- An audit event that records a failure must have `Failed`, `Error` or `Rejected` in its event type (for example `ComanageProvisioningFailed`, `PosixUsernameBuildFailed`). The trace view reads a step's outcome from its name (`internal/tracing/status.go`), so a failure named otherwise shows as ok.

## Tests

Two tiers: unit tests and integration tests, separated by a build tag.

### Unit tests

```bash
go test ./...
```

Runs every package's unit tests. Handlers and stores use `httptest` and mocks.
No DB, no network. This is what should pass on every commit.

### Integration tests

Integration test files carry `//go:build integration` at the top, so the default
`go test ./...` skips them silently. To run them, pass the tag and provide the
required env vars. The core packages share one database and truncate tables, so they must run one package at a
time with `-p 1`:

```bash
# General integration tests (server, core service, identity resolver).
# Reuses the dev compose stack on :5432 (custos_test database).
export CORE_TEST_DATABASE_DSN='postgres://admin:admin@localhost:5432/custos_test?sslmode=disable'
go test -tags integration -p 1 ./internal/... ./pkg/...
```

Run this in a shell where `.env` is not sourced. The AMIE tests read `DATABASE_DSN` and the `AMIE_*` variables.
With `.env` loaded, they would truncate the development database.

The AMIE connector ships its own integration stack because its tests mutate a
lot of state and we don't want to corrupt the dev DB. It brings up an isolated
PostgreSQL on `:5433` and a mock AMIE server on `:8181`, runs the suite, then
tears it down:

```bash
make integration-test-amie
```

(equivalent to `scripts/run-amie-integration-tests.sh`). The `:5433` is
deliberate. It's a separate stack defined in `dev-ops/local-amie/`, independent
of the `:5432` dev compose. If a run leaves the stack up, `make integration-test-amie-down` removes it.

If you see a test file you expect to run but `go test ./...` reports zero tests
for the package, check the first line for `//go:build integration`.

### Portal tests

See [web/README.md](web/README.md).

## Submitting changes

1. Open an issue describing the change, if one does not already exist.
2. Create a topic branch off `master`.
3. Make focused, well-scoped commits with clear messages.
4. Run the checks in [Build and check](#build-and-check), and the portal checks if you changed `web/`.
5. Open a pull request and link the related issue.
