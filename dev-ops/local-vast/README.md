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

# local-vast

A stand-in for the VAST VMS REST API, run on a test cluster's file server so the VAST-Provisioner can be
tested end to end: Custos → COmanage → LDAP → directories users see on the cluster.

| VMS call | Stand-in |
|---|---|
| `POST /api/token/` | A token for `VAST_USER` / `VAST_PASSWORD`; anything else is `403` |
| `POST /api/folders/create_folder/` | Creates the directory, owner and group resolved through NSS, mode applied exactly, setgid included (VMS drops it); `503` if it exists or an owner is unknown |
| `POST /api/folders/stat_path/` | 200 if the path exists, else `503` |
| `POST /api/quotas/` | Logged; not enforced |

## Install

```sh
GOOS=linux GOARCH=amd64 go build -o local-vast ./dev-ops/local-vast
scp local-vast local-vast.service <host>:
ssh <host> sudo install -m 755 local-vast /usr/local/bin/
ssh <host> sudo install -m 644 local-vast.service /etc/systemd/system/
# /etc/local-vast.env, mode 600: VAST_USER=..., VAST_PASSWORD=..., LISTEN=<private ip>:8080
ssh <host> sudo systemctl enable --now local-vast
```

The host needs NSS to resolve COmanage's groups as well as its people (for SSSD, `ldap_group_search_base`), the
parent directories of the configured paths, and the port open to the Custos host only. Point the connector at it:

```yaml
        vast:
          url: "http://<private ip>:8080"
          username: "<VAST_USER>"
          password: "${VAST_API_PASSWORD}"
          tenant_id: 1
```
