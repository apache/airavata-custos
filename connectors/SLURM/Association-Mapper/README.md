# SLURM Association-Mapper

SLURM association creation logic lives in this plugin. It is triggered when the allocation manager has processed an allocation request and released it to downstream handlers. It talks to `slurmrestd` to manage accounts, associations, and TRES limits.

This package is part of the root `github.com/apache/airavata-custos` module.

## Prerequisites

- Go **1.24+**
- A reachable `slurmrestd` endpoint (for integration runs) plus a SLURM user name and JWT token

## Layout

```
.
├── internal/operations/   # slurmrestd client + accounts/associations/TRES
└── pkg/operations/
```

Import path: `github.com/apache/airavata-custos/connectors/SLURM/Association-Mapper/internal/operations`.

## Test

```bash
# from the repository root
go test ./connectors/SLURM/Association-Mapper/...
go vet ./connectors/SLURM/Association-Mapper/...
```

Tests are hermetic and use `httptest` — no live `slurmrestd` required.
