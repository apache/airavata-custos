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

# Storage

A **storage** is a directory core decides a holder gets on a cluster. A storage connector makes it on a storage system;
the [VAST-Provisioner](../connectors/Storage/VAST-Provisioner/README.md) is the only one so far.

## Targets

Each cluster's `targets`, under `core.storage.clusters` in [`config/custos.yaml`](../config/custos.yaml), say how its
storages are made. Each key is a storage type, such as `user_home`. The path's placeholders say whose storage it is:

| `path` names | One per |
|---|---|
| `{user}` | Provisioned account |
| `{allocation}` | Active allocation with a provisioned group |
| `{user}` and `{allocation}` | Active membership of a provisioned account |
| `{name}` alone | User a connector holds; see [Connector holders](#connector-holders) |

Custos does not start when a path names none of these, names `{name}` with another, or repeats another target's path,
when `owner` or `group` names a placeholder the path lacks, or when a `mode` is not octal.

| Key | Purpose |
|---|---|
| `path`, `owner`, `group` | Templates over `{user}`, the account's username, `{allocation}`, the allocation's group, and `{name}`, the holder's name |
| `mode` | Octal permission bits, such as `2770` |
| `hard_limit_size`, `hard_limit_inodes` | Optional quota, in bytes and inodes, when `hard_limit_size` is set |

## Storages

Core makes or changes storages and publishes `storage::create` for each, a parent before what is under it:

| When | Storages |
|---|---|
| An account is provisioned | Its `{user}` storages, and its member storages in each of its projects |
| An allocation's group is provisioned | Its `{allocation}` storages, then each provisioned member's |
| A provisioned account's membership is created or updated | That member's storages |
| A user is created or changed | Their `{name}` storages, when a connector holds them |
| A storage connector's sweep | Every one the targets give the cluster's holders that is missing or differs |

| When | Then |
|---|---|
| A user leaves `ACTIVE` | Deactivates their accounts' `{user}` and member storages |
| An allocation or membership is set `INACTIVE` | Deactivates the storages it gave |
| An unprovisioned account, or an `INACTIVE` allocation, is deleted | Its deactivated storages stay `pending_deletion`, unlinked |
| An `INACTIVE` membership is deleted | Its deactivated storages stay `pending_deletion` |
| A user is deleted, or a connector stops holding them | Deactivates their `{name}` storages |

Making storages also deactivates those of the same holders the targets no longer give, so a sweep deactivates every storage
left without a holder. Retiring marks storages `pending_deletion` for an administrator to review, and a provision or
sweep activates them.

| Route | Privilege | Purpose |
|---|---|---|
| `GET /storages` | `core:clusters:read` | List storages with each one's quota and use, read from the storage connector |
| `DELETE /storages/{id}` | `core:clusters:write` | Remove a pending storage and those under it, once all of them are pending |
| `PUT /storages/{id}` | `core:clusters:write` | Update a storage: set its temporary quota with `quota_bytes`, `quota_until` (at most 90 days ahead) and `reason` |

A temporary quota is audited under `<cluster>:<path>`, so the record outlives the storage. After `quota_until`, the
next sweep restores the target's quota. A storage without a quota of its own counts against the one above it.

## Connector holders

A connector holds users for a `{name}` type with `ClaimStorageHolders`. It says which users hold a storage and the
name each fills in; core makes, deactivates and sweeps them like the rest. Until a connector claims the type, core leaves
its storages as they are. The [temp account connector](../connectors/TempAccount) holds `trial_home` for each pending
or active trial user, named by their email's local part. Two users with the same name share no storage: the first keeps
it. A name that is empty, `.`, `..` or holds `/` gives no storage.

## Events

Core publishes these and subscribes to none; a storage connector acts on them.

| Event | Asks the connector to |
|---|---|
| `storage::create` | Make the storage |
| `storage::update` | Bring the storage to its new state; today, its quota |
| `storage::delete` | Remove the storage with what is under it |

A connector applies events after core commits. A delete that fails every retry leaves the directory on the storage
system until `POST /events/deliveries/{id}/retry` delivers it again.
