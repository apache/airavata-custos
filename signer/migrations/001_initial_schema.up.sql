SET NAMES utf8mb4;
SET time_zone = '+00:00';

CREATE TABLE IF NOT EXISTS client_ssh_configs
(
    tenant_id                  VARCHAR(255) NOT NULL,
    client_id                  VARCHAR(255) NOT NULL,
    client_secret              VARCHAR(512) NOT NULL,
    target_host                VARCHAR(255) NOT NULL,
    target_port                INT          NOT NULL DEFAULT 22,
    max_ttl_seconds            INT          NOT NULL DEFAULT 86400,
    allowed_key_types          JSON         NOT NULL,
    source_address_restriction VARCHAR(255) NULL,
    denied_extensions          JSON         NULL,
    principal_source           VARCHAR(20)  NOT NULL DEFAULT 'noop',
    enabled                    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at                 TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at                 TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6)
                                                     ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (tenant_id, client_id),
    INDEX idx_client_configs_enabled (enabled),
    INDEX idx_client_configs_target_host (target_host)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS certificate_issuance_logs
(
    id                     BIGINT       NOT NULL AUTO_INCREMENT,
    tenant_id              VARCHAR(255) NOT NULL,
    client_id              VARCHAR(255) NOT NULL,
    serial_number          BIGINT       NOT NULL,
    key_id                 VARCHAR(255) NOT NULL,
    principal              VARCHAR(255) NOT NULL,
    user_email             VARCHAR(255) NULL,
    public_key_fingerprint VARCHAR(255) NOT NULL,
    ca_fingerprint         VARCHAR(255) NOT NULL,
    valid_after            TIMESTAMP(6) NOT NULL,
    valid_before           TIMESTAMP(6) NOT NULL,
    issued_at              TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    source_ip              VARCHAR(45)  NULL,
    granted_extensions     JSON         NOT NULL,
    force_command          TEXT         NULL,
    user_access_token_hash VARCHAR(255) NULL,
    request_metadata       JSON         NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uq_issuance_serial (serial_number),
    INDEX idx_issuance_tenant_client (tenant_id, client_id),
    INDEX idx_issuance_keyid (key_id),
    INDEX idx_issuance_principal (principal),
    INDEX idx_issuance_user_email (user_email),
    INDEX idx_issuance_ca_fingerprint (ca_fingerprint),
    INDEX idx_issuance_issued_at (issued_at),
    INDEX idx_issuance_valid_range (valid_after, valid_before)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS revocation_events
(
    id             BIGINT       NOT NULL AUTO_INCREMENT,
    tenant_id      VARCHAR(255) NOT NULL,
    client_id      VARCHAR(255) NOT NULL,
    serial_number  BIGINT       NULL,
    key_id         VARCHAR(255) NULL,
    ca_fingerprint VARCHAR(255) NULL,
    revoked_at     TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    reason         VARCHAR(255) NOT NULL,
    revoked_by     VARCHAR(255) NOT NULL,

    PRIMARY KEY (id),
    INDEX idx_revocation_tenant_client (tenant_id, client_id),
    INDEX idx_revocation_serial (serial_number),
    INDEX idx_revocation_keyid (key_id),
    INDEX idx_revocation_ca_fingerprint (ca_fingerprint),
    INDEX idx_revocation_revoked_at (revoked_at)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;
