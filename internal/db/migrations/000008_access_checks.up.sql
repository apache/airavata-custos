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

CREATE TABLE IF NOT EXISTS access_checks
(
    id                    VARCHAR(255) NOT NULL,
    compute_allocation_id VARCHAR(255) NOT NULL,
    user_id               VARCHAR(255) NOT NULL,
    check_type            VARCHAR(32)  NOT NULL,
    status                VARCHAR(16)  NOT NULL,
    detail                VARCHAR(512) NULL,
    last_checked_at       TIMESTAMP(6) NOT NULL,
    last_ok_at            TIMESTAMP(6) NULL,
    failing_since         TIMESTAMP(6) NULL,
    created_at            TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at            TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_access_checks_target (compute_allocation_id, user_id, check_type),
    KEY idx_access_checks_user (user_id),
    CONSTRAINT fk_access_checks_allocation FOREIGN KEY (compute_allocation_id)
        REFERENCES compute_allocations (id) ON DELETE CASCADE,
    CONSTRAINT fk_access_checks_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS access_check_events
(
    id              VARCHAR(255) NOT NULL,
    access_check_id VARCHAR(255) NOT NULL,
    event_type      VARCHAR(16)  NOT NULL,
    occurred_at     TIMESTAMP(6) NOT NULL,
    PRIMARY KEY (id),
    KEY idx_access_check_events_check (access_check_id, occurred_at),
    CONSTRAINT fk_access_check_events_check FOREIGN KEY (access_check_id)
        REFERENCES access_checks (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
