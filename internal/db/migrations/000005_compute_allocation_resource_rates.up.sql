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

CREATE TABLE IF NOT EXISTS compute_allocation_resource_rates
(
    id                             VARCHAR(255)     NOT NULL,
    compute_allocation_resource_id VARCHAR(255)     NOT NULL,
    rate                           DOUBLE           NOT NULL DEFAULT 0,
    start_time                     TIMESTAMP(6)     NOT NULL,
    end_time                       TIMESTAMP(6)     NOT NULL,
    created_at                     TIMESTAMP(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at                     TIMESTAMP(6)     NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    KEY idx_carr_rates_resource (compute_allocation_resource_id),
    KEY idx_carr_rates_window (compute_allocation_resource_id, start_time, end_time),
    CONSTRAINT fk_carr_rates_resource FOREIGN KEY (compute_allocation_resource_id)
        REFERENCES compute_allocation_resources (id) ON DELETE CASCADE
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci;
