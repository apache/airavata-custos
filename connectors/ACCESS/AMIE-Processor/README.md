# access-amie

ACCESS-CI AMIE packet processing service for Apache Airavata Custos.

This Go service polls the [ACCESS-CI](https://access-ci.org/) AMIE (Account Management Information Exchange) API for allocation packets, processes them through typed handlers, manages person/project/account lifecycle, and replies to AMIE. It is the ACCESS allocation source adapter within the Custos allocations platform.

## Prerequisites

- Go 1.22+
- MariaDB — use the shared dev stack: `docker compose -f compose/docker-compose.yml up db` (creates the `access_ci` database automatically via `compose/dbinit/init-db.sh`)

## Quick Start

### 1. Start the database

```bash
# From the repo root
docker compose -f compose/docker-compose.yml up db -d
```

### 2. Configure

```bash
# Required for production — set your AMIE credentials:
export AMIE_SITE_CODE="YOUR_SITE"
export AMIE_API_KEY="your-api-key"
```

The default `config.yaml` works for local development with the Docker Compose MariaDB defaults.

For local dev without a real ACCESS endpoint, point the service at the local mock AMIE server in [`devtools/amie/`](../devtools/amie/README.md)

```bash
export AMIE_BASE_URL="http://localhost:8180"
export AMIE_SITE_CODE="TESTSITE"
export AMIE_API_KEY="dev"
```

### 3. Build

```bash
cd allocations/access-amie
go build -o bin/access-amie .
```

### 4. Run

```bash
./bin/access-amie
```

The service will connect to MariaDB, run migrations automatically, start the HTTP server on port 8083, the AMIE poller (every 30s), and the event processor (every 5s).

### 5. Verify

```bash
curl http://localhost:8083/health     # Health check (includes AMIE API status)
curl http://localhost:8083/ready      # Readiness check (DB ping)
curl http://localhost:8083/metrics    # Prometheus metrics
```

## Testing

Three layers, used for different things:

### 1. Unit tests (correctness, no external services)

```bash
make test           # all tests, verbose
make test-short     # short mode
```

Every package under this connector ships unit tests against `testify/mock`. No DB, no AMIE server, no network. Run on every commit.

### 2. Integration tests (real DB + mock AMIE server)

Run the per-handler suite (~11 handler tests) plus the pipeline suite (`baseline_integration_test`, `comanage_trace_chain_test`, `tracing_clean_slate_test`). All gated by build tag `integration` and require these env vars set:

- `DATABASE_DSN`, `AMIE_BASE_URL`, `AMIE_SITE_CODE`, `AMIE_API_KEY`, `AMIE_CLUSTER_ID`

The canonical runner brings up an isolated DB on `:3307` + mock AMIE on `:8181`, applies all migrations, fires every integration test, then tears down:

```bash
# From repo root
make integration-test-amie
```

Equivalent script: `scripts/run-amie-integration-tests.sh`.

#### What the pipeline tests fire

Only one fixture exists today: `testdata/scenarios/baseline.yaml` — a deterministic 9-packet flow covering all major handler paths:

| Packet | Asserts |
|---|---|
| `request_project_create` | PI user + project + allocation + PI cluster account |
| `request_account_create` | user + cluster account + membership |
| `data_project_create` / `data_account_create` | PERSIST_DNS rows |
| `request_user_modify` | user + DN updates |
| `request_person_merge` | survivor / retiree consolidation |
| `request_account_inactivate` / `_reactivate` | membership status flips |
| `request_project_inactivate` / `_reactivate` | project + all-member flips |
| `inform_transaction_complete` | TRANSACTION_COMPLETE row |

The three pipeline tests reuse the same fixture but assert different properties:
- `baseline_integration_test` — DB shape + idempotency on rerun.
- `comanage_trace_chain_test` — proves AMIE and COmanage audit rows share a `trace_id` over the unified `audit_events` table (uses mock COmanage REST; never hits the real registry).
- `tracing_clean_slate_test` — clean-slate DB + full audit coverage + `/audit/*` endpoint smoke against an in-process httptest server.

To add a new scenario: drop a YAML next to `baseline.yaml` and call `pipe.fireScenario(t, "<name>")` from a new `*_integration_test.go` file.

### 3. Mock-server REST scenarios + k6 (load / soak / observability)

For traffic generation against a **running** server (NOT a test — no assertions). Used to play in the admin UI, watch Grafana, or stress the worker.

Start the mock AMIE server standalone:

```bash
cd connectors/ACCESS/AMIE-Processor/mock-server
source venv/bin/activate
python3 mock-amie-server.py            # :8180
```

Then queue packets via REST:

```bash
curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=mixed'         # success + failure mix
curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=success_only'
curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=failures_only'
curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=heavy'          # large batch
curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=dev_email'      # needs DEV_EMAIL env
```

For sustained traffic at a configurable rate, use k6 with `mock-server/amie-traffic.js` — see `mock-server/README.md` for stage configuration.

## Observability

The service exports Prometheus metrics at `/metrics`. A pre-built Grafana dashboard is available at `compose/grafana/dashboards/amie-service.json` showing packet processing stats, failures, retries, and processing duration percentiles.

To run the metrics stack:

```bash
# From the repo root
docker compose -f dev-ops/compose/docker-compose.yml up db prometheus grafana -d
```

Then open Grafana at `http://localhost:3000` (admin/admin). The AMIE dashboard loads automatically.

For request-flow tracing, every AMIE audit row carries `trace_id` / `span_id` / `parent_span_id` and lives in the core `audit_events` table (source=`amie`). The connector-specific `packet_id` / `event_id` references live in `amie_audit_extras` joined on `audit_event_id`. The admin trace view at `/audit/traces*` queries `audit_events` directly and renders the hierarchy from `parent_span_id`. For per-packet drill-down, see `GET /connectors/amie/packets/{packet_id}/audits`.

## Architecture

### Packet Processing Pipeline

```
AMIE API ──poll──> Poller ──persist──> DB (packets + events)
                                            │
                              Processor ◄───┘
                                │
                          Router (type-based dispatch)
                                │
                    ┌───────────┼───────────┐
                    ▼           ▼           ▼
              Handler A    Handler B    Handler N
              (services)   (services)   (services)
                    │           │           │
                    └───────────┼───────────┘
                                │
                          Reply to AMIE
```

1. **Poller** polls the AMIE API every 30s for in-progress packets
2. Each new packet is persisted with a processing event (status: NEW)
3. **Processor** picks up pending events every 5s (batch of 50)
4. Events are processed inside a database transaction
5. The **Router** dispatches to the correct handler by packet type
6. Handlers call domain services, then send a reply to AMIE
7. On failure: exponential backoff retry (30s, 60s, 600s cap, max 3 attempts)

### Packet Handlers

| Handler | AMIE Packet Type | Reply Type |
|---------|-----------------|------------|
| RequestProjectCreate | `request_project_create` | `notify_project_create` |
| RequestAccountCreate | `request_account_create` | `notify_account_create` |
| RequestProjectInactivate | `request_project_inactivate` | `notify_project_inactivate` |
| RequestProjectReactivate | `request_project_reactivate` | `notify_project_reactivate` |
| RequestAccountInactivate | `request_account_inactivate` | `notify_account_inactivate` |
| RequestAccountReactivate | `request_account_reactivate` | `notify_account_reactivate` |
| RequestPersonMerge | `request_person_merge` | `inform_transaction_complete` |
| RequestUserModify | `request_user_modify` | `inform_transaction_complete` |
| DataProjectCreate | `data_project_create` | `inform_transaction_complete` |
| DataAccountCreate | `data_account_create` | `inform_transaction_complete` |
| InformTransactionComplete | `inform_transaction_complete` | *(none — terminal)* |
| NoOp | *(catch-all)* | *(none)* |

## Configuration Reference

| Setting | YAML Key | Env Var | Default |
|---------|----------|---------|---------|
| HTTP port | `server.port` | `SERVER_PORT` | `8083` |
| Database DSN | `database.dsn` | `DATABASE_DSN` | `admin:admin@tcp(localhost:3306)/access_ci?parseTime=true&loc=UTC` |
| AMIE API base URL | `amie.base_url` | `AMIE_BASE_URL` | `https://a3mdev.xsede.org/amie-api-test` |
| AMIE site code | `amie.site_code` | `AMIE_SITE_CODE` | *(required)* |
| AMIE API key | `amie.api_key` | `AMIE_API_KEY` | *(required)* |
| Poll interval | `amie.poll_interval` | — | `30s` |
| Worker interval | `amie.worker_interval` | — | `5s` |
| Log level | `log.level` | `LOG_LEVEL` | `info` |
| Log format | `log.format` | `LOG_FORMAT` | `text` |
| Provisioner type | `provisioner.type` | — | `noop` |

## Go Workspace

This module is part of a Go workspace defined in `allocations/go.work`:

```
allocations/
  go.work              Workspace root
  provisioner/         Shared HPC provisioner interface (zero deps)
  access-amie/         This module (ACCESS-CI AMIE adapter)
```

The `provisioner/` module defines the `Provisioner` interface for HPC cluster account and project provisioning. Currently using a no-op stub; a future SLURM provisioner will implement it.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](../../LICENSE) for details.
