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
