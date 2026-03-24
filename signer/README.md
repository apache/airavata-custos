# Custos Signer

A multi-tenant SSH certificate authority for science gateways and research computing. Custos Signer issues short-lived, identity-bound OpenSSH user certificates so researchers can access HPC resources without managing long-lived SSH keys.

It validates user identity through institutional OIDC providers (CILogon, Keycloak, etc.), enforces access policies (TTL limits, allowed key types, principal validation), signs certificates with per-tenant CAs stored in HashiCorp Vault, and records every transaction for audit. Certificates expire automatically -- no key distribution, no manual revocation, no cleanup.

Part of [Apache Airavata Custos](https://airavata.apache.org/custos/), a security middleware platform for science gateways.

## Prerequisites

- Go 1.22+
- MariaDB 10.6+
- Vault / OpenBao with KV v2 secrets engine enabled

---

## Quick Start

```bash
cd signer
go build -o custos-signer .
cp config.example.yaml config.yaml          # edit with your settings
./custos-signer migrate --config config.yaml # apply database schema
./custos-signer serve --config config.yaml   # start on :8084
```

For development (auto-applies migrations on startup):

```bash
./custos-signer serve --config config.yaml --auto-migrate
```

---

## Configuration

Copy and edit the example configuration:

```bash
cp config.example.yaml config.yaml
```

The service loads `config.yaml` by default. Override with `--config /path/to/config.yaml`.

### Environment Variable Overrides

Environment variables take precedence over YAML values.

| Env Var | Config Path | Description |
|---------|------------|-------------|
| `DB_HOST` | `database.host` | Database hostname |
| `DB_PORT` | `database.port` | Database port |
| `DB_NAME` | `database.name` | Database name |
| `DB_USERNAME` | `database.username` | Database user |
| `DB_PASSWORD` | `database.password` | Database password |
| `VAULT_ADDRESS` | `vault.address` | Vault API address |
| `VAULT_TOKEN` | `vault.token` | Vault authentication token |
| `DEV_MODE` | `dev_mode.enabled` | Enable dev mode (disables OIDC validation) |
| `DEV_DEFAULT_EMAIL` | `dev_mode.default_email` | Default email in dev mode |
| `ALLOWED_ISSUERS` | `signer.auth.allowed_issuers` | Comma-separated list of allowed OIDC issuers |
| `LOG_LEVEL` | `logging.level` | Log level: debug, info, warn, error |

---

## Database Setup

### Create the database

```sql
CREATE DATABASE custos_signer CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

If using the project's Docker Compose, the `custos_signer` database is created automatically by `compose/dbinit/init-db.sh`.

### Migrations

Migrations are managed separately from the server. Apply them explicitly:

```bash
./custos-signer migrate --config config.yaml
```

This applies any pending migrations and exits. It's idempotent — safe to run repeatedly. The `migrate` subcommand tracks applied versions in a `schema_migrations` table and only applies what's new.

You can also apply SQL directly:

```bash
mysql -u admin -p custos_signer < migrations/001_initial_schema.up.sql
```

The initial migration creates three tables:
- `client_ssh_configs` — client configuration and credentials
- `certificate_issuance_logs` — audit log for certificate issuances
- `revocation_events` — audit log for certificate revocations

### Insert a client configuration

```sql
INSERT INTO client_ssh_configs
  (tenant_id, client_id, client_secret, target_host, target_port,
   max_ttl_seconds, allowed_key_types, enabled)
VALUES
  ('tenant1', 'webapp', '$2a$10$<bcrypt_hash_of_secret>', 'login.example.org', 22,
   86400, '["ed25519","rsa","ecdsa"]', TRUE);
```

Generate the bcrypt hash with:

```bash
htpasswd -nbBC 10 "" "your-secret" | cut -d: -f2
```

---

## Vault Setup

### Enable KV v2 secrets engine

```bash
vault secrets enable -path=ssh-ca kv-v2
```

The service auto-generates Ed25519 CA key pairs on first signing request for each tenant-client pair. No manual key setup is required.

Ensure the configured token has read/write access to `ssh-ca/data/*`.

---

## Running the Service

### CLI Usage

```
custos-signer <command> [options]

Commands:
  serve      Start the HTTP server (default)
  migrate    Apply pending database migrations and exit

Options:
  --config string    Path to configuration file (default "config.yaml")
  --auto-migrate     Automatically apply migrations on startup (serve only)
```

Running without a subcommand defaults to `serve`.

### Development

```bash
./custos-signer serve --config config.yaml --auto-migrate
```

### Production

```bash
./custos-signer migrate --config config.yaml   # apply pending migrations
./custos-signer serve --config config.yaml      # start the server
```

The service listens on the port specified in configuration (default: 8084).

### Docker

```bash
docker build -t custos-signer .

# Apply migrations first
docker run --rm -v $(pwd)/config.yaml:/config/config.yaml \
  custos-signer migrate --config /config/config.yaml

# Then start the server
docker run -p 8084:8084 \
  -v $(pwd)/config.yaml:/config/config.yaml \
  -e DB_PASSWORD=secret \
  -e VAULT_TOKEN=hvs.xxx \
  custos-signer
```

The Docker image uses a `scratch` base and runs as non-root user (UID 65534).

### Shutdown Behavior

The service handles SIGTERM and SIGINT for graceful shutdown:
1. Stops accepting new connections.
2. Drains in-flight requests (up to `shutdown_timeout_seconds`, default 30s).
3. Closes database connection pool.
4. Exits.

---

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/sign` | Client credentials + OIDC token | Sign an SSH public key |
| `POST` | `/api/v1/revoke` | Client credentials | Revoke a certificate |
| `GET` | `/api/v1/jwks` | Client credentials | Get CA public keys in JWK format |
| `GET` | `/api/v1/health` | None | Health check |
| `POST` | `/api/v1/admin/rotate-ca` | Client credentials | Rotate CA keys |
| `GET` | `/api/v1/certificates` | OIDC Bearer | List certificates for authenticated user |
| `GET` | `/api/v1/certificates/{serial}` | OIDC Bearer | Get certificate details |
| `GET` | `/api/v1/userinfo` | OIDC Bearer | Get authenticated user profile |
| `GET` | `/metrics` | None | Prometheus metrics |

Client credentials are passed via `X-Client-Id` (format: `{tenant_id}:{client_id}`) and `X-Client-Secret` headers. OIDC Bearer endpoints use `Authorization: Bearer <token>`.

### API Examples

**Health Check**

```bash
curl -s http://localhost:8084/api/v1/health | jq .
```

**Sign a Certificate**

```bash
ssh-keygen -t ed25519 -f /tmp/testkey -N ""

curl -s -X POST http://localhost:8084/api/v1/sign \
  -H "Content-Type: application/json" \
  -H "X-Client-Id: tenant1:webapp" \
  -H "X-Client-Secret: your-secret" \
  -d '{
    "principal": "jdoe",
    "ttl_seconds": 3600,
    "public_key": "'"$(cat /tmp/testkey.pub)"'",
    "user_access_token": "<oidc-access-token>"
  }' | jq .
```

**Revoke a Certificate**

```bash
curl -s -X POST http://localhost:8084/api/v1/revoke \
  -H "Content-Type: application/json" \
  -H "X-Client-Id: tenant1:webapp" \
  -H "X-Client-Secret: your-secret" \
  -d '{"serial_number": 42, "reason": "User deprovisioned"}' | jq .
```

**Get CA Public Keys (JWK format)**

```bash
curl -s http://localhost:8084/api/v1/jwks \
  -H "X-Client-Id: tenant1:webapp" \
  -H "X-Client-Secret: your-secret" | jq .
```

**Rotate CA Keys**

```bash
curl -s -X POST http://localhost:8084/api/v1/admin/rotate-ca \
  -H "X-Client-Id: tenant1:webapp" \
  -H "X-Client-Secret: your-secret" | jq .
```

---

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| "failed to connect to database" | MariaDB unreachable or wrong credentials | Check host/port, verify DB user permissions, ensure database exists |
| "failed to create vault client" | Vault unreachable or bad token | Check `VAULT_ADDRESS`, ensure token is valid, verify KV v2 enabled |
| HTTP 401 on authenticated requests | Bad client credentials | Verify `X-Client-Id` is `tenant_id:client_id` format, check bcrypt hash matches |
| HTTP 403 "Client is disabled" | Client row has `enabled=false` | Set `enabled=true` in `client_ssh_configs` |
| HTTP 401 "Issuer not allowed" | OIDC issuer not in allowed list | Add issuer to `allowed_issuers`, or set to empty list to accept all |
| HTTP 503 "CA key storage unavailable" | Vault down during signing | Check Vault status and network connectivity |
| Token validation errors in dev | No OIDC provider configured | Set `DEV_MODE=true` for development (returns default identity) |

---

## License

Apache License 2.0
