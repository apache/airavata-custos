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

# Writing a Connector

A connector is an adapter that brings external state (e.g. an HPC allocation
authority, an identity registry, a SLURM cluster) into Custos. This guide walks
through what a connector looks like, how the runtime loads it, and what to
write end-to-end.

If you only need a glossary or the runtime model, jump to the
[domain glossary](#domain-glossary) or the
[architecture doc](../architecture.md).

---

## What you are building

A connector is a Go package. It is compiled into the main Custos server
(`cmd/server`) and loaded at startup. Connectors share the server's database
handle, event bus, HTTP router, and identity context. There is no separate
process and no remote call between the connector and core.

The common shapes are:

| Shape | Example | What it does |
|---|---|---|
| HTTP-only | `connectors/TempAccount` | Mounts a few `/connectors/<name>/*` endpoints that call into `pkg/service`. |
| Event-subscriber | `connectors/COmanage/Identity-Provisioner`, `connectors/SLURM/Association-Mapper` | Subscribes to core events on `pkg/events.Bus`. Reacts by talking to an external system. No HTTP routes, no goroutine of its own. |
| Background worker | `connectors/SLURM/Usage-Monitor` | Runs a long-lived goroutine that polls an external system on an interval and publishes domain events. No HTTP routes; does not subscribe. |
| Mixed (HTTP + workers + own schema) | `connectors/ACCESS/AMIE-Processor` | All of the above plus its own embedded migrations and an external API client. |

Pick the smallest shape that fits.

---

## Runtime model

Custos runs as a single binary. `cmd/server/main.go` opens the database, runs
migrations, builds the service layer, then calls
`connectors.LoadConnectorsFromConfig` to wire every enabled connector
into the same process.

Each connector exposes one entry point: a `LoadConnector` function with this
signature (see `internal/connectors/loader.go`).

```go
func(
    ctx context.Context,
    database *sqlx.DB,
    eventBus *events.Bus,
    coreService *service.Service,
    wg *sync.WaitGroup,
    router *identity.Router,
    connectorConfig *config.ConnectorConfig,
) error
```

What each argument is for:

- `ctx` — cancelled on shutdown. Pass it into background goroutines so they exit cleanly.
- `database` — the shared `*sqlx.DB`. Use it directly, or use `db.MigrateConnectorFS` to apply per-connector schema.
- `eventBus` — `pkg/events.Bus`. Subscribe here to react to core events; publish here when you want core or other connectors to react to yours.
- `coreService` — the central service layer. Call its methods (e.g. `coreService.CreateUser`, `coreService.CreateComputeAllocationMembership`) instead of going to stores yourself.
- `wg` — register every long-running goroutine with `wg.Add(1)` and `defer wg.Done()`. The server waits up to 30s for this group to drain on shutdown.
- `router` — `*identity.Router`. Register routes with `Public`, `RequireAuth`, or `RequirePrivilege`.
- `connectorConfig` — the YAML block keyed under your connector's name in `config/custos.yaml`. Use `GetStringField`, `GetNestedConfig`, etc. on `*config.ConnectorConfig` to read fields.

Return `nil` on success. Returning an error fails the whole server startup, so
prefer logging and returning `nil` for "config missing, skip me silently"
(the COmanage loader is the reference).

---

## Step-by-step: adding a new connector

For a connector named `lustre-sync`.

### 1. Create the package directory

```
connectors/Lustre/Sync/
├── pkg/lustresync/
│   └── loader.go        // LoadConnector entry point
├── internal/
│   └── handlers.go      // HTTP handlers (if any)
└── README.md
```

If you also need database state, `db/` with `//go:embed` migrations; if you
need an external HTTP client, `internal/client/`; if you need workers,
`internal/worker/`. Mirror the shape of the closest existing connector.

### 2. Write `LoadConnector`

A minimum HTTP-only loader (`connectors/TempAccount/pkg/tempaccount/loader.go`
is the canonical small example):

```go
package lustresync

import (
    "context"
    "sync"

    "github.com/jmoiron/sqlx"

    "github.com/apache/airavata-custos/connectors/Lustre/Sync/internal"
    "github.com/apache/airavata-custos/internal/config"
    "github.com/apache/airavata-custos/pkg/events"
    "github.com/apache/airavata-custos/pkg/identity"
    "github.com/apache/airavata-custos/pkg/service"
)

func LoadConnector(
    ctx context.Context,
    _ *sqlx.DB,
    eventBus *events.Bus,
    coreService *service.Service,
    wg *sync.WaitGroup,
    router *identity.Router,
    connectorConfig *config.ConnectorConfig,
) error {
    handlers := internal.NewHandlers(coreService)
    handlers.RegisterRoutes(router)
    return nil
}
```

### 3. Register in the central loader

Open `internal/connectors/loader.go` and add two lines.

Import:

```go
"github.com/apache/airavata-custos/connectors/Lustre/Sync/pkg/lustresync"
```

Add an entry to the `connectorLoaders` map:

```go
connectorLoaders := map[string]func(...){
    // ...existing entries...
    "lustre-sync": lustresync.LoadConnector,
}
```

The string `"lustre-sync"` is what your `config/custos.yaml` block will
reference as its `type:` field. Pick a kebab-case identifier; the convention is
`<connector-name>` or `<system>-<role>`.

### 4. Add a YAML config block

In `config/custos.yaml`:

```yaml
connectors:
  # ...existing blocks...
  lustre-sync:
    type: "lustre-sync"       # must match the string in connectorLoaders
    enabled: true
    api:
      url: "${LUSTRE_API_URL}"
      token: "${LUSTRE_API_TOKEN}"
```

Top-level fields (`type`, `enabled`) are parsed by
`internal/config/config.go`. Everything else (e.g. `api.url`, `api.token`)
lands inside `ConnectorConfig.Config` and you read it in your loader. Use
`${VAR}` interpolation for secrets so they come from `.env`.

A disabled connector is skipped; an unknown `type:` is logged and skipped.

### 5. Read the config in your loader

Existing connectors all read YAML the same way: call
`connectorConfig.GetNestedConfig("group")`, check the error, then type-assert
each leaf field. Example:

```go
if api, err := connectorConfig.GetNestedConfig("api"); err == nil {
    if url, ok := api["url"].(string); ok {
        cfg.URL = url
    }
    if token, ok := api["token"].(string); ok {
        cfg.Token = token
    }
}
```

There are also flat `GetStringField`, `GetIntField`, and `GetDurationField`
helpers on `*config.ConnectorConfig`, but they only see top-level fields and
no shipping connector currently uses them. Stick with `GetNestedConfig` plus
type assertions.

Two patterns, pick one:

- **YAML-only.** Simplest. Recommended for new connectors. TempAccount is the example.
- **YAML with env fallback.** Read YAML, fall back to env vars if any required field is missing. The COmanage loader is the canonical reference: see `LoadConnector` for the orchestration, `loadConfigFromConnectorConfig` for the YAML reader, and `loadConfigFromEnv` for the env fallback in `connectors/COmanage/Identity-Provisioner/pkg/comanage/loader.go`.

If required config is missing, log and `return nil` so the connector skips
without failing the whole server. Do not panic. Do not return an error unless
the misconfiguration is genuinely fatal (e.g., embedded migrations failed to
apply).

### 6. Register routes (if you have any)

The router has three modes. Pick the right one per route.

```go
// Public: no JWT required.
router.Public("GET /healthz", h.health)

// RequireAuth: a verified caller must be on the context. No specific privilege.
router.RequireAuth("GET /connectors/lustre-sync/me", h.getSelf)

// RequirePrivilege: verified caller AND the named privilege key. 403 if absent.
router.RequirePrivilege(
    "POST /connectors/lustre-sync/sync",
    models.PrivilegeLustreWrite,
    h.runSync,
)
```

API surface in `pkg/identity/middleware.go`. Mount everything under
`/connectors/<your-name>/...` to keep namespaces clean.

Privilege keys live in `pkg/models/privilege.go`. Today the catalog is
`amie:read`, `amie:write`, `hpc:read`, `hpc:write`, `signer:read`,
`signer:write`, `privileges:grant`, `roles:manage`. Reuse an existing key if
it fits. If you genuinely need a new connector-specific key, add it as a
constant in `pkg/models/privilege.go` and append it to `KnownPrivileges()`.

### 7. Database state (if you need it)

If your connector owns its own tables:

**a. Put numbered migration pairs in `connectors/Lustre/Sync/db/migrations/`:**

```
db/migrations/000001_initial_schema.up.sql
db/migrations/000001_initial_schema.down.sql
```

Every `.sql` file must carry the Apache license header.

**b. Embed them with `embed.FS`** — `connectors/Lustre/Sync/db/embed.go`:

```go
package db

import "embed"

//go:embed migrations/*.sql
var migrationFS embed.FS

// MigrationFS exposes the embedded migrations to the host's migration runner.
func MigrationFS() embed.FS { return migrationFS }
```

**c. Apply them at the top of `LoadConnector`,** before you build stores or
spawn workers:

```go
import (
    lustredb "github.com/apache/airavata-custos/connectors/Lustre/Sync/db"
    "github.com/apache/airavata-custos/internal/db"
)

func LoadConnector(ctx context.Context, database *sqlx.DB, ...) error {
    if err := db.MigrateConnectorFS(database, lustredb.MigrationFS(), "migrations", "lustre"); err != nil {
        return err
    }
    // ...rest of loader...
}
```

The fourth argument (`"lustre"`) is the connector slug. It becomes the suffix
of the per-connector version-tracking table: `schema_migrations_lustre`. Pick
a short, stable identifier; you cannot change it later without orphaning
your migration history.

Per-connector tracking means connectors version their schema independently of
core and of one another. AMIE is the existing example: see
`connectors/ACCESS/AMIE-Processor/db/embed.go` for the embed and the top of
`connectors/ACCESS/AMIE-Processor/pkg/amie/loader.go` for the apply call.

### 8. Background goroutines (if you need them)

If you have a poller, processor, or any other long-running goroutine:

```go
wg.Add(1)
go func() {
    defer wg.Done()
    poller.Run(ctx) // honour ctx.Done() inside Run
}()
```

Shutdown is a two-phase drain (see `cmd/server/main.go`):

1. The HTTP server stops accepting new connections and has 15 seconds to
   finish in-flight requests. Your handlers should be done by then.
2. Then the server cancels `ctx` and waits up to 30 seconds for the
   WaitGroup (i.e. your goroutines) to return.

So a long-running worker has up to 30 seconds after `ctx` cancels to exit.
If your worker can take longer to wind down safely, restructure it so the
critical section completes inside that window, or persist enough state that
the next boot can resume.

### 9. Event bus (if you publish or subscribe)

`pkg/events.Bus` is an in-process publish/subscribe channel. The core service
layer publishes domain events whenever state changes (a user is created, an
allocation is updated, etc.). Connectors subscribe to react.

Event types are typed constants, one file per entity, in `pkg/events/`:
`user_subscribe.go` (`UserCreateEvent`, `UserUpdateEvent`, `UserDeleteEvent`),
`compute_allocation_subscribe.go`, `project_subscribe.go`, and so on. Each
file documents the payload type the dispatcher sends with the event.

**Subscribing.** Register inside `LoadConnector` so the subscription is in
place before traffic arrives:

```go
type ClusterUserSubscriber struct {
    bus  *events.Bus
    core *service.Service
}

func (s *ClusterUserSubscriber) RegisterSubscribers() {
    s.bus.Subscribe(events.ComputeClusterUserCreateEvent, s.handleCreate)
}

func (s *ClusterUserSubscriber) handleCreate(ctx context.Context, _ events.Event, payload interface{}) {
    cu, ok := payload.(*models.ComputeClusterUser)
    if !ok {
        slog.Error("payload type mismatch", "type", payload)
        return
    }
    // ...react...
}
```

The COmanage `ClusterUserSubscriber` is the smallest working example.
Subscriber handlers run asynchronously, one goroutine per dispatch; a panic
in one handler does not crash the bus or other subscribers.

**Publishing.** All current publishes come from the core service layer
(`pkg/service/*`). A connector can publish too if it needs to broadcast
something to other connectors or to core; today none do, but the API
supports it. The convention is to publish from your own service layer (not
from handlers or stores):

```go
eventBus.Publish(ctx, events.SomeEvent, payload)
```

`Publish` is fire-and-forget. Use `PublishSync` only when a downstream
subscriber must finish before your function returns (rare).

### 10. Audit rows (and `audit_extras` if you have your own references)

Every meaningful action should leave a row in the core `audit_events` table.
Two paths to write one.

**You called a `coreService` method that audits itself.** Nothing more to do.
`coreService.CreateUser`, `CreateComputeAllocationMembership`, and similar
service methods write an audit row inside their transaction and stamp it
with `trace_id` / `span_id` from your request context. Just pass `r.Context()`
through unmodified.

**You did something audit-worthy that wasn't a core service mutation.** Write
the row yourself with `coreService.CreateAuditEvent`:

```go
_, _ = svc.CreateAuditEvent(ctx, &models.AuditEvent{
    EventType:  "LustreQuotaSyncStarted",
    EntityType: "ComputeAllocation",
    EntityID:   alloc.ID,
    Source:     "lustre",
})
```

`trace_id`, `span_id`, and `parent_span_id` are read from `ctx` and stamped
automatically.

**If you need connector-specific references on the audit row** (the way AMIE
keeps `packet_id` and `event_id`), create a `<connector>_audit_extras` table.
Same shape as AMIE's:

```sql
CREATE TABLE amie_audit_extras (
    audit_event_id VARCHAR(255) NOT NULL,
    packet_id      VARCHAR(255) NOT NULL,
    event_id       VARCHAR(255) NULL,
    PRIMARY KEY (audit_event_id),
    CONSTRAINT fk_amie_audit_extras_event
        FOREIGN KEY (audit_event_id) REFERENCES audit_events(id) ON DELETE CASCADE
    -- plus any FKs to your connector's own tables
);
```

Write the extras row in the **same transaction** as the audit row so the two
stay consistent. AMIE composes a connector-side audit service around the
core audit store plus its own extras store. See
`connectors/ACCESS/AMIE-Processor/store/audit_store.go` for the extras-store
shape and `connectors/ACCESS/AMIE-Processor/service/audit.go` for the
composition.

For the full convention (what columns the core `audit_events` table carries,
why extras tables exist, how the unified trace view joins them), see the
[architecture doc](../architecture.md#audit-conventions).

---

## OpenAPI / Swagger generation

**Required for any connector that registers HTTP routes.** CI runs
`make verify-no-drift`; an out-of-date spec blocks merges.

Custos generates per-connector OpenAPI specs from swag annotations in Go
source. You annotate your handlers, declare a `//go:generate` directive, and
`make gen-api` writes the spec.

### Annotate your loader and handlers

At the top of your loader package (above `LoadConnector`), set the API
metadata:

```go
// @title       Lustre Sync Connector API
// @version     0.1.0
// @description REST endpoints for the Lustre Sync connector, all under /connectors/lustre-sync/.
// @host        localhost:8080
// @BasePath    /
// @securityDefinitions.apikey BearerAuth
// @in          header
// @name        Authorization
func LoadConnector(...) error { ... }
```

Annotate each handler with `@Summary`, `@Tags`, `@Param`, `@Success`, `@Router`
etc. Copy the shape from `connectors/ACCESS/AMIE-Processor/server/handlers.go`
or any AMIE handler.

### Declare the generate directive

Two lines, in the same file as `LoadConnector` (see
`connectors/ACCESS/AMIE-Processor/pkg/amie/loader.go` for the pattern):

```go
//go:generate go run github.com/swaggo/swag/cmd/swag init -g loader.go -d .,../../server,../../store -o ../../api --outputTypes yaml --parseDependency --useStructName
//go:generate mv ../../api/swagger.yaml ../../api/lustre-sync.openapi.yaml
```

What each swag flag does:

- `-g loader.go` — the entry-point file with the `@title` / `@host` / etc. annotations
- `-d .,../../server,../../store` — additional dirs to scan for handler annotations. List every package whose handler or model types appear in the spec
- `-o ../../api` — output directory (writes `swagger.yaml`)
- `--outputTypes yaml` — yaml only, no json
- `--parseDependency` — follow imported types so external structs render properly
- `--useStructName` — use Go struct names in the spec rather than fully-qualified import paths

The `mv` step renames swag's default `swagger.yaml` to a connector-specific
filename so multiple specs can live side by side. The output goes to
`connectors/Lustre/Sync/api/lustre-sync.openapi.yaml`.

### Generate and verify

```bash
# from repo root
make gen-api          # runs `go generate ./...` over the whole repo
make verify-no-drift  # fails if the committed spec is stale
```

Commit the generated `.openapi.yaml` alongside your code change.

---

## Testing

Two tiers: unit tests (default) and integration tests (build tag).

### Unit tests

Use `httptest` and table-driven tests. No DB, no network. These run in the
default `go test ./...` from the repo root.

For routes that need a verified caller on context, build the context yourself
using `pkg/identity`. The AMIE handlers test
(`connectors/ACCESS/AMIE-Processor/server/handlers_test.go`) is the canonical
example:

```go
import (
    "github.com/apache/airavata-custos/pkg/identity"
    "github.com/apache/airavata-custos/pkg/models"
)

ctx := identity.WithCaller(req.Context(), &identity.Caller{UserID: "u-1"})
ctx = identity.WithPrivilegesForTest(ctx, []models.PrivilegeKey{models.PrivilegeHPCWrite})
req = req.WithContext(ctx)
srv.ServeHTTP(rr, req)
```

The handler then sees a caller with the named privileges, without going
through the JWT verifier.

### Integration tests

Mark the file with `//go:build integration` on the first line. Then provide
the database DSN via env var and run with the tag:

```bash
export CORE_TEST_DATABASE_DSN='admin:admin@tcp(localhost:3306)/custos?...&multiStatements=true'
go test -tags integration ./connectors/Lustre/Sync/...
```

The dev compose stack on `:3306` is sufficient. If your tests need destructive
fixtures (truncate tables, seed deterministic rows), spin up your own
isolated stack under `dev-ops/local-<connector>/` so the dev DB stays
untouched. AMIE is the existing example.

[CONTRIBUTING.md](../../CONTRIBUTING.md) covers the full test pyramid,
including the AMIE-specific isolated stack on `:3307`.

---

## Common pitfalls

- **Returning an error for "config missing".** If your connector's required config (API URL, credentials, etc.) isn't set in `config/custos.yaml` or the environment, do NOT return an error from `LoadConnector`. Any error here fails the **whole server startup** — core and every other connector are blocked, not just yours. Instead, log a warning and `return nil`. Your connector silently does nothing; the rest of the server boots cleanly.

- **Not registering background goroutines with `wg`.** When you spawn a long-running goroutine in `LoadConnector`, you must call `wg.Add(1)` first and `defer wg.Done()` inside it. The server uses that WaitGroup to track background work and gives it up to 30 seconds to drain on shutdown. If your goroutine isn't on the WaitGroup, the server doesn't know it exists; it cancels `ctx` and exits while your goroutine is still mid-flight. Anything in progress gets cut off — a transaction may commit half-done, an external reply may never get sent, audit rows may be missing entries. The pattern is:
  ```go
  wg.Add(1)
  go func() {
      defer wg.Done()
      yourWorker.Run(ctx) // honour ctx.Done() inside
  }()
  ```
- **Mounting routes outside `/connectors/<name>/...`.** Other connectors collide; nothing tells you until production.
- **Bypassing `coreService` for domain mutations.** Use `coreService.CreateUser`, `CreateComputeAllocationMembership`, etc., so transactions and audit rows stay consistent. Composing connector-side stores around `internal/store` types is fine when you are wrapping core (e.g. AMIE builds an audit service over `corestore.NewAuditEventStore`), but never call a core store to mutate users, projects, or allocations directly — go through `coreService`.
- **Forgetting to regenerate the OpenAPI spec.** `make verify-no-drift` will fail CI.
- **Skipping license headers** on new `.go` and `.sql` files. Same CI failure mode.

## Domain glossary

Domain terms (HPC, AMIE, COmanage, allocation, PI, DN, site code, etc.) live
in their own document so connector READMEs and external readers can reference
them: see [`docs/glossary.md`](../glossary.md).

---

## Where to look next

- Smallest working sample: `connectors/TempAccount/` (~225 lines total).
- Event-subscriber sample: `connectors/COmanage/Identity-Provisioner/`.
- Full-shape sample: `connectors/ACCESS/AMIE-Processor/` (HTTP + workers + DB + external client).
- Router API: `pkg/identity/middleware.go`.
- Central loader and the registration map: `internal/connectors/loader.go`.
- Privilege catalog: `pkg/models/privilege.go`.
- Architecture and audit conventions: [`docs/architecture.md`](../architecture.md).
