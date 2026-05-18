# Local Installation

This guide covers running Apache Airavata Custos locally for development and testing. For coding conventions, build commands, and contribution workflow, see [CONTRIBUTING.md](CONTRIBUTING.md).

## Prerequisites

- Go **1.24+**
- Docker and Docker Compose (for the local MariaDB and supporting services)
- `git`

## Running locally with a database

The server requires a MariaDB/MySQL database. The easiest way to get one is via the bundled Docker Compose stack.

### 1. Start MariaDB

```bash
cd dev-ops/compose
docker compose up -d db
```

This starts MariaDB on `localhost:3306` and runs `dbinit/init-db.sh`, which creates the `custos` database and the `admin` user (password `admin`).

To stop it later:

```bash
docker compose down
```

Use `docker compose down -v` if you also want to wipe the database volume.

### 2. Run the API server

From the repository root:

```bash
export DATABASE_DSN='admin:admin@tcp(localhost:3306)/custos?parseTime=true&charset=utf8mb4&multiStatements=true'
export HTTP_ADDR=':8080'
go run ./cmd/server
```

On startup, the server:

1. Opens the database connection (`internal/db`).
2. Runs the embedded migrations (`internal/db/migrations/`).
3. Wires the event bus, service layer, and HTTP router.
4. Listens on `HTTP_ADDR`.

### Environment variables

| Variable             | Default                                                            | Description                                  |
|----------------------|--------------------------------------------------------------------|----------------------------------------------|
| `DATABASE_DSN`       | _(required)_                                                       | Go MySQL DSN; must allow multi-statements    |
| `HTTP_ADDR`          | `:8080`                                                            | Listen address for the HTTP API              |
| `DB_MAX_OPEN_CONNS`  | `25`                                                               | Max open DB connections                      |
| `DB_MAX_IDLE_CONNS`  | `5`                                                                | Max idle DB connections                      |

### Optional services

The Compose file also defines Keycloak, Vault, Adminer, Prometheus, and Grafana. Start them as needed, for example:

```bash
docker compose up -d adminer       # DB UI at http://localhost:18080
```
