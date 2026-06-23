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

-- Project-level governance roles (PI / CO_PI / ALLOCATION_MANAGER). MEMBER
-- status is derived from compute_allocation_memberships, so it is not stored
-- here. One row per (project, user).
CREATE TABLE project_memberships (
    project_id CHAR(36) NOT NULL,
    user_id    CHAR(36) NOT NULL,
    role       VARCHAR(32) NOT NULL,
    added_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (project_id, user_id),
    KEY idx_project_memberships_role (role),
    CONSTRAINT fk_project_memberships_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT fk_project_memberships_user    FOREIGN KEY (user_id)    REFERENCES users(id)    ON DELETE CASCADE,
    CONSTRAINT chk_project_memberships_role   CHECK (role IN ('PI', 'CO_PI', 'ALLOCATION_MANAGER'))
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;

-- Carry forward existing special roles that 000013 wrote onto
-- compute_allocation_memberships. The same user can appear on multiple
-- allocations of a project; collapse to one row per (project, user).
INSERT INTO project_memberships (project_id, user_id, role, added_time)
SELECT a.project_id, m.user_id,
       MIN(m.role)       AS role,
       MIN(m.start_time) AS added_time
FROM compute_allocation_memberships m
JOIN compute_allocations a ON a.id = m.compute_allocation_id
WHERE m.role IN ('PI', 'CO_PI', 'ALLOCATION_MANAGER')
GROUP BY a.project_id, m.user_id;

ALTER TABLE compute_allocation_memberships
    DROP KEY idx_compute_allocation_memberships_role,
    DROP COLUMN role;
