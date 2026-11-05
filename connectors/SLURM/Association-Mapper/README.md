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

# SLURM Association-Mapper

The SLURM Association-Mapper turns Custos allocations into SLURM accounts and associations through `slurmrestd`. Each
allocation becomes an account, and each active member gets an association on it to charge jobs. Removed or inactive
members lose theirs. Enable it on a SLURM cluster whose accounts Custos manages.

## How it works

Events apply each change as it happens. A periodic sweep applies any change an event missed or applied too early.

| Event | Action |
|---|---|
| `compute_allocation::scheduler_request` | Create the SLURM account, named after the allocation in lowercase, under its organization, then mark it done so members can become ready |
| `compute_allocation::update` | Remove all members' associations when the allocation leaves `ACTIVE`, restore them when it returns |
| `compute_allocation::delete` | Remove all members' associations. The account stays, so its job history remains attributable. |
| `compute_allocation_resource_mapping::create` | Set the account's `GrpTRES` and `GrpTRESMins` from the mapped resource amount and time |
| `compute_allocation_membership::ready` | Write the member's association |
| `compute_allocation_membership::deactivate` | Remove the member's association on that allocation only |
| `compute_allocation_membership_resource_override::create` | Rewrite the member's association with limits from all of their overrides |

An association uses QoS `normal` and no partition, so the cluster's partition settings decide where jobs run. Overrides
on two GPU resources become one GPU limit, their sum. CPU and GPU limits stay separate.

Before writing an association, the connector makes `none` the user's default account. `none` runs no jobs, so each job
must name the allocation it charges.

Core publishes `ready` only once the member's account is provisioned, by an identity connector or, with none enabled,
by core on approval. The cluster's accounts must then exist by other means.

The sweep waits `association_provision_grace` after provisioning, until the account reaches the directory and the name
cache. An association written earlier makes `slurmctld` cache a failed uid lookup and reject the user's jobs.

The sweep runs at startup, then every `association_reconcile_interval`. Per cluster, it compares SLURM's associations
with those the allocations call for. It writes what is missing or drifted and removes stale member associations. It
skips associations on accounts Custos does not manage and the account-level records that carry an allocation's limits.
An empty desired set usually means a failed lookup, so the sweep removes nothing and logs a warning.

## Configuration

Settings are in the `connectors.slurm-mapper` block of [`config/custos.yaml`](../../../config/custos.yaml).

| YAML key | Variable | Purpose |
|---|---|---|
| `slurm_api.url` | `SLURM_API` | `slurmrestd` base URL, such as `https://slurm.example.org:6820` |
| `slurm_api.username` | `SLURM_USER` | SLURM user, sent as `X-SLURM-USER-NAME` |
| `slurm_api.token` | `SLURM_TOKEN` | JWT for that user, sent as `X-SLURM-USER-TOKEN` |
| `slurm_api.version` | `SLURM_API_VERSION` | Minor version only, such as `40` for `v0.0.40`. `slurmrestd -d list` shows supported versions. |
| `association_reconcile_interval` | | Sweep interval, default `5m`. It bounds the wait after a lost event. |
| `association_provision_grace` | | How long the sweep holds back a just-provisioned account, default `30s` |

A missing or empty YAML value falls back to the variable in the second column. If any `slurm_api` setting is still
empty, the connector skips loading. It logs
`SLURM API credentials not fully provided, skipping SLURM Association Mapper connector` and reports the gap on
`GET /connectors`.

Check the endpoint and token with a ping:

```bash
export SLURM_API='https://slurm.example.org:6820'
export SLURM_USER='slurm'
export SLURM_TOKEN="$(scontrol token lifespan=3600 | awk -F= '{print $2}')"
export SLURM_API_VERSION='40'

curl -sS \
  -H "X-SLURM-USER-NAME: $SLURM_USER" \
  -H "X-SLURM-USER-TOKEN: $SLURM_TOKEN" \
  "$SLURM_API/slurm/v0.0.$SLURM_API_VERSION/ping" | jq
```

## Operations

The event handlers write audit rows with `source=slurm`. The sweep writes none.

| Entity | Event types |
|---|---|
| `compute_allocation` | `ComputeAllocationCreationSucceeded`, `ComputeAllocationCreationFailed`, `ComputeAllocationProvisioned`, `ComputeAllocationDeprovisioned`, `ComputeAllocationUpdateFailed`, `ComputeAllocationDeletionFailed` |
| `compute_allocation_resource_mapping` | `ComputeAllocationResourceMappingCreationSucceeded`, `ComputeAllocationResourceMappingCreationFailed` |
| `compute_allocation_membership` | `ComputeAllocationMembershipCreationSucceeded`, `ComputeAllocationMembershipCreationFailed`, `ComputeAllocationMembershipDeprovisioned`, `ComputeAllocationMembershipUpdateFailed` |
| `compute_allocation_membership_resource_override` | `ComputeAllocationMembershipResourceOverrideCreationSucceeded`, `ComputeAllocationMembershipResourceOverrideCreationFailed` |

The [event bus](../../../docs/architecture.md#event-bus) retries failed handlers. A resource with an empty
`resource_type` fails at once, because a retry cannot fix it.

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| A new member gets `Invalid account` right after being added | The association was written before the cluster could resolve the user | Raise `association_provision_grace` |
| A member waits minutes for access | The sweep held the association back for `association_provision_grace` | Expected. Lower `association_reconcile_interval` to shorten the wait |
| All calls return 401 | The JWT expired | Issue a new token with `scontrol token` |

## Development

The `slurmrestd` client is `connectors/SLURM/Rest-Client/pkg/client`. [`dev-ops/local-slurm`](../../../dev-ops/local-slurm/README.md)
runs a local SLURM 24.05 cluster.

The tests use `httptest` and need no `slurmrestd`:

```bash
# From the repository root
go test ./connectors/SLURM/Association-Mapper/...
go vet ./connectors/SLURM/Association-Mapper/...
```
