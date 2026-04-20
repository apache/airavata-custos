-- Licensed to the Apache Software Foundation (ASF) under one
-- or more contributor license agreements.  See the NOTICE file
-- distributed with this work for additional information
-- regarding copyright ownership.  The ASF licenses this file
-- to you under the Apache License, Version 2.0 (the
-- "License"); you may not use this file except in compliance
-- with the License.  You may obtain a copy of the License at
--
--   http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing,
-- software distributed under the License is distributed on an
-- "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
-- KIND, either express or implied.  See the License for the
-- specific language governing permissions and limitations
-- under the License.

SET NAMES utf8mb4;
SET time_zone = '+00:00';

CREATE TABLE IF NOT EXISTS persons
(
    id               VARCHAR(255) NOT NULL,
    access_global_id VARCHAR(255) NOT NULL,
    first_name       VARCHAR(255) NOT NULL,
    last_name        VARCHAR(255) NOT NULL,
    email            VARCHAR(255) NOT NULL,
    organization     VARCHAR(255) NULL,
    org_code         VARCHAR(255) NULL,
    nsf_status_code  VARCHAR(32)  NULL,
    is_active        BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at       TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_persons_amie_global_id (access_global_id),
    KEY idx_persons_active (is_active),
    KEY idx_persons_email (email)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS person_global_ids
(
    id        BIGINT       NOT NULL AUTO_INCREMENT,
    person_id VARCHAR(255) NOT NULL,
    global_id VARCHAR(255) NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_global_ids_person FOREIGN KEY (person_id) REFERENCES persons (id) ON DELETE CASCADE,
    UNIQUE KEY uq_person_global_id (global_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS person_dns
(
    id        BIGINT       NOT NULL AUTO_INCREMENT,
    person_id VARCHAR(255) NOT NULL,
    dn        VARCHAR(512) NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_dns_person FOREIGN KEY (person_id) REFERENCES persons (id) ON DELETE CASCADE,
    UNIQUE KEY uq_person_dn (person_id, dn)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS cluster_accounts
(
    id         VARCHAR(255) NOT NULL,
    person_id  VARCHAR(255) NOT NULL,
    username   VARCHAR(255) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_accounts_username (username),
    CONSTRAINT fk_accounts_person FOREIGN KEY (person_id) REFERENCES persons (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS projects
(
    id           VARCHAR(255) NOT NULL,
    grant_number VARCHAR(255) NOT NULL,
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at   TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_projects_grant_number (grant_number),
    KEY idx_projects_active (is_active)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS project_memberships
(
    id                 VARCHAR(255) NOT NULL,
    project_id         VARCHAR(255) NOT NULL,
    cluster_account_id VARCHAR(255) NOT NULL,
    role               VARCHAR(32)  NULL,
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
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS amie_packets
(
    id           VARCHAR(255) NOT NULL,
    amie_id      BIGINT       NOT NULL,
    type         VARCHAR(64)  NOT NULL,
    status       VARCHAR(32)  NOT NULL,
    raw_json     TEXT         NOT NULL,
    received_at  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    decoded_at   TIMESTAMP(6) NULL,
    processed_at TIMESTAMP(6) NULL,
    retries      INT          NOT NULL DEFAULT 0,
    last_error   TEXT         NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uq_packets_amie_id (amie_id),
    KEY idx_amie_packets_status (status),
    KEY idx_amie_packets_received_at (received_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS amie_processing_events
(
    id            VARCHAR(255) NOT NULL,
    packet_id     VARCHAR(255) NOT NULL,
    type          VARCHAR(64)  NOT NULL,
    status        VARCHAR(32)  NOT NULL,
    attempts      INT          NOT NULL DEFAULT 0,
    payload       LONGBLOB     NOT NULL,
    created_at    TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    started_at    TIMESTAMP(6) NULL,
    finished_at   TIMESTAMP(6) NULL,
    last_error    TEXT         NULL,
    next_retry_at TIMESTAMP(6) NULL DEFAULT NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_amie_events_packet FOREIGN KEY (packet_id) REFERENCES amie_packets (id) ON DELETE CASCADE,
    UNIQUE KEY uq_amie_events_packet_type (packet_id, type),
    KEY idx_amie_events_status (status),
    KEY idx_amie_events_packet_id (packet_id),
    KEY idx_amie_events_type (type),
    KEY idx_amie_events_next_retry_at (next_retry_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS amie_processing_errors
(
    id          BIGINT       NOT NULL AUTO_INCREMENT,
    packet_id   VARCHAR(255) NULL,
    event_id    VARCHAR(255) NULL,
    occurred_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    summary     TEXT         NOT NULL,
    detail      TEXT         NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_errors_amie_packet FOREIGN KEY (packet_id) REFERENCES amie_packets (id) ON DELETE SET NULL,
    CONSTRAINT fk_errors_amie_event FOREIGN KEY (event_id) REFERENCES amie_processing_events (id) ON DELETE SET NULL,
    KEY idx_errors_amie_packet_id (packet_id),
    KEY idx_errors_amie_event_id (event_id),
    KEY idx_errors_amie_occurred_at (occurred_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS amie_audit_log
(
    id          BIGINT       NOT NULL AUTO_INCREMENT,
    packet_id   VARCHAR(255) NOT NULL,
    event_id    VARCHAR(255) NULL,
    action      VARCHAR(64)  NOT NULL,
    entity_type VARCHAR(64)  NOT NULL,
    entity_id   VARCHAR(255) NULL,
    summary     TEXT         NULL,
    created_at  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT fk_audit_packet FOREIGN KEY (packet_id) REFERENCES amie_packets (id) ON DELETE CASCADE,
    CONSTRAINT fk_audit_event FOREIGN KEY (event_id) REFERENCES amie_processing_events (id) ON DELETE SET NULL,
    KEY idx_audit_packet_id (packet_id),
    KEY idx_audit_action (action),
    KEY idx_audit_created_at (created_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
