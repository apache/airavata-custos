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

CREATE TABLE IF NOT EXISTS access_events
(
    code                  VARCHAR(64)  NOT NULL,
    compute_allocation_id VARCHAR(255) NOT NULL,
    organization_id       VARCHAR(255) NOT NULL,
    created_at            TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at            TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (code),
    KEY idx_access_events_allocation (compute_allocation_id),
    KEY idx_access_events_organization (organization_id),
    CONSTRAINT fk_access_events_allocation FOREIGN KEY (compute_allocation_id)
        REFERENCES compute_allocations (id) ON DELETE RESTRICT,
    CONSTRAINT fk_access_events_organization FOREIGN KEY (organization_id)
        REFERENCES organizations (id) ON DELETE RESTRICT
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS access_requests
(
    id               VARCHAR(255) NOT NULL,
    oidc_sub         VARCHAR(255) NOT NULL,
    email            VARCHAR(255) NOT NULL,
    name             VARCHAR(255) NOT NULL,
    institution      VARCHAR(255) NOT NULL,
    desired_username VARCHAR(64)  NULL,
    event_code       VARCHAR(64)  NOT NULL,
    reason          TEXT         NULL,
    status          VARCHAR(32)  NOT NULL DEFAULT 'PENDING',
    approver_id     VARCHAR(255) NULL,
    deny_reason     TEXT         NULL,
    expires_at      TIMESTAMP(6) NULL,
    created_user_id VARCHAR(255) NULL,
    timestamp       TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_access_requests_sub (oidc_sub),
    KEY idx_access_requests_status (status),
    CONSTRAINT fk_access_requests_event FOREIGN KEY (event_code)
        REFERENCES access_events (code) ON DELETE RESTRICT,
    CONSTRAINT fk_access_requests_approver FOREIGN KEY (approver_id)
        REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_access_requests_created_user FOREIGN KEY (created_user_id)
        REFERENCES users (id) ON DELETE RESTRICT
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS access_request_events
(
    id                VARCHAR(255) NOT NULL,
    access_request_id VARCHAR(255) NOT NULL,
    event_type        VARCHAR(32)  NOT NULL,
    description       TEXT         NULL,
    timestamp         TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_access_request_events_request (access_request_id, timestamp),
    CONSTRAINT fk_access_request_events_request FOREIGN KEY (access_request_id)
        REFERENCES access_requests (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
