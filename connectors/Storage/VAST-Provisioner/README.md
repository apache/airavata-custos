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
views, through the REST API of VMS, the VAST Management System.
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

Core defines storage policies and storages and decides when a storage is made; see [Storage](../../../docs/storage.md).
This connector acts on the storage events listed there and subscribes to nothing else. On a create it makes the folder,
its view and its quota; on an update it sets the quota; on a delete it removes the view and the quota and moves the
folder to VAST's trash.

At startup, then once per `reconcile_interval`, it has core create the cluster's missing storages, makes every active
storage again so one missing on VAST comes back, and reports each quota's use to core.

A storage with `quota_bytes` also gets a quota named after its path. A storage whose type has `view: true` in the cluster's `vast.storages`
also becomes a view, nested under the view that holds it. The view uses the cluster's `vast.view_policy` and
`vast.view_protocols`. The directory comes first, because VMS ignores owner fields on a view.

The connector leaves existing directories, views and quotas alone. A quota whose limits differ from the storage's is
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
          storages:
            - type: user
              view: false
            - type: project
              view: true
            - type: project_user
              view: true
            - type: project_shared
              view: true
```

| Key | Purpose |
|---|---|
| `reconcile_interval` | Sweep interval, a positive Go duration. Default `5m`. |
| `clusters[].id` | Custos compute cluster ID. Events for clusters not listed are ignored. |
| `clusters[].vast.url`, `username`, `password`, `tenant_id` | VMS endpoint, service account and tenant |
| `clusters[].vast.view_policy` | View policy for the storage types with `view: true` |
| `clusters[].vast.view_protocols` | Protocols for those views, comma-separated, such as `NFS,NFS4` |
| `clusters[].vast.storages[].type` | A storage type |
| `clusters[].vast.storages[].view` | `true` when VAST also exports the type's directories as views. Unlisted types are directories only. |

Unlike other connectors, a bad configuration stops the server instead of skipping the connector. Custos does not start
when:

- no cluster is listed
- a cluster lacks `id` or a `vast` setting
- a cluster still has `targets`, which are now core storage policies
- a `vast.storages` entry's `type` is not a storage type
- `reconcile_interval` is not a positive duration
- a cluster has a `vast.storages` view without `view_policy` and `view_protocols`, or VMS has no such policy

The connector looks up the policy once, at startup.

## Upgrading

Folders this connector tracked before storages moved to core carry over without a mode or a policy. Add the
cluster's storage policies and the next sweep restores both; a folder no policy defines stays as recorded.

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| Custos does not start and names the VAST provisioner | A cluster entry is incomplete or the view policy is missing in VMS | Fix the entry, or create the policy in VMS |
| Creates retry and VMS reports an unknown owner | VMS cannot resolve the user or group | Add the LDAP that holds the cluster's accounts as a VMS user-directory provider |
| Project folders lack setgid | VMS drops it | Set the bit outside the connector |
| A user's folders are missing | Core has not created them yet | See [when storages are created](../../../docs/storage.md#storages) |

## Development

[`dev-ops/local-vast`](../../../dev-ops/local-vast/README.md) stands in for the VMS REST API, for end-to-end tests
without VAST.

```bash
go test ./connectors/Storage/VAST-Provisioner/...
```
