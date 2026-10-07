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

# Unix-Provisioner

Creates directories over SSH on an admin host that mounts the cluster's storage, an alternative to the
[VAST-Provisioner](../VAST-Provisioner/README.md) for storage without a management API. It creates no quotas. Nothing
is changed once it exists, and nothing is deleted.

Each cluster's `targets` are templates over `{user}`, the account's username, and `{allocation}`, the allocation's
group; the ones a target uses decide what it waits for. Targets are created in config order, all of a trigger's in one
SSH session. A directory whose owner the host cannot resolve yet is removed again, and the bus retries it. Events for
clusters not listed are ignored.

| Trigger | Creates |
|---|---|
| `compute_cluster_user::update` with `provisioned_at` set | The account's directories, then its allocations' |
| `compute_allocation::create` and `::update` | The allocation's directories, then each active, provisioned member's |
| `compute_allocation_membership::create` and `::update` | The same, for the membership's allocation |
| Startup, then every 24 hours | The same, for every account and allocation on each cluster, at most an hour per cluster |

Configuration is in [`CONFIG.md`](../../../CONFIG.md#storage-unix-provisioner).

## Admin host

| Prerequisite | Why |
|---|---|
| The storage is mounted at the configured paths, with root squash off | `mkdir`, `chown` and `chmod` run as the SSH user |
| NSS resolves the Custos users and groups | Owners are given by name |
| The Custos host's key is in the SSH user's `authorized_keys`, limited with `from="<custos host>"`, and SSH is open to that host only | The connector logs in non-interactively (`BatchMode`) |
| The admin host's key is in the Custos service user's `known_hosts` | Host keys are not accepted on first use |
