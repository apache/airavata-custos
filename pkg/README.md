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
| `models` | `github.com/apache/airavata-custos/pkg/models` | Shared domain types (`User`, `Organization`, `Project`, allocations) |
| `service` | `github.com/apache/airavata-custos/pkg/service` | **High-level API** for creating, reading, updating, and deleting entities |

The supporting `internal/db` and `internal/store` packages are private to this
module and are not importable from outside the repository.

---

## Requirements

- Go 1.24+
- MySQL 8+ or MariaDB 10.5+

---

## Quick Start

Every service method takes a [`context.Context`](https://pkg.go.dev/context)
as its first argument. The context carries cancellation, deadlines, and
request-scoped values down into the database driver, so an in-flight query is
aborted if the caller goes away or a deadline passes. You typically derive it
from one of:

- `context.Background()` — root context for `main`, init, tests, scripts.
- `r.Context()` — inside an HTTP handler; cancelled when the client disconnects.
- The `ctx` argument of a gRPC handler.
- `context.WithTimeout(parent, d)` — adds a deadline; **always `defer cancel()`**.
- `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` — cancelled on Ctrl-C / SIGTERM.

Never pass `nil`; use `context.Background()` (or `context.TODO()`) if you have
nothing better.

```go
import (
    "context"
    "time"

    "github.com/apache/airavata-custos/pkg/models"
    "github.com/apache/airavata-custos/pkg/service"
)

svc := service.New(database) // *sqlx.DB

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

## Read / Update / Delete

```go
org,  err := svc.GetOrganization(ctx, orgID)
user, err := svc.GetUserByEmail(ctx, "ada@example.edu")
proj, err := svc.GetProject(ctx, projID)

users, err    := svc.ListUsersByOrganization(ctx, orgID)
projects, err := svc.ListProjectsByPI(ctx, userID)

err = svc.UpdateUser(ctx, user)
err = svc.DeleteProject(ctx, projID)
```

## Errors

| Sentinel | Returned when |
|----------|---------------|
| `service.ErrNotFound` | A `GetX` call finds no matching record |
| `service.ErrAlreadyExists` | Duplicate email or duplicate `originated_id` |
| `service.ErrInvalidInput` | Missing required field, or unknown FK reference |

Use `errors.Is(err, service.ErrNotFound)` to check.

## Schema

Tables created by the embedded migrations in `/internal/db/migrations/`:

```
organizations
users          (FK → organizations.id)
projects       (FK → users.id  via project_pi_id)
```

`parseTime=true` must be set in the DSN so MySQL `TIMESTAMP` columns scan into
`time.Time` correctly.

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

