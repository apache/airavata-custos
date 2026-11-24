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

# AMIE Processor

The AMIE Processor keeps a site's Custos records in step with ACCESS-CI. Enable it on a site that serves ACCESS
allocations.

ACCESS sends account changes over AMIE (Account Management Information Exchange). Each message is a *packet*, such as
an awarded project, an added user or a merged person. The connector applies each packet to Custos users, projects and
allocations, then replies to ACCESS.

The [glossary](../../../docs/glossary.md) defines domain terms such as PI, DN, site code, survivor and retiree.

## How it works

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

1. The poller asks the AMIE API for in-progress packets once per `poll_interval` (30s by default).
2. It stores each new packet with a processing event in status `NEW`.
3. The processor picks up to 50 due events once per `worker_interval` (5s by default).
4. Each event runs in one database transaction.
5. The router dispatches the event to the handler for its packet type.
6. The handler calls the core service, then replies to AMIE.
7. A failed event is retried with exponential backoff from 30s, capped at 600s, for at most 3 attempts.
8. An invalid packet is not retried. The connector ends its transaction at ACCESS with a failure reply.

A `request_account_create` reply waits for an administrator to approve the cluster account. Until then the packet
sits in `WAITING_APPROVAL`. An account already approved gets its reply at once. A denial sends ACCESS a failure with the
administrator's reason. The packet then becomes `REFUSED`.

### Packet handlers

| Handler                   | AMIE packet type              | Reply type                    |
|---------------------------|-------------------------------|-------------------------------|
| RequestProjectCreate      | `request_project_create`      | `notify_project_create`       |
| RequestAccountCreate      | `request_account_create`      | `notify_account_create`       |
| RequestProjectInactivate  | `request_project_inactivate`  | `notify_project_inactivate`   |
| RequestProjectReactivate  | `request_project_reactivate`  | `notify_project_reactivate`   |
| RequestAccountInactivate  | `request_account_inactivate`  | `notify_account_inactivate`   |
| RequestAccountReactivate  | `request_account_reactivate`  | `notify_account_reactivate`   |
| RequestPersonMerge        | `request_person_merge`        | `inform_transaction_complete` |
| RequestUserModify         | `request_user_modify`         | `inform_transaction_complete` |
| DataProjectCreate         | `data_project_create`         | `inform_transaction_complete` |
| DataAccountCreate         | `data_account_create`         | `inform_transaction_complete` |
| InformTransactionComplete | `inform_transaction_complete` | *(none, terminal)*            |
| NoOp                      | *(catch-all)*                 | *(none)*                      |

### Usernames

The connector generates each ACCESS account's local username with `pkg/posix`:

1. `BuildBase` builds `<prefix>-<first initial><last name>`, lowercase ASCII and digits only. With one name empty it
   uses the other. The prefix is `POSIX_USERNAME_PREFIX`, `custos` by default. A name too long for the 32-character
   POSIX limit is truncated and `PosixUsernameTruncated` is written. With both names empty it writes
   `PosixUsernameBuildFailed` and fails.
2. The handler creates the cluster account with the base name. If `(compute_cluster_id, local_username)` is taken, it
   tries `base2`, `base3` and so on, up to `MaxCollisionSuffix` (999).
3. If all suffixes are taken, it writes `PosixUsernameAllocatorExhausted` and fails.

Core onboarding uses the same generator when a user leaves the username empty. It writes none of these audit events.
With both names empty, it uses the email local part instead of failing.

### Allocation types

A `request_project_create` carries an `AllocationType` and an amount in service units (SUs). The AMIE specification
(section `notify_project_create`) defines what each type means for the site.

| Type         | No allocation yet                 | Existing allocation                           |
|--------------|-----------------------------------|-----------------------------------------------|
| `new`        | create the allocation             | add the SUs                                   |
| `transfer`   | create, as a supplement           | add the SUs; a negative amount is deducted    |
| `supplement` | create                            | add the SUs, dates unchanged                  |
| `renewal`    | create, with the given dates      | replace the SUs and dates from the start date |
| `extension`  | not expected                      | change the end date only                      |
| `adjustment` | not expected                      | add or deduct, like a transfer                |
| `advance`    | create                            | add the SUs                                   |

Negative SUs for a project with no allocation is an invalid packet.

## Configuration

The connector reads the `connectors.amie-processor` block of [`config/custos.yaml`](../../../config/custos.yaml). A
missing or empty YAML value falls back to the variable in the same row.

| YAML key                   | Fallback variable      | Default | Purpose                                          |
|----------------------------|------------------------|---------|--------------------------------------------------|
| `credentials.base_url`     | `AMIE_BASE_URL`        |         | AMIE API endpoint, or the local mock             |
| `credentials.site_code`    | `AMIE_SITE_CODE`       |         | ACCESS site code (`TESTSITE` for the mock)       |
| `credentials.api_key`      | `AMIE_API_KEY`         |         | API key issued by ACCESS (`dev` for the mock)    |
| `cluster.id`               | `AMIE_CLUSTER_ID`      |         | ID of the compute cluster this site serves       |
| `polling.poll_interval`    | `AMIE_POLL_INTERVAL`   | `30s`   | How often the poller asks AMIE for packets       |
| `polling.worker_interval`  | `AMIE_WORKER_INTERVAL` | `5s`    | How often the processor picks up due events      |
| `polling.poller_enabled`   | `AMIE_POLLER_ENABLED`  | `true`  | `false` turns off polling; stored events still run |
| `timeouts.connect_timeout` | `AMIE_CONNECT_TIMEOUT` | `5s`    | HTTP connect timeout                             |
| `timeouts.read_timeout`    | `AMIE_READ_TIMEOUT`    | `60s`   | HTTP read timeout                                |
| `resource_aliases`         |                        | `{}`    | Maps ACCESS resource names to catalog resources  |

Without all three credentials the connector does not start. It logs a warning and reports the gap on `GET /connectors`.
Without a cluster ID it starts, but the project and account create handlers fail.

A packet's `ResourceList` names match the cluster's resource catalog by name. Map any other ACCESS name under
`resource_aliases`. An alias names one catalog resource or a list, and the allocation attaches to each:

```yaml
resource_aliases:
  my-cpu.site.access-ci.org: cpu-general
  my-gpu.site.access-ci.org: [gpu-a100, gpu-h100]
```

The connector shares the core database connection and migrates its own tables at startup. The core server sets the log
level and format.

## Operations

The connector adds these routes to the API port (default `:8080`). [docs/privileges.md](../../../docs/privileges.md)
lists the privilege each one needs.

| Route | Purpose |
|---|---|
| `GET /connectors/amie/packets`, `GET /connectors/amie/packets/{id}` | List and read packets |
| `GET /connectors/amie/packets/{id}/events` | A packet's processing events |
| `GET /connectors/amie/packets/{packet_id}/audits` | A packet's audit rows |
| `GET /connectors/amie/stats` | Packet counts |
| `GET /connectors/amie/replies`, `GET /connectors/amie/unmapped` | Placeholders returning an empty page |
| `POST /connectors/amie/packets/{id}/retry`, `.../packets/{id}/resolve`, `.../replies/{id}/retry`, `.../unmapped/{id}/link` | Reserved, return `501 Not Implemented` |

Audit rows carry `source=amie`. `amie_audit_extras` holds each row's packet and event IDs ([audit conventions](../../../docs/architecture.md#audit-conventions)).

The connector records Prometheus metrics in the default registry: `amie_packets_received_total`,
`amie_packets_processed_total`, `amie_events_retry_total`, `amie_packet_processing_duration_seconds` and
`amie_poller_packets_fetched`. The server has no `/metrics` route, so nothing scrapes them. The Grafana dashboard
`dev-ops/compose/grafana/dashboards/amie-service.json` and the `access-amie` Prometheus job are unused.

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| Log shows `AMIE credentials not fully provided, skipping AMIE connector` | A credential is empty | Set all three credentials |
| Project and account packets fail; log warns that the cluster id is not set | No `cluster.id` | Set `cluster.id` or `AMIE_CLUSTER_ID` |
| A packet sits in `WAITING_APPROVAL` | Its cluster account awaits approval | Approve or deny the account |
| An allocation has no resource | `ResourceList` names no catalog resource | Add the name under `resource_aliases` |
| An event is `PERMANENTLY_FAILED` | Three attempts failed | Fix the cause found in the packet's events and audits. The retry route returns `501` |

## Development

### Run against the mock AMIE server

The [mock server](./mock-server/README.md) queues packets on request. Use it to fill a running server or drive
traffic. It asserts nothing.

1. Start the database:

   ```bash
   # From the repo root
   docker compose -f dev-ops/compose/docker-compose.yml up db -d
   ```

2. Start the mock as [its README](./mock-server/README.md#mock-server) describes.
3. Point the connector at the mock:

   ```bash
   export AMIE_BASE_URL="http://localhost:8180"
   export AMIE_SITE_CODE="TESTSITE"
   export AMIE_API_KEY="dev"
   ```

4. Start the Custos server as [INSTALL.md](../../../INSTALL.md) describes. The connector runs inside it.
5. Queue packets:

   ```bash
   curl -X POST 'http://localhost:8180/test/TESTSITE/scenarios?type=baseline'   # the 12-packet fixture
   ```

The mock's README lists the other scenarios and the k6 load test.

### Tests

Unit tests need no database, AMIE server or network:

```bash
go test ./connectors/ACCESS/AMIE-Processor/...
```

Integration tests carry the `integration` build tag. They need `DATABASE_DSN`, `AMIE_BASE_URL`, `AMIE_SITE_CODE`,
`AMIE_API_KEY` and `AMIE_CLUSTER_ID`.

| Package | Covers |
|---|---|
| `handler/` | Each handler |
| `worker/` | The poller and processor |
| `pipeline/` | The whole pipeline |

The runner starts an isolated database on `:5433` and the mock on `:8181`
([`dev-ops/local-amie`](../../../dev-ops/local-amie/README.md)). It runs the suite, then tears both down:

```bash
# From the repo root
make integration-test-amie        # or scripts/run-amie-integration-tests.sh
```

The pipeline test, `TestPipeline_BaselineDeterminism`, fires the `baseline` scenario and asserts the resulting row
counts. [`testdata/scenarios/README.md`](./testdata/scenarios/README.md) describes the fixture and what it leaves
untested.
To add a scenario, add a generator to the mock and a YAML next to `baseline.yaml`, then call
`pipe.fireScenario(t, "<name>")` from a new `*_integration_test.go` file.
