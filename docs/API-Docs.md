# Custos API Documentation

HTTP/JSON API exposed by `cmd/server`. All endpoints accept and return
`application/json` and use UTF-8.

- **Base URL:** `http://<host>:<port>` (default port `8080`, configurable via `HTTP_ADDR`)
- **Auth:** none currently enforced (deploy behind a trusted ingress / auth proxy)
- **Content-Type:** `application/json` is required on every request that has a body
- **Unknown fields:** request bodies with unknown JSON fields are rejected with `400`

---

## Conventions

### Identifiers

- `id` fields are server-generated UUIDs when omitted from a create request.
- `originated_id` is an optional external identifier (e.g. ACCESS Record ID) — when supplied, it must be unique within its entity type.

### Timestamps

All timestamps are RFC 3339 / ISO 8601 with timezone, e.g. `2026-05-16T12:34:56.789Z`. The server emits UTC.

### Error format

Errors are returned with an appropriate HTTP status code and a JSON body:

```json
{ "error": "human-readable message" }
```

| Status | Meaning | Triggered by |
|--------|---------|--------------|
| `400 Bad Request` | Malformed JSON, unknown field, missing required field, or unknown foreign-key reference | request body validation, `service.ErrInvalidInput` |
| `404 Not Found` | Requested record does not exist | `service.ErrNotFound` |
| `409 Conflict` | Duplicate `email` or duplicate `originated_id` | `service.ErrAlreadyExists` |
| `500 Internal Server Error` | Unexpected server / database failure (driver message is logged, never returned) | any other error |

---

## Health

### `GET /healthz`

Liveness probe. Always returns `200` when the process is accepting connections.

**Response 200**

```json
{ "status": "ok" }
```

---

## Organizations

### `POST /organizations`

Create a new organization.

**Required fields:** `name`
**Optional fields:** `id` (auto-generated if omitted), `originated_id`

**Request**

```json
{
  "name": "University of Example",
  "originated_id": "ACCESS-ORG-001"
}
```

**Response 201**

```json
{
  "id": "8c4a1b2e-7d4f-4b6a-9a0c-2f3b9d1c8e21",
  "originated_id": "ACCESS-ORG-001",
  "name": "University of Example"
}
```

**Errors**

- `400` — `name` is required.
- `409` — an organization with the supplied `originated_id` already exists.

#### Example

```bash
curl -s -X POST http://localhost:8080/organizations \
  -H 'Content-Type: application/json' \
  -d '{"name":"University of Example","originated_id":"ACCESS-ORG-001"}'
```

---

### `GET /organizations/{id}`

Retrieve an organization by its ID.

**Response 200**

```json
{
  "id": "8c4a1b2e-7d4f-4b6a-9a0c-2f3b9d1c8e21",
  "originated_id": "ACCESS-ORG-001",
  "name": "University of Example"
}
```

**Errors**

- `404` — no organization matches the supplied ID.

---

## Users

### `POST /users`

Create a new user.

**Required fields:** `organization_id`, `email`
**Optional fields:** `id`, `first_name`, `last_name`, `middle_name`

The referenced `organization_id` must already exist; emails must be unique.

**Request**

```json
{
  "organization_id": "8c4a1b2e-7d4f-4b6a-9a0c-2f3b9d1c8e21",
  "first_name": "Ada",
  "last_name": "Lovelace",
  "email": "ada@example.edu"
}
```

**Response 201**

```json
{
  "id": "f0c5a4d1-2b9e-4a7c-8d31-1c5b6e3d9f02",
  "organization_id": "8c4a1b2e-7d4f-4b6a-9a0c-2f3b9d1c8e21",
  "first_name": "Ada",
  "last_name": "Lovelace",
  "email": "ada@example.edu"
}
```

**Errors**

- `400` — `email`, `organization_id` missing, or `organization_id` does not exist.
- `409` — a user with this `email` already exists.

#### Example

```bash
curl -s -X POST http://localhost:8080/users \
  -H 'Content-Type: application/json' \
  -d '{
        "organization_id":"8c4a1b2e-7d4f-4b6a-9a0c-2f3b9d1c8e21",
        "first_name":"Ada",
        "last_name":"Lovelace",
        "email":"ada@example.edu"
      }'
```

---

### `GET /users/{id}`

Retrieve a user by its ID.

**Response 200**

```json
{
  "id": "f0c5a4d1-2b9e-4a7c-8d31-1c5b6e3d9f02",
  "organization_id": "8c4a1b2e-7d4f-4b6a-9a0c-2f3b9d1c8e21",
  "first_name": "Ada",
  "last_name": "Lovelace",
  "email": "ada@example.edu"
}
```

**Errors**

- `404` — no user matches the supplied ID.

---

## Projects

### `POST /projects`

Create a new project.

**Required fields:** `title`, `project_pi_id`
**Optional fields:** `id`, `origination`, `originated_id`, `created_time` (defaults to current UTC time)

The referenced `project_pi_id` must be an existing user. `originated_id`, when supplied, must be unique across projects.

**Request**

```json
{
  "title": "Climate Simulation 2026",
  "origination": "ACCESS",
  "originated_id": "ACCESS-PRJ-9000",
  "project_pi_id": "f0c5a4d1-2b9e-4a7c-8d31-1c5b6e3d9f02"
}
```

**Response 201**

```json
{
  "id": "3a8c2e7b-9d1f-4f5a-bc02-7a4d9e6c1bb1",
  "originated_id": "ACCESS-PRJ-9000",
  "title": "Climate Simulation 2026",
  "origination": "ACCESS",
  "project_pi_id": "f0c5a4d1-2b9e-4a7c-8d31-1c5b6e3d9f02",
  "created_time": "2026-05-16T17:21:04.512Z"
}
```

**Errors**

- `400` — `title`, `project_pi_id` missing, or the PI user does not exist.
- `409` — a project with this `originated_id` already exists.

#### Example

```bash
curl -s -X POST http://localhost:8080/projects \
  -H 'Content-Type: application/json' \
  -d '{
        "title":"Climate Simulation 2026",
        "origination":"ACCESS",
        "originated_id":"ACCESS-PRJ-9000",
        "project_pi_id":"f0c5a4d1-2b9e-4a7c-8d31-1c5b6e3d9f02"
      }'
```

---

### `GET /projects/{id}`

Retrieve a project by its ID.

**Response 200**

```json
{
  "id": "3a8c2e7b-9d1f-4f5a-bc02-7a4d9e6c1bb1",
  "originated_id": "ACCESS-PRJ-9000",
  "title": "Climate Simulation 2026",
  "origination": "ACCESS",
  "project_pi_id": "f0c5a4d1-2b9e-4a7c-8d31-1c5b6e3d9f02",
  "created_time": "2026-05-16T17:21:04.512Z"
}
```

**Errors**

- `404` — no project matches the supplied ID.

---

## Compute Clusters

A compute cluster represents a physical or logical HPC resource (e.g. a
Slurm cluster) where allocations can be provisioned.

### `POST /compute-clusters`

Create a new compute cluster.

**Required fields:** `name`
**Optional fields:** `id` (auto-generated if omitted)

`name` must be unique across compute clusters.

**Request**

```json
{ "name": "Delta" }
```

**Response 201**

```json
{
  "id": "9b0a7f1c-2c5d-4e1b-9a0f-22e8a5c2dcb1",
  "name": "Delta"
}
```

**Errors**

- `400` — `name` is required.
- `409` — a compute cluster with this `name` already exists.

---

### `GET /compute-clusters`

List all compute clusters.

**Response 200**

```json
[
  { "id": "9b0a7f1c-2c5d-4e1b-9a0f-22e8a5c2dcb1", "name": "Delta" },
  { "id": "1d4e6a3b-7c8f-49b2-bd34-7c1f9a4e5d10", "name": "Phoenix" }
]
```

---

### `GET /compute-clusters/{id}`

Retrieve a single compute cluster by its ID.

**Errors**

- `404` — no compute cluster matches the supplied ID.

---

## Compute Allocations

A compute allocation grants a project a budget of Service Units (SUs) on a
specific compute cluster for a bounded time window.

### `POST /compute-allocations`

Create a new compute allocation.

**Required fields:** `project_id`, `name`, `compute_cluster_id`
**Optional fields:** `id`, `status` (defaults to `ACTIVE`), `initial_su_amount`, `start_time`, `end_time`

Both `project_id` and `compute_cluster_id` must reference existing records.
`status` is one of `ACTIVE`, `INACTIVE`, `DELETED`.

**Request**

```json
{
  "project_id": "3a8c2e7b-9d1f-4f5a-bc02-7a4d9e6c1bb1",
  "name": "Q2 2026 Climate Run",
  "compute_cluster_id": "9b0a7f1c-2c5d-4e1b-9a0f-22e8a5c2dcb1",
  "initial_su_amount": 100000,
  "start_time": "2026-04-01T00:00:00Z",
  "end_time":   "2026-06-30T23:59:59Z"
}
```

**Response 201**

```json
{
  "id": "2f6a8c1d-3e4b-4a7d-8c91-aa12bb34cc56",
  "project_id": "3a8c2e7b-9d1f-4f5a-bc02-7a4d9e6c1bb1",
  "name": "Q2 2026 Climate Run",
  "status": "ACTIVE",
  "compute_cluster_id": "9b0a7f1c-2c5d-4e1b-9a0f-22e8a5c2dcb1",
  "initial_su_amount": 100000,
  "start_time": "2026-04-01T00:00:00Z",
  "end_time":   "2026-06-30T23:59:59Z"
}
```

**Errors**

- `400` — required field missing, or `project_id` / `compute_cluster_id` does not exist.

---

### `GET /compute-allocations/{id}`

Retrieve a compute allocation by its ID.

**Errors**

- `404` — no compute allocation matches the supplied ID.

---

## Compute Allocation Resources

A compute allocation resource describes a hardware capability (e.g.
`GPU B200`, `CPU`) that can be attached to one or more allocations.

### `POST /compute-allocation-resources`

Create a new compute allocation resource.

**Required fields:** `name`, `resource_type`
**Optional fields:** `id`, `resource_amount`

**Request**

```json
{
  "name": "GPU B200",
  "resource_type": "GPU",
  "resource_amount": 8
}
```

**Response 201**

```json
{
  "id": "c0a1b2c3-d4e5-46f7-8899-aabbccddeeff",
  "name": "GPU B200",
  "resource_type": "GPU",
  "resource_amount": 8
}
```

**Errors**

- `400` — `name` or `resource_type` is missing.

---

### `GET /compute-allocation-resources`

List all compute allocation resources.

**Response 200**

```json
[
  {
    "id": "c0a1b2c3-d4e5-46f7-8899-aabbccddeeff",
    "name": "GPU B200",
    "resource_type": "GPU",
    "resource_amount": 8
  }
]
```

---

### `GET /compute-allocation-resources/{id}`

Retrieve a compute allocation resource by its ID.

**Errors**

- `404` — no resource matches the supplied ID.

---

## Compute Allocation ↔ Resource Mappings

A many-to-many join: an allocation can have many resources attached, and a
resource can be attached to many allocations. Mappings are unique per
(allocation, resource) pair, and are cascade-deleted when either parent is
removed.

### `POST /compute-allocations/{id}/resources`

Attach an existing resource to a compute allocation.

**Path parameters:** `{id}` — the compute allocation ID.
**Required body fields:** `compute_allocation_resource_id`

**Request**

```json
{ "compute_allocation_resource_id": "c0a1b2c3-d4e5-46f7-8899-aabbccddeeff" }
```

**Response 201**

```json
{
  "id": "7e1d2c3b-4a5f-4b6c-9d8e-0011223344ff",
  "compute_allocation_id": "2f6a8c1d-3e4b-4a7d-8c91-aa12bb34cc56",
  "compute_allocation_resource_id": "c0a1b2c3-d4e5-46f7-8899-aabbccddeeff"
}
```

**Errors**

- `400` — `compute_allocation_resource_id` missing, or either the allocation or the resource does not exist.
- `409` — this resource is already attached to the allocation.

---

### `DELETE /compute-allocations/{id}/resources/{resourceId}`

Detach a resource from a compute allocation.

**Response 204** — empty body on success.

**Errors**

- `404` — no such mapping exists.

---

### `GET /compute-allocations/{id}/resources`

List every compute allocation resource currently attached to the given
compute allocation.

**Response 200**

```json
[
  {
    "id": "c0a1b2c3-d4e5-46f7-8899-aabbccddeeff",
    "name": "GPU B200",
    "resource_type": "GPU",
    "resource_amount": 8
  }
]
```

---

### `GET /compute-allocation-resources/{id}/allocations`

List every compute allocation that has the given resource attached.

**Response 200**

```json
[
  {
    "id": "2f6a8c1d-3e4b-4a7d-8c91-aa12bb34cc56",
    "project_id": "3a8c2e7b-9d1f-4f5a-bc02-7a4d9e6c1bb1",
    "name": "Q2 2026 Climate Run",
    "status": "ACTIVE",
    "compute_cluster_id": "9b0a7f1c-2c5d-4e1b-9a0f-22e8a5c2dcb1",
    "initial_su_amount": 100000,
    "start_time": "2026-04-01T00:00:00Z",
    "end_time":   "2026-06-30T23:59:59Z"
  }
]
```

---

## Compute Allocation Resource Rates

A rate captures how many Service Units (SUs) are charged per unit of a
compute allocation resource over a bounded time window. Multiple rates can
exist for the same resource; usage at any instant is charged using the rate
whose `[start_time, end_time)` window contains that instant.

Rates are cascade-deleted when their parent resource is deleted.

### `POST /compute-allocation-resource-rates`

Create a new rate for a compute allocation resource.

**Required fields:** `compute_allocation_resource_id`, `rate`, `start_time`, `end_time`
**Optional fields:** `id`

Validation:

- `compute_allocation_resource_id` must reference an existing resource.
- `rate` must be ≥ 0.
- `start_time` must be strictly before `end_time`.

**Request**

```json
{
  "compute_allocation_resource_id": "c0a1b2c3-d4e5-46f7-8899-aabbccddeeff",
  "rate": 2.0,
  "start_time": "2026-01-01T00:00:00Z",
  "end_time":   "2026-12-31T23:59:59Z"
}
```

**Response 201**

```json
{
  "id": "55aa66bb-77cc-88dd-99ee-001122334455",
  "compute_allocation_resource_id": "c0a1b2c3-d4e5-46f7-8899-aabbccddeeff",
  "rate": 2.0,
  "start_time": "2026-01-01T00:00:00Z",
  "end_time":   "2026-12-31T23:59:59Z"
}
```

**Errors**

- `400` — required field missing, invalid time window, negative `rate`, or unknown `compute_allocation_resource_id`.

---

### `GET /compute-allocation-resource-rates/{id}`

Retrieve a rate by its ID.

**Errors**

- `404` — no rate matches the supplied ID.

---

### `GET /compute-allocation-resources/{id}/rates`

List every rate ever defined for the given compute allocation resource,
ordered by `start_time` ascending.

**Response 200**

```json
[
  {
    "id": "55aa66bb-77cc-88dd-99ee-001122334455",
    "compute_allocation_resource_id": "c0a1b2c3-d4e5-46f7-8899-aabbccddeeff",
    "rate": 2.0,
    "start_time": "2026-01-01T00:00:00Z",
    "end_time":   "2026-12-31T23:59:59Z"
  }
]
```

---

### `GET /compute-allocation-resources/{id}/rates/effective`

Return the rate currently in effect for the given resource. By default the
server uses the current time; supply `?at=<RFC 3339 timestamp>` to query an
arbitrary instant.

A rate is "effective" at instant *t* when `start_time <= t < end_time`. If
multiple rates overlap *t*, the one with the most recent `start_time` wins.

**Examples**

```http
GET /compute-allocation-resources/c0a1.../rates/effective
GET /compute-allocation-resources/c0a1.../rates/effective?at=2026-05-16T12:00:00Z
```

**Response 200**

```json
{
  "id": "55aa66bb-77cc-88dd-99ee-001122334455",
  "compute_allocation_resource_id": "c0a1b2c3-d4e5-46f7-8899-aabbccddeeff",
  "rate": 2.0,
  "start_time": "2026-01-01T00:00:00Z",
  "end_time":   "2026-12-31T23:59:59Z"
}
```

**Errors**

- `400` — `at` query parameter is not a valid RFC 3339 timestamp.
- `404` — no rate is effective for the resource at the supplied instant.

---

## Compute Allocation Diffs

A diff is an append-only audit record of a change applied to a compute
allocation — for example a usage update or a status transition. Diffs are
cascade-deleted when their parent allocation is deleted.

### `POST /compute-allocation-diffs`

Record a new diff against a compute allocation.

**Required fields:** `compute_allocation_id`, `diff_type`, `status`
**Optional fields:** `id`, `new_su_amount` (defaults to `0`), `timestamp` (defaults to the server's current UTC time), `description`

`diff_type` is a free-form short code such as `USAGE_UPDATE` or
`ALLOCATION_STATUS_CHANGE`. `status` must be one of `ACTIVE`, `INACTIVE`,
`DELETED`.

**Request**

```json
{
  "compute_allocation_id": "2f6a8c1d-3e4b-4a7d-8c91-aa12bb34cc56",
  "diff_type": "USAGE_UPDATE",
  "new_su_amount": 90000,
  "status": "ACTIVE",
  "description": "Charged 10000 SUs for completed jobs"
}
```

**Response 201**

```json
{
  "id": "44bb55cc-66dd-77ee-88ff-aabbccddeeff",
  "compute_allocation_id": "2f6a8c1d-3e4b-4a7d-8c91-aa12bb34cc56",
  "diff_type": "USAGE_UPDATE",
  "new_su_amount": 90000,
  "status": "ACTIVE",
  "timestamp": "2026-05-16T17:42:11.918Z",
  "description": "Charged 10000 SUs for completed jobs"
}
```

**Errors**

- `400` — required field missing, or `compute_allocation_id` does not exist.

---

### `GET /compute-allocation-diffs/{id}`

Retrieve a single diff by its ID.

**Errors**

- `404` — no diff matches the supplied ID.

---

### `DELETE /compute-allocation-diffs/{id}`

Remove a diff record. Intended for administrative cleanup; diffs are
otherwise append-only.

**Response 204** — empty body on success.

---

### `GET /compute-allocations/{id}/diffs`

List every diff ever recorded against the given compute allocation, ordered
by `timestamp` ascending.

**Response 200**

```json
[
  {
    "id": "44bb55cc-66dd-77ee-88ff-aabbccddeeff",
    "compute_allocation_id": "2f6a8c1d-3e4b-4a7d-8c91-aa12bb34cc56",
    "diff_type": "USAGE_UPDATE",
    "new_su_amount": 90000,
    "status": "ACTIVE",
    "timestamp": "2026-05-16T17:42:11.918Z",
    "description": "Charged 10000 SUs for completed jobs"
  }
]
```

---

### `GET /compute-allocations/{id}/diffs/latest`

Return the most recent diff (highest `timestamp`) for the given allocation.

**Errors**

- `404` — the allocation has no diffs recorded.

---

## Compute Allocation Change Requests

A change request represents a user- or admin-initiated proposal to mutate a
compute allocation — e.g. asking for additional Service Units or to change
its status. Each request carries a lifecycle (`change_status`: `PENDING`,
`APPROVED`, `REJECTED`, etc.). Change requests are cascade-deleted when their
parent allocation is deleted. Every create, update, and delete of a change
request transactionally appends an entry to its event log (see below); the
event log is intentionally **not** cascade-deleted so the audit trail
survives the deletion of the parent change request.

### `POST /compute-allocation-change-requests`

Submit a new change request.

**Required fields:** `compute_allocation_id`, `requester_id`
**Optional fields:** `id`, `requested_su_amount`, `requested_status`, `reason`, `change_status` (defaults to `PENDING`), `approver_id`, `timestamp` (defaults to the server's current UTC time)

**Request**

```json
{
  "compute_allocation_id": "2f6a8c1d-3e4b-4a7d-8c91-aa12bb34cc56",
  "requested_su_amount": 120000,
  "requested_status": "ACTIVE",
  "reason": "Need more SUs for upcoming HPC runs",
  "requester_id": "11112222-3333-4444-5555-666677778888"
}
```

**Response 201**

```json
{
  "id": "9988aabb-ccdd-eeff-0011-223344556677",
  "compute_allocation_id": "2f6a8c1d-3e4b-4a7d-8c91-aa12bb34cc56",
  "requested_su_amount": 120000,
  "requested_status": "ACTIVE",
  "reason": "Need more SUs for upcoming HPC runs",
  "change_status": "PENDING",
  "requester_id": "11112222-3333-4444-5555-666677778888",
  "timestamp": "2026-05-16T17:42:11.918Z"
}
```

**Errors**

- `400` — required field missing, or `compute_allocation_id` does not exist.

---

### `GET /compute-allocation-change-requests/{id}`

Retrieve a single change request by its ID.

**Errors**

- `404` — no change request matches the supplied ID.

---

### `PUT /compute-allocation-change-requests/{id}`

Replace mutable fields of a change request. Typically used by an approver to
transition `change_status` (e.g. to `APPROVED` or `REJECTED`) and stamp
`approver_id`. Omitted fields are preserved from the existing record.

**Request**

```json
{
  "change_status": "APPROVED",
  "approver_id": "aaaa-bbbb-cccc-dddd-eeee"
}
```

**Errors**

- `400` — request id missing.
- `404` — no change request matches the supplied ID.

---

### `DELETE /compute-allocation-change-requests/{id}`

Remove a change request and (cascading) its event log.

**Response 204** — empty body on success.

---

### `GET /compute-allocations/{id}/change-requests`

List every change request ever submitted against the given allocation,
ordered by `timestamp` ascending.

---

### `GET /users/{id}/change-requests`

List every change request submitted by the given user, ordered by
`timestamp` ascending.

---

## Compute Allocation Change Request Events

Events are an append-only audit trail of state transitions applied to a
change request — typically `CREATED`, `APPROVED`, `REJECTED`, `UPDATED`,
`DELETED`, or arbitrary workflow markers. Create / update / delete of a
change request each emit an event automatically; clients may also append
custom events via the endpoint below. Events are **not** cascade-deleted
when their parent change request is removed, so the audit trail is
preserved indefinitely.

### `POST /compute-allocation-change-request-events`

Append a new event to a change request.

**Required fields:** `compute_allocation_change_request_id`, `event_type`
**Optional fields:** `id`, `description`, `timestamp` (defaults to the server's current UTC time)

**Request**

```json
{
  "compute_allocation_change_request_id": "9988aabb-ccdd-eeff-0011-223344556677",
  "event_type": "APPROVED",
  "description": "Change request approved by admin"
}
```

**Response 201**

```json
{
  "id": "ee11ff22-3344-5566-7788-99aabbccddee",
  "compute_allocation_change_request_id": "9988aabb-ccdd-eeff-0011-223344556677",
  "event_type": "APPROVED",
  "description": "Change request approved by admin",
  "timestamp": "2026-05-16T18:00:00.000Z"
}
```

**Errors**

- `400` — required field missing, or `compute_allocation_change_request_id` does not exist.

---

### `GET /compute-allocation-change-request-events/{id}`

Retrieve a single event by its ID.

**Errors**

- `404` — no event matches the supplied ID.

---

### `DELETE /compute-allocation-change-request-events/{id}`

Remove an event record. Intended for administrative cleanup; events are
otherwise append-only.

**Response 204** — empty body on success.

---

### `GET /compute-allocation-change-requests/{id}/events`

List every event recorded against the given change request, ordered by
`timestamp` ascending.

---

### `GET /compute-allocation-change-requests/{id}/events/latest`

Return the most recent event for the given change request.

**Errors**

- `404` — the change request has no events recorded.

---

## End-to-end example

```bash
BASE=http://localhost:8080

ORG_ID=$(curl -s -X POST $BASE/organizations \
  -H 'Content-Type: application/json' \
  -d '{"name":"University of Example","originated_id":"ACCESS-ORG-001"}' \
  | jq -r .id)

USER_ID=$(curl -s -X POST $BASE/users \
  -H 'Content-Type: application/json' \
  -d "{\"organization_id\":\"$ORG_ID\",\"first_name\":\"Ada\",\"last_name\":\"Lovelace\",\"email\":\"ada@example.edu\"}" \
  | jq -r .id)

PROJ_ID=$(curl -s -X POST $BASE/projects \
  -H 'Content-Type: application/json' \
  -d "{\"title\":\"Climate Simulation 2026\",\"origination\":\"ACCESS\",\"originated_id\":\"ACCESS-PRJ-9000\",\"project_pi_id\":\"$USER_ID\"}" \
  | jq -r .id)

CLUSTER_ID=$(curl -s -X POST $BASE/compute-clusters \
  -H 'Content-Type: application/json' \
  -d '{"name":"Delta"}' | jq -r .id)

ALLOC_ID=$(curl -s -X POST $BASE/compute-allocations \
  -H 'Content-Type: application/json' \
  -d "{\"project_id\":\"$PROJ_ID\",\"name\":\"Q2 2026 Climate Run\",\"compute_cluster_id\":\"$CLUSTER_ID\",\"initial_su_amount\":100000}" \
  | jq -r .id)

RES_ID=$(curl -s -X POST $BASE/compute-allocation-resources \
  -H 'Content-Type: application/json' \
  -d '{"name":"GPU B200","resource_type":"GPU","resource_amount":8}' | jq -r .id)

# Attach the resource to the allocation.
curl -s -X POST $BASE/compute-allocations/$ALLOC_ID/resources \
  -H 'Content-Type: application/json' \
  -d "{\"compute_allocation_resource_id\":\"$RES_ID\"}" | jq

# Define a rate for the resource.
curl -s -X POST $BASE/compute-allocation-resource-rates \
  -H 'Content-Type: application/json' \
  -d "{
        \"compute_allocation_resource_id\":\"$RES_ID\",
        \"rate\":2.0,
        \"start_time\":\"2026-01-01T00:00:00Z\",
        \"end_time\":\"2026-12-31T23:59:59Z\"
      }" | jq

# Look up the currently-effective rate.
curl -s $BASE/compute-allocation-resources/$RES_ID/rates/effective | jq

# Record a usage diff against the allocation.
curl -s -X POST $BASE/compute-allocation-diffs \
  -H 'Content-Type: application/json' \
  -d "{
        \"compute_allocation_id\":\"$ALLOC_ID\",
        \"diff_type\":\"USAGE_UPDATE\",
        \"new_su_amount\":90000,
        \"status\":\"ACTIVE\",
        \"description\":\"Charged 10000 SUs for completed jobs\"
      }" | jq

# Inspect the diff history.
curl -s $BASE/compute-allocations/$ALLOC_ID/diffs | jq
curl -s $BASE/compute-allocations/$ALLOC_ID/diffs/latest | jq

# Bidirectional lookups.
curl -s $BASE/compute-allocations/$ALLOC_ID/resources | jq
curl -s $BASE/compute-allocation-resources/$RES_ID/allocations | jq

curl -s $BASE/projects/$PROJ_ID | jq
```

---

## Running the server

```bash
export DATABASE_DSN='custos:secret@tcp(127.0.0.1:3306)/custos?parseTime=true&charset=utf8mb4'
# optional
export HTTP_ADDR=:8080
export DB_MAX_OPEN_CONNS=25
export DB_MAX_IDLE_CONNS=5

go run ./cmd/server
```

| Environment variable | Default | Purpose |
|----------------------|---------|---------|
| `DATABASE_DSN` | *(required)* | MySQL/MariaDB DSN. `parseTime=true` is mandatory. |
| `HTTP_ADDR` | `:8080` | Address the HTTP server binds to. |
| `DB_MAX_OPEN_CONNS` | `25` | Maximum open database connections. |
| `DB_MAX_IDLE_CONNS` | `5` | Maximum idle database connections. |

Migrations from `internal/db/migrations/` are applied automatically on startup.

The server handles `SIGINT` / `SIGTERM` gracefully, draining in-flight requests
for up to 15 seconds before exiting.
