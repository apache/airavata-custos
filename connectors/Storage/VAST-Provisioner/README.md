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

The VAST-Provisioner creates each user's and project's directories on a cluster's VAST storage, with quotas and NFS
views, through the REST API of VMS, the VAST Management System. It acts on core's storage requests, sent once the account or allocation
group is provisioned, so VAST can resolve the owners. With COmanage, the COmanage Identity-Provisioner provisions them.
With no identity connector enabled, core marks both provisioned itself, and VMS must resolve the owners from another
provider.
Enable it when a cluster's home, scratch or project space lives on VAST.

## Prerequisites on VMS

A VAST administrator sets these before you enable the connector.

| Prerequisite | Why |
|---|---|
| The LDAP that holds the cluster's accounts and allocation groups is a VMS user-directory provider | VMS resolves owners only from its providers. Otherwise owners are unknown and each create retries. |
| The service account may create folders, views and quotas | The connector logs in with it |
| The view policy in `vast.view_policy` exists | Each view the connector creates uses it |
| The service account may delete quotas, views and folders | A quota with other limits is recreated. An approved deletion removes views and quotas. |
| The trash folder is enabled | An approved deletion moves the folder to the trash |

VMS drops setgid from a new directory's mode: a `2770` request yields `0770`. Set the bit outside the connector on each
folder that needs it. Otherwise files in it do not inherit the allocation's group.

## How it works

Each cluster has a `targets` list of directories, set under [Configuration](#configuration). An entry's `path`,
`owner` and `group` are templates over two placeholders:

| Placeholder | Value |
|---|---|
| `{user}` | The account's username |
| `{allocation}` | The allocation's group |

The placeholders decide which event creates the entry:

| Trigger | Runs |
|---|---|
| `compute_cluster_user::storage_request` | The account's directories, then marks storage done, even on an unlisted cluster |
| `compute_allocation::group_provision` | The project's directories, once the group's first member puts it on the cluster |
| `compute_allocation_membership::storage_request` | That member's project directory, once the member is ready and the group provisioned |
| `compute_cluster_user::delete` | Marks the account's directories for deletion, including its directory in each project |
| `compute_cluster_user::deactivate` | Marks the account's directories for deletion |
| `compute_allocation::delete` | Marks the project's directories, and its members' directories inside them, for deletion |
| Startup, then once per `reconcile_interval` | Everything, for each provisioned account and each allocation on each cluster |

Each event creates only its own entries, so a failed delivery retries alone.

An entry with `hard_limit_size` also gets a quota named after its path. An entry with `view: true` also becomes a view,
nested under the view that holds it. The view uses the cluster's `vast.view_policy` and `vast.view_protocols`. The
directory comes first, because VMS ignores owner fields on a view.

The connector leaves existing directories, views and quotas alone. A quota whose limits differ from the configuration is
recreated.

## Configuration

Settings are in the `connectors.vast-provisioner` block of [`config/custos.yaml`](../../../config/custos.yaml):

```yaml
  vast-provisioner:
    type: "vast-storage-provisioner"
    enabled: true
    reconcile_interval: "5m"
    clusters:
      - id: "${CUSTOS_CLUSTER_ID}"
        vast:
          url: "${VAST_VMS_URL}"
          username: "${VAST_API_USER}"
          password: "${VAST_API_PASSWORD}"
          tenant_id: ${VAST_TENANT_ID}
          view_policy: "${VAST_VIEW_POLICY}"
          view_protocols: "${VAST_VIEW_PROTOCOLS}"   # comma-separated, such as NFS,NFS4
        targets:
          - path: "/home/{user}"
            owner: "{user}"
            group: "{user}"
            mode: "0700"
            hard_limit_size: 107374182400      # 100 GiB
            hard_limit_inodes: 1000000
          - path: "/project/{allocation}/{user}"
            owner: "{user}"
            group: "{allocation}"
            mode: "2700"
            view: true
```

| Key | Purpose |
|---|---|
| `reconcile_interval` | Sweep interval, a positive Go duration. Default `5m`. |
| `clusters[].id` | Custos compute cluster ID. Events for clusters not listed are ignored. |
| `clusters[].vast.url`, `username`, `password`, `tenant_id` | VMS endpoint, service account and tenant |
| `clusters[].vast.view_policy` | View policy for targets with `view: true` |
| `clusters[].vast.view_protocols` | Protocols for those views, comma-separated, such as `NFS,NFS4` |
| `clusters[].targets[].path`, `owner`, `group` | Templates over `{user}` and `{allocation}`. Targets are created in list order. |
| `clusters[].targets[].mode` | Octal permission bits, such as `2770` |
| `clusters[].targets[].hard_limit_size`, `hard_limit_inodes` | Optional quota, in bytes and inodes. A quota is created when `hard_limit_size` is set. |
| `clusters[].targets[].view` | `true` also exports the directory as a view |

Unlike other connectors, a bad configuration stops the server instead of skipping the connector. Custos does not start
when:

- no cluster is listed
- a cluster lacks `id`, a `vast` setting or `targets`
- a `mode` is not octal
- `reconcile_interval` is not a positive duration
- a cluster has `view: true` targets without `view_policy` and `view_protocols`, or VMS has no such policy

The connector looks up the policy once, at startup.

## Operations

Deleting an account or allocation marks its folders `pending_deletion` for an administrator to review. It deletes
nothing.

| Route | Privilege | Purpose |
|---|---|---|
| `GET /connectors/vast/folders` | `core:clusters:read` | List folders with each one's quota and use |
| `DELETE /connectors/vast/folders/{id}` | `core:clusters:write` | Remove a folder's views and quotas and move it to VAST's trash, once every folder under it is pending deletion |
| `PUT /connectors/vast/folders/{id}/quota` | `core:clusters:write` | Set a temporary hard limit with `hard_limit`, `until` (at most 90 days ahead) and `reason` |

A temporary limit is audited under `<cluster>:<path>`, so the record outlives the folder. After `until`, the sweep
restores the configured limit.

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| Custos does not start and names the VAST provisioner | A cluster entry is incomplete or the view policy is missing in VMS | Fix the entry, or create the policy in VMS |
| Creates retry and VMS reports an unknown owner | VMS cannot resolve the user or group | Add the LDAP that holds the cluster's accounts as a VMS user-directory provider |
| Project folders lack setgid | VMS drops it | Set the bit outside the connector |
| A user's folders are missing | Their account is not provisioned yet | They appear on provisioning, or at the next sweep |

## Development

[`dev-ops/local-vast`](../../../dev-ops/local-vast/README.md) stands in for the VMS REST API, for end-to-end tests
without VAST.

The folder registry's integration test needs `CORE_TEST_DATABASE_DSN` ([CONTRIBUTING.md](../../../CONTRIBUTING.md)) and
skips without it.

```bash
go test ./connectors/Storage/VAST-Provisioner/...
```
