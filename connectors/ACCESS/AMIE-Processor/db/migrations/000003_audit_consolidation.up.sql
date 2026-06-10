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

-- Retire amie_audit_log. AMIE now writes audit rows to the core audit_events
-- table; connector-specific references (packet_id, event_id) move to a
-- side table joined on audit_event_id.

DROP TABLE IF EXISTS amie_audit_log;

CREATE TABLE IF NOT EXISTS amie_audit_extras
(
    audit_event_id VARCHAR(255) NOT NULL,
    packet_id      VARCHAR(255) NOT NULL,
    event_id       VARCHAR(255) NULL,
    PRIMARY KEY (audit_event_id),
    CONSTRAINT fk_amie_audit_extras_event  FOREIGN KEY (audit_event_id) REFERENCES audit_events(id) ON DELETE CASCADE,
    CONSTRAINT fk_amie_audit_extras_packet FOREIGN KEY (packet_id) REFERENCES amie_packets(id) ON DELETE CASCADE,
    CONSTRAINT fk_amie_audit_extras_procev FOREIGN KEY (event_id) REFERENCES amie_processing_events(id) ON DELETE SET NULL,
    KEY idx_amie_audit_extras_packet_id (packet_id),
    KEY idx_amie_audit_extras_event_id  (event_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
