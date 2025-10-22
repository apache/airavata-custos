-- =================================================================
-- Flyway Migration Script for Custos Signer Service
-- Target Database: MySQL / MariaDB
-- =================================================================

SET NAMES utf8mb4;
SET time_zone = '+00:00';

-- -----------------------------
-- TABLE: client_ssh_configs
-- Maps clientId to target system and security policies
-- Designed for future expansion to one-to-many via separate client_targets table
-- -----------------------------
CREATE TABLE client_ssh_configs
(
    tenant_id                  VARCHAR(255) NOT NULL,
    client_id                  VARCHAR(255) NOT NULL,
    client_secret              VARCHAR(512) NOT NULL,               -- Encrypted client secret
    target_host                VARCHAR(255) NOT NULL,               -- SSH target hostname
    target_port                INT          NOT NULL DEFAULT 22,    -- SSH port
    max_ttl_seconds            INT          NOT NULL DEFAULT 86400, -- Max certificate TTL (24h)
    allowed_key_types          JSON         NOT NULL,               -- e.g., ["ed25519", "rsa", "ecdsa"]
    source_address_restriction VARCHAR(255) NULL,                   -- CIDR or IP for source-address critical option
    critical_options           JSON         NULL,                   -- Additional critical options
    extensions                 JSON         NULL,                   -- Certificate extensions
    enabled                    BOOLEAN      NOT NULL DEFAULT TRUE,  -- Active/inactive flag
    created_at                 TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at                 TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (tenant_id, client_id),
    INDEX idx_client_configs_enabled (enabled),
    INDEX idx_client_configs_target_host (target_host)

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- -----------------------------
-- TABLE: certificate_issuance_logs
-- Comprehensive audit log for certificate issuance
-- -----------------------------
CREATE TABLE certificate_issuance_logs
(
    id                     BIGINT       NOT NULL AUTO_INCREMENT,
    tenant_id              VARCHAR(255) NOT NULL,
    client_id              VARCHAR(255) NOT NULL,
    serial_number          BIGINT       NOT NULL,
    key_id                 VARCHAR(255) NOT NULL, -- SHA256 of public key
    principal              VARCHAR(255) NOT NULL,
    public_key_fingerprint VARCHAR(255) NOT NULL,
    ca_fingerprint         VARCHAR(255) NOT NULL,
    valid_after            TIMESTAMP(6) NOT NULL,
    valid_before           TIMESTAMP(6) NOT NULL,
    issued_at              TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    source_ip              VARCHAR(45)  NULL,     -- Client IP from gRPC context (IPv4/IPv6)
    user_access_token_hash VARCHAR(255) NULL,     -- Token hash for correlation
    request_metadata       JSON         NULL,     -- Additional context

    PRIMARY KEY (id),
    UNIQUE KEY uq_issuance_serial (serial_number),
    INDEX idx_issuance_tenant_client (tenant_id, client_id),
    INDEX idx_issuance_keyid (key_id),
    INDEX idx_issuance_principal (principal),
    INDEX idx_issuance_ca_fingerprint (ca_fingerprint),
    INDEX idx_issuance_issued_at (issued_at),
    INDEX idx_issuance_valid_range (valid_after, valid_before)

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- -----------------------------
-- TABLE: revocation_events
-- Tracks all certificate and CA revocation events
-- -----------------------------
CREATE TABLE revocation_events
(
    id             BIGINT       NOT NULL AUTO_INCREMENT,
    tenant_id      VARCHAR(255) NOT NULL,
    client_id      VARCHAR(255) NOT NULL,
    serial_number  BIGINT       NULL,     -- Revoke by serial (nullable for CA revocation)
    key_id         VARCHAR(255) NULL,     -- Revoke by key ID (nullable for CA revocation)
    ca_fingerprint VARCHAR(255) NULL,     -- Revoke all certs from this CA (nullable for single cert)
    revoked_at     TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    reason         VARCHAR(255) NOT NULL,
    revoked_by     VARCHAR(255) NOT NULL, -- Admin or system identifier

    PRIMARY KEY (id),
    INDEX idx_revocation_tenant_client (tenant_id, client_id),
    INDEX idx_revocation_serial (serial_number),
    INDEX idx_revocation_keyid (key_id),
    INDEX idx_revocation_ca_fingerprint (ca_fingerprint),
    INDEX idx_revocation_revoked_at (revoked_at)

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

-- -----------------------------
-- TABLE: ssh_key_store
-- Optional table for DatabaseKeyStore backend (disabled by default)
-- Stores encrypted SSH private keys for persistent key storage
-- -----------------------------
CREATE TABLE ssh_key_store
(
    id                    BIGINT       NOT NULL AUTO_INCREMENT,
    user_id               VARCHAR(255) NOT NULL,
    context_id            VARCHAR(255) NOT NULL,
    encrypted_private_key TEXT         NOT NULL, -- AES-256-GCM encrypted private key
    public_key            TEXT         NOT NULL, -- Public key in OpenSSH format
    expires_at            TIMESTAMP(6) NOT NULL,
    created_at            TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),

    PRIMARY KEY (id),
    UNIQUE KEY uq_key_store_user_context (user_id, context_id),
    INDEX idx_key_store_expires_at (expires_at)

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;
