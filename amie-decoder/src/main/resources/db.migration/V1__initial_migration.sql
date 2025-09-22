-- =================================================================
-- Flyway Migration Script for AMIE Packet Decoder
-- Target Database: MySQL / MariaDB
-- =================================================================

SET NAMES utf8mb4;
SET
time_zone = '+00:00';

-- -----------------------------
-- TABLE: packets
-- Stores each unique AMIE packet received from the polling endpoint
-- -----------------------------
CREATE TABLE packets
(
    id           VARCHAR(255) NOT NULL,
    amie_id      BIGINT       NOT NULL, -- packet_rec_id from AMIE
    type         VARCHAR(64)  NOT NULL, -- eg. request_account_create
    status       VARCHAR(32)  NOT NULL, -- NEW, DECODED, PROCESSED, FAILED
    raw_json     JSON         NOT NULL,
    received_at  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    decoded_at   TIMESTAMP(6) NULL,
    processed_at TIMESTAMP(6) NULL,
    retries      INT          NOT NULL DEFAULT 0,
    last_error   TEXT NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uq_packets_amie_id (amie_id),
    KEY          idx_packets_status (status),
    KEY          idx_packets_received_at (received_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
    last_error  TEXT NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_events_packet FOREIGN KEY (packet_id) REFERENCES packets (id) ON DELETE CASCADE,
    UNIQUE KEY uq_events_packet_type (packet_id, type),

    KEY         idx_events_status (status),
    KEY         idx_events_packet_id (packet_id),
    KEY         idx_events_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

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
    summary     VARCHAR(255) NOT NULL,
    detail      TEXT NULL,

    PRIMARY KEY (id),
    CONSTRAINT fk_errors_packet FOREIGN KEY (packet_id) REFERENCES packets (id) ON DELETE SET NULL,
    CONSTRAINT fk_errors_event FOREIGN KEY (event_id) REFERENCES processing_events (id) ON DELETE SET NULL,

    CHECK (packet_id IS NOT NULL OR event_id IS NOT NULL),

    KEY         idx_errors_packet_id (packet_id),
    KEY         idx_errors_event_id (event_id),
    KEY         idx_errors_occurred_at (occurred_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
