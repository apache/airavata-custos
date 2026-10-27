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

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE IF NOT EXISTS organizations
(
    id            VARCHAR(255) NOT NULL,
    originated_id VARCHAR(255) NOT NULL,
    name          VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS idx_organizations_originated_id ON organizations (originated_id);

CREATE TABLE IF NOT EXISTS users
(
    id              VARCHAR(255) NOT NULL,
    organization_id VARCHAR(255) NOT NULL,
    first_name      VARCHAR(255) NOT NULL,
    last_name       VARCHAR(255) NOT NULL,
    middle_name     VARCHAR(255) NOT NULL DEFAULT '',
    email           CITEXT       NOT NULL CHECK (length(email) <= 255),
    status          VARCHAR(32)  NOT NULL DEFAULT 'PENDING',
    type            VARCHAR(32)  NOT NULL DEFAULT 'CLUSTER_LOCAL',
    created_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at      TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT uq_users_email UNIQUE (email),
    CONSTRAINT fk_users_organization FOREIGN KEY (organization_id) REFERENCES organizations (id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_users_organization_id ON users (organization_id);
CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);

-- email captures the address the source IdP configured for this identity.
-- oidc_sub is nullable: not every external identity issues an OIDC subject
-- (AMIE binds by external_id only). UNIQUE permits multiple NULLs but blocks
-- collisions on real values across IdPs.
CREATE TABLE IF NOT EXISTS user_identities
(
    id          VARCHAR(255) NOT NULL,
    user_id     VARCHAR(255) NOT NULL,
    source      VARCHAR(64)  NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    email       CITEXT       NULL DEFAULT NULL CHECK (email IS NULL OR length(email) <= 255),
    oidc_sub    VARCHAR(255) NULL DEFAULT NULL,
    metadata    TEXT         NULL DEFAULT NULL,
    created_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    updated_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    PRIMARY KEY (id),
    CONSTRAINT uq_user_identities_source_external UNIQUE (source, external_id),
    CONSTRAINT uq_user_identities_oidc_sub UNIQUE (oidc_sub),
    CONSTRAINT fk_user_identities_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_user_identities_user ON user_identities (user_id);

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE OR REPLACE TRIGGER trg_user_identities_updated_at
    BEFORE UPDATE ON user_identities
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
