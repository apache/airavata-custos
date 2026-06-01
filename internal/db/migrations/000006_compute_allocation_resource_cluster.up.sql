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

ALTER TABLE compute_allocation_resources
    ADD COLUMN compute_cluster_id VARCHAR(255) NOT NULL,
    ADD KEY idx_compute_allocation_resources_cluster (compute_cluster_id),
    ADD UNIQUE KEY uq_compute_allocation_resources_cluster_name (compute_cluster_id, name),
    ADD CONSTRAINT fk_compute_allocation_resources_cluster FOREIGN KEY (compute_cluster_id)
        REFERENCES compute_clusters (id) ON DELETE RESTRICT;
