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

CREATE TABLE IF NOT EXISTS compute_allocation_change_requests
(
    id                    VARCHAR(255) NOT NULL,
    compute_allocation_id VARCHAR(255) NOT NULL,
    requested_su_amount   BIGINT       NOT NULL DEFAULT 0,
    requested_status      VARCHAR(64)  NOT NULL,
    reason                TEXT         NOT NULL,
    change_status         VARCHAR(64)  NOT NULL,
    requester_id          VARCHAR(255) NOT NULL,
    approver_id           VARCHAR(255) NOT NULL DEFAULT '',
    timestamp             TIMESTAMP(6) NOT NULL,
    created_at            TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_compute_allocation_change_requests_allocation (compute_allocation_id, timestamp),
    KEY idx_compute_allocation_change_requests_status (change_status),
    KEY idx_compute_allocation_change_requests_requester (requester_id),
    CONSTRAINT fk_compute_allocation_change_requests_allocation FOREIGN KEY (compute_allocation_id)
        REFERENCES compute_allocations (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS compute_allocation_change_request_events
(
    id                                  VARCHAR(255) NOT NULL,
    compute_allocation_change_request_id VARCHAR(255) NOT NULL,
    event_type                          VARCHAR(64)  NOT NULL,
    description                         TEXT         NOT NULL,
    timestamp                           TIMESTAMP(6) NOT NULL,
    created_at                          TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_compute_allocation_change_request_events_request (compute_allocation_change_request_id, timestamp),
    KEY idx_compute_allocation_change_request_events_type (event_type)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
