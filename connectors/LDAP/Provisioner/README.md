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
| 1 | Validate `local_username` | Empty or non-POSIX-safe → `LDAPProvisioningFailed` audit, return. |
| 2 | Resolve `models.User` | `core.GetUser(user_id)` — needed for cn / sn / mail. |
| 3 | **Cache lookup** | `core.ListUserIdentitiesForUser(user_id)` filtered to `source="ldap:<clusterID>"`. Cache hit → jump to step 6 with the cached uidNumber. |
| 4 | **LDAP lookup** | `client.FindPosixAccount(local_username)` — if the entry already exists (out-of-band or a prior run that failed to cache), read its uidNumber, cache it in `user_identities`, jump to step 6. |
| 5 | **Atomic Allocate + Add** | `client.AllocateAndAddPosixAccount(minUID, build)` holds the client mutex across the LDAP `max(uidNumber)+1` scan and the subsequent `Add`. On `EntryAlreadyExists` (concurrent writer took the DN), adopt via `tryAdoptExisting`. On `constraintViolation` (LDAP-side uidNumber uniqueness), bump the floor and retry up to 3 times. On success, cache in `user_identities` and emit `LDAPAccountCreated`. |
| 6 | **Modify to sync** | With a known uidNumber, `client.ModifyPosixAccount` refreshes cn / sn / givenName / homeDirectory / loginShell / mail. Numeric IDs are intentionally not modified. Emit `LDAPAccountUpdated`. |
| 7 | **Primary posixGroup** | When `LDAP_GROUP_BASE_DN` is set, ensure `cn=<local_username>,<GroupBaseDN>` exists with `gidNumber=<uidNumber>`. `FindPosixGroup` first, `AddPosixGroup` if absent. Concurrent creation is tolerated (`EntryAlreadyExists` is treated as success). Skipped entirely when `GroupBaseDN` is empty (auto-private-groups on the client side). Emit `LDAPGroupCreated` when a new group is written. |

## UID / GID allocation — design

The direct-LDAP path mirrors the invariant the COmanage connector encodes:
**the identity registry owns the uidNumber; Custos reads and caches.**

- **COmanage path**: COmanage's identifier-assignment plugin is the
  registry. The connector reads uidNumber from the CoPerson composite
  (`extractIdentifier(composite, "uidnumber")`) and caches the
  `comanage_id` in `user_identities(source="comanage")`.
- **LDAP path** (this connector): LDAP itself is the registry. The
  connector reads `max(uidNumber) + 1` under `LDAP_BASE_DN` and caches
  the assigned number in `user_identities(source="ldap:<CustosClusterID>")`.
  The cache key is scoped per cluster so a Custos deployment servicing
  multiple clusters keeps each cluster's uidNumber independent.

`gidNumber = uidNumber` — the one-group-per-user pattern the COmanage
connector uses on the CoGroup identifier ([`ensure_posix_account.go:98`](../../COmanage/Identity-Provisioner/internal/operations/ensure_posix_account.go#L98)).

Allocation floor is configurable via `min_uid` (default `50000`) —
sites typically reserve 0–999 for system, 1000–49999 for local, and
50000+ for auto-provisioned identities.

## Known limitations of the naive allocator (deferred to a follow-up)

The `max(uidNumber) + 1` approach is intentionally the v1 shape here —
matches COmanage's schema-less posture and keeps the surface small
enough for review. Two real correctness gaps that a durable allocator
would close:

1. **UID reuse after entry deletion.** If a `posixAccount` entry is
   removed from LDAP, `max` drops and the next allocation may re-hand
   that number to a new user. On HPC clusters where numeric uids
   stamp file ownership, a new user could inherit files owned by the
   deleted one.
2. **Cross-process races** rely on the target LDAP server having a
   `uidNumber` uniqueness constraint configured. Without one, silent
   duplicates are possible when two Custos instances allocate
   concurrently before either has committed.

Mitigations already in this PR:

- `user_identities(source="ldap:<clusterID>")` persists across LDAP
  entry deletion, so a *re-provisioned Custos user* gets their prior
  uid back via the cache-hit path — not a new one. This protects
  users Custos still knows about, though not users deleted from
  Custos or provisioned out-of-band.
- In-process races are prevented by holding the client's mutex across
  the allocate/write pair (`AllocateAndAddPosixAccount`).
- Cross-process constraint violations trigger up to 3 retries with an
  ever-higher floor before emitting `LDAPProvisioningFailed`.

The right long-term fix is either a persistent monotonic counter in
this connector's own table (AMIE-style; ~150 LOC) or delegating uid
assignment to a server-side plugin (389 DS DNA / FreeIPA). Which of
those the connector should target is a design question for the
mentor and is intentionally left open until after v1 lands.

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

```bash
go test ./connectors/LDAP/Provisioner/...
```

All unit tests use interface fakes (no live LDAP / DB). Coverage:

- **`internal/client/client_test.go`** — Find, Add, Modify, allocation
  math, atomic allocate/add, constraint-violation detection, connection
  reuse and reconnect-after-error.
- **`internal/operations/ensure_posix_account_test.go`** — new-user
  allocate+Add, cached-uid Modify, adoption of pre-existing entries,
  concurrent-Add adoption (`EntryAlreadyExists`), retry-with-higher-floor,
  retry-exhaustion, propagation of non-retryable errors, per-cluster
  cache isolation, corrupt-cache tolerance, DN-metacharacter rejection.

## Out of scope for this PR

- **Durable monotonic UID allocator.** See "Known limitations of the
  naive allocator" above. Design decision (persistent counter vs.
  server-side plugin like 389 DS DNA) deferred to a follow-up PR.
- Membership deactivation / status changes (only `Create` handled today).
- Secondary-group / allocation-level groupings — v1 creates only the
  per-user primary posixGroup.
- Reconciliation loop (drift between LDAP and Custos is invisible today).
- Metrics.
- Integration tests against a real LDAP container in CI.
