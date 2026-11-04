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

-- POSIX attributes this connector assigned, one row per cluster account. A
-- deactivated account keeps its row so its uid_number is never reused, except
-- by the same Custos user (user_id) coming back.
CREATE TABLE IF NOT EXISTS ldap_posix_accounts
(
    compute_cluster_user_id VARCHAR(255)   PRIMARY KEY,
    user_id                 VARCHAR(255)   NOT NULL,
    uid_number              BIGINT         NOT NULL,
    gid_number              BIGINT         NOT NULL,
    username                VARCHAR(32)    NOT NULL,
    home_directory          TEXT           NOT NULL,
    login_shell             VARCHAR(255)   NOT NULL,
    entry_dn                TEXT           NOT NULL,
    status                  VARCHAR(16)    NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at              TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at              TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    CONSTRAINT uq_ldap_posix_accounts_uid_number UNIQUE (uid_number),
    CONSTRAINT uq_ldap_posix_accounts_username UNIQUE (username)
);
