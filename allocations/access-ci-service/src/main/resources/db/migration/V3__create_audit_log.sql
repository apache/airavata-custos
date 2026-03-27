--  Licensed to the Apache Software Foundation (ASF) under one or more
--  contributor license agreements.  See the NOTICE file distributed with
--  this work for additional information regarding copyright ownership.
--  The ASF licenses this file to You under the Apache License, Version 2.0
--  (the "License"); you may not use this file except in compliance with
--  the License.  You may obtain a copy of the License at
--
--      http://www.apache.org/licenses/LICENSE-2.0
--
--  Unless required by applicable law or agreed to in writing, software
--  distributed under the License is distributed on an "AS IS" BASIS,
--  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
--  See the License for the specific language governing permissions and
--  limitations under the License.

-- Audit log for handler actions during AMIE packet processing.

CREATE TABLE amie_audit_log
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
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci;
