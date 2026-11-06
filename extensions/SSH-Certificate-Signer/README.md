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

# Custos Signer

Custos Signer is an SSH certificate authority for science gateways.
It issues short-lived OpenSSH user certificates bound to a user's OIDC identity.
This README is for site operators and for developers of gateways that call it.

The signer is a separate Go module (`github.com/apache/airavata-custos/signer`) and its own service.

## How it works

A gateway calls `POST /api/v1/sign` with four inputs: its client credentials, the user's OIDC access token, the user's
SSH public key and the principal. The principal is the Unix username the certificate is valid for.

1. The signer authenticates the client against its row in the client table, `client_ssh_configs`. The secret is stored as a bcrypt hash.
2. It validates the user's token with the issuer's JWKS. The issuer must be in `allowed_issuers`, unless that list is
   empty.
3. It enforces the client's policy: maximum TTL, allowed key types and source address. It grants the standard
   extensions, minus those denied.
4. It validates the principal per `principal_source`, below.
5. It signs with the CA key from Vault and logs the issuance in `certificate_issuance_logs`.

| `principal_source` | Principal check |
|---|---|
| `noop` | Accepts the principal |
| `ldap` | Finds the LDAP entry for the OIDC subject and requires the principal to equal its `username_attribute`. Credentials are in Vault. |
| `comanage` | Not implemented. Rejects every request with `COMANAGE_NOT_IMPLEMENTED`. |

A deployment has one Ed25519 CA, shared by all clients. It is generated in Vault by the first client route that reads it.

## Prerequisites

- Go 1.27+
- PostgreSQL 17+
- Vault / OpenBao with KV v2 secrets engine enabled

## Setup

### Compose stack

The dev compose stack runs the signer with everything it needs: PostgreSQL, OpenBao, Keycloak, OpenLDAP and a test
login node. From the repository root:

```bash
cd dev-ops/compose
docker compose up -d db keycloak openldap openbao openbao-init signer signer-seed sshd-login
./signer-e2e.sh    # token, sign, SSH login, and the negative checks
```

The rest of this section is the manual setup.

### Database

```sql
CREATE DATABASE custos_signer;
```

If using the project's Docker Compose, the `custos_signer` database is created automatically by `dev-ops/compose/dbinit/init-db.sh`.

Migrations are managed separately from the server. Migrations are read from `migrations/`, so run the signer from
this directory. Apply them explicitly:

```bash
./custos-signer migrate --config config.yaml
```

This applies any pending migrations and exits. It's idempotent — safe to run repeatedly. The `migrate` subcommand tracks applied versions in a `schema_migrations` table and only applies what's new.

You can also apply SQL directly:

```bash
psql -U admin -d custos_signer < migrations/001_initial_schema.up.sql
```

The initial migration creates three tables:
- `client_ssh_configs` — client configuration and credentials
- `certificate_issuance_logs` — audit log for certificate issuances
- `revocation_events` — audit log for certificate revocations

Register a client:

```sql
INSERT INTO client_ssh_configs
  (client_id, client_secret, target_host, target_port,
   max_ttl_seconds, allowed_key_types, enabled)
VALUES
  ('webapp', '$2a$10$<bcrypt_hash_of_secret>', 'login.example.org', 22,
   86400, '["ed25519","rsa","ecdsa"]', TRUE);
```

Generate the bcrypt hash with:

```bash
htpasswd -nbBC 10 "" "your-secret" | cut -d: -f2
```

The table also holds `source_address_restriction`, `denied_extensions` and `principal_source` (default `noop`).

### Vault

```bash
vault secrets enable -path=ssh-ca kv-v2
```

The path must match `vault.mount_path`. The signer creates CA keys on demand. No manual key setup is required.

Ensure the configured token has read/write access to `ssh-ca/data/*`.

For `principal_source = 'ldap'`, store each client's LDAP settings at `ssh-ca/clients/<client_id>/validation`:

| Key | Value |
|---|---|
| `type` | `ldap` |
| `ldap_url`, `bind_dn`, `bind_password`, `base_dn` | LDAP connection and bind settings |
| `search_filter` | Search filter. `%s` becomes the OIDC subject. |
| `username_attribute` | Optional. Default `uid`. |

### Configuration

```bash
cd extensions/SSH-Certificate-Signer
go build -o custos-signer .
cp config.example.yaml config.yaml          # edit with your settings
```

The service loads `config.yaml` by default. Override with `--config /path/to/config.yaml`.
In [`config.example.yaml`](config.example.yaml), `vault.token` and `allowed_issuers` are samples. The built-in defaults are empty.

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
| `CORE_API_BASE_URL` | `signer.core_api_base_url` | Custos Core URL used to resolve administrator privileges |
| `LOG_LEVEL` | `logging.level` | Log level: debug, info, warn, error |

## Running the service

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

Development:

```bash
./custos-signer serve --config config.yaml --auto-migrate
```

Production:

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

The image is a multi-stage build on an Alpine base, ships the migrations, and runs as a non-root user (UID 65534).

The service handles SIGTERM and SIGINT for graceful shutdown:
1. Stops accepting new connections.
2. Drains in-flight requests (up to `shutdown_timeout_seconds`, default 30s).
3. Closes database connection pool.
4. Exits.

## API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/v1/sign` | Client credentials and the user's OIDC token | Sign an SSH public key |
| `POST` | `/api/v1/revoke` | Client credentials | Revoke a certificate by `serial_number`, `key_id` or `ca_fingerprint` |
| `GET` | `/api/v1/jwks` | Client credentials | CA public keys as JWKs |
| `GET` | `/api/v1/ca-public-key` | Client credentials | CA public key |
| `POST` | `/api/v1/admin/rotate-ca` | Client credentials | Rotate the CA keys |
| `GET` | `/api/v1/certificates` | OIDC Bearer | Certificates of the signed-in user |
| `GET` | `/api/v1/certificates/{serial}` | OIDC Bearer | One certificate |
| `POST` | `/api/v1/certificates/{serial}/revoke` | OIDC Bearer | Revoke an owned active certificate, or any active certificate with `signer:certificates:write` |
| `GET` | `/api/v1/userinfo` | OIDC Bearer | The signed-in user's profile |
| `GET` | `/api/v1/admin/certificates` | OIDC Bearer and `signer:certificates:read` | All certificates |
| `GET` | `/api/v1/admin/certificates/{serial}` | OIDC Bearer and `signer:certificates:read` | Any certificate |
| `GET` | `/api/v1/health` | None | Health check |
| `GET` | `/metrics` | None | Prometheus metrics, when `metrics.enabled` |

Client credentials go in the `X-Client-Id` and `X-Client-Secret` headers. OIDC Bearer
routes take `Authorization: Bearer <token>`.

The signer reads the caller's privileges from `GET /me` at `core_api_base_url`:

| Action | Needs |
|---|---|
| Admin routes | `signer:certificates:read` |
| Revoke another user's certificate | `signer:certificates:write` |
| Revoke one's own certificate | No privilege |

The Custos server does not register the `signer:` privileges ([docs/privileges.md](../../docs/privileges.md#other-connectors)).
Calls that need one return 403.

`GET /api/v1/admin/certificates` uses keyset pagination. Pass `limit`, and `cursor` set to the last `next_cursor`.
It returns no total or offset.

> **Revocation.** The signer records revocations. It does not generate or distribute OpenSSH KRLs, so login nodes still accept a revoked certificate.

### Examples

**Health Check**

```bash
curl -s http://localhost:8084/api/v1/health | jq .
```

**Sign a Certificate**

```bash
ssh-keygen -t ed25519 -f /tmp/testkey -N ""

curl -s -X POST http://localhost:8084/api/v1/sign \
  -H "Content-Type: application/json" \
  -H "X-Client-Id: webapp" \
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
  -H "X-Client-Id: webapp" \
  -H "X-Client-Secret: your-secret" \
  -d '{"serial_number": 42, "reason": "User deprovisioned"}' | jq .
```

**Get CA Public Keys (JWK format)**

```bash
curl -s http://localhost:8084/api/v1/jwks \
  -H "X-Client-Id: webapp" \
  -H "X-Client-Secret: your-secret" | jq .
```

**Rotate CA Keys**

```bash
curl -s -X POST http://localhost:8084/api/v1/admin/rotate-ca \
  -H "X-Client-Id: webapp" \
  -H "X-Client-Secret: your-secret" | jq .
```

## Troubleshooting

| Error | Cause | Fix |
|-------|-------|-----|
| "failed to connect to database" | PostgreSQL unreachable or wrong credentials | Check host/port, verify DB user permissions, ensure database exists |
| "failed to create vault client" | Vault unreachable or bad token | Check `VAULT_ADDRESS`, ensure token is valid, verify KV v2 enabled |
| Migrations not found | The working directory is not this one | Run the signer from `extensions/SSH-Certificate-Signer` |
| HTTP 401 on authenticated requests | Bad client credentials | Verify `X-Client-Id` is the client's `client_id`, check bcrypt hash matches |
| HTTP 403 "Client is disabled" | Client row has `enabled=false` | Set `enabled=true` in `client_ssh_configs` |
| HTTP 401 "Issuer not allowed" | OIDC issuer not in allowed list | Add issuer to `allowed_issuers`, or set to empty list to accept all |
| HTTP 403 on admin routes | The caller lacks `signer:certificates:read` in the Custos server | See the privilege note under [API](#api) |
| HTTP 503 "CA key storage unavailable" | Vault down during signing | Check Vault status and network connectivity |
| Token validation errors in dev | No OIDC provider configured | Set `DEV_MODE=true` for development (returns default identity) |

## License

Apache License 2.0
