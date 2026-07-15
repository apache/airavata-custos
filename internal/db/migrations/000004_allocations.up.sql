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

CREATE TABLE IF NOT EXISTS compute_allocations
(
    id                 VARCHAR(255) NOT NULL,
    project_id         VARCHAR(255) NOT NULL,
    name               VARCHAR(255) NOT NULL,
    status             VARCHAR(64)  NOT NULL,
    compute_cluster_id VARCHAR(255) NOT NULL,
    initial_su_amount  BIGINT       NOT NULL DEFAULT 0,
    start_time         TIMESTAMP(6) NOT NULL,
    end_time           TIMESTAMP(6) NOT NULL,
    created_at         TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at         TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_compute_allocations_project (project_id),
    KEY idx_compute_allocations_cluster (compute_cluster_id),
    CONSTRAINT fk_compute_allocations_project FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE RESTRICT,
    CONSTRAINT fk_compute_allocations_cluster FOREIGN KEY (compute_cluster_id) REFERENCES compute_clusters (id) ON DELETE RESTRICT
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS compute_allocation_resources
(
    id                 VARCHAR(255) NOT NULL,
    name               VARCHAR(255) NOT NULL,
    resource_type      VARCHAR(64)  NOT NULL,
    resource_amount    BIGINT       NOT NULL DEFAULT 0,
    compute_cluster_id VARCHAR(255) NOT NULL,
    created_at         TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at         TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_compute_allocation_resources_cluster_name (compute_cluster_id, name),
    KEY idx_compute_allocation_resources_type (resource_type),
    KEY idx_compute_allocation_resources_cluster (compute_cluster_id),
    CONSTRAINT fk_compute_allocation_resources_cluster FOREIGN KEY (compute_cluster_id)
        REFERENCES compute_clusters (id) ON DELETE RESTRICT
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS compute_allocation_resource_mappings
(
    id                             VARCHAR(255) NOT NULL,
    compute_allocation_id          VARCHAR(255) NOT NULL,
    compute_allocation_resource_id VARCHAR(255) NOT NULL,
    resource_amount                BIGINT       NOT NULL DEFAULT 0,
    resource_time                  BIGINT       NOT NULL DEFAULT 0,
    created_at                     TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_carm_pair (compute_allocation_id, compute_allocation_resource_id),
    KEY idx_carm_allocation (compute_allocation_id),
    KEY idx_carm_resource (compute_allocation_resource_id),
    CONSTRAINT fk_carm_allocation FOREIGN KEY (compute_allocation_id)
        REFERENCES compute_allocations (id) ON DELETE CASCADE,
    CONSTRAINT fk_carm_resource FOREIGN KEY (compute_allocation_resource_id)
        REFERENCES compute_allocation_resources (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS compute_allocation_resource_rates
(
    id                             VARCHAR(255) NOT NULL,
    compute_allocation_resource_id VARCHAR(255) NOT NULL,
    rate                           DOUBLE       NOT NULL DEFAULT 0,
    start_time                     TIMESTAMP(6) NOT NULL,
    end_time                       TIMESTAMP(6) NOT NULL,
    created_at                     TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at                     TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_carr_rates_resource (compute_allocation_resource_id),
    KEY idx_carr_rates_window (compute_allocation_resource_id, start_time, end_time),
    CONSTRAINT fk_carr_rates_resource FOREIGN KEY (compute_allocation_resource_id)
        REFERENCES compute_allocation_resources (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS compute_allocation_diffs
(
    id                    VARCHAR(255) NOT NULL,
    compute_allocation_id VARCHAR(255) NOT NULL,
    diff_type             VARCHAR(64)  NOT NULL,
    new_su_amount         BIGINT       NOT NULL DEFAULT 0,
    status                VARCHAR(64)  NOT NULL,
    timestamp             TIMESTAMP(6) NOT NULL,
    description           TEXT         NOT NULL,
    created_at            TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_compute_allocation_diffs_allocation (compute_allocation_id, timestamp),
    KEY idx_compute_allocation_diffs_type (diff_type),
    CONSTRAINT fk_compute_allocation_diffs_allocation FOREIGN KEY (compute_allocation_id)
        REFERENCES compute_allocations (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

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
    id                                   VARCHAR(255) NOT NULL,
    compute_allocation_change_request_id VARCHAR(255) NOT NULL,
    event_type                           VARCHAR(64)  NOT NULL,
    description                          TEXT         NOT NULL,
    timestamp                            TIMESTAMP(6) NOT NULL,
    created_at                           TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_compute_allocation_change_request_events_request (compute_allocation_change_request_id, timestamp),
    KEY idx_compute_allocation_change_request_events_type (event_type)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS compute_allocation_memberships
(
    id                    VARCHAR(255) NOT NULL,
    compute_allocation_id VARCHAR(255) NOT NULL,
    user_id               VARCHAR(255) NOT NULL,
    start_time            TIMESTAMP(6) NOT NULL,
    end_time              TIMESTAMP(6) NOT NULL,
    membership_status     VARCHAR(64)  NOT NULL,
    created_at            TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_compute_allocation_memberships_allocation_user (compute_allocation_id, user_id),
    KEY idx_compute_allocation_memberships_user (user_id),
    KEY idx_compute_allocation_memberships_status (membership_status),
    CONSTRAINT fk_compute_allocation_memberships_allocation FOREIGN KEY (compute_allocation_id)
        REFERENCES compute_allocations (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS compute_allocation_usages
(
    id                             VARCHAR(255) NOT NULL,
    compute_allocation_id          VARCHAR(255) NOT NULL,
    used_raw_amount                DOUBLE       NOT NULL DEFAULT 0,
    used_su_amount                 DOUBLE       NOT NULL DEFAULT 0,
    calculated_time                TIMESTAMP(6) NOT NULL,
    user_id                        VARCHAR(255) NOT NULL DEFAULT '',
    job_id                         VARCHAR(255) NOT NULL DEFAULT '',
    compute_allocation_resource_id VARCHAR(255) NOT NULL DEFAULT '',
    created_at                     TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_compute_allocation_usages_allocation (compute_allocation_id, calculated_time),
    KEY idx_compute_allocation_usages_user (user_id),
    KEY idx_compute_allocation_usages_job (job_id),
    KEY idx_compute_allocation_usages_resource (compute_allocation_resource_id)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS compute_allocation_membership_resource_overrides
(
    id                               VARCHAR(255) NOT NULL,
    compute_allocation_membership_id VARCHAR(255) NOT NULL,
    compute_allocation_resource_id   VARCHAR(255) NOT NULL,
    override_resource_amount         BIGINT       NOT NULL DEFAULT 0,
    override_resource_time           BIGINT       NOT NULL DEFAULT 0,
    created_at                       TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at                       TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    UNIQUE KEY uq_camro_membership_resource (compute_allocation_membership_id, compute_allocation_resource_id),
    KEY idx_camro_resource (compute_allocation_resource_id),
    CONSTRAINT fk_camro_membership FOREIGN KEY (compute_allocation_membership_id)
        REFERENCES compute_allocation_memberships (id) ON DELETE CASCADE,
    CONSTRAINT fk_camro_resource FOREIGN KEY (compute_allocation_resource_id)
        REFERENCES compute_allocation_resources (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
