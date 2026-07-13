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

# Workflows

Files are named `<area>-<purpose>.yml`, where the area matches the part of
the repo the workflow validates:

| Area prefix | Covers |
|---|---|
| `core-` | Go backend: `cmd/`, `internal/`, `pkg/`, and unit tests repo-wide |
| `web-` | the portal under `web/` |
| `amie-` | the AMIE connector and its integration stack |

Current workflows:

- `core-verify.yml` — formatting, build, vet, unit tests for all Go packages,
  and the `./internal/...` integration suite against a MariaDB service
  container. Runs on Go changes.
- `web-verify.yml` — frozen-lockfile install, typecheck, lint, unit tests,
  and a production build. Runs on `web/` changes.
- `amie-integration-tests.yml` — the AMIE integration suite with its own
  compose stack (mock AMIE server + MariaDB). Runs on connector changes.
  Not duplicated by `core-verify.yml`.

When adding a workflow, keep the area prefix, add path filters so unrelated
PRs skip it, and list it here.
