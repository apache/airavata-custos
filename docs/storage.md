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

Core defines two entities. A **storage policy** says how a cluster's storages of one type are made. A **storage** is
one made from a policy for an owner. A storage connector makes storages on a storage system; the
[VAST-Provisioner](../connectors/Storage/VAST-Provisioner/README.md) is the only one so far.

## Storages

Core makes a storage from each policy of a type for a user, a project, or both, and publishes `storage::create`:

| When | Storages |
|---|---|
| An account is provisioned | The account's own and its own in each of its projects |
| An allocation's group is provisioned, or the allocation is updated once it is | The project's, then each provisioned member's |
| A provisioned account's membership is created or updated | That member's in the project |
| `POST /storages?cluster_id=`, and a storage connector's sweep | Every one the cluster's policies give its owners that is missing or differs |

Deleting an account or allocation marks its storages `pending_deletion` for an administrator to review. It deletes
nothing. `DELETE /storages/{id}` removes a pending storage and those under it once all of them are pending.

`PUT /storages/{id}` sets an administrator's quota in place of the policy's. It is audited under `<cluster>:<path>`, so
the record outlives the storage. After `quota_until`, the next sweep or `POST /storages` restores the policy's
quota. A storage without a quota of its own counts against the one above it and takes none.

## Storage policies

An administrator adds a policy with `POST /storage-policies`:

```json
{"cluster_id": "<cluster id>", "type": "user", "path_pattern": "/home/{user}", "owner": "{user}", "group": "{user}",
 "mode": "0700", "quota_bytes": 107374182400, "quota_inodes": 1000000}
```

`path_pattern`, `owner` and `group` are templates over `{user}`, the account's username, and `{allocation}`, the
allocation's group. `mode` is the octal permission bits. A zero quota means none, and `quota_inodes` needs
`quota_bytes`. The type fixes which placeholders `path_pattern` uses, and `owner` and `group` may use no others:

| Type | `path_pattern` uses |
|---|---|
| `user` | `{user}` |
| `project`, `project_shared` | `{allocation}` |
| `project_user` | both |

A cluster may hold several policies of one type, such as home and scratch. `PUT /storage-policies/{id}` replaces a
policy's templates, mode and quotas; its cluster and type stay. A storage takes the change on its next create; a new
`path_pattern` makes new storages and leaves the old ones.
Deleting a policy removes no storage.

## Events

Core publishes these and subscribes to none; a storage connector acts on them and reports each storage's use back.

| Event | Asks the connector to |
|---|---|
| `storage::create` | Make the storage, a parent before what is under it |
| `storage::update` | Bring the storage to its new state; today, its quota |
| `storage::delete` | Remove the storage; core sends one per storage, deepest first |
