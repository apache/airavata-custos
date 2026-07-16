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

-- entity_type records the kind of resource the row is about ("user", "role",
-- "compute_cluster_user", "packet", etc.). source names the subsystem that
-- produced the event.
CREATE TABLE IF NOT EXISTS audit_events
(
    id             VARCHAR(255) NOT NULL,
    event_type     VARCHAR(128) NOT NULL,
    event_time     TIMESTAMP(6) NOT NULL,
    entity_id      VARCHAR(255) NOT NULL,
    entity_type    VARCHAR(64)  NOT NULL DEFAULT '',
    details        TEXT         NOT NULL,
    source         VARCHAR(64)  NOT NULL DEFAULT 'core',
    trace_id       CHAR(32) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
    span_id        CHAR(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
    parent_span_id CHAR(16) CHARACTER SET ascii COLLATE ascii_bin NOT NULL DEFAULT '',
    PRIMARY KEY (id),
    KEY idx_audit_events_entity (entity_id, event_time),
    KEY idx_audit_events_entity_type (entity_type),
    KEY idx_audit_events_type (event_type, event_time),
    KEY idx_audit_events_time (event_time),
    KEY idx_audit_events_source (source),
    KEY idx_audit_events_trace (trace_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
