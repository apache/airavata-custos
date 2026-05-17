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

CREATE TABLE IF NOT EXISTS compute_allocation_resource_mappings
(
    id                             VARCHAR(255) NOT NULL,
    compute_allocation_id          VARCHAR(255) NOT NULL,
    compute_allocation_resource_id VARCHAR(255) NOT NULL,
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
