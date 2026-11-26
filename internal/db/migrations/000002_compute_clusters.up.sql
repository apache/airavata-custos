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
    local_username     VARCHAR(32) NOT NULL,
    access_level       VARCHAR(16) NOT NULL DEFAULT 'USER',
    provisioned_at     TIMESTAMPTZ(6) NULL,
    -- An admin with the right privilege has to approve the account before it is created on the cluster.
    -- A user an admin adds is approved at once, a user from an allocation source waits for review.
    approval_status    VARCHAR(16) NOT NULL DEFAULT 'PENDING',
    reviewed_at        TIMESTAMPTZ(6) NULL,
    reviewed_by        VARCHAR(255) NULL,
    review_note        TEXT NULL,
    created_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at         TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT chk_compute_cluster_users_access_level CHECK (access_level IN ('USER', 'ADMIN')),
    CONSTRAINT chk_compute_cluster_users_approval_status CHECK (approval_status IN ('PENDING', 'APPROVED', 'DENIED')),
    CONSTRAINT uq_compute_cluster_users_pair UNIQUE (compute_cluster_id, user_id),
    CONSTRAINT uq_compute_cluster_users_local_username UNIQUE (compute_cluster_id, local_username),
    CONSTRAINT fk_compute_cluster_users_cluster FOREIGN KEY (compute_cluster_id) REFERENCES compute_clusters (id) ON DELETE CASCADE,
    CONSTRAINT fk_compute_cluster_users_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_compute_cluster_users_reviewed_by FOREIGN KEY (reviewed_by) REFERENCES users (id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_compute_cluster_users_user ON compute_cluster_users (user_id);
CREATE INDEX IF NOT EXISTS idx_compute_cluster_users_approval_status ON compute_cluster_users (approval_status);

-- A QoS definition Custos owns and writes to the cluster's scheduler. slurm_pending_since is set while a change waits to be written.
CREATE TABLE IF NOT EXISTS compute_cluster_qos
(
    id                  VARCHAR(255) NOT NULL,
    compute_cluster_id  VARCHAR(255) NOT NULL,
    name                VARCHAR(255) NOT NULL,
    priority            BIGINT NOT NULL DEFAULT 0,
    max_wall_minutes    BIGINT NULL,
    max_jobs_per_user   BIGINT NULL,
    max_tres_per_user   TEXT NOT NULL DEFAULT '',
    flags               TEXT[] NOT NULL DEFAULT '{}',
    is_default          BOOLEAN NOT NULL DEFAULT false,
    slurm_pending_since TIMESTAMPTZ(6) NULL,
    created_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at          TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT uq_compute_cluster_qos_name UNIQUE (compute_cluster_id, name),
    CONSTRAINT fk_compute_cluster_qos_cluster FOREIGN KEY (compute_cluster_id) REFERENCES compute_clusters (id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_compute_cluster_qos_default ON compute_cluster_qos (compute_cluster_id) WHERE is_default;

CREATE OR REPLACE TRIGGER trg_compute_clusters_updated_at
    BEFORE UPDATE ON compute_clusters
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE TRIGGER trg_compute_cluster_users_updated_at
    BEFORE UPDATE ON compute_cluster_users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE TRIGGER trg_compute_cluster_qos_updated_at
    BEFORE UPDATE ON compute_cluster_qos
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
