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

# COmanage Identity-Provisioner

Connector that bridges Custos to a COmanage Registry. When the core service
emits `ComputeClusterUserCreateEvent` for a cluster this connector services,
the orchestrator ensures the user has a fully provisioned POSIX identity in
COmanage: a CoPerson, a per-user CoGroup, the matching Identifiers, and a
`UnixClusterAccount` block on the target UnixCluster.

## How it loads

The loader is wired from `internal/connectors/loader.go` alongside SLURM and
AMIE. On startup:

1. `comanage.LoadConnector` reads seven required env vars:
   `COMANAGE_REGISTRY_URL`, `COMANAGE_CO_ID`, `COMANAGE_API_USER`,
   `COMANAGE_API_KEY`, `COMANAGE_PERSON_ID_TYPE`,
   `COMANAGE_UNIX_CLUSTER_ID`, `CUSTOS_CLUSTER_ID`.
2. If any are missing it logs `comanage provisioner: required env vars not set; skipping`
   and returns `nil`. This is the documented way to disable the connector in a
   deployment without code changes.

`COMANAGE_PERSON_ID_TYPE` is the Identifier Type configured in the registry's
Identifier Types screen that the connector uses to look up and tag a CoPerson.
Set it to whatever string your COmanage instance uses for the per-person ID.
3. If all six are present it constructs an HTTP client and a
   `ClusterUserSubscriber`, then registers it against the shared event bus.
   The subscriber filters by `CustosClusterID` so a deployment that services
   multiple clusters can run multiple connector instances side by side.

See `config.example.yaml` for the full env var reference (required + optional).

## What the orchestrator does

For each event, the orchestrator runs the following calls in order. Steps that
find an existing record are short-circuited; steps that create something
tolerate "already exists" responses so re-runs after a partial failure are
idempotent.

| # | Step | What it does | Endpoint |
|---|------|--------------|----------|
| 1 | Lookup CoPerson | Look up `comanage_id` stored in `user_identities(source="comanage")`, fall back to REST email search, otherwise POST a new CoPerson. | Core `GET /people/<comanage_id>`, REST `co_people.json`, Core `POST /people` |
| 2 | Store `comanage_id` | Persist the COmanage identifier in `user_identities` so step 1 finds it next time. | Core service `CreateUserIdentity` |
| 3 | Read composite | Pull the full Core composite to read `uidnumber` and `CoPerson.meta.id`. | Core `GET /people/<comanage_id>` |
| 4 | Per-user CoGroup | Find or create a CoGroup named after the user's local username (`GroupType:"CL"`, `Auto:false`). | REST `co_groups.json` |
| 5 | groupname Identifier | Attach a `type:"uid"` Identifier (the local username string) to the CoGroup. | REST `identifiers.json` |
| 6 | gidnumber Identifier | Attach a `type:"gidnumber"` Identifier (numeric, mirrors `uidnumber`) to the CoGroup. | REST `identifiers.json` |
| 7 | CoGroupMember | Join the CoPerson to the CoGroup as member + owner. | REST `co_group_members.json` |
| 8 | UnixClusterGroup | Attach the CoGroup to the configured UnixCluster. A 4xx here is treated as "already attached". | REST `unix_cluster_groups.json` |
| 9 | UnixClusterAccount | Re-GET a fresh composite, merge a `UnixClusterAccount` block, then full-composite PUT. The merge round-trips unmodeled fields as `json.RawMessage` so `deleteOmitted` cannot drop attributes the connector does not understand. | Core `GET /people/<comanage_id>` then `PUT /people/<comanage_id>` |

Step 9 uses `sync_mode:"M"` (Manual) so a downstream provisioning plugin
cannot overwrite the block when Custos is the source of truth.

## Audit events

The connector emits the following `audit_events.event_type` values via the
core service. `audit_events.entity_id` is always the `compute_cluster_users.id`.

| Event type | When |
|------------|------|
| `ComanageCoPersonCreated` | A new CoPerson was POSTed in step 1. Lookups that resolved to an existing CoPerson do not fire this. |
| `ComanageClusterAccountAttached` | The sequence completed and the user has a UnixClusterAccount block on the configured UnixCluster. |
| `ComanageProvisioningFailed` | Any step returned an error. The `details` field carries `step=<name> err=<message>` so the dead-letter is tractable from the audit log alone. |

The AMIE-Processor side of this flow also emits:

| Event type | When |
|------------|------|
| `PosixUsernameTruncated` | The base username derived from the user's name was longer than the POSIX cap and got truncated. |
| `PosixUsernameUnbuildable` | First and last name both normalized to empty so no candidate could be built. |
| `PosixUsernameAllocatorExhausted` | All collision suffixes were exhausted. The cluster-user row is not created. |

## End-to-end audit walkthrough

A successful provisioning of a new user produces this sequence for a single
`compute_cluster_users.id`:

1. AMIE side: `PosixUsernameTruncated` may fire if the user's given and family
   name exceeded the POSIX cap.
2. AMIE side: the cluster-user row is committed and the event bus delivers
   `ComputeClusterUserCreateEvent`.
3. COmanage side: `ComanageCoPersonCreated` fires if step 1 went through the
   POST path. Idempotent re-runs against an existing CoPerson skip this event.
4. COmanage side: `ComanageClusterAccountAttached` fires after step 9, with
   `details` of the form `comanage_id=<id> username=<local> uid=<n>`.

A failure produces `ComanageProvisioningFailed` with
`details=step=<n> err=<msg>` in place of the success event.

## Username allocation

Local POSIX usernames are allocated by `pkg/posix`:

1. `BuildBase` derives a candidate from `Prefix() + first-initial + last-name`,
   normalised to lowercase ASCII, capped at the POSIX length limit. Truncation
   fires a `PosixUsernameTruncated` audit event. If both names normalize to
   empty, the allocator emits `PosixUsernameUnbuildable` and returns an error.
2. The AMIE handler attempts `CreateComputeClusterUser` with the base. If the
   composite UNIQUE on `(compute_cluster_id, local_username)` rejects it, the
   handler retries with `base + "2"`, `base + "3"`, and so on up to
   `MaxCollisionSuffix`.
3. If suffixes are exhausted a `PosixUsernameAllocatorExhausted` event is
   emitted and the handler returns an error.

## Local development

The connector reads config from env vars. To run it against a test registry:

```bash
export COMANAGE_REGISTRY_URL=https://<your-registry>/registry
export COMANAGE_CO_ID=<numeric CO id>
export COMANAGE_API_USER=<api user for that CO>
export COMANAGE_API_KEY=<api key for that user>
export COMANAGE_PERSON_ID_TYPE=<CoPerson Identifier Type, per your registry>
export COMANAGE_UNIX_CLUSTER_ID=<numeric UnixCluster id>
export CUSTOS_CLUSTER_ID=<UUID from the compute_clusters table>
go run ./cmd/server
```

Without those vars set the connector skips registration silently, so the rest
of Custos runs unaffected.

## Test surface

```bash
# unit
go test ./connectors/COmanage/Identity-Provisioner/...

# integration (requires a real registry; skipped by default)
go test -tags integration ./connectors/COmanage/Identity-Provisioner/...
```

The unit tests cover:

- HTTP client transport: 5xx retry with exponential backoff, error
  classification (`ErrNotFound`, `ErrAuth401`, `HTTPError`).
- Core API and REST wrappers via `httptest.NewServer`.
- The `compose.go` merge layer: a fixture composite is merged with a
  `UnixClusterAccount` block, then asserted to preserve every key from the
  original.
