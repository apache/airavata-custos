# Privilege Reference

The authoritative list of privilege keys, their owning component, and the
endpoints they gate. Keep this in sync with `pkg/models/privilege.go` and
each connector's `privileges.go`.

## Key format

```
<owner>:<scope>:<action>
```

| Segment | Meaning | Examples |
|---|---|---|
| `owner` | The component that defines and consumes the privilege. `core` for the Custos core; the connector short name otherwise. | `core`, `amie` |
| `scope` | The resource the privilege applies to inside that owner. | `clusters`, `packets`, `replies` |
| `action` | The verb the privilege gates. Use `read` for reads, `write` for mutations, `grant` / `manage` for meta operations. | `read`, `write`, `grant`, `manage` |

A privilege key with a different shape (e.g. `amie:read`) is legacy and
should be migrated.

## Core (`pkg/models/privilege.go`)

| Privilege | Endpoints it gates |
|---|---|
| `core:clusters:read` | `GET /compute-clusters`, `GET /compute-clusters/{id}` |
| `core:clusters:write` | `POST /compute-clusters`, `PUT /compute-clusters/{id}`, `DELETE /compute-clusters/{id}` |
| `core:allocations:read` | _(catalog only; allocation routes still run on plain auth)_ |
| `core:allocations:write` | _(catalog only; allocation routes still run on plain auth)_ |
| `core:projects:read` | _(catalog only; project routes still run on plain auth)_ |
| `core:projects:write` | _(catalog only; project routes still run on plain auth)_ |
| `core:traces:read` | _(catalog only; `/audit/*` routes still run on plain auth)_ |
| `core:privileges:grant` | `GET /privileges/catalog`, `GET /users/{id}/privileges`, `GET /privileges/{key}/holders`, `POST /users/{id}/privileges`, `DELETE /users/{id}/privileges/{key}` |
| `core:roles:manage` | `GET /roles`, `POST /roles`, `PUT /roles/{id}`, `DELETE /roles/{id}`, and the role-assignment endpoints under `/users/{id}/roles` |

Keys marked _catalog only_ are registered and granted like any other and
callers use them to decide what to show, but the matching routes have not
been switched from `RequireAuth` to `RequirePrivilege` yet. That switch is
tracked as a follow-up.

The bootstrap super_admin role carries every privilege registered at
startup, core and connector alike. Individual grants beyond that are made
by an admin holding `core:privileges:grant`.

## AMIE connector (`connectors/ACCESS/AMIE-Processor/server/privileges.go`)

| Privilege | Endpoints it gates |
|---|---|
| `amie:packets:read` | `GET /connectors/amie/packets`, `GET /connectors/amie/packets/{id}`, `GET /connectors/amie/packets/{id}/events`, `GET /connectors/amie/packets/{id}/audits`, `GET /connectors/amie/stats` |
| `amie:packets:write` | `POST /connectors/amie/packets/{id}/retry`, `POST /connectors/amie/packets/{id}/resolve` |
| `amie:replies:read` | `GET /connectors/amie/replies` |
| `amie:replies:write` | `POST /connectors/amie/replies/{id}/retry` |
| `amie:unmapped:read` | `GET /connectors/amie/unmapped` |
| `amie:unmapped:write` | _(reserved, no endpoint yet)_ |

## Suggested role bundles

These bundles are conventions, not enforced. Use them as starting points
when defining roles in your deployment.

| Role | Privileges |
|---|---|
| `super_admin` (bootstrap) | Every key registered at startup. |

## Adding a new privilege

1. Pick a key in the `<owner>:<scope>:<action>` shape.
2. Declare it in the owning package:
   - Core: add a constant to `pkg/models/privilege.go` and register it
     via `models.Register(...)`.
   - Connector: add a constant to the connector's `server/privileges.go`
     and register it in the connector's init path.
3. Gate the endpoint with `router.RequirePrivilege(pattern, KEY, handler)`.
4. Add a row to the appropriate table above.

The runtime privilege catalog (`GET /privileges/catalog`) returns whatever
the registry holds at startup, so registering is what makes a key real.
