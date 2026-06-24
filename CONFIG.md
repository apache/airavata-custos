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

# Configuration Guide

Custos server can be configured using a YAML configuration file instead of environment variables. This guide explains how to set up and use the configuration system.

## Quick Start

1. Place your configuration file at `config/custos.yaml`
2. Start the server:
   ```bash
   ./custos
   ```

The server will automatically load the configuration file. If the config file is not found at the resolved path, the server exits with an error — make sure `config/custos.yaml` exists (or override `CONFIG_PATH`).

## Configuration File Location

By default, the server looks for the configuration file at `config/custos.yaml`. You can customize this location using the `CONFIG_PATH` environment variable:

```bash
CONFIG_PATH=/etc/custos/config.yaml ./custos
```

## Configuration Structure

### Core Configuration

The `core` section contains essential server settings:

```yaml
core:
  database:
    url: "admin:admin@tcp(localhost:3306)/custos?parseTime=true&charset=utf8mb4"
  api:
    port: 8080
  log_level: "info"
```

- **database.url**: MariaDB / MySQL DSN (the server uses the `go-sql-driver/mysql` driver). Required.
- **api.port**: HTTP API port (default: 8080)
- **log_level**: Logging level (info, debug, warn, error)

### Auth Configuration

The verifier and caller resolver read this block. The loader refuses to boot when either OIDC field is empty; the cache TTL is capped at 60s.

```yaml
core:
  auth:
    oidc:
      issuer:   "${OIDC_ISSUER_URL}"
      audience: "${OIDC_AUDIENCE}"
    cache_ttl: "30s"
```

| Key | Default | Notes |
|-----|---------|-------|
| `core.auth.oidc.issuer` | *(required)* | OIDC issuer URL. The verifier discovers JWKS from `<issuer>/.well-known/openid-configuration`. Empty → boot fails with `core.auth.oidc.issuer is required`. |
| `core.auth.oidc.audience` | *(required)* | Expected `aud` claim. Usually the client_id registered with the IdP. Empty → boot fails with `core.auth.oidc.audience is required`. |
| `core.auth.cache_ttl` | `30s` | TTL of the in-process caller + privilege cache, keyed by OIDC sub. Any `time.Duration` string (`s`, `m`). Zero or negative → `30s`. Values above `60s` are capped at `60s` at boot and a WARN is logged. |

Public routes (those that bypass JWT verification) are declared in code via `router.Public(...)`. There is no YAML allowlist.

### Connector Configuration

The `connectors` section defines which connectors are enabled and their settings.

#### SLURM Association Mapper

```yaml
connectors:
  slurm-mapper:
    type: "slurm-association-mapper"
    enabled: true
    slurm_api:
      url: "https://slurm-api.example.com"
      version: "0.0.38"
      username: "slurm_admin"
      token: "${SLURM_TOKEN}"
```

#### SLURM Usage Monitor

```yaml
  slurm-usage-monitor:
    type: "slurm-usage-monitor"
    enabled: true
    slurm_api:
      url: "https://slurm-api.example.com"
      version: "0.0.38"
      username: "slurm_admin"
      token: "${SLURM_TOKEN}"
    cluster_id: "slurm-cluster"
```

#### COmanage Identity Provisioner

```yaml
  comanage-provisioner:
    type: "comanage-identity-provisioner"
    enabled: true
    registry:
      url: "https://comanage.example.org"
      co_id: 1
      api_user: "comanage_api_user"
      api_key: "${COMANAGE_API_KEY}"
    unix_cluster:
      id: 10
      person_id_type: "eppn"
    provisioning:
      custos_cluster_id: "cluster-001"
      default_shell: "/bin/bash"
      homedir_prefix: "/home/"
      http_timeout: "30s"
```

#### AMIE Processor

```yaml
  amie-processor:
    type: "amie-processor"
    enabled: true
    credentials:
      base_url: "https://amie.xsede.org"
      site_code: "XSEDE"
      api_key: "${AMIE_API_KEY}"
    cluster:
      id: "cluster-001"
    polling:
      poll_interval: "30s"
      worker_interval: "5s"
      poller_enabled: true
    timeouts:
      connect_timeout: "5s"
      read_timeout: "20s"
```

## Environment Variable Substitution

The configuration parser supports environment variable substitution using the `${VAR_NAME}` syntax. This allows you to:

1. Keep sensitive values (API keys, passwords) out of version control
2. Use different values for different deployment environments
3. Reference environment variables directly in the config file

### Example

```yaml
core:
  database:
    url: "${DATABASE_URL}"
```

In your shell:
```bash
export DATABASE_URL="admin:admin@tcp(localhost:3306)/custos?parseTime=true&charset=utf8mb4"
./custos
```

If an environment variable referenced in the config file is not set, it will remain unexpanded:

```yaml
token: "${MISSING_TOKEN}"  # Will not be replaced if MISSING_TOKEN is not set
```

## Environment Overrides

A small set of process-level knobs aren't in the YAML — they're read directly from the environment at boot.

| Variable | Default | Purpose |
|----------|---------|---------|
| `CONFIG_PATH` | `config/custos.yaml` | Override the config file location. |
| `DB_MAX_OPEN_CONNS` | `25` | Maximum open database connections. |
| `DB_MAX_IDLE_CONNS` | `5` | Maximum idle database connections. |

## Enabling/Disabling Connectors

The configuration supports any number of connectors with any names. Each connector is identified by its `type` field which determines which loader is used:

```yaml
connectors:
  connector-name:  # Can be any identifier
    type: "slurm-association-mapper"  # Determines which loader to use
    enabled: true
    # Connector-specific configuration
```

To disable a connector, set `enabled: false`:

```yaml
connectors:
  slurm-mapper:
    type: "slurm-association-mapper"
    enabled: false
    # Rest of configuration is still required but will be ignored
```

When a connector is disabled:
- It will not be loaded on server startup
- An info log message will indicate it's disabled
- The server will continue to function normally

### Supported Connector Types

- `slurm-association-mapper` - SLURM Association Mapper
- `slurm-usage-monitor` - SLURM Usage Monitor
- `comanage-identity-provisioner` - COmanage Identity Provisioner
- `amie-processor` - AMIE Processor

New connector types can be added by:
1. Implementing the connector package
2. Adding the loader function to `internal/connectors/loader.go`
3. Mapping the type to the loader in `connectorLoaders` map

## Configuration Validation

The configuration is validated during parsing. Common issues:

1. **Missing required fields**: If `core.database.url` is missing, the server will fail to start
2. **Invalid YAML syntax**: Check your YAML indentation and structure
3. **Undefined environment variables**: Variables referenced with `${VAR}` will be left as-is if not set

## Logging

Configuration loading is logged at the INFO level:

```
{
  "time": "2026-06-12T10:30:45Z",
  "level": "INFO",
  "msg": "loaded config",
  "path": "config/custos.yaml"
}
```

Connector loading logs indicate which connectors are enabled or disabled:

```
{
  "time": "2026-06-12T10:30:46Z",
  "level": "INFO",
  "msg": "loading SLURM Association Mapper connector"
}
```

## Adding New Connectors

To add a new connector type:

1. Implement the connector package with a `LoadConnector` function
2. Update `internal/connectors/loader.go` to import the new connector
3. Add the type-to-loader mapping in the `connectorLoaders` map:

```go
connectorLoaders := map[string]func(context.Context, *sqlx.DB, *events.Bus, *service.Service, *sync.WaitGroup, *http.ServeMux, *config.ConnectorConfig) error{
    "my-new-connector": mynewconnector.LoadConnector,
    // ... existing connectors
}
```

4. Add the connector to your config file with the corresponding type

No code changes are required elsewhere — the configuration system is fully extensible.

## Example Complete Configuration

See [config/custos.yaml](config/custos.yaml) for a complete example configuration with all connectors.
