# Custos Signer Service and SDK

A secure SSH certificate signing service and Java SDK for Apache Airavata Custos, enabling short-lived, identity-bound
SSH access without static keys.

## Overview

The Custos Signer Service provides a centralized, secure way to issue short-lived SSH certificates based on
OIDC-authenticated user identities. This eliminates the need for static SSH keys while maintaining strong security
through certificate-based authentication.

### Key Features

- **Short-lived SSH Certificates**: Issue certificates with configurable TTL (default 24 hours)
- **OIDC Integration**: Authenticate users via CILogon or other OIDC providers
- **Multi-tenant Support**: Isolated certificate authorities per tenant
- **Automatic CA Rotation**: Scheduled rotation with overlap periods
- **Key Revocation Lists (KRL)**: Real-time certificate revocation
- **Audit Logging**: Comprehensive audit trail for compliance
- **Ansible Automation**: Automated host configuration and KRL distribution
- **Modular SDK**: Lightweight core with optional persistence backends
- **Single Tenant Architecture**: Clean, industry-standard SDK design

[//]: # (TODO - redraw the architecture diagram)

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Airavata      │    │  Custos Signer  │    │   Target SSH    │
│   Application   │───▶│     Service     │───▶│    Servers      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Custos SDK    │    │   OpenBao/Vault │    │   Ansible       │
│   (Java)        │    │   (CA Storage)  │    │   (Host Config) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Quick Start

### Prerequisites

- Java 17+
- Maven 3.8+
- Docker and Docker Compose
- OpenBao/Vault (for CA key storage)
- MariaDB/MySQL (for audit logs and configuration)

### Building

```bash
# Build the entire project
mvn clean install

# Build only the signer modules
cd signer
mvn clean install
```

### Running with Docker Compose

```bash
# Start all services
cd compose
docker-compose up -d

# Check service status
docker-compose ps

# View logs
docker-compose logs signer-service
```

### Service Endpoints

- **HTTP (Health/Metrics)**: http://localhost:8084
- **gRPC**: localhost:9095
- **Vault UI**: http://localhost:8201
- **Adminer (DB)**: http://localhost:18080

## Service Configuration

### Database Setup

The service uses MariaDB for storing:

- Client configurations and policies
- Certificate issuance audit logs
- Revocation events

Database schema is managed via Flyway migrations in `src/main/resources/db/migration/`.

### OpenBao/Vault Configuration

CA keys are stored in OpenBao/Vault using the following schema:

```
ssh-ca/{tenant}/{clientId}/current    # Active CA private key
ssh-ca/{tenant}/{clientId}/next       # Next rotation CA key
ssh-ca/{tenant}/{clientId}/metadata   # Rotation schedule, serial counter
```

### Client Configuration

Clients are configured in the database with the following information:

```sql
INSERT INTO client_ssh_configs (tenant_id, client_id, client_secret, target_host, target_port, max_ttl_seconds,
                                allowed_key_types, enabled)
VALUES ('airavata-prod', 'airavata-hpcA', '$2a$10$encrypted_secret', 'hpc.example.com', 22, 86400,
        '["ed25519", "rsa", "ecdsa"]', true);
```

## SDK Architecture

The Custos Signer SDK is designed with a modular architecture for flexibility and lightweight deployment:

### Core Module (`signer-sdk-core`)

- **Lightweight**: Only essential dependencies (gRPC, sshj, BouncyCastle, Jackson, slf4j)
- **Self-contained**: Includes `InMemoryKeyStore` for basic key storage
- **No Spring Boot**: Avoids heavy Spring dependencies for simple use cases
- **Industry Standard**: Follows patterns from AWS SDK, GCP SDK, Azure SDK

### Optional Modules

- **`signer-sdk-dbstore`**: Database-backed key storage with Spring Data JPA
- **`signer-sdk-vaultstore`**: Vault/OpenBao-backed key storage with Spring Vault

### Usage Patterns

**Basic Usage (Core Only)**:

```xml

<dependency>
    <groupId>org.apache.custos</groupId>
    <artifactId>custos-signer-sdk-core</artifactId>
</dependency>
```

**With Database Storage**:

```xml

<dependency>
    <groupId>org.apache.custos</groupId>
    <artifactId>custos-signer-sdk-core</artifactId>
</dependency>
<dependency>
<groupId>org.apache.custos</groupId>
<artifactId>custos-signer-sdk-dbstore</artifactId>
</dependency>
```

**With Vault Storage**:

```xml

<dependency>
    <groupId>org.apache.custos</groupId>
    <artifactId>custos-signer-sdk-core</artifactId>
</dependency>
<dependency>
<groupId>org.apache.custos</groupId>
<artifactId>custos-signer-sdk-vaultstore</artifactId>
</dependency>
```

## SDK Usage

### Basic Usage

```java
// Configure the SDK (tenant-id at SDK level)
SdkConfiguration config = new SdkConfiguration.Builder()
                .tenantId("airavata-prod")
                .signerServiceAddress("signer.custos.org:9095")
                .tlsEnabled(true)
                .addClient("hpcA", "airavata-hpcA", "client-secret")
                .addClient("hpcB", "airavata-hpcB", "client-secret")
                .build();

// Create SSH client
SshClient sshClient = SshClient.builder()
        .configuration(config)
        .build();

// Open SSH session
try(
SshSession session = sshClient.openSession("hpcA", "user", 3600, oidcToken)){
// Execute commands
CommandResult result = session.exec("ls -la /home/user");
    System.out.

println(result.getStdout());

        // Upload files
        session.

upload("/local/data.txt","/remote/data.txt",0644);

// Download files
    session.

download("/remote/output.txt","/local/output.txt");
}
```

### Request Certificate Materials

Get all SSH connection materials (private key, public key, certificate, and metadata) to use with your preferred SSH client or library:

```java
// Request certificate materials
CertificateMaterials materials = sshClient.requestCertificateMaterials(
    "hpcA", "user", 3600, oidcToken
);

// Use materials with OpenSSH command-line or other SSH libraries
java.nio.file.Path keyFile = java.nio.file.Files.createTempFile("ssh-key", "");
java.nio.file.Path certFile = java.nio.file.Files.createTempFile("ssh-cert", "");

java.nio.file.Files.write(keyFile, materials.getPrivateKeyPem().getBytes());
java.nio.file.Files.write(certFile, materials.getOpensshCert().getBytes());

// Use with OpenSSH
// ssh -i keyFile -o CertificateFile=certFile user@host

// Or use KeyPair object directly with other SSH libraries
KeyPair keyPair = materials.getKeyPair();
byte[] certBytes = materials.getCertBytes();
```

**OpenSSH Certificate Format**: The `opensshCert` field contains the certificate in OpenSSH string format:
```
ssh-ed25519-cert-v01@openssh.com <base64-encoded-certificate> <comment>
```
This format can be written directly to a file and used with OpenSSH command-line tools using the `-o CertificateFile=` option.

### Configuration File

Create `custos-sdk.yml`:

```yaml
sdk:
  tenant-id: airavata-prod  # Single tenant-id for all clients
  signer:
    address: signer.custos.org:9095
    tls:
      enabled: true
      trust-store: classpath:signer-ca.p12
      trust-store-password: ${TRUSTSTORE_PASSWORD}
  keystore:
    backend: in-memory
  clients:
    - alias: hpcA
      client-id: airavata-hpcA
      client-secret: ${HPC_A_SECRET}
    - alias: hpcB
      client-id: airavata-hpcB
      client-secret: ${HPC_B_SECRET}
```

### Tenant Architecture

The SDK uses a **single tenant-id per SDK instance** architecture:

- **One SDK = One Tenant**: Each SDK instance serves one tenant only
- **Cleaner Configuration**: No repetition of tenant-id across clients
- **Security Boundary**: Clear isolation between tenants

This design ensures that developers work within a single tenant context, making the SDK more intuitive and secure.

**Note**: The SDK supports multiple clients per tenant (e.g., different HPC systems). While this adds some configuration
complexity, it provides flexibility for complex enterprise scenarios. Future versions may include convenience methods
for single-client use cases.

## gRPC API

### Sign Certificate

```protobuf
rpc Sign(SignRequest) returns (SignResponse);

message SignRequest {
  string tenant_id = 1;
  string client_id = 2;
  string principal = 3;
  int32 ttl_seconds = 4;
  bytes public_key = 5;
  string user_access_token = 6;
}

message SignResponse {
  bytes certificate = 1;
  int64 serial_number = 2;
  int64 valid_after = 3;
  int64 valid_before = 4;
  string ca_fingerprint = 5;
  string target_host = 6;
  int32 target_port = 7;
  string target_username = 8;
}
```

### Revoke Certificate

```protobuf
rpc Revoke(RevokeRequest) returns (RevokeResponse);

message RevokeRequest {
  string tenant_id = 1;
  string client_id = 2;
  int64 serial_number = 3;      // Optional
  string key_id = 4;           // Optional
  string ca_fingerprint = 5;   // Optional
  string reason = 6;
}
```

### Get JWKS

```protobuf
rpc GetJWKS(GetJWKSRequest) returns (GetJWKSResponse);

message GetJWKSResponse {
  string current_key = 1;
  string next_key = 2;
  string current_fingerprint = 3;
  string next_fingerprint = 4;
  int64 rotation_scheduled_at = 5;
}
```

## Host Configuration

### Ansible Role

Use the included Ansible role to configure SSH servers:

```yaml
---
- hosts: compute_nodes
  become: yes
  roles:
    - role: custos-signer-trust
      vars:
        tenant_id: airavata-prod
        client_id: airavata-hpcA
        ca_public_key: "{{ lookup('file', 'ca-keys/hpcA.pub') }}"
        signer_service_url: https://signer.custos.org
```

### Manual Configuration

1. **Install CA Public Key**:
   ```bash
   mkdir -p /etc/ssh/ca_keys
   cp ca-public-key.pub /etc/ssh/ca_keys/airavata-prod_airavata-hpcA.pub
   chmod 644 /etc/ssh/ca_keys/airavata-prod_airavata-hpcA.pub
   ```

2. **Configure SSH Daemon**:
   ```bash
   echo "TrustedUserCAKeys /etc/ssh/ca_keys/*.pub" >> /etc/ssh/sshd_config
   echo "RevokedKeys /etc/ssh/revoked.krl" >> /etc/ssh/sshd_config
   systemctl reload sshd
   ```

3. **Set up KRL Fetch**:
   ```bash
   curl -o /usr/local/bin/custos-fetch-krl.sh https://signer.custos.org/api/krl/airavata-prod/airavata-hpcA
   chmod +x /usr/local/bin/custos-fetch-krl.sh
   
   # Add to crontab
   echo "*/5 * * * * /usr/local/bin/custos-fetch-krl.sh" | crontab -
   ```

## Security Considerations

### Certificate Lifecycle

- **Default TTL**: 24 hours (configurable per client)
- **Maximum TTL**: Enforced by policy (default 24 hours)
- **Key Types**: Ed25519 (preferred), RSA, ECDSA
- **CA Rotation**: Every 90 days with 2-hour overlap

### Authentication

- **Client Authentication**: Client ID + Secret in gRPC metadata
- **User Authentication**: OIDC tokens (CILogon integration)
- **Transport Security**: TLS 1.2+ for gRPC (configurable)

### Audit and Compliance

- **Database Audit**: All certificate operations logged
- **Structured Logging**: JSON logs for SIEM integration
- **Revocation Tracking**: Complete revocation audit trail
- **Policy Enforcement**: Dynamic policy validation

## Monitoring and Operations

### Health Checks

```bash
# Service health
curl http://localhost:8084/actuator/health

# Metrics
curl http://localhost:8084/actuator/metrics

# Prometheus metrics
curl http://localhost:8084/actuator/prometheus
```

### Key Metrics

- `custos.signer.certificates.issued` - Certificates issued
- `custos.signer.certificates.revoked` - Certificates revoked
- `custos.signer.ca.rotations` - CA rotations performed
- `custos.signer.requests.duration` - Request processing time

### Logging

Structured JSON logs are written to:

- Console (for development)
- `/var/log/custos/signer-service.log` (for production)

Key log events:

- Certificate issuance
- Certificate revocation
- CA rotations
- Authentication failures
- Policy violations

## Development

### Project Structure

```
signer/
├── signer-service/          # Spring Boot gRPC service
│   ├── src/main/java/
│   │   ├── config/         # Spring configuration
│   │   ├── grpc/           # gRPC service implementation
│   │   ├── ca/             # Certificate signing logic
│   │   ├── vault/          # OpenBao integration
│   │   ├── policy/         # Policy enforcement
│   │   ├── auth/           # Authentication
│   │   └── audit/          # Audit logging
│   ├── src/main/resources/
│   │   ├── db/migration/   # Flyway migrations
│   │   └── ansible-role/   # Ansible automation
│   └── Dockerfile
├── signer-sdk-core/         # Core SDK (lightweight)
│   ├── src/main/java/
│   │   ├── client/         # gRPC client
│   │   ├── config/         # Configuration
│   │   ├── keystore/       # KeyStoreProvider interface + InMemoryKeyStore
│   │   └── ssh/            # SSH operations
│   └── pom.xml
├── signer-sdk-dbstore/      # Optional database keystore
│   └── src/main/java/
│       └── keystore/dbstore/ # DatabaseKeyStore implementation
└── signer-sdk-vaultstore/   # Optional vault keystore
    └── src/main/java/
        └── keystore/vaultstore/ # VaultKeyStore implementation
```

### Running Tests

```bash
# Unit tests
mvn test

# Integration tests (requires Docker)
mvn verify -P integration-tests

# SDK tests
cd signer-sdk-core
mvn test

# Optional module tests
cd ../signer-sdk-dbstore
mvn test

cd ../signer-sdk-vaultstore
mvn test
```

### Building Distribution

```bash
# Build service distribution
cd signer-service
mvn clean package

# Distribution will be in target/
ls target/apache-airavata-custos-signer-service-*.tar.gz
```

## Troubleshooting

### Common Issues

1. **Certificate Authentication Fails**:
    - Check CA public key is installed on target host
    - Verify certificate hasn't expired
    - Check KRL for revoked certificates

2. **gRPC Connection Issues**:
    - Verify TLS configuration
    - Check client credentials
    - Ensure service is running on correct port

3. **Database Connection Issues**:
    - Verify database credentials
    - Check Flyway migrations completed
    - Ensure database exists

4. **Vault Connection Issues**:
    - Verify Vault token is valid
    - Check Vault address configuration
    - Ensure CA key paths exist

### Debug Mode

Enable debug logging:

```yaml
logging:
  level:
    org.apache.custos.signer: debug
    audit: info
```

### Performance Tuning

- **Database**: Optimize connection pool settings
- **Vault**: Use connection pooling for high throughput
- **gRPC**: Tune keepalive and message size limits
- **JVM**: Adjust heap size and GC settings

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Support

- **Documentation**: [Custos Documentation](https://custos.apache.org)
- **Issues**: [GitHub Issues](https://github.com/apache/airavata-custos/issues)
- **Mailing List**: [dev@custos.apache.org](mailto:dev@custos.apache.org)
