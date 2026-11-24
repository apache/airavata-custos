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

-- Versioned user agreements. User rows shouldn't be deleted, so every user here has RESTRICT FK reference.
-- TODO - adjust the rest of the tables
CREATE TABLE IF NOT EXISTS user_agreements
(
    id            VARCHAR(255) NOT NULL,
    title         VARCHAR(255) NOT NULL,
    version_label VARCHAR(32)  NOT NULL,
    summary       TEXT         NOT NULL DEFAULT '',
    body_md       TEXT         NOT NULL,
    body_hash     VARCHAR(64)  NOT NULL,
    status        VARCHAR(16)  NOT NULL DEFAULT 'DRAFT',
    change_note   TEXT         NOT NULL DEFAULT '',
    created_by    VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    published_by  VARCHAR(255) NULL,
    published_at  TIMESTAMPTZ(6) NULL,
    updated_at    TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    removed_by    VARCHAR(255) NULL,
    removed_at    TIMESTAMPTZ(6) NULL,
    PRIMARY KEY (id),
    CONSTRAINT uq_user_agreements_version UNIQUE (version_label),
    CONSTRAINT chk_user_agreements_status CHECK (status IN ('DRAFT', 'PUBLISHED', 'REMOVED')),
    CONSTRAINT fk_user_agreements_created_by FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_user_agreements_published_by FOREIGN KEY (published_by) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_user_agreements_removed_by FOREIGN KEY (removed_by) REFERENCES users (id) ON DELETE RESTRICT
);

-- Only drafts can be deleted. A published version is marked REMOVED instead.
CREATE OR REPLACE FUNCTION refuse_non_draft_user_agreement_delete() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status <> 'DRAFT' THEN
        RAISE EXCEPTION 'user agreement % is %, only drafts can be deleted', OLD.id, OLD.status;
    END IF;
    RETURN OLD;
END $$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER trg_user_agreements_delete_drafts_only
    BEFORE DELETE ON user_agreements
    FOR EACH ROW EXECUTE FUNCTION refuse_non_draft_user_agreement_delete();

CREATE TABLE IF NOT EXISTS user_agreement_acceptances
(
    id           VARCHAR(255) NOT NULL,
    user_id      VARCHAR(255) NOT NULL,
    agreement_id VARCHAR(255) NOT NULL,
    accepted_at  TIMESTAMPTZ(6) NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
    body_hash    VARCHAR(64)  NOT NULL,
    PRIMARY KEY (id),
    CONSTRAINT uq_user_agreement_acceptances UNIQUE (user_id, agreement_id),
    CONSTRAINT fk_user_agreement_acceptances_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
    CONSTRAINT fk_user_agreement_acceptances_agreement FOREIGN KEY (agreement_id) REFERENCES user_agreements (id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_user_agreement_acceptances_agreement ON user_agreement_acceptances (agreement_id);
