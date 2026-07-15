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

CREATE TABLE IF NOT EXISTS projects
(
    id            VARCHAR(255) NOT NULL,
    originated_id VARCHAR(255) NOT NULL,
    title         VARCHAR(255) NOT NULL,
    origination   VARCHAR(255) NOT NULL,
    project_pi_id VARCHAR(255) NOT NULL,
    status        VARCHAR(32)  NOT NULL DEFAULT 'ACTIVE',
    created_time  TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_projects_originated_id (originated_id),
    KEY idx_projects_pi (project_pi_id),
    KEY idx_projects_status (status),
    CONSTRAINT fk_projects_pi FOREIGN KEY (project_pi_id) REFERENCES users (id) ON DELETE RESTRICT
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- Project-level governance roles (PI / CO_PI / ALLOCATION_MANAGER). MEMBER
-- status is derived from compute_allocation_memberships, so it is not stored
-- here. One row per (project, user).
CREATE TABLE IF NOT EXISTS project_memberships
(
    project_id VARCHAR(255) NOT NULL,
    user_id    VARCHAR(255) NOT NULL,
    role       VARCHAR(32)  NOT NULL,
    added_time TIMESTAMP(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (project_id, user_id),
    KEY idx_project_memberships_role (role),
    CONSTRAINT fk_project_memberships_project FOREIGN KEY (project_id) REFERENCES projects (id) ON DELETE CASCADE,
    CONSTRAINT fk_project_memberships_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT chk_project_memberships_role CHECK (role IN ('PI', 'CO_PI', 'ALLOCATION_MANAGER'))
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
