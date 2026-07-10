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

# SLURM Association-Mapper

SLURM association creation logic lives in this plugin. It is triggered when the allocation manager has processed an allocation request and released it to downstream handlers. It talks to `slurmrestd` to manage accounts, associations, and TRES limits.

This package is part of the root `github.com/apache/airavata-custos` module.

## Prerequisites

- Go **1.24+**
- A reachable `slurmrestd` endpoint (for integration runs) plus a SLURM user name and JWT token

## Layout

```
.
├── internal/subscribers/   # event subscribers wiring core events to slurmrestd
└── pkg/smapper/            # connector loader
```

The slurmrestd client lives in the sibling Rest-Client module at `connectors/SLURM/Rest-Client/pkg/client`.

Import path: `github.com/apache/airavata-custos/connectors/SLURM/Rest-Client/pkg/client`.

## Configuration

The connector reads its `slurmrestd` connection details from the following environment variables:

| Variable      | Description                                                                 |
|---------------|-----------------------------------------------------------------------------|
| `SLURM_API`   | Base URL of the `slurmrestd` endpoint, e.g. `https://slurm.example.org:6820` |
| `SLURM_USER`  | SLURM user name to authenticate as (sent in the `X-SLURM-USER-NAME` header)  |
| `SLURM_TOKEN` | JWT token for that user (sent in the `X-SLURM-USER-TOKEN` header)            |
| `SLURM_API_VERSION` | API version of `slurmrestd`. You can check it from `slurmrestd -d list` | 

Example:

```bash
export SLURM_API='https://slurm.example.org:6820'
export SLURM_USER='slurm'
export SLURM_TOKEN="$(scontrol token lifespan=3600 | awk -F= '{print $2}')"
export SLURM_API_VERSION='40'
```

All three are required for any live `slurmrestd` interaction. Tests do not need them.

You can check the functionality of the SLURM REST API through 
```bash
curl -sS \
  -H "X-SLURM-USER-NAME: $SLURM_USER" \
  -H "X-SLURM-USER-TOKEN: $SLURM_TOKEN" \
  "$SLURM_API/slurm/v0.0.$SLURM_API_VERSION/ping" | jq
```

## Test

```bash
# from the repository root
go test ./connectors/SLURM/Association-Mapper/...
go vet ./connectors/SLURM/Association-Mapper/...
```

Tests are hermetic and use `httptest` — no live `slurmrestd` required.
