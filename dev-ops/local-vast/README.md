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
| `POST`, `GET /api/quotas/`, `DELETE /api/quotas/{id}/` | Recorded in memory, not enforced; a second quota by the same name is `503` |
| `POST /api/views/` | `exportfs -o $EXPORT_OPTIONS $EXPORT_TO:<path>` |
| `GET /api/views/` | `exportfs -s`, each export as a view whose id is its directory's inode |
| `DELETE /api/views/{id}/` | `exportfs -u` of that export; the directory stays |
| `GET /api/viewpolicies/` | A policy with id 1 for any name |

## Install

```sh
GOOS=linux GOARCH=amd64 go build -o local-vast ./dev-ops/local-vast
scp local-vast local-vast.service <host>:
ssh <host> sudo install -m 755 local-vast /usr/local/bin/
ssh <host> sudo install -m 644 local-vast.service /etc/systemd/system/
# /etc/local-vast.env, mode 600: VAST_USER=..., VAST_PASSWORD=..., LISTEN=<private ip>:8080
#   and EXPORT_TO=<cluster subnet> for the host's NFS server, optionally EXPORT_OPTIONS (default rw,sync,no_subtree_check,root_squash)
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

## vms

A POSIX shell command line for the VMS REST API, real or this stand-in, needing `ssh`, `curl` and `jq`. `vms` alone
opens a prompt.

### Install

```sh
mkdir -p ~/.local/bin
curl -fsSL https://raw.githubusercontent.com/apache/airavata-custos/master/dev-ops/local-vast/vms -o ~/.local/bin/vms
chmod +x ~/.local/bin/vms
```

### Configure

```console
$ vms config
VMS URL []: https://vms.example.org
ssh arguments reaching a host that reaches VMS, - for none []: -J me@login.example.org me@admin.example.org
username (empty keeps the saved one): admin
password:
vms: tunnel through -J me@login.example.org me@admin.example.org on 127.0.0.1:1080
logged in to https://vms.example.org; settings in ~/.config/vms
```

With an ssh route, the first command opens a SOCKS tunnel that later commands reuse until `vms disconnect`;
`vms browse` opens the VMS web interface through it.

| File in `~/.config/vms` | Holds |
|---|---|
| `vms.conf` | `VMS` and `VMS_SSH` as shell assignments |
| `login.json` | `{"username": ..., "password": ...}`, owner-only |
| `tunnel.sock` | Control socket of the open tunnel |

| Variable | Effect |
|---|---|
| `VMS_CONFIG` | Settings directory, for one per cluster |
| `VMS_TOKEN` | Bearer token used instead of logging in |
| `VMS_PROXY_PORT` | Local SOCKS port, default 1080 |
| `VMS_BROWSER` | Browser `vms browse` starts |

### Commands

`vms help` lists them. `RESOURCE` is a path under `/api/`, such as `views/42`, documented at `$VMS/docs`. `get`
arguments become the query string; the other methods send a JSON body of `key=value` strings and `key:=json` values,
or stdin with `-`. Deletes ask first at a terminal; any failure exits 1.

```sh
vms info /project/foo
vms mkdir /project/foo 2000008 2000008 2770 && vms quota /project/foo 10995116277760 && vms mkview /project/foo
vms get users/query username=alice
vms patch clusters/3 enable_trash:=true
vms post views - < view.json
```
