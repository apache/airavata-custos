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

CREATE TABLE IF NOT EXISTS compute_clusters
(
    id         VARCHAR(255) NOT NULL,
    name       VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT uq_compute_clusters_name UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS compute_cluster_users
(
    id                 VARCHAR(255) NOT NULL,
    compute_cluster_id VARCHAR(255) NOT NULL,
    user_id            VARCHAR(255) NOT NULL,
    local_username     VARCHAR(255) NOT NULL,
    provisioned_at     TIMESTAMPTZ(6) NULL,
    created_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT uq_compute_cluster_users_pair UNIQUE (compute_cluster_id, user_id),
    CONSTRAINT uq_compute_cluster_users_local_username UNIQUE (compute_cluster_id, local_username),
    CONSTRAINT fk_compute_cluster_users_cluster FOREIGN KEY (compute_cluster_id) REFERENCES compute_clusters (id) ON DELETE CASCADE,
    CONSTRAINT fk_compute_cluster_users_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_compute_cluster_users_user ON compute_cluster_users (user_id);

CREATE OR REPLACE TRIGGER trg_compute_clusters_updated_at
    BEFORE UPDATE ON compute_clusters
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE TRIGGER trg_compute_cluster_users_updated_at
    BEFORE UPDATE ON compute_cluster_users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
