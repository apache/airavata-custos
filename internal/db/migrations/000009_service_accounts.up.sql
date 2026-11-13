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

CREATE TABLE IF NOT EXISTS service_accounts
(
    id          VARCHAR(255)   NOT NULL,
    key         VARCHAR(50)    NOT NULL, -- The service account's readable ID. It is in the token and in audit records.
    name        VARCHAR(100)   NOT NULL,
    description TEXT           NOT NULL DEFAULT '',
    status      VARCHAR(16)    NOT NULL DEFAULT 'ACTIVE',
    secret_hash VARCHAR(64)    NOT NULL,
    secret_hint VARCHAR(12)    NOT NULL, -- First part of the secret to identify which key is being used
    created_by  VARCHAR(255)   NOT NULL,
    created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT uq_service_accounts_key UNIQUE (key), -- Deleted service accounts keep their key, so it is never reused.
    CONSTRAINT chk_service_accounts_status CHECK (status IN ('ACTIVE', 'DISABLED', 'DELETED')),
    CONSTRAINT fk_service_accounts_created_by FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT
);

-- A deleted service account's name can be used again.
CREATE UNIQUE INDEX IF NOT EXISTS uq_service_accounts_name ON service_accounts (name) WHERE status <> 'DELETED';

CREATE TABLE IF NOT EXISTS service_account_privileges
(
    service_account_id VARCHAR(255) NOT NULL,
    privilege          VARCHAR(64)  NOT NULL,
    PRIMARY KEY (service_account_id, privilege),
    CONSTRAINT fk_service_account_privileges_account FOREIGN KEY (service_account_id) REFERENCES service_accounts (id) ON DELETE RESTRICT
);
