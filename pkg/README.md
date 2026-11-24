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

# pkg

Public Go libraries shared across this repository's modules. Following the
[golang-standards/project-layout](https://github.com/golang-standards/project-layout)
convention, only externally-usable code lives here; private implementation
details live in the top-level [`/internal`](../internal/) tree.

Module root: `github.com/apache/airavata-custos` (root `go.mod`).

## Packages

| Package | Import path | Purpose |
|---------|-------------|---------|
| `models` | `github.com/apache/airavata-custos/pkg/models` | Shared domain types (`User`, `Organization`, `Project`, allocations, roles, audit and traces) |
| `service` | `github.com/apache/airavata-custos/pkg/service` | **High-level API** for creating, reading, updating, and deleting entities |
| `events` | `github.com/apache/airavata-custos/pkg/events` | The event bus that carries domain events from the service to the connectors through the database |
| `identity` | `github.com/apache/airavata-custos/pkg/identity` | The verified caller, the OIDC token verifier and the router that applies route policies |
| `posix` | `github.com/apache/airavata-custos/pkg/posix` | Builds and validates POSIX usernames for cluster accounts |
| `common` | `github.com/apache/airavata-custos/pkg/common` | HTTP helpers for handlers: JSON responses, paging and service error mapping |

The supporting `internal/db` and `internal/store` packages are private to this
module and are not importable from outside the repository.

## Using the service

Every service method takes a [`context.Context`](https://pkg.go.dev/context)
as its first argument. The context carries cancellation, deadlines, and
request-scoped values down into the database driver, so an in-flight query is
aborted if the caller goes away or a deadline passes. You typically derive it
from one of:

- `context.Background()` — root context for `main`, init, tests, scripts.
- `r.Context()` — inside an HTTP handler; cancelled when the client disconnects.
- `context.WithTimeout(parent, d)` — adds a deadline; **always `defer cancel()`**.
- `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` — cancelled on Ctrl-C / SIGTERM.

Never pass `nil`; use `context.Background()` (or `context.TODO()`) if you have
nothing better.

Create the bus first, then the service on the same database. `bus.Run(ctx)` delivers events to subscribers, as
[the event bus](../docs/architecture.md#event-bus) describes.

```go
import (
    "context"
    "time"

    "github.com/apache/airavata-custos/pkg/events"
    "github.com/apache/airavata-custos/pkg/models"
    "github.com/apache/airavata-custos/pkg/service"
)

// database is a *sqlx.DB with the core migrations applied.
bus, err := events.New(context.Background(), database)
svc := service.New(database, bus)

// Derive a context with a 10s budget for this call chain.
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

org, err := svc.CreateOrganization(ctx, &models.Organization{
    Name:         "University of Example",
    OriginatedID: "ACCESS-ORG-001",
})

user, err := svc.CreateUser(ctx, &models.User{
    OrganizationID: org.ID,
    FirstName:      "Ada",
    LastName:       "Lovelace",
    Email:          "ada@example.edu",
})

project, err := svc.CreateProject(ctx, &models.Project{
    Title:        "Climate Simulation 2026",
    Origination:  "ACCESS",
    OriginatedID: "ACCESS-PRJ-9000",
    ProjectPIID:  user.ID,
})
```

- If `ID` is left empty the service generates a UUID.
- If `CreatedTime` on a project is zero, the service sets it to `time.Now().UTC()`.
- The populated entity is returned.

Reads, updates and deletes follow the same pattern:

```go
org,  err := svc.GetOrganization(ctx, orgID)
user, err := svc.GetUserByEmail(ctx, "ada@example.edu")
proj, err := svc.GetProject(ctx, projID)

users, err    := svc.ListUsersByOrganization(ctx, orgID)
projects, err := svc.ListProjectsForParticipant(ctx, userID)

err = svc.UpdateUser(ctx, user)
err = svc.DeleteProject(ctx, projID)
```

## Errors

The service wraps one of these sentinels, so check with `errors.Is(err, service.ErrNotFound)`.
`common.WriteServiceError` maps them to HTTP status codes.

| Sentinel | Returned when | HTTP |
|---|---|---|
| `service.ErrInvalidInput` | a required field is missing, or a reference points at a missing record | 400 |
| `service.ErrNotFound` | a get finds no record | 404 |
| `service.ErrAlreadyExists` | a record with the same email or `originated_id` exists | 409 |
| `service.ErrInUse` | a record cannot be removed while others depend on it | 409 |
| `service.ErrPIChange` | an update would change a project's PI | 409 |

`common.WriteServiceError` also maps a PostgreSQL unique violation to 409, a foreign-key or length violation to 400,
and any other error to 500.

## Schema

The tables come from [`internal/db/migrations/`](../internal/db/migrations/). The DSN is a PostgreSQL URL, such as
`postgres://user:pass@localhost:5432/custos?sslmode=disable`.

## Testing with a mock service

The `service` package exposes a `CoreService` interface (see
[`pkg/service/interface.go`](service/interface.go)) that the concrete
`*service.Service` satisfies. Callers should depend on this interface — or one
of the narrower per-domain interfaces in the same file — so tests can swap in
a mock without a database.

A mock implementation, `CoreServiceMock`, is generated by
[matryer/moq](https://github.com/matryer/moq) into
[`pkg/service/mock.go`](service/mock.go). Each method has a corresponding
`XxxFunc` field; set only the funcs your test needs, and use the matching
`XxxCalls()` helper to assert how the code under test invoked the service.

```go
import (
    "context"
    "testing"

    "github.com/apache/airavata-custos/pkg/models"
    "github.com/apache/airavata-custos/pkg/service"
)

func TestSomething(t *testing.T) {
    svc := &service.CoreServiceMock{
        GetComputeClusterFunc: func(ctx context.Context, id string) (*models.ComputeCluster, error) {
            return &models.ComputeCluster{ID: id, Name: "test-cluster"}, nil
        },
        GetProjectFunc: func(ctx context.Context, id string) (*models.Project, error) {
            return &models.Project{ID: id, ProjectPIID: "pi-id"}, nil
        },
    }

    // ... exercise the code under test, passing svc where a service.CoreService is expected ...

    if got := len(svc.GetComputeClusterCalls()); got != 1 {
        t.Fatalf("GetComputeCluster called %d times, want 1", got)
    }
}
```

Methods whose `Func` field is left nil will panic when invoked, surfacing
unexpected calls loudly.

### Regenerating the mock

After adding or changing a method on any interface in `interface.go`, regenerate
`mock.go`:

```sh
go generate ./pkg/service/...
```

The `//go:generate` directive at the top of `interface.go` invokes moq via
`go run`, so no project-wide install is required. The Go toolchain will fetch
the version of moq it needs on demand.

