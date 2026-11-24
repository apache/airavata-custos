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

-- Who listens to what. A connector saves its rows when it subscribes, and they
-- stay across system restarts, so an event fired before the connector loads still
-- gets an event delivery row.
CREATE TABLE IF NOT EXISTS event_subscriptions
(
    subscriber VARCHAR(64)    NOT NULL,
    event_type VARCHAR(128)   NOT NULL,
    created_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (subscriber, event_type)
);

-- One row per published event, written in the same transaction as the change
-- it came from, and only when there's a subscriber.
-- The source is what component published it, same values as `audit_events.source` set by the bus.
CREATE TABLE IF NOT EXISTS events
(
    id         VARCHAR(255)   NOT NULL,
    event_type VARCHAR(128)   NOT NULL,
    payload    JSONB          NOT NULL,
    source     VARCHAR(64)    NOT NULL,
    trace_id   CHAR(32)       NOT NULL CHECK (trace_id <> ''),
    span_id    CHAR(16)       NOT NULL CHECK (span_id <> ''),
    created_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_events_type ON events (event_type, created_at);
CREATE INDEX IF NOT EXISTS idx_events_source ON events (source);
CREATE INDEX IF NOT EXISTS idx_events_time ON events (created_at);
CREATE INDEX IF NOT EXISTS idx_events_trace ON events (trace_id);

-- One row per event subscription. The worker updates it after each try.
-- Failed tries go to audit_events.
CREATE TABLE IF NOT EXISTS event_deliveries
(
    id          VARCHAR(255)   NOT NULL,
    event_id    VARCHAR(255)   NOT NULL,
    subscriber  VARCHAR(64)    NOT NULL,
    status      VARCHAR(32)    NOT NULL,
    attempts    INT            NOT NULL DEFAULT 0,
    next_run_at TIMESTAMPTZ(6) NOT NULL,
    last_error  TEXT           NULL,
    created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    finished_at TIMESTAMPTZ(6) NULL,
    PRIMARY KEY (id),
    CONSTRAINT fk_event_deliveries_event FOREIGN KEY (event_id) REFERENCES events (id),
    CONSTRAINT chk_event_deliveries_status CHECK (status IN ('PENDING', 'SUCCEEDED', 'FAILED'))
);
CREATE INDEX IF NOT EXISTS idx_event_deliveries_due ON event_deliveries (status, next_run_at);
CREATE INDEX IF NOT EXISTS idx_event_deliveries_event ON event_deliveries (event_id);
