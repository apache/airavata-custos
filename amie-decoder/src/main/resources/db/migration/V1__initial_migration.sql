-- =================================================================
-- Flyway Migration Script for AMIE Packet Decoder
-- Target Database: MySQL / MariaDB
-- =================================================================

SET NAMES utf8mb4;
SET time_zone = '+00:00';

-- -----------------------------
-- TABLE: persons
-- Stores a unique record for each person received from AMIE.
-- -----------------------------
CREATE TABLE persons
(
    id               VARCHAR(255) NOT NULL, -- Local user ID - UserPersonID
    access_global_id VARCHAR(255) NOT NULL, -- UserGlobalID from AMIE
    first_name       VARCHAR(255) NOT NULL,
    last_name        VARCHAR(255) NOT NULL,
    email            VARCHAR(255) NOT NULL,
    organization     VARCHAR(255) NULL,
    org_code         VARCHAR(255) NULL,
    nsf_status_code  VARCHAR(32)  NULL,
    created_at       TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at       TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (id),
    UNIQUE KEY uq_persons_amie_global_id (access_global_id)

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;


-- -----------------------------
-- TABLE: person_dns
-- Stores the list of Distinguished Names (DNs) for a person.
-- -----------------------------
CREATE TABLE person_dns
(
    id        BIGINT       NOT NULL AUTO_INCREMENT,
    person_id VARCHAR(255) NOT NULL,
    dn        VARCHAR(512) NOT NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_dns_person FOREIGN KEY (person_id) REFERENCES persons (id) ON DELETE CASCADE,
    UNIQUE KEY uq_person_dn (person_id, dn)

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;


-- -----------------------------
-- TABLE: cluster_accounts
-- Stores the actual provisioned usernames on the cluster.
-- -----------------------------
CREATE TABLE cluster_accounts
(
    id         VARCHAR(255) NOT NULL,
    person_id  VARCHAR(255) NOT NULL,
    username   VARCHAR(255) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (id),
    UNIQUE KEY uq_accounts_username (username),
    CONSTRAINT fk_accounts_person FOREIGN KEY (person_id) REFERENCES persons (id) ON DELETE CASCADE

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;


-- -----------------------------
-- TABLE: projects
-- Stores unique information about each project/allocation
-- -----------------------------
CREATE TABLE projects
(
    id           VARCHAR(255) NOT NULL, -- AMIE ProjectID (local), e.g., "PRJ-TRA258601"
    grant_number VARCHAR(255) NOT NULL,
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at   TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (id),
    UNIQUE KEY uq_projects_grant_number (grant_number),
    KEY idx_projects_active (is_active)

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;


-- -----------------------------
-- TABLE: project_memberships
-- Links cluster accounts to projects with roles (user memberships on projects).
-- -----------------------------
CREATE TABLE project_memberships
(
    id                 VARCHAR(255) NOT NULL,
    project_id         VARCHAR(255) NOT NULL, -- FK to projects.id
    cluster_account_id VARCHAR(255) NOT NULL, -- FK to cluster_accounts.id
    role               VARCHAR(32)  NULL,     -- PI, USER
    is_active          BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at         TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at         TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (id),
    CONSTRAINT fk_membership_project FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE,
    CONSTRAINT fk_membership_account FOREIGN KEY (cluster_account_id) REFERENCES cluster_accounts (id) ON DELETE CASCADE,
    UNIQUE KEY uq_project_account (project_id, cluster_account_id),
    KEY idx_memberships_project (project_id),
    KEY idx_memberships_account (cluster_account_id),
    KEY idx_memberships_active (is_active)

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;


-- Stores each unique AMIE packet received from the polling endpoint
-- -----------------------------
CREATE TABLE packets
(
    id           VARCHAR(255) NOT NULL,
    amie_id      BIGINT       NOT NULL, -- packet_rec_id from AMIE
    type         VARCHAR(64)  NOT NULL, -- eg. request_account_create
    status       VARCHAR(32)  NOT NULL, -- NEW, DECODED, PROCESSED, FAILED
    raw_json     TEXT         NOT NULL,
    received_at  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    decoded_at   TIMESTAMP(6) NULL,
    processed_at TIMESTAMP(6) NULL,
    retries      INT          NOT NULL DEFAULT 0,
    last_error   TEXT         NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uq_packets_amie_id (amie_id),
    KEY idx_packets_status (status),
    KEY idx_packets_received_at (received_at)

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;


-- -----------------------------
-- TABLE: processing_events
-- Individual processing steps for a given packet
-- -----------------------------
CREATE TABLE processing_events
(
    id          VARCHAR(255) NOT NULL,
    packet_id   VARCHAR(255) NOT NULL,
    type        VARCHAR(64)  NOT NULL, -- DECODE_PACKET, SYNC_USER, ACK, etc.
    status      VARCHAR(32)  NOT NULL, -- PENDING, RUNNING, DONE, FAILED
    attempts    INT          NOT NULL DEFAULT 0,
    payload     LONGBLOB     NOT NULL, -- Protobuf serialized DTO
    created_at  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    started_at  TIMESTAMP(6) NULL,
    finished_at TIMESTAMP(6) NULL,
    last_error  TEXT         NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_events_packet FOREIGN KEY (packet_id) REFERENCES packets (id) ON DELETE CASCADE,
    UNIQUE KEY uq_events_packet_type (packet_id, type),

    KEY idx_events_status (status),
    KEY idx_events_packet_id (packet_id),
    KEY idx_events_type (type)

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;


-- -----------------------------
-- TABLE: processing_errors
-- Processing error logs
-- -----------------------------
CREATE TABLE processing_errors
(
    id          BIGINT       NOT NULL AUTO_INCREMENT,
    packet_id   VARCHAR(255) NULL,
    event_id    VARCHAR(255) NULL,
    occurred_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    summary     TEXT         NOT NULL,
    detail      TEXT         NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_errors_packet FOREIGN KEY (packet_id) REFERENCES packets (id) ON DELETE SET NULL,
    CONSTRAINT fk_errors_event FOREIGN KEY (event_id) REFERENCES processing_events (id) ON DELETE SET NULL,

    KEY idx_errors_packet_id (packet_id),
    KEY idx_errors_event_id (event_id),
    KEY idx_errors_occurred_at (occurred_at)

) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;
