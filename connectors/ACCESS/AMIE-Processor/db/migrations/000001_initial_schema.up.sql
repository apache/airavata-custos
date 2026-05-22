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

-- AMIE-protocol-specific tables. Domain entities (User/Project/etc.) live in
-- core. This migration creates only the AMIE-Processor's local state.

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
    UNIQUE KEY uq_amie_packets_amie_id (amie_id),
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
    CONSTRAINT fk_amie_errors_packet FOREIGN KEY (packet_id) REFERENCES amie_packets (id) ON DELETE SET NULL,
    CONSTRAINT fk_amie_errors_event FOREIGN KEY (event_id) REFERENCES amie_processing_events (id) ON DELETE SET NULL,
    KEY idx_amie_errors_packet_id (packet_id),
    KEY idx_amie_errors_event_id (event_id),
    KEY idx_amie_errors_occurred_at (occurred_at)
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
    CONSTRAINT fk_amie_audit_packet FOREIGN KEY (packet_id) REFERENCES amie_packets (id) ON DELETE CASCADE,
    CONSTRAINT fk_amie_audit_event FOREIGN KEY (event_id) REFERENCES amie_processing_events (id) ON DELETE SET NULL,
    KEY idx_amie_audit_packet_id (packet_id),
    KEY idx_amie_audit_action (action),
    KEY idx_amie_audit_created_at (created_at)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- AMIE-side DN registry. AMIE delivers DnList fields that contain DNs across
-- the federation (not just this site's downstream cluster).
--
-- user_id references core users.id by value; no FK because connector schema
-- stays independent of core.
CREATE TABLE IF NOT EXISTS amie_user_dns
(
    id         VARCHAR(255) NOT NULL,
    user_id    VARCHAR(255) NOT NULL,
    dn         VARCHAR(512) NOT NULL,
    created_at TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_amie_user_dns_dn (dn),
    KEY idx_amie_user_dns_user (user_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
