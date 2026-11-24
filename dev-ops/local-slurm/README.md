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

# local-slurm

A SLURM 24.05 cluster in Docker with accounting and `slurmrestd`.
Use it to develop the SLURM connectors against a real REST API.
The cluster `artisan` has one `compute` partition on nodes `c1` and `c2`, each with 4 CPUs and 2 GPUs.

| Service | Purpose | Host port |
|---|---|---|
| `mariadb`, `slurmdbd` | Accounting database and daemon | |
| `slurmctld` | Controller | |
| `c1`, `c2` | Compute nodes | |
| `login` | Login node with `testuser`, `testuser2` and `testuser3` (UIDs 1000 to 1002) | `2222` (`LOGIN_SSH_PORT`) |
| `slurmrestd` | REST API with JWT auth | `6820` (`REST_PORT`) |

## Usage

```bash
make up            # build the base image, then start the cluster
make token         # print a REST API token for root
make down          # stop the cluster and drop its volumes
```

| Target | Effect |
|---|---|
| `make build` | Build the images without starting them |
| `make keys` | Generate the munge and JWT keys |
| `make logs` | Follow the logs |
| `make login` | Open a shell on the login node |
| `make smoke` | Submit a `hostname` job and show its accounting record |
| `make token2`, `token3`, `token4` | Print a token for `testuser`, `testuser2` or `testuser3` |
| `make associations` | List associations with their `GrpTRES`, `GrpTRESMins` and QoS |
| `make allocations` | List jobs from the last five minutes |

Point the [SLURM Association-Mapper](../../connectors/SLURM/Association-Mapper/README.md) at it:

```bash
export SLURM_API='http://localhost:6820'
export SLURM_USER='root'
export SLURM_TOKEN="$(make -s token | awk -F= '{print $2}')"
export SLURM_API_VERSION='40'
```
