# Custos configuration

The server reads `config/custos.yaml` at boot. `${VAR}` substitution keeps
secrets out of the file; an unset var stays as the literal `${VAR}` so
missing-secret failures surface at the consumer, not at load time.

## Core

| Key | Default | Notes |
|-----|---------|-------|
| `core.database.url` | *(required)* | MariaDB / MySQL DSN. `parseTime=true` is mandatory. |
| `core.api.port` | `8080` | HTTP API port. |
| `core.log_level` | `info` | One of `debug`, `info`, `warn`, `error`. |

## Auth

The verifier + caller resolver read this block. The loader refuses to boot
when either OIDC field is empty; the cache TTL is capped at 60s.

| Key | Default | Notes |
|-----|---------|-------|
| `core.auth.oidc.issuer` | *(required)* | OIDC issuer URL. The verifier discovers JWKS from `<issuer>/.well-known/openid-configuration`. Empty → boot fails with `core.auth.oidc.issuer is required`. |
| `core.auth.oidc.audience` | *(required)* | Expected `aud` claim. Usually the client_id registered with the IdP. Empty → boot fails with `core.auth.oidc.audience is required`. |
| `core.auth.cache_ttl` | `30s` | TTL of the in-process caller + privilege cache, keyed by OIDC sub. Units: any `time.Duration` string (`s`, `m`). Zero or negative → `30s`. Values above `60s` are capped at `60s` at boot and a WARN is logged. |
| `core.auth.public_paths` | `[]` | List of path components that bypass JWT verification entirely. Typically `/healthz` and `/ready`. Routes registered via `router.Public(...)` are merged with this list at server bootstrap. |

### Example

```yaml
core:
  auth:
    oidc:
      issuer:   "${OIDC_ISSUER_URL}"
      audience: "${OIDC_AUDIENCE}"
    cache_ttl: "30s"
    public_paths:
      - "/healthz"
      - "/ready"
```

## Environment overrides

| Variable | Default | Purpose |
|----------|---------|---------|
| `DB_MAX_OPEN_CONNS` | `25` | Maximum open database connections. |
| `DB_MAX_IDLE_CONNS` | `5` | Maximum idle database connections. |
| `CONFIG_PATH` | `config/custos.yaml` | Override the config file location. |
