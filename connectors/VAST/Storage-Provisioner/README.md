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

# VAST Storage-Provisioner

Creates directories and quotas on a VAST cluster through the VMS REST API, once the COmanage Identity-Provisioner
on the same Custos cluster has provisioned the account. Nothing is changed once it exists, and nothing is deleted.

The directories come from the `storage` list in the config. Each entry's `path`, `owner` and `group` are templates
over `{user}`, the account's username. Entries are created in config order, with a quota named after the path.

| Trigger | Runs |
|---|---|
| `compute_cluster_user::update` with `provisioned_at` set | The account's directories |
| Startup, then every 24 hours | The same, for every account on the cluster |

Configuration is in [`CONFIG.md`](../../../CONFIG.md#vast-storage-provisioner).

## Prerequisites on VMS

VAST admins make these changes before the connector is enabled.

| Prerequisite | Why |
|---|---|
| The COmanage LDAP is a VMS user-directory provider | VMS resolves owners only from its providers; Custos users are otherwise unknown and every create retries |
| The service account may create folders and quotas | The account the connector logs in with |
