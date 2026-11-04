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

# LDAP Provisioner

Connector that provisions POSIX identities directly in an LDAP directory
when a cluster account is approved, and removes them when it is deleted. Parallels
the [COmanage Identity-Provisioner](../../COmanage/Identity-Provisioner/):
a site runs one or the other depending on whether it fronts LDAP with a
COmanage Registry. This is the direct-to-LDAP path for sites that don't
use COmanage.

## How it loads

The loader is wired from `internal/connectors/loader.go` alongside
COmanage, SLURM, AMIE, and Temp-Account. On startup:

1. `ldap.LoadConnector` reads five required values from the connector's
   YAML config: `directory.url`, `directory.bind_dn`,
   `directory.bind_password`, `directory.base_dn`, and
   `provisioning.custos_cluster_id`. Optionally: `directory.group_base_dn`
   (enables posixGroup provisioning). Use `${VAR}` in the YAML to take a
   value from the environment.
2. If any are missing it logs `ldap provisioner: required config not set; skipping`
   and returns `nil` — the documented way to disable the connector in a
   deployment without code changes.
3. If all are present it constructs an LDAP client and a
   `ClusterUserSubscriber`, then registers it against the shared event
   bus. The subscriber filters by `CustosClusterID` so a deployment
   servicing multiple clusters can run multiple connector instances side
   by side.

See `config.example.yaml` for the full config reference.

## What the orchestrator does

For each approved cluster account whose `ComputeClusterID` matches
`CustosClusterID`, the orchestrator:

| # | Step | What it does |
|---|------|--------------|
| 1 | Validate `local_username` | Not POSIX-safe → `LDAPProvisioningFailed` audit, return. |
| 2 | Resolve `models.User` | `core.GetUser(user_id)` — needed for cn / sn / mail. |
| 3 | **Stored attributes** | Look up the `ldap_posix_accounts` row by `compute_cluster_users.id`. When there is none, the username must be free, both in the table and in the directory; otherwise the run fails with `step=claim_username`. Adopting an existing account has to be an explicit import. A free username gets a new row and uidNumber. |
| 4 | **Add or Modify** | No entry → `client.AddPosixAccount(...)`, emit `LDAPAccountCreated`. On `constraintViolation` (out-of-band writer took the uidNumber), the row is renumbered and the add retried up to 3 times. An existing entry must carry the stored uidNumber; it is then refreshed with `client.ModifyPosixAccount` (cn / sn / givenName / homeDirectory / loginShell / mail), emit `LDAPAccountUpdated`. |
| 5 | **Primary posixGroup** | When `group_base_dn` is set, ensure `cn=<local_username>,<GroupBaseDN>` exists with `gidNumber=<uidNumber>`. `FindPosixGroup` first, `AddPosixGroup` if absent. Concurrent creation is tolerated (`EntryAlreadyExists` is treated as success). Skipped entirely when `GroupBaseDN` is empty (auto-private-groups on the client side). Emit `LDAPGroupCreated` when a new group is written. |
| 6 | **Mark provisioned** | `core.MarkComputeClusterUserProvisioned`, which starts the downstream connectors. |

When a cluster account is deleted, the posixAccount entry is removed and
the row is marked `INACTIVE`, emitting `LDAPAccountDeactivated`. When
`group_base_dn` is set, the primary posixGroup is deleted too, emitting
`LDAPGroupDeleted`. An approval for an `INACTIVE` row is ignored.

## UID / GID allocation — design

The connector keeps the attributes it assigns in its own
`ldap_posix_accounts` table, keyed on `compute_cluster_users.id` (which is
already unique per cluster and user), with `UNIQUE (uid_number)` and
`UNIQUE (username)`. A new row gets one above the highest `uid_number` in
the table, and at least the floor. Deactivated rows keep their number, so
a new user cannot inherit a deleted user's uid and therefore cannot
suddenly own files stamped with that numeric uid on the cluster.

`gidNumber = uidNumber` — the one-group-per-user pattern the COmanage
connector uses on the CoGroup identifier ([`ensure_posix_account.go:98`](../../COmanage/Identity-Provisioner/internal/operations/ensure_posix_account.go#L98)).

Allocation floor is configurable via `min_uid` (default `50000`) — sites
typically reserve 0–999 for system, 1000–49999 for local, and 50000+
for auto-provisioned identities. The client rejects any uidNumber or
gidNumber below it.

Race handling:

- **Concurrent assignments**, in one process or across Custos instances,
  take a Postgres advisory lock, so each gets a distinct number.
- **Constraint violations** on the LDAP-side `uidNumber` uniqueness
  constraint (if configured) still happen when an out-of-band writer
  holds the number. The orchestrator renumbers and retries up to 3
  times before emitting `LDAPProvisioningFailed`.

## Primary groups

When `group_base_dn` is set, the connector creates a `posixGroup`
entry at `cn=<LocalUsername>,<group_base_dn>` with
`gidNumber=<uidNumber>` alongside the account. This mirrors the
COmanage flow — COmanage creates a per-user CoGroup with a `gidnumber`
identifier as part of provisioning.

When `group_base_dn` is empty, the group step is skipped entirely.
This works on systems using automatic-private-groups (the default on
RHEL / Fedora), where each user's primary group is derived from their
uid without a matching LDAP `posixGroup` entry. Strict SSSD setups that
require `getgrgid()` to resolve the primary group must set
`group_base_dn`.

## What gets written

Entry at `uid=<LocalUsername>,<LDAP_BASE_DN>` with object classes
`top, person, organizationalPerson, inetOrgPerson, posixAccount` and:

| LDAP attribute | Source |
|---|---|
| `uid` | `ComputeClusterUser.LocalUsername` |
| `uidNumber` | `ldap_posix_accounts.uid_number` |
| `gidNumber` | Same value as `uidNumber` |
| `cn` | `User.FirstName + " " + User.LastName` |
| `sn` | `User.LastName`; else `User.FirstName` (and no `givenName`); else the username — `sn` is required |
| `givenName` | `User.FirstName` |
| `mail` | `User.Email` |
| `homeDirectory` | `<homedir_prefix>/<LocalUsername>` (default `/home/<LocalUsername>`) |
| `loginShell` | `default_shell` (default `/bin/bash`) |

## Idempotency

Guaranteed at two levels:

- **Per-event**: the stored row means a re-fired event converges on the
  existing entry rather than creating a duplicate.
- **Per-uidNumber**: `ModifyPosixAccount` intentionally does not touch
  `uid`, `uidNumber`, or `gidNumber`. A change in any of those would be a
  different account; identity churn on the directory side is out of scope
  for this connector.

## Audit events

Emitted via `coreService.CreateAuditEvent`. `entity_id` is always the
`compute_cluster_users.id`.

| Event type | When |
|------------|------|
| `LDAPProvisioningStarted` | Subscription marker — event accepted for this connector's cluster. |
| `LDAPAccountCreated` | `AddPosixAccount` succeeded. Details include DN, username, uid, gid. |
| `LDAPAccountUpdated` | `ModifyPosixAccount` succeeded. Details include DN, username. |
| `LDAPGroupCreated` | `AddPosixGroup` succeeded. Details include DN, cn, gid. Only when `group_base_dn` is set. |
| `LDAPAccountDeactivated` | The account was deleted from the directory. Details include DN, username, uid. |
| `LDAPGroupDeleted` | The primary posixGroup was deleted on deactivation. Details include DN, cn, gid. |
| `LDAPProvisioningFailed` | Any step returned an error. Details carry `step=<name> err=<message>` so the failure is tractable from the audit table alone. |

## Local development

Add the connector block from `config.example.yaml` to the server's config
file, then:

```bash
go run ./cmd/server
```

Without a complete block the connector skips registration silently, so the
rest of Custos runs unaffected.

For a quick LDAP server run OpenLDAP in Docker:

```bash
docker run -d --name openldap -p 389:389 \
  -e LDAP_ADMIN_PASSWORD=admin \
  -e LDAP_ROOT=dc=example,dc=edu \
  bitnami/openldap:latest
```

Then approve a cluster account (e.g. via the AMIE processor's
provisioning flow or the admin portal) and verify with `ldapsearch`:

```bash
ldapsearch -H ldap://localhost:389 -x -D cn=admin,dc=example,dc=edu -w admin \
  -b ou=people,dc=example,dc=edu '(objectClass=posixAccount)'
```

## Test surface

Unit tests (no external services):

```bash
go test ./connectors/LDAP/Provisioner/...
```

Integration tests for `ldap_posix_accounts` (requires a live Postgres):

```bash
export CORE_TEST_DATABASE_DSN='postgres://admin:admin@localhost:5432/custos_test?sslmode=disable'
go test -tags integration -p 1 ./connectors/LDAP/Provisioner/internal/store/...
```

Covers numbering above every row and the floor, username uniqueness,
concurrent-assignment distinctness, and renumbering.

All unit tests use interface fakes (no live LDAP / DB). Coverage:

- **`internal/client/client_test.go`** — Find, Add, Modify, Delete,
  allocation math, min_uid enforcement, constraint-violation detection,
  connection reuse and reconnect-after-error.
- **`internal/operations/ensure_posix_account_test.go`** — new-user
  Add, stored-row Modify, conflicts on a username held in the table or
  the directory, retry on constraint violation, retry-exhaustion,
  propagation of non-retryable errors, deactivation, correlation_id in
  every log line.

## Seeding the floor at startup

On boot the loader runs a single LDAP scan for `max(existing uidNumber)`
under `BaseDN` and uses the greater of that value + 1 and the configured
`min_uid` as the floor, so a fresh deployment starts above any entries
already in LDAP from out-of-band provisioning. If LDAP is unreachable at
boot it falls back to `min_uid` and logs the fallback.

## Out of scope for this PR

- Importing an existing directory entry into Custos.
- Secondary-group / allocation-level groupings — v1 creates only the
  per-user primary posixGroup.
- Reconciliation loop (drift between LDAP and Custos is invisible today).
- Metrics.
- Integration tests against a real LDAP container in CI.
