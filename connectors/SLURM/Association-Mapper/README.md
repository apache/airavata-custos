# SLURM Association-Mapper

SLURM association creation logic lives in this plugin. It is triggered when the allocation manager has processed an allocation request and released it to downstream handlers. It talks to `slurmrestd` to manage accounts, associations, and TRES limits.

## Prerequisites

- Go **1.24+**
- A reachable `slurmrestd` endpoint (for integration runs) plus a SLURM user name and JWT token

## Layout

```
.
├── main.go                       # entry point
├── internal/operations/          # slurmrestd client + accounts/associations/TRES
├── go.mod
└── Makefile
```

Module path: `github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper`.

## Build

```bash
# from this directory
make build           # produces bin/association-mapper
# or directly:
go build -o bin/association-mapper .
```

## Run

```bash
make run             # build, then ./bin/association-mapper
```

The service starts, logs `association-mapper started`, and blocks until it receives `SIGINT` or `SIGTERM`.

## Test

```bash
make test            # go test ./...
go vet ./...         # static checks
```

Tests are hermetic and use `httptest` — no live `slurmrestd` required.

## Common make targets

| Target  | Description                          |
|---------|--------------------------------------|
| `build` | Compile the binary into `bin/`       |
| `run`   | Build and run                        |
| `test`  | Run all unit tests                   |
| `tidy`  | `go mod tidy`                        |
| `clean` | Remove `bin/`                        |
