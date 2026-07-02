# Privilege Reference

The authoritative list of privilege keys, their owning component, and the
endpoints they gate. Keep this in sync with `pkg/models/privilege.go` and
each connector's `privileges.go`.

## Key format

```
<owner>:<scope>:<action>
```

| Segment  | Meaning                                                                                                                | Examples                           |
|----------|------------------------------------------------------------------------------------------------------------------------|------------------------------------|
| `owner`  | The component that defines and consumes the privilege. `core` for the Custos core; the connector short name otherwise. | `core`, `amie`                     |
| `scope`  | The resource the privilege applies to inside that owner.                                                               | `clusters`, `packets`, `replies`   |
| `action` | The verb the privilege gates. Use `read` for reads, `write` for mutations, `grant` / `manage` for meta operations.     | `read`, `write`, `grant`, `manage` |

A privilege key with a different shape (e.g. `amie:read`) is legacy and
should be migrated.

## Core (`pkg/models/privilege.go`)

| Privilege                | Endpoints it gates                                                                                                                                                                                                                                                                                                                                                                                                                      |
|--------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `core:clusters:read`     | `GET` endpoints under `/compute-clusters*` and `/compute-cluster-users/{id}`, plus `GET /users/{id}/compute-cluster-users`                                                                                                                                                                                                                                                                                                              |
| `core:clusters:write`    | `POST /compute-clusters`, `POST /compute-cluster-users`, `PUT /compute-cluster-users/{id}`, `DELETE /compute-cluster-users/{id}`                                                                                                                                                                                                                                                                                                        |
| `core:allocations:read`  | Every `GET` across the allocation family: `/compute-allocations*`, `/compute-allocation-resources*`, `/compute-allocation-resource-rates*`, `/compute-allocation-diffs*`, `/compute-allocation-change-requests*`, `/compute-allocation-change-request-events*`, `/compute-allocation-memberships*`, `/compute-allocation-membership-resource-overrides*`, `/compute-allocation-usages*`, and the per-user views under `/users/{id}/...` |
| `core:allocations:write` | Every `POST` / `PUT` / `DELETE` across the same allocation family                                                                                                                                                                                                                                                                                                                                                                       |
| `core:projects:read`     | `GET /projects`, `GET /projects/{id}`, `GET /projects/{id}/members`                                                                                                                                                                                                                                                                                                                                                                     |
| `core:projects:write`    | `POST /projects`, `PUT /projects/{id}/status`                                                                                                                                                                                                                                                                                                                                                                                           |
| `core:users:read`        | `GET /users/{id}`, `GET` endpoints under `/user-identities*`, `GET /users/{id}/user-identities` (user identities are user records, so they share the `users` scope)                                                                                                                                                                                                                                                                     |
| `core:users:write`       | `POST /users`, `PUT /users/{id}/status`, `POST /users/merge`, `POST /user-identities`, `PUT /user-identities/{id}`, `DELETE /user-identities/{id}`                                                                                                                                                                                                                                                                                      |
| `core:organizations:read`  | `GET /organizations/{id}`                                                                                                                                                                                                                                                                                                                                                                                                             |
| `core:organizations:write` | `POST /organizations`                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `core:traces:read`       | `GET /audit/traces`, `GET /audit/traces/{trace_id}`, `GET /audit/events`, `GET /audit/sources`                                                                                                                                                                                                                                                                                                                                          |
| `core:privileges:grant`  | `GET /privileges/catalog`, `GET /users/{id}/privileges`, `GET /privileges/{key}/holders`, `POST /users/{id}/privileges`, `DELETE /users/{id}/privileges/{key}`                                                                                                                                                                                                                                                                          |
| `core:roles:manage`      | `GET /roles`, `POST /roles`, `PUT /roles/{id}`, `DELETE /roles/{id}`, and the role-assignment endpoints under `/users/{id}/roles`                                                                                                                                                                                                                                                                                                       |

Only three routes skip the privilege check by design: `GET /healthz` is
public, and `GET /me` / `GET /user/privileges` accept any verified caller
so users can read their own profile and effective privilege set.

The bootstrap super_admin role carries every privilege registered at
startup, core and connector alike. Individual grants beyond that are made
by an admin holding `core:privileges:grant`.

## AMIE connector (`connectors/ACCESS/AMIE-Processor/server/privileges.go`)

| Privilege             | Endpoints it gates                                                                                                                                                                        |
|-----------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `amie:packets:read`   | `GET /connectors/amie/packets`, `GET /connectors/amie/packets/{id}`, `GET /connectors/amie/packets/{id}/events`, `GET /connectors/amie/packets/{id}/audits`, `GET /connectors/amie/stats` |
| `amie:packets:write`  | `POST /connectors/amie/packets/{id}/retry`, `POST /connectors/amie/packets/{id}/resolve`                                                                                                  |
| `amie:replies:read`   | `GET /connectors/amie/replies`                                                                                                                                                            |
| `amie:replies:write`  | `POST /connectors/amie/replies/{id}/retry`                                                                                                                                                |
| `amie:unmapped:read`  | `GET /connectors/amie/unmapped`                                                                                                                                                           |
| `amie:unmapped:write` | _(reserved, no endpoint yet)_                                                                                                                                                             |

## TempAccount connector (`connectors/TempAccount/internal/privileges.go`)

| Privilege                    | Endpoints it gates                                                                                                                                        |
|------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| `temp-account:accounts:read`  | `GET /connectors/temp-account/membership/{user_id}`                                                                                                       |
| `temp-account:accounts:write` | `POST /connectors/temp-account/create`, `POST /connectors/temp-account/assign-allocation`, `POST /connectors/temp-account/update-allocation`, `DELETE /connectors/temp-account/remove/{user_id}` |

## Suggested role bundles

These bundles are conventions, not enforced. Use them as starting points
when defining roles in your deployment.

| Role                      | Privileges                       |
|---------------------------|----------------------------------|
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
