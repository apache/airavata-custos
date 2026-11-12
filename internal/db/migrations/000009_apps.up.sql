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

CREATE TABLE IF NOT EXISTS apps
(
    id          VARCHAR(255)   NOT NULL,
    key         VARCHAR(50)    NOT NULL, -- The app's readable ID. It is in the token and on every certificate.
    name        VARCHAR(100)   NOT NULL,
    description TEXT           NOT NULL DEFAULT '',
    status      VARCHAR(16)    NOT NULL DEFAULT 'ACTIVE',
    mode        VARCHAR(16)    NOT NULL,
    secret_hash VARCHAR(64)    NOT NULL,
    secret_hint VARCHAR(12)    NOT NULL, -- First part of the secret to identify which key is being used
    created_by  VARCHAR(255)   NOT NULL,
    created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT uq_apps_key UNIQUE (key),
    CONSTRAINT uq_apps_name UNIQUE (name),
    CONSTRAINT chk_apps_status CHECK (status IN ('ACTIVE', 'DISABLED')),
    CONSTRAINT chk_apps_mode CHECK (mode IN ('SERVICE', 'DELEGATED')),
    CONSTRAINT fk_apps_created_by FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS app_privileges
(
    app_id    VARCHAR(255) NOT NULL,
    privilege VARCHAR(100) NOT NULL,
    PRIMARY KEY (app_id, privilege),
    CONSTRAINT fk_app_privileges_app FOREIGN KEY (app_id) REFERENCES apps (id) ON DELETE RESTRICT
);

-- A row means the user has allowed the app. Removing it is recorded in audit_events.
CREATE TABLE IF NOT EXISTS app_grants
(
    app_id     VARCHAR(255)   NOT NULL,
    user_id    VARCHAR(255)   NOT NULL,
    granted_at TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (app_id, user_id),
    CONSTRAINT fk_app_grants_app FOREIGN KEY (app_id) REFERENCES apps (id) ON DELETE RESTRICT,
    CONSTRAINT fk_app_grants_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT
);
