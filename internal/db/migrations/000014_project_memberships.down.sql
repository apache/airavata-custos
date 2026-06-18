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

ALTER TABLE compute_allocation_memberships
    ADD COLUMN role VARCHAR(32) NOT NULL DEFAULT 'MEMBER' AFTER membership_status,
    ADD KEY idx_compute_allocation_memberships_role (role);

-- Project-level role applies to every allocation membership of the user
-- under that project.
UPDATE compute_allocation_memberships m
JOIN compute_allocations a   ON a.id = m.compute_allocation_id
JOIN project_memberships pm  ON pm.project_id = a.project_id
                            AND pm.user_id    = m.user_id
SET m.role = pm.role;

DROP TABLE project_memberships;
