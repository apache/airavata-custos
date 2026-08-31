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
when the core service emits `ComputeClusterUserCreateEvent`. Parallels
the [COmanage Identity-Provisioner](../../COmanage/Identity-Provisioner/):
a site runs one or the other depending on whether it fronts LDAP with a
COmanage Registry. This is the direct-to-LDAP path for sites that don't
use COmanage.

## How it loads

The loader is wired from `internal/connectors/loader.go` alongside
COmanage, SLURM, AMIE, and Temp-Account. On startup:

1. `ldap.LoadConnector` reads five required config values (YAML or env):
   `LDAP_URL`, `LDAP_BIND_DN`, `LDAP_BIND_PASSWORD`, `LDAP_BASE_DN`, and
   `CUSTOS_CLUSTER_ID`. Optionally: `LDAP_GROUP_BASE_DN` (enables
   posixGroup provisioning).
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

For each `ComputeClusterUserCreateEvent` whose `ComputeClusterID` matches
`CustosClusterID`, the orchestrator:

| # | Step | What it does |
|---|------|--------------|
| 1 | Validate `local_username` | Empty → `LDAPProvisioningFailed` audit, return. |
| 2 | Resolve `models.User` | `core.GetUser(user_id)` — needed for cn / sn / mail. |
| 3 | **Cache lookup** | `core.ListUserIdentitiesForUser(user_id)` filtered to `source="ldap"`. Cache hit → jump to step 6 with the cached uidNumber. |
| 4 | **LDAP lookup** | `client.FindPosixAccount(local_username)` — if the entry already exists (out-of-band or a prior run that failed to cache), read its uidNumber, cache it in `user_identities`, jump to step 6. |
| 5 | **Allocate + Add** | `uidSeq.Allocate(clusterID)` pulls the next uidNumber from the persistent counter (single atomic UPDATE, InnoDB row lock — no LDAP scan). Then `client.AddPosixAccount(...)`. On `EntryAlreadyExists` (concurrent writer took the DN), adopt via `tryAdoptExisting`. On `constraintViolation` (out-of-band writer took the uidNumber), retry up to 3 times — the counter yields a fresh value each time. On success, cache in `user_identities` and emit `LDAPAccountCreated`. |
| 6 | **Modify to sync** | With a known uidNumber, `client.ModifyPosixAccount` refreshes cn / sn / givenName / homeDirectory / loginShell / mail. Numeric IDs are intentionally not modified. Emit `LDAPAccountUpdated`. |
| 7 | **Primary posixGroup** | When `LDAP_GROUP_BASE_DN` is set, ensure `cn=<local_username>,<GroupBaseDN>` exists with `gidNumber=<uidNumber>`. `FindPosixGroup` first, `AddPosixGroup` if absent. Concurrent creation is tolerated (`EntryAlreadyExists` is treated as success). Skipped entirely when `GroupBaseDN` is empty (auto-private-groups on the client side). Emit `LDAPGroupCreated` when a new group is written. |

## UID / GID allocation — design

The direct-LDAP path mirrors the invariant the COmanage connector encodes:
**the identity registry owns the uidNumber; Custos reads and caches.**

- **COmanage path**: COmanage's identifier-assignment plugin is the
  registry. The connector reads uidNumber from the CoPerson composite
  (`extractIdentifier(composite, "uidnumber")`) and caches the
  `comanage_id` in `user_identities(source="comanage")`.
- **LDAP path** (this connector): the identity registry is a persistent
  monotonic counter in the connector's own `ldap_uid_sequence` table
  (one row per cluster). Every allocation is a single atomic
  `UPDATE ... SET next_uid = LAST_INSERT_ID(next_uid + 1)` — InnoDB
  row locking serialises concurrent allocators across processes so
  no LDAP-side uidNumber uniqueness constraint is required to prevent
  collisions. The counter never regresses on entry deletion, so a new
  user cannot inherit a deleted user's uid and therefore cannot
  suddenly own files stamped with that numeric uid on the cluster.
  Assigned numbers are also cached in
  `user_identities(source="ldap:<CustosClusterID>")` so
  re-provisioning the same user is O(1). The cache key is scoped per
  cluster so a Custos deployment servicing multiple clusters keeps
  each cluster's uidNumber independent — the same Custos user can have
  different POSIX uids on different clusters.

`gidNumber = uidNumber` — the one-group-per-user pattern the COmanage
connector uses on the CoGroup identifier ([`ensure_posix_account.go:98`](../../COmanage/Identity-Provisioner/internal/operations/ensure_posix_account.go#L98)).

Allocation floor is configurable via `min_uid` (default `50000`) — sites
typically reserve 0–999 for system, 1000–49999 for local, and 50000+
for auto-provisioned identities.

Race handling:

- **In-process and cross-process** are handled by the same mechanism:
  `store.UIDSequence.Allocate` performs a single-statement atomic
  `UPDATE ... SET next_uid = LAST_INSERT_ID(next_uid + 1)` on the
  cluster's row in `ldap_uid_sequence`. InnoDB row locking serialises
  concurrent callers whether they're two goroutines in one process or
  two Custos instances in a multi-instance deployment — the counter
  hands each caller a distinct value.
- **Constraint violations** on the LDAP-side `uidNumber` uniqueness
  constraint (if configured) still happen when an out-of-band writer
  claims a uid the counter hasn't yet reached. The orchestrator
  retries up to 3 times, pulling the next counter value each attempt,
  before emitting `LDAPProvisioningFailed`.

## Primary groups

When `LDAP_GROUP_BASE_DN` is set, the connector creates a `posixGroup`
entry at `cn=<LocalUsername>,<LDAP_GROUP_BASE_DN>` with
`gidNumber=<uidNumber>` alongside the account. This mirrors the
COmanage flow — COmanage creates a per-user CoGroup with a `gidnumber`
identifier as part of provisioning.

When `LDAP_GROUP_BASE_DN` is empty, the group step is skipped entirely.
This works on systems using automatic-private-groups (the default on
RHEL / Fedora), where each user's primary group is derived from their
uid without a matching LDAP `posixGroup` entry. Strict SSSD setups that
require `getgrgid()` to resolve the primary group must set
`LDAP_GROUP_BASE_DN`.

## What gets written

Entry at `uid=<LocalUsername>,<LDAP_BASE_DN>` with object classes
`top, person, organizationalPerson, inetOrgPerson, posixAccount` and:

| LDAP attribute | Source |
|---|---|
| `uid` | `ComputeClusterUser.LocalUsername` |
| `uidNumber` | Allocated (max+1 from LDAP, floor `min_uid`) or read from cache |
| `gidNumber` | Same value as `uidNumber` |
| `cn` | `User.FirstName + " " + User.LastName` |
| `sn` | `User.LastName` |
| `givenName` | `User.FirstName` |
| `mail` | `User.Email` |
| `homeDirectory` | `<homedir_prefix>/<LocalUsername>` (default `/home/<LocalUsername>`) |
| `loginShell` | `default_shell` (default `/bin/bash`) |

## Idempotency

Guaranteed at two levels:

- **Per-event**: cache lookup + LDAP lookup + retry loop mean a re-fired
  event converges on the existing entry rather than creating a duplicate.
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
| `LDAPGroupCreated` | `AddPosixGroup` succeeded. Details include DN, cn, gid. Only when `LDAP_GROUP_BASE_DN` is set. |
| `LDAPProvisioningFailed` | Any step returned an error. Details carry `step=<name> err=<message>` so the dead-letter is tractable from the audit table alone. |

## Local development

```bash
export LDAP_URL=ldap://localhost:389
export LDAP_BIND_DN=cn=admin,dc=example,dc=edu
export LDAP_BIND_PASSWORD=admin
export LDAP_BASE_DN=ou=people,dc=example,dc=edu
export LDAP_VERIFY_SSL=false          # dev only
export CUSTOS_CLUSTER_ID=<UUID of the compute_clusters row this connector services>
# optional:
export LDAP_GROUP_BASE_DN=ou=groups,dc=example,dc=edu   # enables posixGroup provisioning
export LDAP_MIN_UID=50000
go run ./cmd/server
```

Without those vars the connector skips registration silently, so the
rest of Custos runs unaffected.

For a quick LDAP server run OpenLDAP in Docker:

```bash
docker run -d --name openldap -p 389:389 \
  -e LDAP_ADMIN_PASSWORD=admin \
  -e LDAP_ROOT=dc=example,dc=edu \
  bitnami/openldap:latest
```

Then trigger a `ComputeClusterUserCreateEvent` (e.g. via the AMIE
processor's provisioning flow or by hand-inserting a
`compute_cluster_users` row) and verify with `ldapsearch`:

```bash
ldapsearch -H ldap://localhost:389 -x -D cn=admin,dc=example,dc=edu -w admin \
  -b ou=people,dc=example,dc=edu '(objectClass=posixAccount)'
```

## Test surface

Unit tests (no external services):

```bash
go test ./connectors/LDAP/Provisioner/...
```

Integration tests for the persistent UID sequence (requires a live
MariaDB/MySQL with the connector migrations applied):

```bash
export LDAP_TEST_DSN='admin:admin@tcp(localhost:3306)/custos?parseTime=true'
go test -tags integration ./connectors/LDAP/Provisioner/internal/store/...
```

Covers Seed idempotency (including the never-regress guarantee),
sequential monotonicity, concurrent-allocator distinctness, per-cluster
isolation, and the "Allocate without Seed" error path.

All unit tests use interface fakes (no live LDAP / DB). Coverage:

- **`internal/client/client_test.go`** — Find, Add, Modify, allocation
  math, constraint-violation detection, connection reuse and
  reconnect-after-error.
- **`internal/operations/ensure_posix_account_test.go`** — new-user
  allocate+Add, cached-uid Modify, adoption of pre-existing entries,
  retry on constraint violation, retry-exhaustion, propagation of
  non-retryable errors, corrupt-cache tolerance.

## One-time seeding at startup

On boot the loader runs a single LDAP scan for `max(existing uidNumber)`
under `BaseDN` and calls `store.UIDSequence.Seed` with the greater of
that value and the configured `min_uid`. `Seed` uses
`INSERT ... ON DUPLICATE KEY UPDATE next_uid = GREATEST(next_uid, VALUES(next_uid))`,
so:

- **Fresh deployment**: the counter starts above any entries already
  in LDAP from out-of-band provisioning.
- **Restart of an existing deployment**: the counter keeps its
  current value; a re-scan cannot lower it.
- **LDAP unreachable at boot**: fall back to the floor
  (`min_uid`, default 50000). The connector logs the fallback and
  continues starting up.

Steady-state allocations do **not** scan LDAP — that path is only
taken at seed time.

## Out of scope for this PR

- Membership deactivation / status changes (only `Create` handled today).
- Secondary-group / allocation-level groupings — v1 creates only the
  per-user primary posixGroup.
- Reconciliation loop (drift between LDAP and Custos is invisible today).
- Metrics.
- Integration tests against a real LDAP container in CI.
