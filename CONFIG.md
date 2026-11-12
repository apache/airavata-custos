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

# Configuration

Reference for the server's YAML file, environment variables, `core` keys and connector blocks. Connector keys are in
each connector's README, linked under [Connectors](#connectors). For a local stack, follow [INSTALL.md](INSTALL.md).
[`config/custos.yaml`](config/custos.yaml) is the complete sample.

## The config file

The server reads one YAML file at startup, `config/custos.yaml` by default. Set `CONFIG_PATH` to read another:

```bash
CONFIG_PATH=/etc/custos/config.yaml ./custos
```

The server exits if it cannot read or parse the file. On success it logs the path:

```json
{
  "time": "2026-06-12T10:30:45Z",
  "level": "INFO",
  "msg": "loaded config",
  "path": "config/custos.yaml"
}
```

## Environment variable substitution

Before parsing, the server replaces each `${VAR_NAME}` in the file with that variable's value.

```yaml
core:
  database:
    url: "${DATABASE_DSN}"
```

```bash
export DATABASE_DSN="postgres://admin:admin@localhost:5432/custos?sslmode=disable"
./custos
```

| Variable state                      | Result in the file |
|-------------------------------------|--------------------|
| set                                 | its value          |
| set to the empty string, or not set | the empty string   |

An unset variable is not an error, so a typo leaves the value empty. Write numeric values unquoted, so YAML
reads them as numbers.

[`.env.example`](.env.example) lists the variables the sample file uses, with local defaults.

## Environment variables read directly

The server reads these from the environment, not the YAML file.

| Variable                       | Default              | Purpose                                                                                                                                                                                                        |
|--------------------------------|----------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `CONFIG_PATH`                  | `config/custos.yaml` | Location of the config file.                                                                                                                                                                                   |
| `DB_MAX_OPEN_CONNS`            | `25`                 | Maximum open database connections.                                                                                                                                                                             |
| `DB_MAX_IDLE_CONNS`            | `5`                  | Maximum idle database connections.                                                                                                                                                                             |
| `CUSTOS_BOOTSTRAP_ADMIN_EMAIL` | unset                | Grants super_admin to the user with this email at startup. Creates a PENDING user if none exists. Skipped when super_admin already has a holder. The first sign-in links the OIDC `sub` to that user by email. |

## Core keys

The `core` block configures the server itself.

```yaml
core:
  database:
    url: "${DATABASE_DSN}"
  api:
    port: 8080
  log_level: "info"
  auth:
    oidc:
      issuer: "${OIDC_ISSUER_URL}"
      audience: "${OIDC_AUDIENCE}"
      apps_audience: "${OIDC_APPS_AUDIENCE}"
    cache_ttl: "30s"
  users:
    reserved_usernames: [ root, daemon, slurm, ... ]
```

| Key                             | Default  | Notes                                                                                                                                                                                          |
|---------------------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `core.database.url`             | required | PostgreSQL DSN, opened with the `pgx` driver. Empty → the server exits with `database.url in config is required`.                                                                              |
| `core.api.port`                 | `8080`   | HTTP port of the API.                                                                                                                                                                          |
| `core.log_level`                | `info`   | `debug`, `info`, `warn` or `error`. Any other value means `info`.                                                                                                                              |
| `core.auth.oidc.issuer`         | required | OIDC issuer URL. The verifier discovers the JWKS from `<issuer>/.well-known/openid-configuration`. Empty → boot fails with `core.auth.oidc.issuer is required`.                                |
| `core.auth.oidc.audience`       | required | Expected `aud` claim, usually the client ID registered with the provider. Empty → boot fails with `core.auth.oidc.audience is required`.                                                       |
| `core.auth.oidc.apps_audience`  | empty    | Client ID of our public client that delegated apps use to get the user's token ([apps](docs/architecture.md#apps)). Empty turns off delegated apps. The same value as `audience` → boot fails. |
| `core.auth.cache_ttl`           | `30s`    | How long a resolved caller and its privileges stay cached. A Go duration such as `30s` or `1m`. Zero or negative → `30s`. Above `60s` → capped at `60s` with a warning.                        |
| `core.users.reserved_usernames` | empty    | Names a user cannot choose as their cluster login. The sample lists system, service and distribution accounts.                                                                                 |

## Connectors

Each `connectors` entry maps a site-chosen name to one connector. `type` selects the loader.

```yaml
connectors:
  slurm-mapper: # any name
    type: "slurm-association-mapper"     # selects the loader
    enabled: true
    # connector keys
```

At startup the server goes through the blocks:

| Block                          | Behavior                                                                                                            |
|--------------------------------|---------------------------------------------------------------------------------------------------------------------|
| `enabled: false`               | Not loaded. Logs `connector is disabled` and stops keeping bus events for that type. The other keys are ignored.    |
| unknown `type`                 | Not loaded. Logs a warning. Startup continues.                                                                      |
| enabled, required keys missing | The connector logs it, records it in its health status, and is skipped. The VAST provisioner fails startup instead. |
| enabled, load error            | Startup fails.                                                                                                      |

`GET /connectors` reports each configured connector's status.

| Type                            | Keys                                                                                              |
|---------------------------------|---------------------------------------------------------------------------------------------------|
| `slurm-association-mapper`      | [SLURM Association-Mapper](connectors/SLURM/Association-Mapper/README.md#configuration)           |
| `slurm-usage-monitor`           | [SLURM Usage Monitor](connectors/SLURM/Usage-Monitor/README.md#configuration)                     |
| `comanage-identity-provisioner` | [COmanage Identity-Provisioner](connectors/COmanage/Identity-Provisioner/README.md#configuration) |
| `vast-storage-provisioner`      | [VAST-Provisioner](connectors/Storage/VAST-Provisioner/README.md#configuration)                   |
| `amie-processor`                | [AMIE Processor](connectors/ACCESS/AMIE-Processor/README.md#configuration)                        |
| `email-notifier`                | [Email Notifier](connectors/Notification/Email/README.md#configuration)                           |
| `temp-account`, `analytics`     | [Temporary accounts and analytics](#temporary-accounts-and-analytics), below                      |

To add a connector type, follow [docs/contributing/writing-a-connector.md](docs/contributing/writing-a-connector.md).

### Temporary accounts and analytics

The `temp-account` and `analytics` connectors take no keys beyond `type` and `enabled`.
