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

# local-amie

Isolated Docker stack for AMIE connector integration tests. Runs on offset
ports so it coexists with the main dev stack in `dev-ops/compose/`.

## Services

| Service     | Container               | Host port |
|-------------|-------------------------|-----------|
| MariaDB     | `custos_amie_test_db`   | `3307`    |
| mock-amie   | `custos_amie_test_mock` | `8181`    |

## Usage

```bash
make up      # build + start, blocks until both are healthy
make down    # stop + remove + drop volumes
make logs    # tail both services
make ps      # status
```

The integration test runner at `scripts/run-amie-integration-tests.sh`
handles up/down for you.
