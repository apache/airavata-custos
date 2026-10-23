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

CREATE TABLE IF NOT EXISTS vast_folders
(
    id                 VARCHAR(255)   PRIMARY KEY,
    compute_cluster_id VARCHAR(255)   NOT NULL,
    path               TEXT           NOT NULL,
    owner_kind         VARCHAR(16)    NOT NULL CHECK (owner_kind IN ('cluster_user', 'allocation')),
    owner_name         VARCHAR(255)   NOT NULL,
    unix_owner         VARCHAR(255)   NOT NULL,
    allocation_id      VARCHAR(255),
    allocation_name    VARCHAR(255),
    start_time         TIMESTAMPTZ(6),
    end_time           TIMESTAMPTZ(6),
    quota_override     BIGINT,
    quota_until        TIMESTAMPTZ(6),
    status             VARCHAR(16)    NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'pending_deletion')),
    CONSTRAINT uq_vast_folders_path UNIQUE (compute_cluster_id, path)
);
