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

The COmanage Identity-Provisioner gives each approved cluster account a POSIX identity in a COmanage Registry. Login
nodes and storage resolve users and groups from the registry's LDAP export. Enable it when a COmanage Registry is the
site's source of Unix accounts.

Custos owns the records the connector writes. Do not edit them in COmanage. The connector does not read changes back.

## How it works

The connector subscribes to core events for one cluster, `provisioning.custos_cluster_id`, and ignores the rest.

| Event | Action |
|---|---|
| `compute_cluster_user::approve` | Provision the account, join the admin group for an `ADMIN` account, then sync the user's allocation groups |
| `user::onboarded` | Provision again once the user has signed in, so the registry gets their OIDC `sub` |
| `compute_cluster_user::deactivate` | Suspend the account in the registry and leave the admin group |
| `compute_cluster_user::delete` | Remove an approved account from the registry, then sync the user's allocation groups |
| `compute_allocation::create` and `::update`; `compute_allocation_membership::create`, `::update` and `::delete` | Sync the allocation's group |
| `compute_allocation::delete` | Delete the allocation's group |

The registry matches a person to their SSH login by the OIDC `sub`. An account approved before the first sign-in has
none. The connector then writes `ComanageProvisioningDeferred` and runs again on `user::onboarded`.

The [event bus](../../../docs/architecture.md#event-bus) retries a failed delivery with backoff. Each step finds an
existing record before it creates one, so a retry is safe.

### Provisioning an account

The connector calls two registry APIs. *Core* is the Core API (`/api/co/<co_id>/core/v1`), which reads and writes a
person as one composite document. *REST* is the v1 REST API, which edits one record at a time.

| # | Step | What it does | Endpoint |
|---|------|--------------|----------|
| 1 | Find or create the CoPerson | Look up `comanage_id` in `user_identities` (`source="comanage"`), else search by email, else POST a new CoPerson. | Core `GET /people/<comanage_id>`, REST `co_people.json`, Core `POST /people` |
| 2 | Store `comanage_id` | Save the identifier in `user_identities` so step 1 finds it next time. | Core service `CreateUserIdentity` |
| 3 | Read the composite | Read `uidnumber` and `CoPerson.meta.id` from the full Core composite. | Core `GET /people/<comanage_id>` |
| 4 | Link the OIDC `sub` | Add the user's `sub` to the person as an `oidcsub` identifier, if it is missing. | Core `PUT /people/<comanage_id>` |
| 5 | Personal CoGroup | Find or create a CoGroup named after the local username (`GroupType:"CL"`, `Auto:false`). | REST `co_groups.json` |
| 6 | UnixClusterGroup | Bind the CoGroup to the configured UnixCluster. Bind first, because the registry publishes a group on identifier and member writes. | REST `unix_cluster_groups.json` |
| 7 | Group name | Attach a `type:"uid"` Identifier holding the local username to the CoGroup. | REST `identifiers.json` |
| 8 | gidnumber | Attach a `type:"gidnumber"` Identifier equal to the person's `uidnumber`. | REST `identifiers.json` |
| 9 | CoGroupMember | Add the CoPerson to the CoGroup as a member, never an owner. | REST `co_group_members.json` |
| 10 | UnixClusterAccount | Read a fresh composite, merge a `UnixClusterAccount` block and the login identifier, then PUT the full composite. Fields Custos does not model go back unchanged, so the registry keeps them. | Core `GET` then `PUT /people/<comanage_id>` |

Step 10 sets `sync_mode:"M"` (Manual), so no downstream plugin overwrites the block. The home directory is
`homedir_prefix` plus the username, and the shell is `default_shell`. The connector then marks the cluster account
provisioned, which other connectors such as the VAST-Provisioner wait for.

An `ADMIN` account also joins the CoGroup `unix_cluster.admin_group`, which grants sudo. An operator creates that group.
Without it, the join writes `ComanageProvisioningFailed` and the bus retries it. The account works meanwhile, without
sudo.

Removing an account takes the `UnixClusterAccount`, the login identifier and the admin group membership off the person,
then deletes the personal CoGroup. The CoPerson stays.

### Allocation groups

Each active allocation with a `posix_group` gets a CoGroup of that name, bound and identified as in steps 6 to 8. Its
gid is 3000000 plus the CoGroup id. It holds exactly the allocation's active members with a provisioned cluster account.
Deleting the allocation deletes the group.

### Usernames

The username comes from core onboarding or the [AMIE Processor](../../ACCESS/AMIE-Processor/README.md#usernames).

## Configuration

The connector reads the `connectors.comanage-provisioner` block of [`config/custos.yaml`](../../../config/custos.yaml).
An incomplete block is ignored whole, and every setting is read from the variables. A variable never fills a single
missing key. If neither is complete, the connector skips loading and the rest of Custos runs. It logs
`comanage provisioner: required config not set; skipping` and reports the gap on `GET /connectors`.

| YAML key | Variable | Required | Purpose |
|---|---|---|---|
| `registry.url` | `COMANAGE_REGISTRY_URL` | yes | Registry base URL, such as `https://<host>/registry` |
| `registry.co_id` | `COMANAGE_CO_ID` | yes | Numeric CO id |
| `registry.api_user` | `COMANAGE_API_USER` | yes | API user for that CO |
| `registry.api_key` | `COMANAGE_API_KEY` | yes | API key for that user |
| `unix_cluster.id` | `COMANAGE_UNIX_CLUSTER_ID` | yes | Numeric UnixCluster id |
| `unix_cluster.person_id_type` | `COMANAGE_PERSON_ID_TYPE` | yes | The CoPerson Identifier Type, as named on the registry's Identifier Types screen |
| `unix_cluster.admin_group` | `COMANAGE_CLUSTER_ADMIN_GROUP` | for `ADMIN` accounts | CoGroup that grants sudo |
| `provisioning.custos_cluster_id` | `CUSTOS_CLUSTER_ID` | yes | ID of the `compute_clusters` row this connector serves |
| `provisioning.default_shell` | `COMANAGE_DEFAULT_SHELL` | no | Login shell, default `/bin/bash` |
| `provisioning.homedir_prefix` | `COMANAGE_HOMEDIR_PREFIX` | no | Home directory prefix, default `/home/` |
| `provisioning.http_timeout` | `COMANAGE_HTTP_TIMEOUT` | no | HTTP timeout, default `30s` |

[`config.example.yaml`](config.example.yaml) lists the variables with example values.

## Operations

Each audit row the connector writes goes to the core `audit_events` table with `source=comanage`. `entity_id` is the
`compute_cluster_users.id`, or the `compute_allocations.id` for the `ComanageAllocationGroup*` events.

| Event type | When |
|------------|------|
| `ComanageProvisioningDeferred` | The user has not signed in. Provisioning waits for `user::onboarded`. |
| `ComanageProvisioningStarted` | Provisioning of an approved account began. |
| `ComanageCoPersonCreated` | Step 1 created a new CoPerson. A lookup that found one does not write this. |
| `ComanageOIDCLinkageCreated` | Step 4 added the `oidcsub` identifier. |
| `ComanageClusterAccountAttached` | Step 10 finished. `details` is `comanage_id=<id> username=<local> uid=<n>`. |
| `ComanageClusterAdminGroupJoined` | An `ADMIN` account joined the admin group. |
| `ComanageClusterAccountSuspended` | A deactivated account was suspended in the registry. |
| `ComanageClusterAccountRemoved` | A deleted account was removed from the registry. |
| `ComanageAllocationGroupCreated` | An active allocation's CoGroup was created. |
| `ComanageAllocationGroupMemberAdded` | A provisioned, active member was added to an allocation's CoGroup. |
| `ComanageAllocationGroupMemberRemoved` | A member the allocation no longer calls for was removed from its CoGroup. |
| `ComanageAllocationGroupDeleted` | A deleted allocation's CoGroup was unbound and deleted. |
| `ComanageProvisioningFailed` | A step failed. `details` is `step=<name> err=<message>`. |

A new user's trace reads, in order:

1. `ComanageProvisioningStarted`
2. `ComanageCoPersonCreated`, if the person is new
3. `ComanageOIDCLinkageCreated`, if the `sub` was missing
4. `ComanageClusterAccountAttached`

A failure writes `ComanageProvisioningFailed` in place of the last event.

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| `ComanageProvisioningDeferred` with `reason=no_oidc_identity` | The user has not signed in to Custos | None. Provisioning runs after the first sign-in |
| `ComanageProvisioningFailed` with `step=missing_uidnumber` | The registry has not assigned a `uidnumber` | Check the registry's identifier assignment for `uidnumber` |
| `ComanageProvisioningFailed` with `step=cluster_admin_group_unset` or `cluster_admin_group_missing` | No `admin_group` configured, or no CoGroup by that name | Set `unix_cluster.admin_group` and create the group in the registry |
| No audit rows at all | The connector skipped loading | Check the log for `required config not set` and `GET /connectors` |
| A user cannot log in after a registry edit | The record was changed in COmanage | Revert it. Custos does not read registry changes back |

## Development

Export the variables and start the server. [INSTALL.md](../../../INSTALL.md) covers the rest.

```bash
export COMANAGE_REGISTRY_URL=https://<registry-host>/registry
export COMANAGE_CO_ID=<numeric CO id>
export COMANAGE_API_USER=<api user for that CO>
export COMANAGE_API_KEY=<api key for that user>
export COMANAGE_PERSON_ID_TYPE=<CoPerson Identifier Type, as named in the registry>
export COMANAGE_UNIX_CLUSTER_ID=<numeric UnixCluster id>
export CUSTOS_CLUSTER_ID=<UUID from the compute_clusters table>
go run ./cmd/server
```

The unit tests need no registry:

```bash
go test ./connectors/COmanage/Identity-Provisioner/...
```

They cover:

- the HTTP client: 5xx retry with backoff, and the `ErrNotFound`, `ErrAuth401` and `HTTPError` classes
- the Core and REST wrappers, against `httptest.NewServer`
- the composite merge in `compose.go`
- the subscribers and orchestrator, against fakes
