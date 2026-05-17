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
    id              VARCHAR(255) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    resource_type   VARCHAR(64)  NOT NULL,
    resource_amount BIGINT       NOT NULL DEFAULT 0,
    created_at      TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at      TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_compute_allocation_resources_type (resource_type)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
