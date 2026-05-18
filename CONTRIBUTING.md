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
- `dev-ops/compose/` — local Docker Compose stack (MariaDB, Keycloak, Vault, Prometheus, Grafana)

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

Run the full test suite from the repo root:

```bash
go test ./...
```

Connector packages (for example `connectors/SLURM/Association-Mapper/...`) use `httptest` and do not require external services.

## Submitting changes

1. Open an issue describing the change, if one does not already exist.
2. Create a topic branch off `main`.
3. Make focused, well-scoped commits with clear messages.
4. Ensure `go build ./...`, `go vet ./...`, and `go test ./...` all pass.
5. Open a pull request and link the related issue.
