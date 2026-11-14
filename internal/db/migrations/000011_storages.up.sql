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

CREATE TABLE IF NOT EXISTS storage_policies
(
    id                 VARCHAR(255) NOT NULL,
    compute_cluster_id VARCHAR(255) NOT NULL,
    type               VARCHAR(16)  NOT NULL,
    path_pattern       TEXT         NOT NULL,
    owner              VARCHAR(255) NOT NULL,
    "group"            VARCHAR(255) NOT NULL,
    mode               TEXT         NOT NULL,
    quota_bytes        BIGINT       NOT NULL DEFAULT 0,
    quota_inodes       BIGINT       NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    CONSTRAINT uq_storage_policies_path_pattern UNIQUE (compute_cluster_id, path_pattern),
    CONSTRAINT fk_storage_policies_cluster FOREIGN KEY (compute_cluster_id) REFERENCES compute_clusters (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS storages
(
    id                 VARCHAR(255)   NOT NULL,
    compute_cluster_id VARCHAR(255)   NOT NULL,
    policy_id          VARCHAR(255),
    type               VARCHAR(16)    NOT NULL,
    path               TEXT           NOT NULL,
    username           VARCHAR(255)   NOT NULL DEFAULT '',
    allocation_id      VARCHAR(255),
    allocation_name    VARCHAR(255),
    start_time         TIMESTAMPTZ(6),
    end_time           TIMESTAMPTZ(6),
    owner              VARCHAR(255)   NOT NULL,
    "group"            VARCHAR(255)   NOT NULL,
    mode               TEXT           NOT NULL,
    quota_bytes        BIGINT         NOT NULL DEFAULT 0,
    quota_inodes       BIGINT         NOT NULL DEFAULT 0,
    quota_until        TIMESTAMPTZ(6),
    used_bytes         BIGINT         NOT NULL DEFAULT 0,
    status             VARCHAR(16)    NOT NULL DEFAULT 'active',
    PRIMARY KEY (id),
    CONSTRAINT chk_storages_status CHECK (status IN ('active', 'pending_deletion')),
    CONSTRAINT uq_storages_path UNIQUE (compute_cluster_id, path),
    CONSTRAINT fk_storages_cluster FOREIGN KEY (compute_cluster_id) REFERENCES compute_clusters (id) ON DELETE CASCADE,
    CONSTRAINT fk_storages_policy FOREIGN KEY (policy_id) REFERENCES storage_policies (id) ON DELETE SET NULL
);

-- Storage routes were gated by the cluster privileges, so their holders keep access.
INSERT INTO role_privileges (role_id, privilege)
SELECT role_id, k || CASE privilege WHEN 'core:clusters:read' THEN 'read' ELSE 'manage' END FROM role_privileges
CROSS JOIN unnest(ARRAY['core:storages:', 'core:storage-policies:']) AS k
WHERE privilege IN ('core:clusters:read', 'core:clusters:write')
ON CONFLICT DO NOTHING;

INSERT INTO user_privileges (id, user_id, privilege, granted_by, reason)
SELECT gen_random_uuid(), user_id, k || CASE privilege WHEN 'core:clusters:read' THEN 'read' ELSE 'manage' END, granted_by, reason FROM user_privileges
CROSS JOIN unnest(ARRAY['core:storages:', 'core:storage-policies:']) AS k
WHERE privilege IN ('core:clusters:read', 'core:clusters:write')
ON CONFLICT DO NOTHING;

INSERT INTO service_account_privileges (service_account_id, privilege)
SELECT service_account_id, k || CASE privilege WHEN 'core:clusters:read' THEN 'read' ELSE 'manage' END FROM service_account_privileges
CROSS JOIN unnest(ARRAY['core:storages:', 'core:storage-policies:']) AS k
WHERE privilege IN ('core:clusters:read', 'core:clusters:write')
ON CONFLICT DO NOTHING;
