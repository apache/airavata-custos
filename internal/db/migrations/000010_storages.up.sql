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

CREATE TABLE IF NOT EXISTS storages
(
    id                      VARCHAR(255)   NOT NULL,
    compute_cluster_id      VARCHAR(255)   NOT NULL,
    compute_cluster_user_id VARCHAR(255),
    allocation_id           VARCHAR(255),
    user_id                 VARCHAR(255),
    type                    TEXT           NOT NULL DEFAULT '',
    path                    TEXT           NOT NULL,
    owner                   VARCHAR(255)   NOT NULL,
    "group"                 VARCHAR(255)   NOT NULL,
    mode                    TEXT           NOT NULL,
    quota_bytes             BIGINT         NOT NULL DEFAULT 0,
    quota_inodes            BIGINT         NOT NULL DEFAULT 0,
    quota_until             TIMESTAMPTZ(6),
    status                  VARCHAR(16)    NOT NULL DEFAULT 'active',
    PRIMARY KEY (id),
    CONSTRAINT chk_storages_status CHECK (status IN ('active', 'pending_deletion')),
    CONSTRAINT uq_storages_path UNIQUE (compute_cluster_id, path),
    -- A deactivated storage outlives its holder for an administrator to review.
    CONSTRAINT fk_storages_cluster_user FOREIGN KEY (compute_cluster_user_id) REFERENCES compute_cluster_users (id) ON DELETE SET NULL,
    CONSTRAINT fk_storages_allocation FOREIGN KEY (allocation_id) REFERENCES compute_allocations (id) ON DELETE SET NULL,
    CONSTRAINT fk_storages_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_storages_cluster_user ON storages (compute_cluster_user_id);
CREATE INDEX IF NOT EXISTS idx_storages_allocation ON storages (allocation_id);
CREATE INDEX IF NOT EXISTS idx_storages_user ON storages (user_id);
