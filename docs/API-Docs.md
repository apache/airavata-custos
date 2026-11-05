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

# Custos REST API

The entry point to the HTTP/JSON API of `cmd/server`: conventions, endpoint groups and an example session. Read it
before you write a client. The generated OpenAPI specs give each request and response field.

## Where the reference lives

The specs are generated from swag annotations on the handlers. [`api/README.md`](../api/README.md) describes the
build.

| Spec | Scope |
|---|---|
| [`api/core.openapi.yaml`](../api/core.openapi.yaml) | Core endpoints |
| [`connectors/ACCESS/AMIE-Processor/api/amie.openapi.yaml`](../connectors/ACCESS/AMIE-Processor/api/amie.openapi.yaml) | `/connectors/amie/*` |
| [`connectors/Analytics/api/analytics.openapi.yaml`](../connectors/Analytics/api/analytics.openapi.yaml) | `/connectors/analytics/*` |

The TempAccount and VAST-Provisioner routes have no spec. Their handlers are the reference.

## Conventions

| Topic | Rule |
|---|---|
| Base URL | `http://<host>:<port>`. The port is `core.api.port` in `config/custos.yaml`, `8080` by default. |
| Authentication | `Authorization: Bearer <OIDC JWT>` on all routes except `GET /healthz`. |
| Authorization | Most routes require a privilege key. [`privileges.md`](privileges.md) maps each route to its key. |
| Encoding | Requests and responses are `application/json` in UTF-8. |
| Unknown fields | A request body with an unknown field gets `400`. |
| Identifiers | `id` is a server-generated UUID when a create request omits it. `originated_id` is an optional ID from the source system, such as an ACCESS record ID, and must be unique per entity type. |
| Timestamps | RFC 3339 with a time zone, such as `2026-05-16T12:34:56.789Z`. The server emits UTC. |
| Paging | Paged list routes take `limit` (default 50, at most 200) and `offset`, and return `{"items": [...], "total": n}`. Other list routes return a JSON array. The spec gives each route's shape. |

## Errors

Handler errors return a status code and a body with one field:

```json
{ "error": "human-readable message" }
```

| Status | Meaning | Raised by |
|---|---|---|
| `400 Bad Request` | Malformed JSON, an unknown or missing field, or a reference to a missing record | Body decoding, `service.ErrInvalidInput`, a foreign-key or length violation |
| `404 Not Found` | The record does not exist, or the caller may not see it | `service.ErrNotFound` |
| `409 Conflict` | A duplicate email or `originated_id`, a change to a project's PI, or a record still in use | `service.ErrAlreadyExists`, `service.ErrPIChange`, `service.ErrInUse`, a unique violation |
| `500 Internal Server Error` | An unexpected failure. The cause is logged; the body says only `internal server error` | Any other error |

Authentication and authorization failures come from the middleware, before any handler, with a different body:

```json
{ "code": "insufficient_privilege", "message": "Caller lacks required privilege" }
```

The [architecture doc](architecture.md#identity-and-authentication) lists each `code`.

## Endpoint groups

| Group | Paths | Notes |
|---|---|---|
| Health | `GET /healthz` | Public, returns `{"status": "ok"}` |
| Caller | `/me`, `/me/agreements/*`, `/me/onboarding*`, `/user/privileges` | The caller's own profile, privileges, consent and onboarding |
| Organizations | `/organizations*` | |
| Users and identities | `/users*`, `/user-identities*` | `POST /users` onboards a researcher or an admin |
| Projects | `/projects*` | Includes project members and status |
| Clusters and cluster accounts | `/compute-clusters*`, `/compute-cluster-users*` | Cluster accounts are approved or denied with `POST /compute-cluster-users/{id}/approve` and `/deny` |
| Allocations | `/compute-allocations*` | Includes the allocation's resources, memberships, diffs, change requests and usage |
| Resources and rates | `/compute-allocation-resources*`, `/compute-allocation-resource-rates*` | |
| Memberships and overrides | `/compute-allocation-memberships*`, `/compute-allocation-membership-resource-overrides*` | |
| Diffs, change requests, usage | `/compute-allocation-diffs*`, `/compute-allocation-change-requests*`, `/compute-allocation-change-request-events*`, `/compute-allocation-usages*` | |
| Privileges and roles | `/privileges/*`, `/roles*`, `/users/{id}/privileges*`, `/users/{id}/roles*` | |
| User agreements | `/agreements*` | |
| Audit and events | `/audit/*`, `/events/*` | Traces, audit rows, event deliveries and subscriptions |
| Connectors | `GET /connectors`, `/connectors/<name>/*` | Load status, and each connector's own routes |

The data behind the allocation groups is described in [`Allocation-Data-Models.md`](Allocation-Data-Models.md).

## Example session

The session creates a project with one allocation on one cluster, attaches a resource and prices it. The caller needs
the matching `write` privileges, which the bootstrap `super_admin` holds. `TOKEN` is the caller's OIDC access token.

```bash
BASE=http://localhost:8080
AUTH="Authorization: Bearer $TOKEN"
JSON='Content-Type: application/json'

# The caller becomes the project's PI.
PI_ID=$(curl -s -H "$AUTH" $BASE/me | jq -r .user.id)

ORG_ID=$(curl -s -X POST $BASE/organizations -H "$AUTH" -H "$JSON" \
  -d '{"name":"Example University","originated_id":"ACCESS-ORG-001"}' | jq -r .id)

CLUSTER_ID=$(curl -s -X POST $BASE/compute-clusters -H "$AUTH" -H "$JSON" \
  -d '{"name":"example-cluster"}' | jq -r .id)

PROJ_ID=$(curl -s -X POST $BASE/projects -H "$AUTH" -H "$JSON" \
  -d "{\"title\":\"Climate Simulation 2026\",\"origination\":\"ACCESS\",\"originated_id\":\"ACCESS-PRJ-9000\",\"project_pi_id\":\"$PI_ID\"}" \
  | jq -r .id)

ALLOC_ID=$(curl -s -X POST $BASE/compute-allocations -H "$AUTH" -H "$JSON" \
  -d "{
        \"project_id\":\"$PROJ_ID\",
        \"name\":\"Q2 2026 Climate Run\",
        \"compute_cluster_id\":\"$CLUSTER_ID\",
        \"initial_su_amount\":100000,
        \"start_time\":\"2026-04-01T00:00:00Z\",
        \"end_time\":\"2026-06-30T23:59:59Z\"
      }" | jq -r .id)

# A resource belongs to a cluster, here a partition with 24 units.
RES_ID=$(curl -s -X POST $BASE/compute-allocation-resources -H "$AUTH" -H "$JSON" \
  -d "{\"name\":\"debug\",\"resource_type\":\"cpu\",\"resource_amount\":24,\"compute_cluster_id\":\"$CLUSTER_ID\"}" \
  | jq -r .id)

# Grant the allocation 24 units for 1440 minutes.
curl -s -X POST $BASE/compute-allocations/$ALLOC_ID/resources -H "$AUTH" -H "$JSON" \
  -d "{\"compute_allocation_resource_id\":\"$RES_ID\",\"resource_amount\":24,\"resource_time\":1440}" | jq

# Charge 2 SU per unit-hour during 2026.
curl -s -X POST $BASE/compute-allocation-resource-rates -H "$AUTH" -H "$JSON" \
  -d "{
        \"compute_allocation_resource_id\":\"$RES_ID\",
        \"rate\":2.0,
        \"start_time\":\"2026-01-01T00:00:00Z\",
        \"end_time\":\"2026-12-31T23:59:59Z\"
      }" | jq

# Read it back.
curl -s -H "$AUTH" $BASE/compute-allocation-resources/$RES_ID/rates/effective | jq
curl -s -H "$AUTH" $BASE/compute-allocations/$ALLOC_ID/resources | jq
curl -s -H "$AUTH" $BASE/projects/$PROJ_ID | jq
```

## Running the server

[INSTALL.md](../INSTALL.md) brings up the server. [CONFIG.md](../CONFIG.md) lists every setting or links to it.
