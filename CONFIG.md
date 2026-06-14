# Configuration Guide

Custos server can be configured using a YAML configuration file instead of environment variables. This guide explains how to set up and use the configuration system.

## Quick Start

1. Place your configuration file at `config/custos.yaml`
2. Start the server:
   ```bash
   ./custos
   ```

The server will automatically load the configuration file. If the config file is not found, it falls back to legacy environment variable configuration.

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
    url: "postgresql://user:password@localhost:5432/custos_db?sslmode=disable"
  api:
    port: 8080
  log_level: "info"
```

- **database.url**: PostgreSQL connection string (required)
- **api.port**: HTTP API port (default: 8080)
- **log_level**: Logging level (info, debug, warn, error)

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
export DATABASE_URL="postgresql://user:password@localhost:5432/custos_db"
./custos
```

If an environment variable referenced in the config file is not set, it will remain unexpanded:

```yaml
token: "${MISSING_TOKEN}"  # Will not be replaced if MISSING_TOKEN is not set
```

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

## Fallback to Environment Variables

For backward compatibility, if the configuration file is not found, the server falls back to the legacy environment variable configuration:

```bash
DATABASE_DSN="..." HTTP_ADDR=":8080" ./custos
```

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
connectorLoaders := map[string]func(context.Context, *sqlx.DB, *events.Bus, *service.Service, *sync.WaitGroup) error{
    "my-new-connector": mynewconnector.LoadConnector,
    // ... existing connectors
}
```

4. Add the connector to your config file with the corresponding type

No code changes are required elsewhere — the configuration system is fully extensible.

## Example Complete Configuration

See [config/custos.yaml](config/custos.yaml) for a complete example configuration with all connectors.
