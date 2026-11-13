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

Custos Signer is an SSH certificate authority. It issues short-lived OpenSSH user certificates.

The signer is internal and has one trusted caller, the Custos server, through its
[Signer connector](../../connectors/Signer/README.md). Users and apps talk to Custos, and Custos decides who may have a
certificate. Do not expose the signer to the public. Whoever holds its service token can get a certificate for any user
that is not reserved.

The signer is a separate Go module (`github.com/apache/airavata-custos/signer`) and its own service.

## How it works

The caller sends `POST /api/v1/sign` with the principal, the user's SSH public key, the user's email and subject, and
`requested_by`, which says who asked. The principal is the Unix username the certificate is valid for. The signer does
not interpret `requested_by`.

1. The signer checks the service token.
2. It refuses a username in `signer.reserved_usernames`, whichever validator is on.
3. It enforces the deployment's maximum TTL and allowed key types, and grants the standard extensions minus those
   denied.
4. It runs the validator set in `signer.validation.principal_validator`, below.
5. It signs with the CA key from Vault and logs the issuance in `certificate_issuance_logs`.

| `principal_validator` | Principal check                                                                                                                                        |
|-----------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| `none`                | Accepts the principal. The caller has already decided.                                                                                                 |
| `ldap`                | A site's own extra check. Finds the LDAP entry for the subject and requires the principal to equal its `username_attribute`. Credentials are in Vault. |

The certificate names who it is for. The CA signature covers all of it.

| Field                                      | Value                                                      |
|--------------------------------------------|------------------------------------------------------------|
| Key ID                                     | `<email>:<requested_by>`. sshd writes it to the login log. |
| Extension `user-sub@custos.apache.org`     | The user's subject                                         |
| Extension `user-email@custos.apache.org`   | The user's email                                           |
| Extension `requested-by@custos.apache.org` | Who asked. Custos sends `user/<id>` or `app/<id>`.         |

A request can also set `source_address`, a comma-separated list of IP addresses or CIDR blocks. It goes in the
certificate's `source-address` critical option, and sshd then accepts the certificate only from those addresses.

A deployment has one Ed25519 CA, stored in Vault in OpenSSH format, the same format `ssh-keygen -t ed25519` writes. The
signer reads it at startup and refuses to start without it,
unless `signer.ca.create_if_missing` is set, in which case it creates one and writes a `ca_created` audit log entry.

## Prerequisites

- Go 1.27+
- PostgreSQL 17+
- Vault / OpenBao with KV v2 secrets engine enabled

## Setup

### Compose stack

The dev compose stack runs the signer with everything it needs: PostgreSQL, OpenBao and a test login node. From the
repository root:

```bash
cd dev-ops/compose
docker compose up -d db openbao openbao-init signer sshd-login
./signer-e2e.sh    # sign, SSH login, list, revoke, and the negative checks
```

The rest of this section is the manual setup.

### Database

```sql
CREATE
    DATABASE custos_signer;
```

If using the project's Docker Compose, the `custos_signer` database is created automatically by
`dev-ops/compose/dbinit/init-db.sh`.

Migrations are managed separately from the server. Migrations are read from `migrations/`, so run the signer from
this directory. Apply them explicitly:

```bash
./custos-signer migrate --config config.yaml
```

This applies any pending migrations and exits. It's idempotent — safe to run repeatedly. The `migrate` subcommand tracks
applied versions in a `schema_migrations` table and only applies what's new.

You can also apply SQL directly:

```bash
psql -U admin -d custos_signer < migrations/001_initial_schema.up.sql
```

The initial migration creates two tables:

- `certificate_issuance_logs` — audit log for certificate issuances
- `revocation_events` — audit log for certificate revocations

### Vault

```bash
vault secrets enable -path=ssh-ca kv-v2
```

The path must match `vault.mount_path`. Store the CA key before starting the signer, or set
`signer.ca.create_if_missing` to let it create one.

Ensure the configured token has read/write access to `ssh-ca/data/*`.

For `principal_validator: ldap`, store the LDAP settings at `ssh-ca/validation`:

| Key                                               | Value                                           |
|---------------------------------------------------|-------------------------------------------------|
| `type`                                            | `ldap`                                          |
| `ldap_url`, `bind_dn`, `bind_password`, `base_dn` | LDAP connection and bind settings               |
| `search_filter`                                   | Search filter. `%s` becomes the user's subject. |
| `username_attribute`                              | Optional. Default `uid`.                        |

### Configuration

```bash
cd extensions/SSH-Certificate-Signer
go build -o custos-signer .
cp config.example.yaml config.yaml          # edit with your settings
```

The service loads `config.yaml` by default. Override with `--config /path/to/config.yaml`.
In [`config.example.yaml`](config.example.yaml), `vault.token` is a sample. The signer refuses to start without a
`service_token` of at least 32 characters, without `root` in `reserved_usernames`, and without a `login_host`.

Environment variables take precedence over YAML values.

| Env Var                | Config Path            | Description                         |
|------------------------|------------------------|-------------------------------------|
| `DB_HOST`              | `database.host`        | Database hostname                   |
| `DB_PORT`              | `database.port`        | Database port                       |
| `DB_NAME`              | `database.name`        | Database name                       |
| `DB_USERNAME`          | `database.username`    | Database user                       |
| `DB_PASSWORD`          | `database.password`    | Database password                   |
| `VAULT_ADDRESS`        | `vault.address`        | Vault API address                   |
| `VAULT_TOKEN`          | `vault.token`          | Vault authentication token          |
| `SIGNER_SERVICE_TOKEN` | `signer.service_token` | The caller's token                  |
| `LOG_LEVEL`            | `logging.level`        | Log level: debug, info, warn, error |

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

| Method | Path                                   | Description                                                                      |
|--------|----------------------------------------|----------------------------------------------------------------------------------|
| `POST` | `/api/v1/sign`                         | Sign an SSH public key                                                           |
| `GET`  | `/api/v1/ca-public-key`                | CA public key                                                                    |
| `GET`  | `/api/v1/certificates`                 | List certificates. Filters: `user_sub`, `requested_by`                           |
| `GET`  | `/api/v1/certificates/{serial}`        | One certificate                                                                  |
| `POST` | `/api/v1/certificates/{serial}/revoke` | Revoke an active certificate. `revoked_by` is required and `reason` can be empty |
| `GET`  | `/api/v1/health`                       | Health check, no token                                                           |

Every route with a token takes `Authorization: Bearer <service token>`.

`GET /api/v1/certificates` uses keyset pagination. Pass `limit`, and `cursor` set to the last `next_cursor`.
It returns no total or offset.

> **Revocation.** The signer records revocations. It does not generate or distribute OpenSSH KRLs, so login nodes still
> accept a revoked certificate.

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
  -H "Authorization: Bearer $SIGNER_SERVICE_TOKEN" \
  -d '{
    "principal": "jdoe",
    "ttl_seconds": 3600,
    "public_key": "'"$(cat /tmp/testkey.pub)"'",
    "user_email": "jdoe@example.edu",
    "user_sub": "http://idp.example.edu/users/9",
    "requested_by": "user/8f3a"
  }' | jq .
```

**Revoke a Certificate**

```bash
curl -s -X POST http://localhost:8084/api/v1/certificates/42/revoke \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $SIGNER_SERVICE_TOKEN" \
  -d '{"reason": "User deprovisioned", "revoked_by": "admin@example.edu"}' | jq .
```

## Troubleshooting

| Error                                 | Cause                                       | Fix                                                                 |
|---------------------------------------|---------------------------------------------|---------------------------------------------------------------------|
| "failed to connect to database"       | PostgreSQL unreachable or wrong credentials | Check host/port, verify DB user permissions, ensure database exists |
| "failed to create vault client"       | Vault unreachable or bad token              | Check `VAULT_ADDRESS`, ensure token is valid, verify KV v2 enabled  |
| Migrations not found                  | The working directory is not this one       | Run the signer from `extensions/SSH-Certificate-Signer`             |
| HTTP 401 on any route                 | Wrong service token                         | Set the same `SIGNER_SERVICE_TOKEN` on the signer and the caller    |
| HTTP 403 `reserved_username`          | The username is in `reserved_usernames`     | Certificates are not issued for these names                         |
| HTTP 503 "CA key storage unavailable" | Vault down during signing                   | Check Vault status and network connectivity                         |

## License

Apache License 2.0
