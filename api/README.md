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

# API specs

Auto-generated OpenAPI specs for Apache Custos.

## Layout

One spec per code-ownership unit:

| Spec | Path | Scope |
|---|---|---|
| Core | `api/core.openapi.yaml` | Core REST endpoints. |
| Connector | `connectors/<NAME>/api/<name>.openapi.yaml` | A connector's endpoints under `/connectors/<name>/`. |

Connector specs ship alongside the connector code; the core spec never
declares connector endpoints. The connectors with specs today are AMIE at
`connectors/ACCESS/AMIE-Processor/api/amie.openapi.yaml` and Analytics at
`connectors/Analytics/api/analytics.openapi.yaml`.

## Generating

Specs are produced from [`swag`](https://github.com/swaggo/swag)
annotations on each handler in the Go source. Regenerate after editing
any annotation:

```
make gen-api
```

The `swag` version is pinned as a `tool` in `go.mod`; `go generate` invokes it through
`go run`, so contributors do not need to install the binary separately.

To check that the committed specs are current, run:

```
make verify-no-drift
```

It regenerates the specs and fails if any committed `*.openapi.yaml` file changed. No CI workflow runs it. Run it
before pushing.

## Consuming

These files are standard OpenAPI documents and load into any
OpenAPI-aware client, viewer, or codegen tool. The portal generates its TypeScript types and Zod schemas from them
with `pnpm gen:api` (`web/scripts/gen-api.mts`).
