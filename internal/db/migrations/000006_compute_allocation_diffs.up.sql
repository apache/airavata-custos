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
