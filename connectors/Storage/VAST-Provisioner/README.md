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

# VAST-Provisioner

Creates directories and quotas on each configured cluster's VAST through the VMS REST API, once the COmanage
Identity-Provisioner has provisioned the account or the allocation's group. Nothing is changed once it exists, and nothing is deleted.

The directories come from each cluster's `targets` list in the config. Each entry's `path`, `owner` and `group` are
templates over `{user}`, the account's username, and `{allocation}`, the allocation's group; the ones it uses decide
what it waits for. Entries are created in config order, with a quota named after the path when `hard_limit_size` is
set. Events for clusters not listed are ignored.

| Trigger | Runs |
|---|---|
| `compute_cluster_user::update` with `provisioned_at` set | The account's directories, then its project directories |
| `compute_allocation::create` and `::update` | The project's directories, then each member's; retried until LDAP has the group |
| `compute_allocation_membership::create` and `::update` | The same, for the membership's allocation |
| Startup, then every 24 hours | The same, for every account and allocation on each cluster |

Configuration is in [`CONFIG.md`](../../../CONFIG.md#storage-vast-provisioner).

## Prerequisites on VMS

VAST admins make these changes before the connector is enabled.

| Prerequisite | Why |
|---|---|
| The COmanage LDAP is a VMS user-directory provider | VMS resolves owners only from its providers; Custos users and `proj-*` groups are otherwise unknown and every create retries |
| The service account may create folders and quotas | The account the connector logs in with |

VMS drops setgid from `create_dir_mode` (a `2770` request yields `0770`), so folders configured with setgid, such as
`/project/{allocation}`, `shared` and `/project/{allocation}/{user}`, come out without it and files in them do not
inherit the allocation's group until the bit is set outside the connector.
