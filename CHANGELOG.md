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

# Changelog

## Unreleased

### Upgrade notes

Storages move from the VAST connector into core. A deployment that ran the VAST connector moves its folders into
`storages` once, after Custos has started and created that table:

```sql
INSERT INTO storages (id, compute_cluster_id, compute_cluster_user_id, allocation_id, path, owner, "group", mode,
    quota_bytes, quota_until, status)
SELECT f.id, f.compute_cluster_id, cu.id, a.id, f.path, split_part(f.unix_owner, ':', 1), split_part(f.unix_owner, ':', 2), '',
    COALESCE(f.quota_override, 0), f.quota_until, f.status
FROM vast_folders f
LEFT JOIN compute_cluster_users cu ON f.owner_kind = 'cluster_user' AND cu.compute_cluster_id = f.compute_cluster_id
    AND cu.local_username = f.owner_name
LEFT JOIN compute_allocations a ON a.id = f.allocation_id
ON CONFLICT (compute_cluster_id, path) DO UPDATE SET quota_bytes = EXCLUDED.quota_bytes, quota_until = EXCLUDED.quota_until
    WHERE EXCLUDED.quota_until > now();
DROP TABLE vast_folders, schema_migrations_vast;
```
