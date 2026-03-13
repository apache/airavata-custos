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

# Custos Identity Server

The Identity Server is the core component of [Apache Airavata Custos](https://github.com/apache/airavata-custos), providing identity and access management, tenant administration, user and group management, and credential storage for science gateways. It exposes a REST API backed by Keycloak for IAM, HashiCorp Vault for secrets management, and MariaDB for persistence.

For the full project overview, other Custos components, and deployment instructions, see the [root README](../README.md).

## Module Structure

```
identity/
├── pom.xml          # Parent POM (custos-identity)
├── core/            # JPA entities, Spring Data repositories, Protobuf definitions, mappers
├── services/        # Business logic -- Keycloak, Vault, CILogon integrations, TokenAuthorizer
├── api/             # REST controllers under /api/v1/*
└── application/     # Spring Boot entry point and configuration (port 8081)
```

Each layer depends only on the one directly below it:

```
application -> api -> services -> core
```

## Prerequisites

- Java 17
- Maven 3.6+
- Docker (for local infrastructure)

## Local Development Setup

### 1. Start Infrastructure

From the repository root, start the required containers:

```sh
cd compose
docker compose up -d
```

This launches:

| Service  | URL                        | Purpose              |
|----------|----------------------------|----------------------|
| Keycloak | http://localhost:8080       | Identity and access management |
| MariaDB  | localhost:3306             | Persistence          |
| Vault    | http://localhost:8200       | Secrets management   |
| Adminer  | http://localhost:18080      | Database admin UI    |

### 2. Configure Vault

1. Open http://localhost:8200 in your browser.
2. Walk through the Vault initialization process.
3. Save the **root token** and **unseal key** -- you will need both.
4. Edit `application/src/main/resources/application.yml` and replace the placeholder value under `spring.cloud.vault.token` with your root token:

```yaml
spring:
  cloud:
    vault:
      token: <your-root-token>
```

### 3. Build

From the repository root, build all modules:

```sh
mvn clean install
```

To build only the Identity Server and its dependencies:

```sh
mvn clean install -pl identity -am
```

### 4. Run

```sh
mvn spring-boot:run -pl identity/application
```

The server starts on port **8081**.

### 5. Super-Tenant Bootstrap

The super-tenant is created automatically on first startup when `custos.bootstrap.enabled` is `true`. The `application-local.yml` profile (active by default in local dev) enables this with default dev credentials.

On startup you will see a log line:

```
Super-tenant bootstrapped. Client ID: <generated-id>. Secret stored in Vault at /secret/<tenantId>/CUSTOS
```

To retrieve the generated credentials, query Vault:

```sh
curl -H "X-Vault-Token: <your-root-token>" http://localhost:8200/v1/secret/<tenantId>/CUSTOS
```

These `client_id` and `client_secret` values are required to authenticate subsequent API calls.

If the super-tenant already exists (from a previous startup), bootstrap is skipped automatically.

## REST API

Base path: `/api/v1`

| Controller            | Path                              | Purpose                             |
|-----------------------|-----------------------------------|-------------------------------------|
| Identity Management   | `/api/v1/identity-management`     | Authentication, tokens, OIDC flows  |
| Tenant Management     | `/api/v1/tenant-management`       | Tenant CRUD, role management        |
| User Management       | `/api/v1/user-management`         | User CRUD, attributes, role assignment |
| Group Management      | `/api/v1/group-management`        | Group CRUD, membership management   |

Swagger UI is available at http://localhost:8081/swagger-ui.html once the server is running.

### Request and Response Format

The API uses Protobuf message definitions (found in `core/src/main/proto/`) serialized as JSON for both request and response bodies. Content type is `application/json`.

### Authentication

A minimal Spring Security configuration provides security headers (X-Frame-Options, X-Content-Type-Options). Endpoint-level authentication is handled by `TokenAuthorizer` on a per-endpoint basis. API calls require a valid access token obtained through the Identity Management endpoints.

## Key Technologies

- **Spring Boot** -- Application framework and REST layer
- **Keycloak** -- Delegated identity and access management
- **HashiCorp Vault** -- Secure credential and secret storage
- **MariaDB** -- Relational persistence (schema auto-generated via Hibernate `ddl-auto: update`)
- **Protocol Buffers** -- API message definitions
- **CILogon** -- Federated authentication for research institutions

## Additional Components

The Custos project includes additional components beyond the Identity Server (such as the ACCESS CI service). Refer to the [root README](../README.md) for information on those components.
