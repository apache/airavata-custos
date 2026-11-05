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

# Privilege Reference

Each privilege key, its owner and the routes it gates. Read it to grant access or to gate a new route. Keep it in
sync with `pkg/models/privilege.go` and each connector's `privileges.go`.

## How privileges work

A user holds a key directly or through a role. A `RequirePrivilege` route returns `403 insufficient_privilege` to a
caller without it.

A key exists only once it is registered with `models.Register(...)`. Grants and role changes reject any other key.
`GET /privileges/catalog` returns the registry as loaded at startup.

## Key format

```
<owner>:<scope>:<action>
```

| Segment | Meaning | Examples |
|---|---|---|
| `owner` | The component that defines and checks the key: `core`, or the connector's short name | `core`, `amie`, `temp-account` |
| `scope` | The resource inside that owner | `clusters`, `packets`, `agreements` |
| `action` | The verb: `read` for reads, `write` for mutations, `grant` or `manage` for meta operations | `read`, `write`, `grant`, `manage` |

## Core (`pkg/models/privilege.go`)

| Privilege | Routes it gates |
|---|---|
| `core:clusters:read` | `GET /compute-clusters`<br>`GET /compute-clusters/{id}`<br>`GET /compute-clusters/{id}/users`<br>`GET /compute-clusters/{id}/users/{userId}`<br>`GET /compute-cluster-users`<br>`GET /compute-cluster-users/{id}`<br>`GET /users/{id}/compute-cluster-users`<br>`GET /connectors`<br>`GET /connectors/vast/folders` |
| `core:clusters:write` | `POST /compute-clusters`<br>`POST /compute-cluster-users`<br>`POST /compute-cluster-users/{id}/approve`<br>`POST /compute-cluster-users/{id}/deny`<br>`PUT /compute-cluster-users/{id}`<br>`DELETE /compute-cluster-users/{id}`<br>`DELETE /connectors/vast/folders/{id}`<br>`PUT /connectors/vast/folders/{id}/quota` |
| `core:clusters:deactivate` | `POST /compute-cluster-users/{id}/deactivate` |
| `core:allocations:read` | Each `GET` across the allocation family: `/compute-allocation-resources*`<br>`/compute-allocation-resource-rates*`<br>`/compute-allocation-diffs*`<br>`/compute-allocation-change-requests*`<br>`/compute-allocation-change-request-events*`<br>`/compute-allocation-memberships*`<br>`/compute-allocation-membership-resource-overrides*`<br>`/compute-allocation-usages*`<br>the sub-resources of `/compute-allocations/{id}` not listed under [routes without a privilege](#routes-without-a-privilege)<br>the per-user views:<br>`GET /users/{id}/compute-allocation-memberships`<br>`GET /users/{id}/compute-allocation-usages`<br>`GET /users/{id}/change-requests` |
| `core:allocations:write` | Each `POST`, `PUT` and `DELETE` across the same allocation family, including `POST /compute-allocations` and `DELETE /compute-allocations/{id}`, except the two below |
| `core:allocations:deactivate` | `POST /compute-allocations/{id}/deactivate`<br>`POST /compute-allocation-memberships/{id}/deactivate` |
| `core:projects:read` | `GET /projects/{id}/members` |
| `core:projects:write` | `POST /projects`, `PUT /projects/{id}`, `PUT /projects/{id}/status`, `DELETE /projects/{id}`, `PUT /projects/{id}/members/{userId}` |
| `core:users:read` | `GET /users`<br>`GET /users/{id}`<br>`GET /user-identities/{id}`<br>`GET /user-identities/sources/{source}/external/{externalId}`<br>`GET /user-identities/oidc-subjects/{oidcSub}`<br>`GET /users/{id}/user-identities` |
| `core:users:write` | `POST /users`<br>`PUT /users/{id}`<br>`PUT /users/{id}/status`<br>`POST /users/merge`<br>`POST /user-identities`<br>`PUT /user-identities/{id}`<br>`DELETE /user-identities/{id}` |
| `core:organizations:read` | `GET /organizations`<br>`GET /organizations/{id}` |
| `core:organizations:write` | `POST /organizations` |
| `core:traces:read` | `GET /audit/traces`<br>`GET /audit/traces/{trace_id}`<br>`GET /audit/events`<br>`GET /audit/sources`<br>`GET /events/deliveries`<br>`GET /events/deliveries/{id}`<br>`GET /events/subscriptions` |
| `core:events:manage` | `POST /events/deliveries/{id}/retry`<br>`GET /events/graph` |
| `core:privileges:grant` | `GET /privileges/catalog`<br>`GET /users/{id}/privileges`<br>`GET /privileges/{key}/holders`<br>`POST /users/{id}/privileges`<br>`DELETE /users/{id}/privileges/{key}` |
| `core:roles:manage` | `GET /roles`<br>`POST /roles`<br>`GET /roles/{id}`<br>`PUT /roles/{id}`<br>`DELETE /roles/{id}`<br>`POST /roles/{id}/privileges`<br>`DELETE /roles/{id}/privileges/{key}`<br>`GET /roles/{id}/holders`<br>`GET /users/{id}/roles`<br>`POST /users/{id}/roles`<br>`DELETE /users/{id}/roles/{roleId}` |
| `core:agreements:read` | `GET /agreements`<br>`GET /agreements/{id}`<br>`GET /agreements/{id}/acceptances` |
| `core:agreements:manage` | `POST /agreements`<br>`PUT /agreements/{id}`<br>`DELETE /agreements/{id}`<br>`POST /agreements/{id}/publish`<br>`POST /agreements/{id}/remove` |

`POST /users` with `portal_admin` or `cluster_admin` set also requires `core:roles:manage`.

## Routes without a privilege

These routes need a verified caller but no key.

| Router method | Routes | Access |
|---|---|---|
| `Public` | `GET /healthz` | Anyone, no token |
| `RequireAuthWithoutUserConsent` | `GET /me`<br>`GET /user/privileges`<br>`GET /me/agreements/pending`<br>`POST /me/agreements/{id}/accept`<br>`GET /me/onboarding/username`<br>`POST /me/onboarding` | The caller's own profile and onboarding, before consent |
| `RequireAuth` | `GET /me/onboarding/status` | The caller's own status |
| `RequireAuth` | `GET /projects`<br>`GET /compute-allocations` | All records with `core:projects:read` or `core:allocations:read`; otherwise the caller's own |
| `RequireScoped` | `GET /projects/{id}` | `core:projects:read`, or a participant in the project; `404` otherwise |
| `RequireScoped` | `GET /compute-allocations/{id}`<br>`GET /compute-allocations/{id}/resources`<br>`GET /compute-allocations/{id}/memberships`<br>`GET /compute-allocations/{id}/usages`<br>`GET /compute-allocations/{id}/usages/total` | `core:allocations:read`, or a participant in the allocation; `404` otherwise |
| `RequireAuth` | `GET /connectors/analytics/contexts`<br>`GET /connectors/analytics/allocations/{id}/usage-summary`<br>`GET /connectors/analytics/allocations/{id}/jobs` | Decided by allocation role alone: PI, CO_PI and ALLOCATION_MANAGER see all members, plain members see their own records |

## AMIE connector (`connectors/ACCESS/AMIE-Processor/server/privileges.go`)

| Privilege | Routes it gates |
|---|---|
| `amie:packets:read` | `GET /connectors/amie/packets`<br>`GET /connectors/amie/packets/{id}`<br>`GET /connectors/amie/packets/{id}/events`<br>`GET /connectors/amie/packets/{packet_id}/audits`<br>`GET /connectors/amie/stats` |
| `amie:packets:write` | `POST /connectors/amie/packets/{id}/retry`<br>`POST /connectors/amie/packets/{id}/resolve` |
| `amie:replies:read` | `GET /connectors/amie/replies` |
| `amie:replies:write` | `POST /connectors/amie/replies/{id}/retry` |
| `amie:unmapped:read` | `GET /connectors/amie/unmapped` |
| `amie:unmapped:write` | `POST /connectors/amie/unmapped/{id}/link` |

## TempAccount connector (`connectors/TempAccount/internal/privileges.go`)

| Privilege | Routes it gates |
|---|---|
| `temp-account:accounts:read` | `GET /connectors/temp-account/membership/{user_id}` |
| `temp-account:accounts:write` | `POST /connectors/temp-account/create`<br>`POST /connectors/temp-account/assign-allocation`<br>`POST /connectors/temp-account/update-allocation`<br>`DELETE /connectors/temp-account/remove/{user_id}` |

## Other connectors

The VAST-Provisioner routes use the core `core:clusters:*` keys, listed above. The remaining connectors gate
no route on a key.

The SSH-Certificate-Signer extension looks for `signer:certificates:read` and `signer:certificates:write` in the
privileges that core `GET /me` returns. Core does not register these keys, so they cannot be granted.

## System roles

Two roles are created by the server and marked `is_system`.

| Role | Privileges | Created |
|---|---|---|
| `super_admin` | All keys registered at startup, core and connector alike | At startup, for the user named by `CUSTOS_BOOTSTRAP_ADMIN_EMAIL` ([CONFIG.md](../CONFIG.md#environment-variables-read-directly)) |
| `admin` | All registered keys except `core:privileges:grant` and `core:roles:manage` | The first time `POST /users` onboards a portal admin. Its keys are set only then |

Two revoke paths guard against locking every administrator out:

| Revoke | Refused when |
|---|---|
| A role assignment | No user would still hold `core:privileges:grant` or `core:roles:manage` |
| A direct grant of `core:privileges:grant` | It is the last direct grant of that key, or the caller revokes their own |

A direct grant of `core:roles:manage` has no such guard.

Other roles are site-defined through `/roles`.

## Adding a privilege

1. Pick a key in the `<owner>:<scope>:<action>` shape.
2. Declare it in the owning package:
    - Core: add a constant to `pkg/models/privilege.go` and to its `Register(...)` call.
    - Connector: add a constant to the connector's `privileges.go` and register it from that file's `init`.
3. Gate the route with `router.RequirePrivilege(pattern, KEY, handler)`.
4. Add a row to the matching table above.

`super_admin` picks up a new key at the next start with `CUSTOS_BOOTSTRAP_ADMIN_EMAIL` set. The `admin` role does
not. Add the key to it through `POST /roles/{id}/privileges`.
