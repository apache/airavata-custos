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
