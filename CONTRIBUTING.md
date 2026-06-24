# Contributing

Thanks for your interest in contributing to Apache Airavata Custos. This document covers the repository layout, build commands, coding conventions, and the pull request workflow.

For instructions on running Custos locally against a database, see [INSTALL.md](INSTALL.md).

## Prerequisites

- Go **1.24+**
- `git`

## Repository layout

- `cmd/server/` — main HTTP API entry point
- `internal/` — server, store, and database wiring (not importable outside the module)
- `pkg/` — public packages (models, service, events)
- `connectors/` — protocol/site-specific connectors (ACCESS, SLURM, …)
- `extensions/` — out-of-process extensions (PAM module, SSH cert signer, …)
- `dev-ops/compose/` — local Docker Compose stack (MariaDB, Adminer, Prometheus, Grafana, Vault)

## Build

From the repository root:

```bash
go build ./...
go vet ./...
go test ./...
```

All three should be clean before opening a pull request.

## Database migrations

Migrations live in `internal/db/migrations/` and are embedded into the binary via `//go:embed`. They are applied automatically on server start.

To add a new migration, create a matching pair following the existing numbering:

```
internal/db/migrations/NNN_short_name.up.sql
internal/db/migrations/NNN_short_name.down.sql
```

Every SQL file must carry the Apache 2.0 license header (see existing files for the exact comment block).

## Coding conventions

- Every new Go or SQL file must include the Apache 2.0 license header used throughout the repo.
- Service-layer errors use the sentinel errors in `pkg/service` (`ErrInvalidInput`, `ErrNotFound`, `ErrAlreadyExists`); HTTP handlers translate them to 400/404/409 respectively.
- Database work happens inside `s.inTx(ctx, func(*sql.Tx) error { … })`. Do not nest transactions.
- HTTP routing uses the Go 1.22 `ServeMux` with `r.PathValue` for path parameters.
- Publish domain events through `pkg/events.Bus` from the service layer; never from handlers or stores.

## Tests

Two tiers: unit tests and integration tests, separated by a build tag.

### Unit tests (default, no external services)

```bash
go test ./...
```

Runs every package's unit tests. Handlers and stores use `httptest` and mocks.
No DB, no network. This is what should pass on every commit.

### Integration tests (build tag, real services)

Integration test files carry `//go:build integration` at the top, so the default
`go test ./...` skips them silently. To run them, pass the tag and provide the
required env vars:

```bash
# General integration tests (server, core service, identity resolver).
# Reuses the dev compose stack on :3306.
export CORE_TEST_DATABASE_DSN='admin:admin@tcp(localhost:3306)/custos?parseTime=true&charset=utf8mb4&multiStatements=true'
go test -tags integration ./...
```

The AMIE connector ships its own integration stack because its tests mutate a
lot of state and we don't want to corrupt the dev DB. It brings up an isolated
MariaDB on `:3307` and a mock AMIE server on `:8181`, runs the suite, then
tears it down:

```bash
make integration-test-amie
```

(equivalent to `scripts/run-amie-integration-tests.sh`). The `:3307` is
deliberate. It's a separate stack defined in `dev-ops/local-amie/`, independent
of the `:3306` dev compose.

If you see a test file you expect to run but `go test ./...` reports zero tests
for the package, check the first line for `//go:build integration`.

## Submitting changes

1. Open an issue describing the change, if one does not already exist.
2. Create a topic branch off `master`.
3. Make focused, well-scoped commits with clear messages.
4. Ensure `go build ./...`, `go vet ./...`, and `go test ./...` all pass.
5. Open a pull request and link the related issue.
